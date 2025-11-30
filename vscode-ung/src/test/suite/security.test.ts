import * as assert from 'node:assert';

/**
 * Security Tests for UNG VSCode Extension
 *
 * Tests encryption detection, password validation, and secure storage operations.
 * Note: These tests use mocks since VSCode's SecretStorage requires a real extension context.
 */

// Mock SecureStorage for testing without VSCode context
class MockSecureStorage {
  private storage: Map<string, string> = new Map();

  async savePassword(password: string): Promise<void> {
    if (!password || password.length === 0) {
      throw new Error('Password cannot be empty');
    }
    this.storage.set('ung-database-password', password);
  }

  async getPassword(): Promise<string | undefined> {
    return this.storage.get('ung-database-password');
  }

  async deletePassword(): Promise<void> {
    this.storage.delete('ung-database-password');
  }

  async hasPassword(): Promise<boolean> {
    const password = await this.getPassword();
    return password !== undefined && password.length > 0;
  }

  clear(): void {
    this.storage.clear();
  }
}

// Encryption detection utilities (matching CLI logic)
const EncryptionUtils = {
  /**
   * Check if data looks like it could be encrypted
   * Encrypted files have: [32-byte salt][12-byte nonce][ciphertext + tag]
   */
  isEncryptedData(data: Uint8Array): boolean {
    const SALT_SIZE = 32;
    const NONCE_SIZE = 12;
    const MIN_SIZE = SALT_SIZE + NONCE_SIZE + 16; // 16 = GCM tag size

    if (data.length < MIN_SIZE) {
      return false;
    }

    // Check for known file format magic bytes (not encrypted)
    // SQLite: "SQLite format 3\0"
    const sqliteMagic = [
      0x53, 0x51, 0x4c, 0x69, 0x74, 0x65, 0x20, 0x66, 0x6f, 0x72, 0x6d, 0x61,
      0x74, 0x20, 0x33, 0x00,
    ];
    let isSqlite = true;
    for (let i = 0; i < sqliteMagic.length && i < data.length; i++) {
      if (data[i] !== sqliteMagic[i]) {
        isSqlite = false;
        break;
      }
    }
    if (isSqlite) {
      return false;
    }

    // Check if first SALT_SIZE bytes look like random data
    const salt = data.slice(0, SALT_SIZE);
    let allZeros = true;
    let allPrintable = true;

    for (const byte of salt) {
      if (byte !== 0) {
        allZeros = false;
      }
      if (byte < 32 || byte > 126) {
        allPrintable = false;
      }
    }

    // If all zeros or all printable ASCII, probably not encrypted
    return !allZeros && !allPrintable;
  },

  /**
   * Validate password meets security requirements
   */
  validatePassword(password: string): { valid: boolean; errors: string[] } {
    const errors: string[] = [];

    if (!password) {
      errors.push('Password is required');
      return { valid: false, errors };
    }

    if (password.length < 8) {
      errors.push('Password must be at least 8 characters');
    }

    if (password.length > 128) {
      errors.push('Password must be at most 128 characters');
    }

    // Check for null bytes (security issue)
    if (password.includes('\0')) {
      errors.push('Password cannot contain null bytes');
    }

    return { valid: errors.length === 0, errors };
  },
};

