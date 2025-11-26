import * as assert from 'node:assert';

suite('Command Test Suite', () => {
  test('Package.json should have valid command definitions', () => {
    // This is a simple validation that the extension structure is correct
    // The actual command registration is tested by VS Code marketplace validation
    const packageJson = require('../../../package.json');

    assert.ok(packageJson.contributes, 'package.json should have contributes');
    assert.ok(
      packageJson.contributes.commands,
      'package.json should have commands'
    );
    assert.ok(
      Array.isArray(packageJson.contributes.commands),
      'commands should be an array'
    );

    // Verify core commands are defined in package.json
    const commandIds = packageJson.contributes.commands.map(
      (c: { command: string }) => c.command
    );
    const coreCommands = ['ung.installCli', 'ung.openDocs', 'ung.recheckCli'];

    for (const cmd of coreCommands) {
      assert.ok(
        commandIds.includes(cmd),
        `Core command ${cmd} should be defined in package.json`
      );
    }
  });

  test('Package.json should have valid view contributions', () => {
    const packageJson = require('../../../package.json');

    assert.ok(packageJson.contributes.views, 'package.json should have views');
    assert.ok(
      packageJson.contributes.viewsContainers,
      'package.json should have viewsContainers'
    );
  });
});
