export interface WysiwygOptions {
	onChange(before: string, after: string, kind: string): void;
	onSeparate(): void;
	onUndo(redo: boolean): void;
	resolveImage(reference: string): string | null;
}

const blockTags = new Set(['P','DIV','H1','H2','H3','H4','H5','H6','UL','OL','BLOCKQUOTE','PRE','TABLE','HR','DETAILS','SUMMARY','SECTION']);

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
	case 'SUP': {
		const citation = node.dataset.hatonoteCitation;
		if (citation) return `[@${citation}]`;
		return inner();
	}
	case 'SPAN': {
		const styles: string[] = [];
		if (node.style.color) styles.push(`color:${node.style.color}`);
		if (node.style.fontSize) styles.push(`font-size:${node.style.fontSize}`);
		return styles.length ? `<span style="${styles.join(';')}">${inner()}</span>` : inner();
	}
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

function serializeSlideColumns(node: HTMLElement): string {
	const columns = Array.from(node.children).filter((child): child is HTMLElement => child instanceof HTMLElement);
	if (!columns.length) return '';
	return columns.map(column => trimBlock(Array.from(column.childNodes).map(serializeBlock).join(''))).join('\n\n<!-- column -->\n\n') + '\n\n';
}

function decodeBase64URL(value: string): string {
	try {
		const normalized = value.replace(/-/g, '+').replace(/_/g, '/');
		const padded = normalized + '='.repeat((4 - normalized.length % 4) % 4);
		const bytes = Uint8Array.from(atob(padded), char => char.charCodeAt(0));
		return new TextDecoder().decode(bytes);
	} catch {
		return '';
	}
}

function serializeBlock(node: Node): string {
	if (node.nodeType === Node.TEXT_NODE) return (node.textContent || '').trim() ? escapeText(node.textContent || '') : '';
	if (!(node instanceof HTMLElement)) return '';
	if (node.tagName === 'DIV' && node.classList.contains('slide-columns')) return serializeSlideColumns(node);
	if (node.tagName === 'SECTION' && node.classList.contains('hatonote-references')) {
		const source = decodeBase64URL(node.dataset.hatonoteSource || '');
		return source ? source.trim() + '\n\n' : '';
	}
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
	case 'DETAILS': {
		const summary = Array.from(node.children).find(child => child.tagName === 'SUMMARY') as HTMLElement | undefined;
		const summaryText = summary ? trimBlock(Array.from(summary.childNodes).map(serializeInline).join('')) : '詳細';
		const body = trimBlock(Array.from(node.childNodes).filter(child => child !== summary).map(serializeBlock).join(''));
		return `<details${node.hasAttribute('open') ? ' open' : ''}>\n<summary>${summaryText}</summary>\n\n${body}\n\n</details>\n\n`;
	}
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
		if (preserveWhitespace) return text ? JSON.stringify(text) : '';
		const value = text.replace(/\s+/g, ' ').trim();
		return value ? JSON.stringify(value) : '';
	}
	if (!(node instanceof HTMLElement)) return '';
	const tag = node.tagName.toLowerCase();
	const preserve = preserveWhitespace || tag === 'pre' || tag === 'code';
	const attrs: string[] = [];
	for (const name of ['href','src','title','start','align','type','checked','class','open','id','data-hatonote-citation','data-hatonote-source']) {
		if (node.hasAttribute(name)) attrs.push(`${name}=${JSON.stringify(node.getAttribute(name) || '')}`);
	}
	if (tag === 'span') {
		const style = [node.style.color ? `color:${node.style.color}` : '', node.style.fontSize ? `font-size:${node.style.fontSize}` : ''].filter(Boolean).join(';');
		if (style) attrs.push(`style=${JSON.stringify(style)}`);
	}
	const children = Array.from(node.childNodes).map(child => canonicalNode(child, preserve)).filter(Boolean).join(',');
	return `<${tag}${attrs.length ? ' ' + attrs.join(' ') : ''}>${children}</${tag}>`;
}

export function equivalentRenderedHTML(left: string, right: string): boolean {
	const a = new DOMParser().parseFromString(left, 'text/html');
	const b = new DOMParser().parseFromString(right, 'text/html');
	return canonicalNode(a.body) === canonicalNode(b.body);
}

export interface RenderedHTMLDifference {
	path: string;
	reason: string;
	before: string;
	after: string;
	tag: string;
	text: string;
}

