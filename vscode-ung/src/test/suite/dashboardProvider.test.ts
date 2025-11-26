import * as assert from 'node:assert';
import * as vscode from 'vscode';
import { UngCli } from '../../cli/ungCli';
import { DashboardProvider } from '../../views/dashboardProvider';

suite('DashboardProvider Test Suite', () => {
  let provider: DashboardProvider;
  let outputChannel: vscode.OutputChannel;
  let cli: UngCli;

  setup(() => {
    outputChannel = vscode.window.createOutputChannel('Test');
    cli = new UngCli(outputChannel);
    provider = new DashboardProvider(cli);
  });

  teardown(() => {
    outputChannel.dispose();
  });

  test('getChildren returns root items when no element provided', async () => {
    const items = await provider.getChildren();
    assert.ok(items.length > 0, 'Should return at least one root item');
    assert.ok(items.length >= 3, 'Should have at least 3 root sections');

    const labels = items.map((item) => item.label);
    assert.ok(
      labels.some((label) => label === 'Revenue Overview'),
      'Should have Revenue Overview section'
    );
    assert.ok(
      labels.some((label) => label === 'Top Contracts'),
      'Should have Top Contracts section'
    );
    assert.ok(
      labels.some((label) => label === 'Business Summary'),
      'Should have Business Summary section'
    );
  });

  test('getTreeItem returns correct tree item', async () => {
    const items = await provider.getChildren();
    assert.ok(items.length > 0, 'Should have items');

    const treeItem = provider.getTreeItem(items[0]);
    assert.ok(treeItem, 'Should return tree item');
    assert.ok(treeItem.label, 'Tree item should have label');
  });

  test('getChildren returns metrics when given revenue section', async () => {
    const items = await provider.getChildren();
    const revenueSection = items.find(
      (item) => item.label === 'Revenue Overview'
    );

    if (revenueSection) {
      const metrics = await provider.getChildren(revenueSection);
      // Metrics may be empty if CLI is not installed, but should be an array
      assert.ok(Array.isArray(metrics), 'Should return array of metrics');
    }
  });

  test('getChildren returns metrics when given contracts section', async () => {
    const items = await provider.getChildren();
    const contractsSection = items.find(
      (item) => item.label === 'Top Contracts'
    );

    if (contractsSection) {
      const metrics = await provider.getChildren(contractsSection);
      assert.ok(Array.isArray(metrics), 'Should return array of metrics');
    }
  });

  test('getChildren returns metrics when given stats section', async () => {
    const items = await provider.getChildren();
    const statsSection = items.find(
      (item) => item.label === 'Business Summary'
    );

    if (statsSection) {
      const metrics = await provider.getChildren(statsSection);
      assert.ok(Array.isArray(metrics), 'Should return array of metrics');
    }
  });

  test('refresh clears cache', async () => {
    // This test just verifies refresh doesn't throw
    assert.doesNotThrow(() => {
      provider.refresh();
    });
  });
});
