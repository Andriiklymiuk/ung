import * as assert from 'node:assert';

/**
 * CLI Tests for UNG VSCode Extension
 *
 * Tests CLI response parsing, data transformation, and command building.
 * These tests use mocks since actual CLI execution requires installation.
 */

// Mock CLI response types (matching actual CLI output format)
interface MockClient {
  id: number;
  name: string;
  email: string;
  address?: string;
  tax_id?: string;
}

interface MockInvoice {
  id: number;
  invoice_num: string;
  amount: number;
  currency: string;
  status: string;
  issued_date: string;
  due_date: string;
}

// MockTrackingSession type used for testing CLI parsing
type MockTrackingSession = {
  id: number;
  project_name: string;
  start_time: string;
  end_time?: string;
  duration?: number;
  hours?: number;
  billable: boolean;
};
// Exported to avoid unused variable warning
export type { MockTrackingSession };

// CLI output parser utilities
const CliParser = {
  /**
   * Parse JSON output from CLI
   */
  parseJson<T>(output: string): T | null {
    try {
      return JSON.parse(output) as T;
    } catch {
      return null;
    }
  },

  /**
   * Parse CLI list output (tab-separated values)
   */
  parseTabular(output: string): string[][] {
    return output
      .split('\n')
      .filter((line) => line.trim().length > 0)
      .map((line) => line.split('\t'));
  },

  /**
   * Check if output indicates success
   */
  isSuccess(output: string): boolean {
    return output.includes('✓') || output.includes('successfully');
  },

  /**
   * Extract ID from success message
   */
  extractId(output: string): number | null {
    const match = output.match(/ID:\s*(\d+)/);
    return match ? Number.parseInt(match[1], 10) : null;
  },

  /**
   * Parse duration string to seconds
   */
  parseDuration(duration: string): number {
    let seconds = 0;
    const hourMatch = duration.match(/(\d+)h/);
    const minMatch = duration.match(/(\d+)m/);
    const secMatch = duration.match(/(\d+)s/);

    if (hourMatch) {
      seconds += Number.parseInt(hourMatch[1], 10) * 3600;
    }
    if (minMatch) {
      seconds += Number.parseInt(minMatch[1], 10) * 60;
    }
    if (secMatch) {
      seconds += Number.parseInt(secMatch[1], 10);
    }

    return seconds;
  },
};

// CLI command builder
const CommandBuilder = {
  /**
   * Build client add command
   */
  clientAdd(
    name: string,
    email: string,
    options?: { address?: string; taxId?: string }
  ): string[] {
    const args = ['client', 'add', '--name', name, '--email', email];
    if (options?.address) {
      args.push('--address', options.address);
    }
    if (options?.taxId) {
      args.push('--tax-id', options.taxId);
    }
    return args;
  },

  /**
   * Build invoice create command
   */
  invoiceCreate(options: {
    clientId: number;
    amount: number;
    currency?: string;
    dueDate?: string;
  }): string[] {
    const args = [
      'invoice',
      'create',
      '--client',
      options.clientId.toString(),
      '--amount',
      options.amount.toString(),
    ];
    if (options.currency) {
      args.push('--currency', options.currency);
    }
    if (options.dueDate) {
      args.push('--due-date', options.dueDate);
    }
    return args;
  },

  /**
   * Build tracking log command
   */
  trackLog(options: {
    contractId?: number;
    client?: string;
    hours: number;
    project?: string;
  }): string[] {
    const args = ['track', 'log', '--hours', options.hours.toString()];
    if (options.contractId) {
      args.push('--contract', options.contractId.toString());
    }
    if (options.client) {
      args.push('--client', options.client);
    }
    if (options.project) {
      args.push('--project', options.project);
    }
    return args;
  },
};

