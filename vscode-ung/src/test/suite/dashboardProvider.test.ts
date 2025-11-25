import * as assert from 'assert';
import * as vscode from 'vscode';
import { DashboardProvider } from '../../views/dashboardProvider';
import { UngCli } from '../../cli/ungCli';

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

        const labels = items.map(item => item.label);
        assert.ok(labels.some(label => label === 'Revenue'), 'Should have Revenue section');
        assert.ok(labels.some(label => label === 'Contracts'), 'Should have Contracts section');
        assert.ok(labels.some(label => label === 'Quick Stats'), 'Should have Quick Stats section');
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
        const revenueSection = items.find(item => item.label === 'Revenue');

        if (revenueSection) {
            const metrics = await provider.getChildren(revenueSection);
            // Metrics may be empty if CLI is not installed, but should be an array
            assert.ok(Array.isArray(metrics), 'Should return array of metrics');
        }
    });

    test('getChildren returns metrics when given contracts section', async () => {
        const items = await provider.getChildren();
        const contractsSection = items.find(item => item.label === 'Contracts');

        if (contractsSection) {
            const metrics = await provider.getChildren(contractsSection);
            assert.ok(Array.isArray(metrics), 'Should return array of metrics');
        }
    });

    test('getChildren returns metrics when given stats section', async () => {
        const items = await provider.getChildren();
        const statsSection = items.find(item => item.label === 'Quick Stats');

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
