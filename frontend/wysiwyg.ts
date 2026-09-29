export interface WysiwygOptions {
	onChange(before: string, after: string, kind: string): void;
	onSeparate(): void;
	onUndo(redo: boolean): void;
	resolveImage(reference: string): string | null;
}

const blockTags = new Set(['P','DIV','H1','H2','H3','H4','H5','H6','UL','OL','BLOCKQUOTE','PRE','TABLE','HR']);

function trimBlock(text: string): string {
	return text.replace(/[ \t]+\n/g, '\n').replace(/\n{3,}/g, '\n\n').trim();
}

function escapeText(text: string): string {
	return text.replace(/\\/g, '\\\\').replace(/([\`*_\[\]~])/g, '\\$1');
}

function inlineCode(text: string): string {
	const longest = Math.max(0, ...Array.from(text.matchAll(/`+/g), match => match[0].length));
	const fence = '`'.repeat(Math.max(1, longest + 1));
	const pad = /^\s|\s$/.test(text) ? ' ' : '';
	return `${fence}${pad}${text}${pad}${fence}`;
}

function serializeInline(node: Node): string {
	if (node.nodeType === Node.TEXT_NODE) return escapeText(node.textContent || '');
	if (!(node instanceof HTMLElement)) return '';
	const inner = () => Array.from(node.childNodes).map(serializeInline).join('');
	switch (node.tagName) {
	case 'STRONG':
	case 'B':
		return `**${inner()}**`;
	case 'EM':
	case 'I':
		return `*${inner()}*`;
	case 'DEL':
	case 'S':
	case 'STRIKE':
		return `~~${inner()}~~`;
	case 'CODE':
		return inlineCode(node.textContent || '');
	case 'A': {
		const href = node.getAttribute('href') || '';
		const title = node.getAttribute('title');
		const suffix = title ? ` "${title.replace(/"/g, '\\"')}"` : '';
		return `[${inner()}](${href}${suffix})`;
	}
	case 'IMG': {
		const src = node.dataset.markdownSrc || node.getAttribute('src') || '';
		const alt = (node.getAttribute('alt') || '').replace(/]/g, '\\]');
		const title = node.getAttribute('title');
		const suffix = title ? ` "${title.replace(/"/g, '\\"')}"` : '';
		return `![${alt}](${src}${suffix})`;
	}
	case 'BR':
		return '  \n';
	case 'INPUT':
		return '';
	default:
		return inner();
	}
}

function listItemText(item: HTMLLIElement): string {
	const nodes = Array.from(item.childNodes).filter(node => !(node instanceof HTMLElement && (node.tagName === 'UL' || node.tagName === 'OL')));
	return nodes.map(serializeInline).join('').trim();
}

function serializeList(list: HTMLUListElement | HTMLOListElement): string {
	const ordered = list.tagName === 'OL';
	const start = ordered ? Number(list.getAttribute('start') || '1') || 1 : 1;
	const lines: string[] = [];
	Array.from(list.children).filter((node): node is HTMLLIElement => node instanceof HTMLLIElement).forEach((item, index) => {
		const checkbox = Array.from(item.children).find(child => child instanceof HTMLInputElement && child.type === 'checkbox') as HTMLInputElement | undefined;
		const marker = ordered ? `${start + index}. ` : '- ';
		const task = checkbox ? `[${checkbox.checked ? 'x' : ' '}] ` : '';
		lines.push(marker + task + listItemText(item));
		for (const nested of Array.from(item.children).filter((child): child is HTMLUListElement | HTMLOListElement => child instanceof HTMLUListElement || child instanceof HTMLOListElement)) {
			for (const line of serializeList(nested).trimEnd().split('\n')) lines.push('  ' + line);
		}
	});
	return lines.join('\n') + '\n\n';
}

function serializeTable(table: HTMLTableElement): string {
	const rows = Array.from(table.rows);
	if (!rows.length) return '';
	const head = Array.from(rows[0].cells).map(cell => trimBlock(Array.from(cell.childNodes).map(serializeInline).join('')).replace(/\|/g, '\\|'));
	if (!head.length) return '';
	const alignment = Array.from(rows[0].cells).map(cell => {
		const align = (cell.getAttribute('align') || '').toLowerCase();
		if (align === 'center') return ':---:';
		if (align === 'right') return '---:';
		if (align === 'left') return ':---';
		return '---';
	});
	const output = [`| ${head.join(' | ')} |`, `| ${alignment.join(' | ')} |`];
	for (const row of rows.slice(1)) {
		const cells = Array.from(row.cells).map(cell => trimBlock(Array.from(cell.childNodes).map(serializeInline).join('')).replace(/\|/g, '\\|'));
		while (cells.length < head.length) cells.push('');
		output.push(`| ${cells.slice(0, head.length).join(' | ')} |`);
	}
	return output.join('\n') + '\n\n';
}

