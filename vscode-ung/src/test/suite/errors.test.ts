import * as assert from 'node:assert';

/**
 * Error Tests for UNG VSCode Extension
 *
 * Tests error handling patterns and validation logic.
 * Uses mocks since actual errors.ts depends on vscode module.
 */

// Mock error types matching the actual implementation
enum ErrorType {
  CLI_NOT_INSTALLED = 'CLI_NOT_INSTALLED',
  CLI_EXECUTION_FAILED = 'CLI_EXECUTION_FAILED',
  PARSE_ERROR = 'PARSE_ERROR',
  VALIDATION_ERROR = 'VALIDATION_ERROR',
  NOT_FOUND = 'NOT_FOUND',
  PERMISSION_DENIED = 'PERMISSION_DENIED',
  NETWORK_ERROR = 'NETWORK_ERROR',
  TIMEOUT = 'TIMEOUT',
  UNKNOWN = 'UNKNOWN',
}

// Mock UngError class matching the actual implementation
class UngError extends Error {
  constructor(
    public type: ErrorType,
    message: string,
    public details?: unknown
  ) {
    super(message);
    this.name = 'UngError';
  }
}

// Mock validation functions matching ErrorHandler
const ErrorHandler = {
  validateInput(value: string | undefined, fieldName: string): string {
    if (!value || value.trim() === '') {
      throw new UngError(
        ErrorType.VALIDATION_ERROR,
        `${fieldName} is required`
      );
    }
    return value.trim();
  },

  validateNumber(
    value: string | undefined,
    fieldName: string,
    min?: number,
    max?: number
  ): number {
    const validatedString = ErrorHandler.validateInput(value, fieldName);
    const num = Number(validatedString);

    if (Number.isNaN(num)) {
      throw new UngError(
        ErrorType.VALIDATION_ERROR,
        `${fieldName} must be a valid number`
      );
    }

    if (min !== undefined && num < min) {
      throw new UngError(
        ErrorType.VALIDATION_ERROR,
        `${fieldName} must be at least ${min}`
      );
    }

    if (max !== undefined && num > max) {
      throw new UngError(
        ErrorType.VALIDATION_ERROR,
        `${fieldName} must be at most ${max}`
      );
    }

    return num;
  },

  fromCliError(error: string, stderr?: string): UngError {
    if (error.includes('not found') || error.includes('command not found')) {
      return new UngError(ErrorType.CLI_NOT_INSTALLED, error);
    }
    if (error.includes('timeout') || error.includes('timed out')) {
      return new UngError(ErrorType.TIMEOUT, error);
    }
    if (error.includes('permission denied') || error.includes('EACCES')) {
      return new UngError(ErrorType.PERMISSION_DENIED, error);
    }
    return new UngError(ErrorType.CLI_EXECUTION_FAILED, error, { stderr });
  },
};