function diagnosticText(node: Node, preserveWhitespace: boolean): string {
	const text = node.textContent || '';
	const value = preserveWhitespace ? text : text.replace(/\s+/g, ' ').trim();
	return value.length > 120 ? value.slice(0, 117) + '...' : value;
}

function diagnosticNode(node: Node): string {
	if (node.nodeType === Node.TEXT_NODE) return JSON.stringify(diagnosticText(node, false));
	if (!(node instanceof HTMLElement)) return node.nodeName;
	const value = node.outerHTML.replace(/\s+/g, ' ').trim();
	return value.length > 240 ? value.slice(0, 237) + '...' : value;
}

function diagnosticAttributes(node: HTMLElement): Map<string,string> {
	const result = new Map<string,string>();
	for (const name of ['href','src','title','start','align','type','checked','class','open','id','data-hatonote-citation','data-hatonote-source']) {
		if (node.hasAttribute(name)) result.set(name, node.getAttribute(name) || '');
	}
	if (node.tagName.toLowerCase() === 'span') {
		const style = [node.style.color ? `color:${node.style.color}` : '', node.style.fontSize ? `font-size:${node.style.fontSize}` : ''].filter(Boolean).join(';');
		if (style) result.set('style', style);
	}
	return result;
}

function meaningfulDiagnosticChildren(node: Node, preserveWhitespace: boolean): Node[] {
	return Array.from(node.childNodes).filter(child => canonicalNode(child, preserveWhitespace) !== '');
}

function renderedHTMLDifferenceNode(left: Node, right: Node, path: string, preserveWhitespace = false): RenderedHTMLDifference | null {
	if (canonicalNode(left, preserveWhitespace) === canonicalNode(right, preserveWhitespace)) return null;
	const leftElement = left instanceof HTMLElement ? left : null;
	const rightElement = right instanceof HTMLElement ? right : null;
	const leftTag = leftElement?.tagName.toLowerCase() || '#text';
	const rightTag = rightElement?.tagName.toLowerCase() || '#text';
	const text = diagnosticText(left, preserveWhitespace);
	const difference = (reason: string): RenderedHTMLDifference => ({path, reason, before:diagnosticNode(left), after:diagnosticNode(right), tag:leftTag, text});
	if (left.nodeType !== right.nodeType || leftTag !== rightTag) return difference(`要素が変化しました（${leftTag} → ${rightTag}）`);
	if (!leftElement || !rightElement) return difference('テキストが変化しました');
	const leftAttributes = diagnosticAttributes(leftElement);
	const rightAttributes = diagnosticAttributes(rightElement);
	for (const name of new Set([...leftAttributes.keys(), ...rightAttributes.keys()])) {
		if (leftAttributes.get(name) !== rightAttributes.get(name)) return difference(`属性 ${name} が変化しました`);
	}
	const preserve = preserveWhitespace || leftTag === 'pre' || leftTag === 'code';
	const leftChildren = meaningfulDiagnosticChildren(left, preserve);
	const rightChildren = meaningfulDiagnosticChildren(right, preserve);
	const count = Math.min(leftChildren.length, rightChildren.length);
	for (let index = 0; index < count; index++) {
		if (canonicalNode(leftChildren[index], preserve) === canonicalNode(rightChildren[index], preserve)) continue;
		const child = leftChildren[index];
		const label = child instanceof HTMLElement ? child.tagName.toLowerCase() : '#text';
		const nested = renderedHTMLDifferenceNode(child, rightChildren[index], `${path} > ${label}[${index + 1}]`, preserve);
		if (nested) return nested;
	}
	if (leftChildren.length !== rightChildren.length) return difference(`子要素数が変化しました（${leftChildren.length} → ${rightChildren.length}）`);
	return difference('DOM構造が変化しました');
}

export function renderedHTMLDifference(left: string, right: string): RenderedHTMLDifference | null {
	if (equivalentRenderedHTML(left, right)) return null;
	const a = new DOMParser().parseFromString(left, 'text/html');
	const b = new DOMParser().parseFromString(right, 'text/html');
	return renderedHTMLDifferenceNode(a.body, b.body, 'body') || {path:'body', reason:'DOM構造が変化しました', before:diagnosticNode(a.body), after:diagnosticNode(b.body), tag:'body', text:diagnosticText(a.body, false)};
}

export function renderedHTMLHasOmittedRawHTML(html: string): boolean {
	return /raw HTML omitted/i.test(html);
}

