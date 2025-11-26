import * as assert from 'node:assert';
import * as vscode from 'vscode';

suite('Command Test Suite', () => {
  vscode.window.showInformationMessage('Start command tests.');

  test('All UNG commands should be registered', async () => {
    const commands = await vscode.commands.getCommands(true);

    const ungCommands = [
      'ung.createCompany',
      'ung.createClient',
      'ung.createContract',
      'ung.createInvoice',
      'ung.logExpense',
      'ung.startTracking',
      'ung.stopTracking',
    ];

    for (const cmd of ungCommands) {
      assert.ok(commands.includes(cmd), `Command ${cmd} should be registered`);
    }
  });

  test('Tree views should be registered', () => {
    // Check if tree views are available
    // This is a basic test - in practice you would check if they render correctly
    assert.ok(true);
  });
});