function serializeBlock(node: Node): string {
	if (node.nodeType === Node.TEXT_NODE) return (node.textContent || '').trim() ? escapeText(node.textContent || '') : '';
	if (!(node instanceof HTMLElement)) return '';
	switch (node.tagName) {
	case 'H1':
	case 'H2':
	case 'H3':
	case 'H4':
	case 'H5':
	case 'H6':
		return `${'#'.repeat(Number(node.tagName.slice(1)))} ${trimBlock(Array.from(node.childNodes).map(serializeInline).join(''))}\n\n`;
	case 'P':
	case 'DIV':
		return trimBlock(Array.from(node.childNodes).map(serializeInline).join('')) + '\n\n';
	case 'UL':
	case 'OL':
		return serializeList(node as HTMLUListElement | HTMLOListElement);
	case 'BLOCKQUOTE': {
		const text = trimBlock(Array.from(node.childNodes).map(serializeBlock).join(''));
		return text.split('\n').map(line => `> ${line}`).join('\n') + '\n\n';
	}
	case 'PRE': {
		const code = node.querySelector(':scope > code');
		const text = (code?.textContent || node.textContent || '').replace(/\n$/, '');
		const language = code?.className.match(/(?:^|\s)language-([^\s]+)/)?.[1] || '';
		const longest = Math.max(0, ...Array.from(text.matchAll(/`{3,}/g), match => match[0].length));
		const fence = '`'.repeat(Math.max(3, longest + 1));
		return `${fence}${language}\n${text}\n${fence}\n\n`;
	}
	case 'TABLE':
		return serializeTable(node as HTMLTableElement);
	case 'HR':
		return '---\n\n';
	default: {
		if (blockTags.has(node.tagName)) return Array.from(node.childNodes).map(serializeBlock).join('');
		const inline = trimBlock(Array.from(node.childNodes).map(serializeInline).join(''));
		return inline ? inline + '\n\n' : '';
	}
	}
}

function markdownFromElement(root: HTMLElement): string {
	const value = trimBlock(Array.from(root.childNodes).map(serializeBlock).join(''));
	return value ? value + '\n' : '';
}

export function markdownFromRenderedHTML(html: string): string {
	const doc = new DOMParser().parseFromString(html, 'text/html');
	return markdownFromElement(doc.body);
}

function canonicalNode(node: Node, preserveWhitespace = false): string {
	if (node.nodeType === Node.TEXT_NODE) {
		const text = node.textContent || '';
		const value = preserveWhitespace ? text : text.replace(/\s+/g, ' ');
		return value.trim() ? JSON.stringify(value.trim()) : '';
	}
	if (!(node instanceof HTMLElement)) return '';
	const tag = node.tagName.toLowerCase();
	const preserve = preserveWhitespace || tag === 'pre' || tag === 'code';
	const attrs: string[] = [];
	for (const name of ['href','src','title','start','align','type','checked','class']) {
		if (node.hasAttribute(name)) attrs.push(`${name}=${JSON.stringify(node.getAttribute(name) || '')}`);
	}
	const children = Array.from(node.childNodes).map(child => canonicalNode(child, preserve)).filter(Boolean).join(',');
	return `<${tag}${attrs.length ? ' ' + attrs.join(' ') : ''}>${children}</${tag}>`;
}

export function equivalentRenderedHTML(left: string, right: string): boolean {
	const a = new DOMParser().parseFromString(left, 'text/html');
	const b = new DOMParser().parseFromString(right, 'text/html');
	return canonicalNode(a.body) === canonicalNode(b.body);
}

export function renderedHTMLHasOmittedRawHTML(html: string): boolean {
	return /raw HTML omitted/i.test(html);
}

// WysiwygEditor は表示用HTMLを直接編集し、変更時だけMarkdownへ戻します。
export class WysiwygEditor {
	private markdown = '';
	private beforeInput = '';
	private commanding = false;

	constructor(
		private root: HTMLElement,
		private content: HTMLElement,
		private toolbar: HTMLElement,
		private blockSelect: HTMLSelectElement,
		private options: WysiwygOptions,
	) {
		this.content.addEventListener('beforeinput', event => {
			const input = event as InputEvent;
			if (input.inputType === 'historyUndo' || input.inputType === 'historyRedo') {
				event.preventDefault();
				this.options.onUndo(input.inputType === 'historyRedo');
				return;
			}
			if (!this.commanding) this.beforeInput = this.markdown;
		});
		this.content.addEventListener('input', event => {
			if (this.commanding) return;
			this.commit((event as InputEvent).inputType || 'wysiwyg', this.beforeInput || this.markdown);
			this.beforeInput = '';
		});
		this.content.addEventListener('change', () => {
			if (!this.commanding) this.commit('change', this.markdown);
		});
		this.content.addEventListener('blur', () => this.options.onSeparate());
		this.content.addEventListener('click', event => {
			if ((event.target as Element).closest('a')) event.preventDefault();
		});
		for (const button of this.toolbar.querySelectorAll<HTMLButtonElement>('[data-wysiwyg-command]')) {
			button.addEventListener('mousedown', event => event.preventDefault());
			button.addEventListener('click', () => this.runCommand(button.dataset.wysiwygCommand || ''));
		}
		this.blockSelect.addEventListener('change', () => {
			this.runCommand('formatBlock', this.blockSelect.value);
			this.blockSelect.value = 'p';
		});
	}

	load(markdown: string, html: string): void {
		const doc = new DOMParser().parseFromString(html, 'text/html');
		for (const image of doc.querySelectorAll<HTMLImageElement>('img')) {
			const reference = image.getAttribute('src') || '';
			image.dataset.markdownSrc = reference;
			const resolved = this.options.resolveImage(reference);
			if (resolved) image.src = resolved;
			else image.removeAttribute('src');
		}
		for (const checkbox of doc.querySelectorAll<HTMLInputElement>('input[type="checkbox"]')) {
			checkbox.disabled = false;
			checkbox.contentEditable = 'false';
		}
		this.commanding = true;
		this.content.innerHTML = doc.body.innerHTML || '<p><br></p>';
		this.commanding = false;
		this.markdown = markdown;
		this.beforeInput = '';
	}

	setActive(active: boolean): void {
		this.root.hidden = !active;
	}

	setEditable(editable: boolean): void {
		this.content.contentEditable = String(editable);
		this.toolbar.querySelectorAll<HTMLButtonElement | HTMLSelectElement>('button,select').forEach(control => control.disabled = !editable);
	}

	focus(): void {
		this.content.focus({preventScroll:true});
	}

	contains(node: EventTarget | null): boolean {
		return node instanceof Node && this.root.contains(node);
	}

	get scrollTop(): number { return this.content.scrollTop; }
	set scrollTop(value: number) { this.content.scrollTop = value; }
	get scrollHeight(): number { return this.content.scrollHeight; }
	get clientHeight(): number { return this.content.clientHeight; }

	insertImage(reference: string, display: string | null, alt = '画像'): void {
		const selection = window.getSelection();
		if (!selection) return;
		this.focus();
		const before = this.markdown;
		const range = selection.rangeCount && this.content.contains(selection.anchorNode) ? selection.getRangeAt(0) : document.createRange();
		if (!this.content.contains(range.commonAncestorContainer)) {
			range.selectNodeContents(this.content);
			range.collapse(false);
		}
		const image = document.createElement('img');
		image.alt = alt;
		image.dataset.markdownSrc = reference;
		if (display) image.src = display;
		range.deleteContents();
		range.insertNode(image);
		range.setStartAfter(image);
		range.collapse(true);
		selection.removeAllRanges();
		selection.addRange(range);
		this.commit('insertImage', before);
	}

	private commit(kind: string, before: string): void {
		const after = markdownFromElement(this.content);
		this.markdown = after;
		if (before !== after) this.options.onChange(before, after, kind);
	}

	private runCommand(command: string, value = ''): void {
		if (!command) return;
		const before = this.markdown;
		this.focus();
		this.commanding = true;
		try {
			switch (command) {
			case 'inlineCode':
				this.wrapInlineCode();
				break;
			case 'createLink': {
				const selection = window.getSelection();
				if (!selection || selection.isCollapsed || !this.content.contains(selection.anchorNode)) break;
				const href = window.prompt('リンク先URLまたは相対パスを入力してください');
				if (href) document.execCommand('createLink', false, href);
				break;
			}
			case 'formatBlock':
				document.execCommand('formatBlock', false, value || 'p');
				break;
			default:
				document.execCommand(command, false);
				break;
			}
		} finally {
			this.commanding = false;
		}
		this.commit(command, before);
	}

	private wrapInlineCode(): void {
		const selection = window.getSelection();
		if (!selection || !selection.rangeCount || !this.content.contains(selection.anchorNode)) return;
		const range = selection.getRangeAt(0);
		const code = document.createElement('code');
		if (range.collapsed) {
			code.textContent = 'コード';
			range.insertNode(code);
			range.selectNodeContents(code);
		} else {
			code.append(range.extractContents());
			range.insertNode(code);
			range.selectNodeContents(code);
		}
		selection.removeAllRanges();
		selection.addRange(range);
	}
}
