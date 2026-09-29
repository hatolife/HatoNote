import assert from 'node:assert/strict';
import { CommandRegistry } from '../commands.ts';

const calls: string[] = [];
const registry = new CommandRegistry();
let enabled = true;

registry.register({id:'document.open',title:'開く',execute:()=>{calls.push('open');}});
registry.register({id:'document.save',title:'保存',enabled:()=>enabled,execute:()=>{calls.push('save');}});

assert.deepEqual(registry.list().map(command => [command.id,command.enabled]), [
	['document.open',true],
	['document.save',true],
]);
assert.equal(await registry.execute('document.save'),true);
enabled=false;
assert.equal(await registry.execute('document.save'),false);
assert.deepEqual(await registry.executeSequence(['document.open','document.save']),['document.open']);
assert.deepEqual(calls,['save','open']);
assert.throws(()=>registry.register({id:'document.open',title:'duplicate',execute:()=>{}}),/duplicate command/);
await assert.rejects(()=>registry.execute('missing'),/unknown command/);
