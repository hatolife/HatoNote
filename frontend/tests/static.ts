import assert from 'node:assert/strict';
import fs from 'node:fs';

const html = fs.readFileSync('dist/index.html', 'utf8');
const app = fs.readFileSync('app.ts', 'utf8');
const settings = fs.readFileSync('../internal/settings/settings.go', 'utf8');

function duplicates(values: readonly string[]): string[] {
	const seen = new Set<string>();
	const duplicate = new Set<string>();
	for (const value of values) {
		if (seen.has(value)) duplicate.add(value);
		seen.add(value);
	}
	return [...duplicate].sort();
}

const ids = [...html.matchAll(/\bid="([^"]+)"/g)].map(match => match[1]);
assert.deepEqual(duplicates(ids), [], 'duplicate HTML ids');

const commandIDs = [...app.matchAll(/commands\.register\(\{id:'([^']+)'/g)].map(match => match[1]);
assert(commandIDs.length > 0, 'no commands registered');
assert.deepEqual(duplicates(commandIDs), [], 'duplicate command ids');

const settingsFields = [...settings.matchAll(/`json:"([^"]+)"`/g)].map(match => match[1]);
const controlNames = [...html.matchAll(/\b(?:input|select)[^>]*\bname="([^"]+)"/g)].map(match => match[1]);
const missingControls = settingsFields.filter(name => !controlNames.includes(name));
assert.deepEqual(missingControls, [], 'settings without UI controls');

const userTemplateIDs = [
	'user-template-list',
	'user-template-scope',
	'user-template-title',
	'user-template-content',
	'user-template-error',
	'user-template-delete',
	'user-template-save',
];
for (const id of userTemplateIDs) {
	assert.equal(ids.filter(value => value === id).length, 1, `template manager id ${id}`);
}

const navigationIDs = ['nav-back','nav-forward','quick-open-dialog','quick-open-input','quick-open-results'];
for (const id of navigationIDs) {
	assert.equal(ids.filter(value => value === id).length, 1, `navigation id ${id}`);
}
for (const command of ['navigation.quickOpen','navigation.back','navigation.forward']) {
	assert(commandIDs.includes(command), `missing navigation command ${command}`);
}

const diagnosticsIDs = ['document-diagnostics-dialog','document-diagnostics-title','document-diagnostics-summary','document-diagnostics-results'];
for (const id of diagnosticsIDs) {
	assert.equal(ids.filter(value => value === id).length, 1, `document diagnostics id ${id}`);
}
assert(commandIDs.includes('document.diagnostics'), 'missing document diagnostics command');

const commandizedFeatureCommands = [
	'image.add',
	'document.convertToSlides',
	'document.convertToMDZ',
	'recovery.open',
	'editor.builtin',
	'editor.wysiwyg',
	'editor.neovim',
	'presentation.startCurrent',
	'presentation.startFirst',
	'slides.previous',
	'slides.next',
];
for (const command of commandizedFeatureCommands) {
	assert(commandIDs.includes(command), `missing feature command ${command}`);
}

// COMMAND-006, COMMAND-038〜047: 操作部品からCommand経路を外さないことを固定します。
function commandBinding(control: string, command: string): RegExp {
	const escapedCommand = command.replaceAll('.', '\\.');
	return new RegExp("element\\('" + control + "'\\)\\.onclick\\s*=\\s*\\(\\)\\s*=>\\s*runCommand\\('" + escapedCommand + "'\\)");
}
for (const [control, command] of [
	['image','image.add'],
	['document-to-slides','document.convertToSlides'],
	['slides-to-document','document.convertToMDZ'],
	['recovery','recovery.open'],
	['welcome-recovery','recovery.open'],
	['slides-present','presentation.startCurrent'],
	['slides-start','presentation.startFirst'],
	['slide-prev','slides.previous'],
	['slide-next','slides.next'],
] as const) {
	assert.match(app, commandBinding(control, command), `control ${control} bypasses command ${command}`);
}
for (const [engine, command] of [
	['builtin','editor.builtin'],
	['wysiwyg','editor.wysiwyg'],
	['neovim','editor.neovim'],
] as const) {
	assert(app.includes(`${engine}:'${command}'`), `editor engine ${engine} bypasses command ${command}`);
}
assert.match(app, /button\.onclick\s*=\s*\(\)\s*=>\s*runCommand\(editorEngineCommands\[button\.dataset\.engine as EditorEngine\]\)/, 'editor engine button bypasses command registry');

console.log(`checked ${ids.length} HTML ids, ${commandIDs.length} commands, ${settingsFields.length} settings`);
