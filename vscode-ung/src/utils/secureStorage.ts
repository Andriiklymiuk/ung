import * as vscode from 'vscode';

const PASSWORD_KEY = 'ung-database-password';

/**
 * Secure storage utility for managing database passwords using VSCode's SecretStorage API.
 *
 * VSCode's SecretStorage uses the operating system's secure credential storage:
 * - macOS: Keychain
 * - Windows: Credential Manager
 * - Linux: Secret Service (libsecret)
 */
export class SecureStorage {
  private secretStorage: vscode.SecretStorage;

  constructor(context: vscode.ExtensionContext) {
    this.secretStorage = context.secrets;
  }

  /**
   * Save the database password to secure storage
   */
  async savePassword(password: string): Promise<void> {
    await this.secretStorage.store(PASSWORD_KEY, password);
  }

  /**
   * Get the database password from secure storage
   * @returns The password if saved, undefined otherwise
   */
  async getPassword(): Promise<string | undefined> {
    return this.secretStorage.get(PASSWORD_KEY);
  }

  /**
   * Delete the database password from secure storage
   */
  async deletePassword(): Promise<void> {
    await this.secretStorage.delete(PASSWORD_KEY);
  }

  /**
   * Check if a password is saved in secure storage
   */
  async hasPassword(): Promise<boolean> {
    const password = await this.getPassword();
    return password !== undefined && password.length > 0;
  }

  /**
   * Prompt the user to enter a password and optionally save it
   * @param options Options for the password prompt
   * @returns The entered password, or undefined if cancelled
   */
  async promptForPassword(options?: {
    title?: string;
    prompt?: string;
    offerToSave?: boolean;
  }): Promise<string | undefined> {
    const {
      title = 'Database Password',
      prompt = 'Enter database password',
      offerToSave = true,
    } = options ?? {};

    const password = await vscode.window.showInputBox({
      prompt,
      password: true,
      title,
      ignoreFocusOut: true,
    });

    if (password && offerToSave) {
      const save = await vscode.window.showQuickPick(
        [
          {
            label: '$(check) Save password',
            value: true,
            description: 'Remember for this VS Code session',
          },
          {
            label: "$(x) Don't save",
            value: false,
            description: 'Ask again next time',
          },
        ],
        {
          placeHolder: 'Would you like to save the password for future use?',
          title: 'Save Password',
        }
      );

      if (save?.value) {
        await this.savePassword(password);
        vscode.window.showInformationMessage('Password saved securely');
      }
    }

    return password;
  }

  /**
   * Get the password, either from storage or by prompting the user
   * @param options Options for getting the password
   * @returns The password, or undefined if cancelled
   */
  async getOrPromptPassword(options?: {
    title?: string;
    prompt?: string;
    offerToSave?: boolean;
  }): Promise<string | undefined> {
    // First try to get from secure storage
    const savedPassword = await this.getPassword();
    if (savedPassword) {
      return savedPassword;
    }

    // If not saved, prompt the user
    return this.promptForPassword(options);
  }
}

/**
 * Singleton instance of SecureStorage
 * Must be initialized with initSecureStorage() before use
 */
let secureStorageInstance: SecureStorage | undefined;

/**
 * Initialize the secure storage with the extension context
 * Call this once during extension activation
 */
export function initSecureStorage(
  context: vscode.ExtensionContext
): SecureStorage {
  secureStorageInstance = new SecureStorage(context);
  return secureStorageInstance;
}

/**
 * Get the secure storage instance
 * @throws Error if not initialized
 */
export function getSecureStorage(): SecureStorage {
  if (!secureStorageInstance) {
    throw new Error(
      'SecureStorage not initialized. Call initSecureStorage() first.'
    );
  }
  return secureStorageInstance;
}