// WysiwygEditor は表示用HTMLを直接編集し、変更時だけMarkdownへ戻します。
export class WysiwygEditor {
	private markdown = '';
	private beforeInput = '';
	private commanding = false;
	private savedRange: Range | null = null;
	private tableTools: HTMLElement | null = null;
	private colorInput: HTMLInputElement | null = null;
	private fontSizeSelect: HTMLSelectElement | null = null;

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
		this.tableTools = this.toolbar.querySelector<HTMLElement>('#wysiwyg-table-tools');
		for (const button of this.toolbar.querySelectorAll<HTMLButtonElement>('[data-wysiwyg-table-command]')) {
			button.addEventListener('mousedown', event => event.preventDefault());
			button.addEventListener('click', () => this.runTableCommand(button.dataset.wysiwygTableCommand || ''));
		}
		this.colorInput = this.toolbar.querySelector<HTMLInputElement>('#wysiwyg-color');
		this.colorInput?.addEventListener('change', () => this.applyInlineStyle('color', this.colorInput!.value));
		this.fontSizeSelect = this.toolbar.querySelector<HTMLSelectElement>('#wysiwyg-font-size');
		this.fontSizeSelect?.addEventListener('change', () => {
			const value = this.fontSizeSelect!.value;
			if (value) this.applyInlineStyle('fontSize', `${value}px`);
			this.fontSizeSelect!.value = '';
		});
		const updateTable = () => this.updateTableTools();
		this.content.addEventListener('keyup', updateTable);
		this.content.addEventListener('mouseup', updateTable);
		this.content.addEventListener('click', updateTable);
		document.addEventListener('selectionchange', () => {
			if (this.root.hidden) return;
			const selection = document.getSelection();
			if (selection?.rangeCount && this.contains(selection.anchorNode)) this.savedRange = selection.getRangeAt(0).cloneRange();
			if (this.contains(selection?.anchorNode || null)) this.updateTableTools();
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
		for (const managed of doc.querySelectorAll<HTMLElement>('.hatonote-references,sup[data-hatonote-citation]')) managed.contentEditable = 'false';
		this.commanding = true;
		this.content.innerHTML = doc.body.innerHTML || '<p><br></p>';
		this.commanding = false;
		this.markdown = markdown;
		this.beforeInput = '';
		this.updateTableTools();
	}

	setActive(active: boolean): void {
		this.root.hidden = !active;
		if (!active && this.tableTools) this.tableTools.hidden = true;
	}

	setEditable(editable: boolean): void {
		this.content.contentEditable = String(editable);
		this.toolbar.querySelectorAll<HTMLButtonElement | HTMLSelectElement | HTMLInputElement>('button,select,input').forEach(control => control.disabled = !editable);
	}

	focus(): void {
		this.content.focus({preventScroll:true});
	}

	contains(node: EventTarget | null): boolean {
		return node instanceof Node && this.root.contains(node);
	}

	rememberSelection(): void {
		const selection = window.getSelection();
		if (selection?.rangeCount && this.content.contains(selection.anchorNode)) this.savedRange = selection.getRangeAt(0).cloneRange();
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

	insertCitation(id: string): void {
		const before = this.markdown;
		const sup = document.createElement('sup');
		sup.dataset.hatonoteCitation = id;
		const link = document.createElement('a');
		link.href = '#hatonote-references';
		link.textContent = '[引用]';
		sup.append(link);
		this.insertNode(sup, true);
		this.commit('insertCitation', before);
	}

	private selectionRange(): Range | null {
		const selection = window.getSelection();
		if (selection?.rangeCount && this.content.contains(selection.anchorNode)) {
			this.savedRange = selection.getRangeAt(0).cloneRange();
			return selection.getRangeAt(0);
		}
		if (this.savedRange && this.content.contains(this.savedRange.commonAncestorContainer)) return this.savedRange.cloneRange();
		return null;
	}

	private insertNode(node: Node, selectContents = false): void {
		this.focus();
		const selection = window.getSelection();
		let range = this.selectionRange();
		if (!range) {
			range = document.createRange();
			range.selectNodeContents(this.content);
			range.collapse(false);
		}
		range.deleteContents();
		range.insertNode(node);
		range.setStartAfter(node);
		range.collapse(true);
		if (selectContents && node instanceof HTMLElement) range.selectNodeContents(node);
		selection?.removeAllRanges();
		selection?.addRange(range);
		this.savedRange = range.cloneRange();
	}

	private insertTable(): void {
		const before = this.markdown;
		const table = document.createElement('table');
		const head = table.createTHead().insertRow();
		for (let col = 0; col < 3; col++) {
			const cell = document.createElement('th');
			cell.textContent = `見出し${col + 1}`;
			head.append(cell);
		}
		const body = table.createTBody();
		for (let row = 0; row < 2; row++) {
			const tr = body.insertRow();
			for (let col = 0; col < 3; col++) tr.insertCell().textContent = 'セル';
		}
		this.insertNode(table);
		const first = table.rows[0]?.cells[0];
		if (first) {
			const range = document.createRange();
			range.selectNodeContents(first);
			const selection = window.getSelection();
			selection?.removeAllRanges();
			selection?.addRange(range);
		}
		this.commit('insertTable', before);
		this.updateTableTools();
	}

	private insertDetails(): void {
		const before = this.markdown;
		const details = document.createElement('details');
		details.open = true;
		const summary = document.createElement('summary');
		summary.textContent = '折りたたみ';
		const paragraph = document.createElement('p');
		paragraph.textContent = '内容';
		details.append(summary, paragraph);
		this.insertNode(details);
		const range = document.createRange();
		range.selectNodeContents(summary);
		const selection = window.getSelection();
		selection?.removeAllRanges();
		selection?.addRange(range);
		this.commit('insertDetails', before);
	}

	private applyInlineStyle(property: 'color' | 'fontSize', value: string): void {
		const range = this.selectionRange();
		if (!range) return;
		const before = this.markdown;
		const span = document.createElement('span');
		span.style[property] = value;
		if (range.collapsed) span.textContent = '文字';
		else span.append(range.extractContents());
		range.insertNode(span);
		range.selectNodeContents(span);
		const selection = window.getSelection();
		selection?.removeAllRanges();
		selection?.addRange(range);
		this.commit('formatStyle', before);
	}

	private currentCell(): HTMLTableCellElement | null {
		const selection = window.getSelection();
		const node = selection?.anchorNode;
		const element = node instanceof Element ? node : node?.parentElement;
		const cell = element?.closest('td,th');
		return cell instanceof HTMLTableCellElement && this.content.contains(cell) ? cell : null;
	}

	private updateTableTools(): void {
		if (!this.tableTools) return;
		this.tableTools.hidden = !this.currentCell();
	}

	private runTableCommand(command: string): void {
		const cell = this.currentCell();
		const row = cell?.parentElement as HTMLTableRowElement | null;
		const table = cell?.closest('table') as HTMLTableElement | null;
		if (!cell || !row || !table) return;
		const before = this.markdown;
		const rowIndex = row.rowIndex;
		const columnIndex = cell.cellIndex;
		const columnCount = Math.max(...Array.from(table.rows).map(item => item.cells.length));
		switch (command) {
		case 'rowBefore':
		case 'rowAfter': {
			const index = command === 'rowBefore' ? rowIndex : rowIndex + 1;
			const inserted = table.insertRow(index);
			for (let col = 0; col < columnCount; col++) {
				const newCell = inserted.insertCell();
				newCell.textContent = 'セル';
				const align = table.rows[index === 0 ? 1 : 0]?.cells[col]?.getAttribute('align');
				if (align) newCell.setAttribute('align', align);
			}
			break;
		}
		case 'rowDelete':
			if (table.rows.length > 1) table.deleteRow(rowIndex);
			break;
		case 'columnBefore':
		case 'columnAfter': {
			const index = command === 'columnBefore' ? columnIndex : columnIndex + 1;
			for (const tableRow of Array.from(table.rows)) {
				const newCell = tableRow.insertCell(Math.min(index, tableRow.cells.length));
				newCell.textContent = tableRow.rowIndex === 0 ? '見出し' : 'セル';
			}
			break;
		}
		case 'columnDelete':
			if (columnCount > 1) for (const tableRow of Array.from(table.rows)) if (tableRow.cells[columnIndex]) tableRow.deleteCell(columnIndex);
			break;
		case 'alignLeft':
		case 'alignCenter':
		case 'alignRight': {
			const align = command === 'alignLeft' ? 'left' : command === 'alignCenter' ? 'center' : 'right';
			for (const tableRow of Array.from(table.rows)) if (tableRow.cells[columnIndex]) tableRow.cells[columnIndex].setAttribute('align', align);
			break;
		}
		default:
			return;
		}
		this.commit('tableEdit', before);
		this.updateTableTools();
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
			case 'insertTable':
				this.commanding = false;
				this.insertTable();
				return;
			case 'insertDetails':
				this.commanding = false;
				this.insertDetails();
				return;
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
