export type LineDiffKind = 'same' | 'add' | 'delete' | 'change';

export interface LineDiffRow {
	kind: LineDiffKind;
	historyLine?: number;
	currentLine?: number;
	history: string;
	current: string;
}

const lookahead = 80;

function lines(text: string): string[] {
	if (!text) return [];
	const normalized = text.replace(/\r\n/g, '\n').replace(/\r/g, '\n');
	const result = normalized.split('\n');
	if (normalized.endsWith('\n')) result.pop();
	return result;
}

function findAhead(values: readonly string[], start: number, target: string): number {
	const end = Math.min(values.length, start + lookahead + 1);
	for (let index = start + 1; index < end; index++) {
		if (values[index] === target) return index;
	}
	return -1;
}

export function diffLines(historyText: string, currentText: string): LineDiffRow[] {
	const history = lines(historyText);
	const current = lines(currentText);
	const result: LineDiffRow[] = [];
	let historyIndex = 0;
	let currentIndex = 0;

	while (historyIndex < history.length || currentIndex < current.length) {
		if (historyIndex >= history.length) {
			result.push({
				kind: 'add',
				currentLine: currentIndex + 1,
				history: '',
				current: current[currentIndex],
			});
			currentIndex++;
			continue;
		}
		if (currentIndex >= current.length) {
			result.push({
				kind: 'delete',
				historyLine: historyIndex + 1,
				history: history[historyIndex],
				current: '',
			});
			historyIndex++;
			continue;
		}

		const historyValue = history[historyIndex];
		const currentValue = current[currentIndex];
		if (historyValue === currentValue) {
			result.push({
				kind: 'same',
				historyLine: historyIndex + 1,
				currentLine: currentIndex + 1,
				history: historyValue,
				current: currentValue,
			});
			historyIndex++;
			currentIndex++;
			continue;
		}

		const currentAnchor = findAhead(current, currentIndex, historyValue);
		const historyAnchor = findAhead(history, historyIndex, currentValue);
		const additions = currentAnchor < 0 ? Number.POSITIVE_INFINITY : currentAnchor - currentIndex;
		const deletions = historyAnchor < 0 ? Number.POSITIVE_INFINITY : historyAnchor - historyIndex;

		if (additions < Number.POSITIVE_INFINITY && additions <= deletions) {
			while (currentIndex < currentAnchor) {
				result.push({
					kind: 'add',
					currentLine: currentIndex + 1,
					history: '',
					current: current[currentIndex],
				});
				currentIndex++;
			}
			continue;
		}
		if (deletions < Number.POSITIVE_INFINITY) {
			while (historyIndex < historyAnchor) {
				result.push({
					kind: 'delete',
					historyLine: historyIndex + 1,
					history: history[historyIndex],
					current: '',
				});
				historyIndex++;
			}
			continue;
		}

		result.push({
			kind: 'change',
			historyLine: historyIndex + 1,
			currentLine: currentIndex + 1,
			history: historyValue,
			current: currentValue,
		});
		historyIndex++;
		currentIndex++;
	}
	return result;
}
