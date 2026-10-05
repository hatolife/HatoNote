import assert from 'node:assert/strict';
import { findMarkdownTableRanges, mapAnchoredScrollRatio } from '../scroll-sync.ts';

const markdown = [
	'# 前',
	'',
	'| No | 内容 |',
	'| --- | --- |',
	'| 1 | a |',
	'| 2 | b |',
	'',
	'本文 | 表ではない',
	'',
	'| A | B |',
	'| :--- | ---: |',
	'| x | y |',
].join('\n');
assert.deepEqual(findMarkdownTableRanges(markdown), [
	{startLine:3,endLine:6},
	{startLine:10,endLine:12},
], 'SCROLL-001: Markdown表の行範囲を抽出できる');

const anchors = [
	{source:0.2,preview:0.1},
	{source:0.4,preview:0.7},
];
assert.ok(Math.abs(mapAnchoredScrollRatio(0.3, anchors) - 0.4) < 1e-9, 'SCROLL-001: 表区間をプレビュー比率へ補間できる');
assert.ok(Math.abs(mapAnchoredScrollRatio(0.4, anchors, true) - 0.3) < 1e-9, 'SCROLL-001: プレビュー比率から表区間を逆変換できる');
assert.equal(mapAnchoredScrollRatio(0, anchors), 0);
assert.equal(mapAnchoredScrollRatio(1, anchors), 1);

console.log('scroll sync checks passed');
