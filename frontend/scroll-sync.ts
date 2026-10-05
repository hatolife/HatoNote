export interface MarkdownTableRange {
	startLine: number;
	endLine: number;
}

export interface ScrollRatioAnchor {
	source: number;
	preview: number;
}

function clampRatio(value: number): number {
	if (!Number.isFinite(value)) return 0;
	return Math.max(0, Math.min(1, value));
}

function hasUnescapedPipe(line: string): boolean {
	let escaped = false;
	for (const char of line) {
		if (escaped) { escaped = false; continue; }
		if (char === '\\') { escaped = true; continue; }
		if (char === '|') return true;
	}
	return false;
}

function splitTableRow(line: string): string[] {
	const cells: string[] = [];
	let cell = '';
	let escaped = false;
	for (const char of line.trim()) {
		if (escaped) { cell += char; escaped = false; continue; }
		if (char === '\\') { cell += char; escaped = true; continue; }
		if (char === '|') { cells.push(cell.trim()); cell = ''; continue; }
		cell += char;
	}
	cells.push(cell.trim());
	if (cells.length > 1 && cells[0] === '') cells.shift();
	if (cells.length > 1 && cells[cells.length - 1] === '') cells.pop();
	return cells;
}

function isTableDelimiter(line: string): boolean {
	if (!hasUnescapedPipe(line)) return false;
	const cells = splitTableRow(line);
	return cells.length > 0 && cells.every(cell => /^:?-{3,}:?$/.test(cell));
}

function isTableBodyLine(line: string): boolean {
	return line.trim() !== '' && hasUnescapedPipe(line);
}

// findMarkdownTableRanges はGFM表のヘッダー行から最終行までを1始まりの行番号で返します。
export function findMarkdownTableRanges(markdown: string): MarkdownTableRange[] {
	const lines = markdown.replace(/\r\n?/g, '\n').split('\n');
	const ranges: MarkdownTableRange[] = [];
	for (let delimiter = 1; delimiter < lines.length; delimiter++) {
		if (!isTableDelimiter(lines[delimiter]) || !hasUnescapedPipe(lines[delimiter - 1])) continue;
		const headerCells = splitTableRow(lines[delimiter - 1]);
		const delimiterCells = splitTableRow(lines[delimiter]);
		if (headerCells.length !== delimiterCells.length) continue;
		let end = delimiter + 1;
		while (end < lines.length && isTableBodyLine(lines[end])) end++;
		ranges.push({startLine:delimiter, endLine:end});
		delimiter = end;
	}
	return ranges;
}

function normalizedAnchors(anchors: ScrollRatioAnchor[]): ScrollRatioAnchor[] {
	const candidates = [{source:0,preview:0}, ...anchors, {source:1,preview:1}]
		.map(anchor => ({source:clampRatio(anchor.source),preview:clampRatio(anchor.preview)}))
		.sort((a,b) => a.source - b.source || a.preview - b.preview);
	const result: ScrollRatioAnchor[] = [];
	for (const anchor of candidates) {
		const previous = result[result.length - 1];
		if (!previous) { result.push(anchor); continue; }
		if (anchor.source <= previous.source || anchor.preview < previous.preview) continue;
		result.push(anchor);
	}
	if (result[result.length - 1]?.source !== 1 || result[result.length - 1]?.preview !== 1) result.push({source:1,preview:1});
	return result;
}

// mapAnchoredScrollRatio は表などで高さ比が変わる区間をアンカー間の線形補間で同期します。
export function mapAnchoredScrollRatio(value: number, anchors: ScrollRatioAnchor[], reverse = false): number {
	const points = normalizedAnchors(anchors);
	const input = clampRatio(value);
	const from = reverse ? 'preview' : 'source';
	const to = reverse ? 'source' : 'preview';
	for (let i = 1; i < points.length; i++) {
		const before = points[i - 1], after = points[i];
		if (input > after[from]) continue;
		const span = after[from] - before[from];
		if (span <= 0) return after[to];
		const position = (input - before[from]) / span;
		return clampRatio(before[to] + position * (after[to] - before[to]));
	}
	return 1;
}