suite('Errors Test Suite', () => {
  suite('UngError Tests', () => {
    test('creates error with type and message', () => {
      const error = new UngError(ErrorType.VALIDATION_ERROR, 'Test error');
      assert.strictEqual(error.type, ErrorType.VALIDATION_ERROR);
      assert.strictEqual(error.message, 'Test error');
      assert.strictEqual(error.name, 'UngError');
    });

    test('creates error with details', () => {
      const details = { field: 'email', value: 'invalid' };
      const error = new UngError(
        ErrorType.VALIDATION_ERROR,
        'Test error',
        details
      );
      assert.deepStrictEqual(error.details, details);
    });

    test('is instanceof Error', () => {
      const error = new UngError(ErrorType.UNKNOWN, 'Test');
      assert.ok(error instanceof Error);
    });
  });

  suite('ErrorType Tests', () => {
    test('has all expected error types', () => {
      const expectedTypes = [
        'CLI_NOT_INSTALLED',
        'CLI_EXECUTION_FAILED',
        'PARSE_ERROR',
        'VALIDATION_ERROR',
        'NOT_FOUND',
        'PERMISSION_DENIED',
        'NETWORK_ERROR',
        'TIMEOUT',
        'UNKNOWN',
      ];

      for (const type of expectedTypes) {
        assert.ok(type in ErrorType, `ErrorType should contain ${type}`);
      }
    });
  });

  suite('CLI Error Classification Tests', () => {
    test('classifies CLI not found errors', () => {
      const error = ErrorHandler.fromCliError('ung: command not found');
      assert.strictEqual(error.type, ErrorType.CLI_NOT_INSTALLED);
    });

    test('classifies timeout errors', () => {
      const error = ErrorHandler.fromCliError('operation timed out');
      assert.strictEqual(error.type, ErrorType.TIMEOUT);
    });

    test('classifies permission errors', () => {
      const error = ErrorHandler.fromCliError('permission denied');
      assert.strictEqual(error.type, ErrorType.PERMISSION_DENIED);
    });

    test('classifies EACCES errors', () => {
      const error = ErrorHandler.fromCliError('EACCES: permission denied');
      assert.strictEqual(error.type, ErrorType.PERMISSION_DENIED);
    });

    test('classifies unknown errors as CLI execution failed', () => {
      const error = ErrorHandler.fromCliError('some unknown error');
      assert.strictEqual(error.type, ErrorType.CLI_EXECUTION_FAILED);
    });
  });

  suite('Input Validation Tests', () => {
    test('validates required string input', () => {
      assert.strictEqual(ErrorHandler.validateInput('test', 'field'), 'test');
      assert.strictEqual(
        ErrorHandler.validateInput('  test  ', 'field'),
        'test'
      );
    });

    test('throws on empty input', () => {
      assert.throws(
        () => ErrorHandler.validateInput('', 'field'),
        /field is required/
      );
    });

    test('throws on undefined input', () => {
      assert.throws(
        () => ErrorHandler.validateInput(undefined, 'field'),
        /field is required/
      );
    });

    test('throws on whitespace-only input', () => {
      assert.throws(
        () => ErrorHandler.validateInput('   ', 'field'),
        /field is required/
      );
    });
  });

  suite('Numeric Validation Tests', () => {
    test('validates valid numbers', () => {
      assert.strictEqual(ErrorHandler.validateNumber('100', 'amount'), 100);
      assert.strictEqual(ErrorHandler.validateNumber('3.14', 'rate'), 3.14);
      assert.strictEqual(ErrorHandler.validateNumber('0', 'count'), 0);
      assert.strictEqual(ErrorHandler.validateNumber('-5', 'offset'), -5);
    });

    test('throws on empty input', () => {
      assert.throws(
        () => ErrorHandler.validateNumber('', 'amount'),
        /amount is required/
      );
    });

    test('throws on non-numeric input', () => {
      assert.throws(
        () => ErrorHandler.validateNumber('abc', 'amount'),
        /must be a valid number/
      );
    });

    test('throws when below minimum', () => {
      assert.throws(
        () => ErrorHandler.validateNumber('5', 'amount', 10),
        /must be at least 10/
      );
    });

    test('throws when above maximum', () => {
      assert.throws(
        () => ErrorHandler.validateNumber('100', 'rate', 0, 50),
        /must be at most 50/
      );
    });

    test('accepts value at minimum', () => {
      assert.strictEqual(ErrorHandler.validateNumber('10', 'amount', 10), 10);
    });

    test('accepts value at maximum', () => {
      assert.strictEqual(ErrorHandler.validateNumber('50', 'rate', 0, 50), 50);
    });
  });

  suite('Error Message Formatting Tests', () => {
    test('formats CLI not installed error', () => {
      const error = new UngError(
        ErrorType.CLI_NOT_INSTALLED,
        'ung command not found'
      );
      assert.ok(error.message.includes('not found'));
    });

    test('formats validation error with field name', () => {
      const error = new UngError(
        ErrorType.VALIDATION_ERROR,
        'Email is required'
      );
      assert.ok(error.message.includes('Email'));
      assert.ok(error.message.includes('required'));
    });

    test('formats timeout error', () => {
      const error = new UngError(
        ErrorType.TIMEOUT,
        'Operation timed out after 30s'
      );
      assert.ok(error.message.includes('timed out'));
    });
  });

  suite('Error Context Tests', () => {
    test('error preserves context', () => {
      const details = {
        operation: 'client.add',
        input: { name: 'Test', email: 'invalid' },
        timestamp: new Date().toISOString(),
      };

      const error = new UngError(
        ErrorType.VALIDATION_ERROR,
        'Invalid email format',
        details
      );

      assert.ok(error.details);
      assert.strictEqual(
        (error.details as { operation: string }).operation,
        'client.add'
      );
    });

    test('error stack trace is available', () => {
      const error = new UngError(ErrorType.UNKNOWN, 'Test error');
      assert.ok(error.stack);
      assert.ok(error.stack.includes('UngError'));
    });
  });
});
