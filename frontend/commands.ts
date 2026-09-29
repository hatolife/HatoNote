export interface Command {
	id: string;
	title: string;
	shortcut?: string;
	enabled?: () => boolean;
	execute: () => void | Promise<void>;
}

export interface CommandState {
	id: string;
	title: string;
	shortcut?: string;
	enabled: boolean;
}

export class CommandRegistry {
	private readonly commands = new Map<string, Command>();

	register(command: Command): void {
		const id = command.id.trim();
		if (!id) throw new Error('command id is required');
		if (this.commands.has(id)) throw new Error(`duplicate command: ${id}`);
		this.commands.set(id, {...command, id});
	}

	list(): CommandState[] {
		return [...this.commands.values()].map(command => ({
			id: command.id,
			title: command.title,
			shortcut: command.shortcut,
			enabled: command.enabled?.() ?? true,
		}));
	}

	async execute(id: string): Promise<boolean> {
		const command = this.commands.get(id);
		if (!command) throw new Error(`unknown command: ${id}`);
		if (!(command.enabled?.() ?? true)) return false;
		await command.execute();
		return true;
	}

	async executeSequence(ids: readonly string[]): Promise<string[]> {
		const executed: string[] = [];
		for (const id of ids) {
			if (await this.execute(id)) executed.push(id);
		}
		return executed;
	}
}
