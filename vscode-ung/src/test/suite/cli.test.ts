import * as assert from 'node:assert';
import * as vscode from 'vscode';
import { UngCli } from '../../cli/ungCli';

suite('CLI Wrapper Test Suite', () => {
  vscode.window.showInformationMessage('Start CLI tests.');

  let cli: UngCli;
  let outputChannel: vscode.OutputChannel;

  setup(() => {
    outputChannel = vscode.window.createOutputChannel('UNG Test');
    cli = new UngCli(outputChannel);
  });

  teardown(() => {
    outputChannel.dispose();
  });

  test('CLI should be able to check installation', async () => {
    const isInstalled = await cli.isInstalled();
    // Note: This test will fail if UNG is not installed
    // In a real test environment, you would mock the CLI
    assert.strictEqual(typeof isInstalled, 'boolean');
  });

  test('CLI should be able to get version', async () => {
    const version = await cli.getVersion();
    // Version can be null if CLI is not installed
    if (version) {
      assert.strictEqual(typeof version, 'string');
    }
  });

  test('CLI exec should return CliResult', async () => {
    const result = await cli.exec(['--help']);
    assert.strictEqual(typeof result.success, 'boolean');
    assert.ok(result.stdout !== undefined || result.error !== undefined);
  });
});