suite('CLI Test Suite', () => {
  suite('JSON Parsing Tests', () => {
    test('parses valid client JSON', () => {
      const json = '{"id":1,"name":"Test Client","email":"test@example.com"}';
      const result = CliParser.parseJson<MockClient>(json);
      assert.ok(result);
      assert.strictEqual(result.id, 1);
      assert.strictEqual(result.name, 'Test Client');
      assert.strictEqual(result.email, 'test@example.com');
    });

    test('parses client array JSON', () => {
      const json = '[{"id":1,"name":"Client A"},{"id":2,"name":"Client B"}]';
      const result = CliParser.parseJson<MockClient[]>(json);
      assert.ok(result);
      assert.strictEqual(result.length, 2);
      assert.strictEqual(result[0].name, 'Client A');
      assert.strictEqual(result[1].name, 'Client B');
    });

    test('returns null for invalid JSON', () => {
      const invalid = 'This is not JSON';
      const result = CliParser.parseJson<MockClient>(invalid);
      assert.strictEqual(result, null);
    });

    test('parses invoice JSON with dates', () => {
      const json = JSON.stringify({
        id: 1,
        invoice_num: 'INV-001',
        amount: 1500.0,
        currency: 'USD',
        status: 'pending',
        issued_date: '2025-01-15',
        due_date: '2025-02-15',
      });
      const result = CliParser.parseJson<MockInvoice>(json);
      assert.ok(result);
      assert.strictEqual(result.invoice_num, 'INV-001');
      assert.strictEqual(result.amount, 1500.0);
    });
  });

  suite('Tabular Parsing Tests', () => {
    test('parses tab-separated output', () => {
      const output =
        'ID\tNAME\tEMAIL\n1\tClient A\ta@test.com\n2\tClient B\tb@test.com';
      const result = CliParser.parseTabular(output);
      assert.strictEqual(result.length, 3);
      assert.deepStrictEqual(result[0], ['ID', 'NAME', 'EMAIL']);
      assert.deepStrictEqual(result[1], ['1', 'Client A', 'a@test.com']);
    });

    test('handles empty lines', () => {
      const output = 'ID\tNAME\n\n1\tTest\n\n';
      const result = CliParser.parseTabular(output);
      assert.strictEqual(result.length, 2);
    });

    test('handles single line', () => {
      const output = 'No clients found';
      const result = CliParser.parseTabular(output);
      assert.strictEqual(result.length, 1);
      assert.deepStrictEqual(result[0], ['No clients found']);
    });
  });

  suite('Success Detection Tests', () => {
    test('detects success with checkmark', () => {
      assert.strictEqual(
        CliParser.isSuccess('✓ Client added successfully (ID: 1)'),
        true
      );
    });

    test('detects success with "successfully" text', () => {
      assert.strictEqual(
        CliParser.isSuccess('Operation completed successfully'),
        true
      );
    });

    test('detects failure without markers', () => {
      assert.strictEqual(
        CliParser.isSuccess('Error: Something went wrong'),
        false
      );
    });
  });

  suite('ID Extraction Tests', () => {
    test('extracts ID from success message', () => {
      const id = CliParser.extractId('✓ Client added successfully (ID: 123)');
      assert.strictEqual(id, 123);
    });

    test('extracts ID with different format', () => {
      const id = CliParser.extractId('Created with ID: 456');
      assert.strictEqual(id, 456);
    });

    test('returns null when no ID found', () => {
      const id = CliParser.extractId('No ID in this message');
      assert.strictEqual(id, null);
    });
  });

  suite('Duration Parsing Tests', () => {
    test('parses hours and minutes', () => {
      assert.strictEqual(CliParser.parseDuration('2h 30m'), 9000);
    });

    test('parses hours only', () => {
      assert.strictEqual(CliParser.parseDuration('3h'), 10800);
    });

    test('parses minutes only', () => {
      assert.strictEqual(CliParser.parseDuration('45m'), 2700);
    });

    test('parses seconds only', () => {
      assert.strictEqual(CliParser.parseDuration('30s'), 30);
    });

    test('parses full duration', () => {
      assert.strictEqual(CliParser.parseDuration('1h 30m 45s'), 5445);
    });

    test('handles zero duration', () => {
      assert.strictEqual(CliParser.parseDuration('0s'), 0);
    });
  });

  suite('Command Builder Tests', () => {
    test('builds basic client add command', () => {
      const cmd = CommandBuilder.clientAdd('Test Client', 'test@example.com');
      assert.deepStrictEqual(cmd, [
        'client',
        'add',
        '--name',
        'Test Client',
        '--email',
        'test@example.com',
      ]);
    });

    test('builds client add command with options', () => {
      const cmd = CommandBuilder.clientAdd('Test Client', 'test@example.com', {
        address: '123 Test St',
        taxId: 'TAX123',
      });
      assert.ok(cmd.includes('--address'));
      assert.ok(cmd.includes('123 Test St'));
      assert.ok(cmd.includes('--tax-id'));
      assert.ok(cmd.includes('TAX123'));
    });

    test('builds invoice create command', () => {
      const cmd = CommandBuilder.invoiceCreate({
        clientId: 1,
        amount: 1500,
        currency: 'EUR',
      });
      assert.ok(cmd.includes('invoice'));
      assert.ok(cmd.includes('create'));
      assert.ok(cmd.includes('--client'));
      assert.ok(cmd.includes('1'));
      assert.ok(cmd.includes('--amount'));
      assert.ok(cmd.includes('1500'));
      assert.ok(cmd.includes('--currency'));
      assert.ok(cmd.includes('EUR'));
    });

    test('builds track log command', () => {
      const cmd = CommandBuilder.trackLog({
        hours: 2.5,
        project: 'Frontend Development',
      });
      assert.ok(cmd.includes('track'));
      assert.ok(cmd.includes('log'));
      assert.ok(cmd.includes('--hours'));
      assert.ok(cmd.includes('2.5'));
      assert.ok(cmd.includes('--project'));
      assert.ok(cmd.includes('Frontend Development'));
    });
  });

  suite('Data Validation Tests', () => {
    test('validates email format', () => {
      const validEmails = ['test@example.com', 'user.name@domain.co.uk'];
      const invalidEmails = ['notanemail', '@nodomain.com', 'no@domain'];

      for (const email of validEmails) {
        assert.ok(
          /^[^\s@]+@[^\s@]+\.[^\s@]+$/.test(email),
          `${email} should be valid`
        );
      }

      for (const email of invalidEmails) {
        assert.ok(
          !/^[^\s@]+@[^\s@]+\.[^\s@]+$/.test(email),
          `${email} should be invalid`
        );
      }
    });

    test('validates amount format', () => {
      const validAmounts = [0, 100, 1234.56, 0.01];
      const invalidAmounts = [-100, Number.NaN, Number.POSITIVE_INFINITY];

      for (const amount of validAmounts) {
        assert.ok(
          typeof amount === 'number' &&
            !Number.isNaN(amount) &&
            Number.isFinite(amount) &&
            amount >= 0,
          `${amount} should be valid`
        );
      }

      for (const amount of invalidAmounts) {
        assert.ok(
          Number.isNaN(amount) || !Number.isFinite(amount) || amount < 0,
          `${amount} should be invalid`
        );
      }
    });

    test('validates date format', () => {
      const validDates = ['2025-01-15', '2024-12-31', '2025-02-28'];
      const invalidDates = ['01-15-2025', '2025/01/15', 'invalid'];

      for (const date of validDates) {
        assert.ok(/^\d{4}-\d{2}-\d{2}$/.test(date), `${date} should be valid`);
      }

      for (const date of invalidDates) {
        assert.ok(
          !/^\d{4}-\d{2}-\d{2}$/.test(date),
          `${date} should be invalid`
        );
      }
    });
  });
});
