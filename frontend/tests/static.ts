import assert from 'node:assert/strict';
import fs from 'node:fs';

const html = fs.readFileSync('dist/index.html', 'utf8');
const app = fs.readFileSync('app.ts', 'utf8');
const nvim = fs.readFileSync('nvim.ts', 'utf8');
const audience = fs.readFileSync('audience.ts', 'utf8');
const audienceHTML = fs.readFileSync('dist/audience.html', 'utf8');
const style = fs.readFileSync('dist/style.css', 'utf8');
const audienceGo = fs.readFileSync('../audience.go', 'utf8');
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

const settingsCategories = [...html.matchAll(/data-settings-category="([^"]+)"/g)].map(match => match[1]);
const settingsPanels = [...html.matchAll(/data-settings-panel="([^"]+)"/g)].map(match => match[1]);
assert.deepEqual(settingsCategories, ['appearance','mdbook','editor','images','templates','markdown','diagrams','storage'], 'settings categories');
assert.deepEqual(settingsPanels, settingsCategories, 'settings category panels');

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

const diagnosticsIDs = ['document-diagnostics-dialog','document-diagnostics-title','document-diagnostics-summary','document-diagnostics-results','log-panel','log-output','log-panel-close'];
for (const id of diagnosticsIDs) {
	assert.equal(ids.filter(value => value === id).length, 1, `document diagnostics id ${id}`);
}
assert(commandIDs.includes('document.diagnostics'), 'missing document diagnostics command');
assert(app.includes('api.DiagnosticLog()'), 'diagnostic log drawer does not load the log');
assert(app.includes('api.LogStatus(message, error)'), 'status messages are not written to the diagnostic log');
assert(style.includes('#log-panel.open'), 'diagnostic log drawer open style is missing');
assert(html.includes('id="window-resize-right"'), 'right edge resize handle is missing');
assert(app.includes("window.WailsInvoke('resize:e-resize')"), 'right edge handle does not start native resize');
assert(style.includes('#window-resize-right'), 'right edge resize handle style is missing');
assert(!html.includes('<h1>HatoNote</h1>'), 'start screen repeats the product name');
assert(!html.includes('ここに .md / .mdz をドロップ'), 'start screen still has a dedicated drop zone');
assert(html.includes('.md / .mdz はウィンドウにD&amp;Dして開けます'), 'start screen drag-and-drop hint is missing');
assert(app.includes("const folderExpansionOverrides = new Map<string, boolean>();"), 'folder expansion overrides are missing');
assert(app.includes("folderExpansionOverrides.get(folder) ?? currentFolders.has(folder)"), 'current file folders are not expanded by default');
assert(app.includes("folder.split('/').pop() + '/'"), 'folder labels do not end with slash');
assert(app.includes("b.textContent = `${'#'.repeat(heading.level)} ${heading.text}`;"), 'markdown heading labels do not include heading markers');
assert(app.includes("state.pages.length > 1 && state.pages.every(name => !name.includes('/'))"), 'page view is not limited to flat multi-page MDZ documents');
assert(app.includes("if (!hasPageView && documentView === 'pages') documentView = 'files';"), 'nested Markdown documents do not fall back to file view');
assert(app.includes("element<HTMLButtonElement>('pages-tab').hidden = !hasPageView;"), 'page tab is not hidden for nested Markdown documents');
assert(app.includes("if (name === current && state.pages.includes(name)) appendMarkdownHeadings(nav, depth);"), 'file tree does not show headings under the current Markdown file');
assert(app.includes("if (view === 'pages' && !pageViewAvailable()) return;"), 'hidden page mode can still be selected for nested Markdown documents');
assert(app.includes("else renderFileTree(element('pages'));"), 'file tree is not refreshed after Markdown headings are rendered');

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

// EDITOR-NVIM-013〜016: Neovim編集中にアプリのグローバルショートカットがキーを横取りしないことを固定します。
assert(app.includes("function neovimOwnsKeyboard(): boolean { return editing && activeEngine === 'neovim'; }"), 'Neovim shortcut guard is missing');
assert(app.includes("if (neovimOwnsKeyboard() || !(event.ctrlKey || event.metaKey)) return;"), 'Ctrl/Cmd shortcuts are not disabled in Neovim');
assert(app.includes("if (neovimOwnsKeyboard() || !event.altKey"), 'Alt shortcuts are not disabled in Neovim');
assert(nvim.includes("(event.ctrlKey || event.metaKey) && event.key.toLowerCase() === 'v'"), 'Neovim is still yielding non-paste shortcuts to the app');

// SCR-PRESENT-008 / SCR-PRESENTER-037: 全画面表示ではアプリ側・投影側ともタイトルバーを残さないことを固定します。
assert(app.includes("document.body.classList.add('presentation-active')"), 'presentation does not enter titlebar-hidden state');
assert(app.includes("document.body.classList.remove('presentation-active')"), 'presentation does not leave titlebar-hidden state');
assert(style.includes('body.presentation-active #titlebar{display:none}'), 'presentation titlebar hide rule is missing');
assert(audienceGo.includes('Frameless: true'), 'audience window is not frameless');
assert(audienceHTML.includes('id="audience-titlebar"'), 'windowed audience titlebar is missing');
assert(audienceHTML.includes('body.fullscreen #audience-titlebar{display:none}'), 'audience fullscreen titlebar hide rule is missing');
assert(audience.includes("document.body.classList.toggle('fullscreen',fullscreen)"), 'audience fullscreen state is not reflected in the DOM');

console.log(`checked ${ids.length} HTML ids, ${commandIDs.length} commands, ${settingsFields.length} settings`);
