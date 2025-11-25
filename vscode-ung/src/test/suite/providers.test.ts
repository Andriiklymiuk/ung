import * as assert from 'assert';
import * as vscode from 'vscode';
import { UngCli } from '../../cli/ungCli';
import { InvoiceProvider } from '../../views/invoiceProvider';
import { ClientProvider } from '../../views/clientProvider';

suite('Provider Test Suite', () => {
    vscode.window.showInformationMessage('Start provider tests.');

    let cli: UngCli;
    let outputChannel: vscode.OutputChannel;

    setup(() => {
        outputChannel = vscode.window.createOutputChannel('UNG Test');
        cli = new UngCli(outputChannel);
    });

    teardown(() => {
        outputChannel.dispose();
    });

    test('InvoiceProvider should be instantiable', () => {
        const provider = new InvoiceProvider(cli);
        assert.ok(provider);
        assert.ok(provider.getChildren);
        assert.ok(provider.getTreeItem);
    });

    test('ClientProvider should be instantiable', () => {
        const provider = new ClientProvider(cli);
        assert.ok(provider);
        assert.ok(provider.getChildren);
        assert.ok(provider.getTreeItem);
    });

    test('Providers should handle empty data gracefully', async () => {
        const invoiceProvider = new InvoiceProvider(cli);
        const children = await invoiceProvider.getChildren();
        assert.ok(Array.isArray(children));
    });
});
