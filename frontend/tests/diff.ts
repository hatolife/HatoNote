import assert from 'node:assert/strict';
import { diffLines } from '../diff.ts';

assert.deepEqual(diffLines('a\nb\nc\n', 'a\nB\nc\n'), [
	{kind:'same',historyLine:1,currentLine:1,history:'a',current:'a'},
	{kind:'change',historyLine:2,currentLine:2,history:'b',current:'B'},
	{kind:'same',historyLine:3,currentLine:3,history:'c',current:'c'},
]);

assert.deepEqual(diffLines('a\nc\n', 'a\nb\nc\n').map(row => row.kind), ['same','add','same']);
assert.deepEqual(diffLines('a\nb\nc\n', 'a\nc\n').map(row => row.kind), ['same','delete','same']);

const multiple = diffLines(
	'one\ntwo\nthree\nfour\nfive\n',
	'one\nTWO\nthree\ninserted\nfour\nFIVE\n',
);
assert.deepEqual(multiple.map(row => row.kind), ['same','change','same','add','same','change']);

const largeHistory = Array.from({length:5000}, (_, i) => 'line-'+i).join('\n');
const largeCurrent = largeHistory.replace('line-2500','changed');
const large = diffLines(largeHistory, largeCurrent);
assert.equal(large.filter(row => row.kind === 'change').length, 1);
