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

console.log(`checked ${ids.length} HTML ids, ${commandIDs.length} commands, ${settingsFields.length} settings`);