suite('Security Test Suite', () => {
  suite('SecureStorage Tests', () => {
    let storage: MockSecureStorage;

    setup(() => {
      storage = new MockSecureStorage();
    });

    teardown(() => {
      storage.clear();
    });

    test('savePassword stores password', async () => {
      await storage.savePassword('TestPassword123!');
      const retrieved = await storage.getPassword();
      assert.strictEqual(retrieved, 'TestPassword123!');
    });

    test('hasPassword returns false when no password saved', async () => {
      const has = await storage.hasPassword();
      assert.strictEqual(has, false);
    });

    test('hasPassword returns true after saving password', async () => {
      await storage.savePassword('SecurePass!');
      const has = await storage.hasPassword();
      assert.strictEqual(has, true);
    });

    test('deletePassword removes saved password', async () => {
      await storage.savePassword('ToBeDeleted');
      await storage.deletePassword();
      const has = await storage.hasPassword();
      assert.strictEqual(has, false);
    });

    test('getPassword returns undefined when not saved', async () => {
      const password = await storage.getPassword();
      assert.strictEqual(password, undefined);
    });

    test('savePassword rejects empty password', async () => {
      await assert.rejects(async () => {
        await storage.savePassword('');
      }, /Password cannot be empty/);
    });

    test('password can be updated', async () => {
      await storage.savePassword('FirstPassword');
      await storage.savePassword('SecondPassword');
      const retrieved = await storage.getPassword();
      assert.strictEqual(retrieved, 'SecondPassword');
    });

    test('password with special characters', async () => {
      const specialPassword = 'P@$$w0rd!#$%^&*()_+-=[]{}|;:,.<>?';
      await storage.savePassword(specialPassword);
      const retrieved = await storage.getPassword();
      assert.strictEqual(retrieved, specialPassword);
    });

    test('password with unicode characters', async () => {
      const unicodePassword = 'пароль密码パスワード🔐';
      await storage.savePassword(unicodePassword);
      const retrieved = await storage.getPassword();
      assert.strictEqual(retrieved, unicodePassword);
    });
  });

  suite('Encryption Detection Tests', () => {
    test('detects encrypted data (random bytes)', () => {
      // Simulate encrypted data with random-looking salt
      const encryptedData = new Uint8Array(100);
      // Fill with non-zero, non-printable bytes
      for (let i = 0; i < encryptedData.length; i++) {
        encryptedData[i] = (i * 7 + 128) % 256;
      }
      assert.strictEqual(EncryptionUtils.isEncryptedData(encryptedData), true);
    });

    test('rejects plain text as not encrypted', () => {
      // Plain text file (all printable ASCII)
      const plainText = new TextEncoder().encode(
        'This is a plain text file with normal content.'
      );
      // Need to pad to minimum size
      const paddedPlain = new Uint8Array(100);
      paddedPlain.set(plainText);
      for (let i = plainText.length; i < 100; i++) {
        paddedPlain[i] = 32; // spaces
      }
      assert.strictEqual(EncryptionUtils.isEncryptedData(paddedPlain), false);
    });

    test('rejects file with all zeros', () => {
      const zeros = new Uint8Array(100);
      assert.strictEqual(EncryptionUtils.isEncryptedData(zeros), false);
    });

    test('rejects file too short for encryption', () => {
      const shortData = new Uint8Array(50); // Less than salt + nonce + tag
      for (let i = 0; i < shortData.length; i++) {
        shortData[i] = (i * 13) % 256;
      }
      assert.strictEqual(EncryptionUtils.isEncryptedData(shortData), false);
    });

    test('minimum valid encrypted size', () => {
      // Minimum size: 32 (salt) + 12 (nonce) + 16 (tag) = 60 bytes
      const minSize = new Uint8Array(60);
      for (let i = 0; i < minSize.length; i++) {
        minSize[i] = (i * 11 + 200) % 256;
      }
      assert.strictEqual(EncryptionUtils.isEncryptedData(minSize), true);
    });

    test('SQLite file header not detected as encrypted', () => {
      // SQLite files start with "SQLite format 3\0"
      const sqliteHeader = new Uint8Array(100);
      const header = 'SQLite format 3\0';
      for (let i = 0; i < header.length; i++) {
        sqliteHeader[i] = header.charCodeAt(i);
      }
      // Fill rest with printable chars
      for (let i = header.length; i < 100; i++) {
        sqliteHeader[i] = 65 + (i % 26); // A-Z
      }
      assert.strictEqual(EncryptionUtils.isEncryptedData(sqliteHeader), false);
    });
  });

  suite('Password Validation Tests', () => {
    test('accepts valid password', () => {
      const result = EncryptionUtils.validatePassword('SecurePassword123!');
      assert.strictEqual(result.valid, true);
      assert.strictEqual(result.errors.length, 0);
    });

    test('rejects empty password', () => {
      const result = EncryptionUtils.validatePassword('');
      assert.strictEqual(result.valid, false);
      assert.ok(result.errors.includes('Password is required'));
    });

    test('rejects short password', () => {
      const result = EncryptionUtils.validatePassword('short');
      assert.strictEqual(result.valid, false);
      assert.ok(
        result.errors.includes('Password must be at least 8 characters')
      );
    });

    test('rejects password with null bytes', () => {
      const result = EncryptionUtils.validatePassword('pass\0word');
      assert.strictEqual(result.valid, false);
      assert.ok(result.errors.includes('Password cannot contain null bytes'));
    });

    test('accepts password with exactly 8 characters', () => {
      const result = EncryptionUtils.validatePassword('12345678');
      assert.strictEqual(result.valid, true);
    });

    test('rejects password longer than 128 characters', () => {
      const longPassword = 'a'.repeat(129);
      const result = EncryptionUtils.validatePassword(longPassword);
      assert.strictEqual(result.valid, false);
      assert.ok(
        result.errors.includes('Password must be at most 128 characters')
      );
    });

    test('accepts password with exactly 128 characters', () => {
      const maxPassword = 'a'.repeat(128);
      const result = EncryptionUtils.validatePassword(maxPassword);
      assert.strictEqual(result.valid, true);
    });

    test('accepts password with special characters', () => {
      const result = EncryptionUtils.validatePassword('P@$$w0rd!#$%^&*()');
      assert.strictEqual(result.valid, true);
    });

    test('accepts password with unicode', () => {
      const result = EncryptionUtils.validatePassword('密码пароль🔐secure');
      assert.strictEqual(result.valid, true);
    });
  });

  suite('Security Edge Cases', () => {
    test('password storage is isolated per key', async () => {
      const storage1 = new MockSecureStorage();
      const storage2 = new MockSecureStorage();

      await storage1.savePassword('password1');
      await storage2.savePassword('password2');

      assert.strictEqual(await storage1.getPassword(), 'password1');
      assert.strictEqual(await storage2.getPassword(), 'password2');
    });

    test('clearing storage removes all data', async () => {
      const storage = new MockSecureStorage();
      await storage.savePassword('testpass');
      storage.clear();
      assert.strictEqual(await storage.hasPassword(), false);
    });

    test('concurrent password operations', async () => {
      const storage = new MockSecureStorage();
      const operations = [];

      for (let i = 0; i < 10; i++) {
        operations.push(storage.savePassword(`password${i}`));
      }

      await Promise.all(operations);
      const hasPassword = await storage.hasPassword();
      assert.strictEqual(hasPassword, true);
    });
  });
});
