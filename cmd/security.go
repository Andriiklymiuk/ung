package cmd

import (
	"fmt"
	"os"

	"github.com/Andriiklymiuk/ung/internal/config"
	"github.com/Andriiklymiuk/ung/internal/db"
	"github.com/spf13/cobra"
)

var securityCmd = &cobra.Command{
	Use:     "security",
	Aliases: []string{"sec", "encrypt"},
	Short:   "Manage database security and encryption",
	Long: `Manage database security settings including encryption.

Examples:
  ung security status              # Show encryption status
  ung security enable              # Enable database encryption
  ung security disable             # Disable database encryption
  ung security change-password     # Change encryption password
  ung security save-password       # Save password to OS keychain
  ung security forget-password     # Remove password from OS keychain`,
}

var securityStatusCmd = &cobra.Command{
	Use:   "status",
	Short: "Show database encryption status",
	Long:  `Display whether the database is encrypted and security settings.`,
	Run:   runSecurityStatus,
}

var securityEnableCmd = &cobra.Command{
	Use:   "enable",
	Short: "Enable database encryption",
	Long: `Enable encryption for the database. This will encrypt the database file at rest
using AES-256-GCM encryption with PBKDF2 key derivation.

The database will be encrypted with a password that you provide. Save your password
to the OS keychain with 'ung security save-password' for seamless access.

Example:
  ung security enable`,
	Run: runSecurityEnable,
}

var securityDisableCmd = &cobra.Command{
	Use:   "disable",
	Short: "Disable database encryption",
	Long: `Disable encryption for the database. This will decrypt the database and
store it in plain text.

⚠️  WARNING: This will make your financial data readable by anyone with
access to the file. Only do this if you understand the security implications.

Example:
  ung security disable        # Will prompt for confirmation
  ung security disable --yes  # Skip confirmation`,
	Run: runSecurityDisable,
}

var disableYesFlag bool

var securityChangePasswordCmd = &cobra.Command{
	Use:   "change-password",
	Short: "Change database encryption password",
	Long: `Change the password used to encrypt the database.

You'll need to provide both the current password and the new password.`,
	Run: runSecurityChangePassword,
}

var securitySavePasswordCmd = &cobra.Command{
	Use:   "save-password",
	Short: "Save password to OS keychain",
	Long: `Save the database encryption password to your operating system's secure credential storage.

On macOS: Keychain
On Windows: Credential Manager
On Linux: Secret Service (requires libsecret)

Once saved, ung will automatically retrieve the password when needed.`,
	Run: runSecuritySavePassword,
}

var securityForgetPasswordCmd = &cobra.Command{
	Use:     "forget-password",
	Aliases: []string{"clear-password", "remove-password"},
	Short:   "Remove password from OS keychain",
	Long: `Remove the saved database password from your operating system's secure credential storage.

After running this command, you'll need to enter the password when prompted.`,
	Run: runSecurityForgetPassword,
}

func init() {
	securityCmd.AddCommand(securityStatusCmd)
	securityCmd.AddCommand(securityEnableCmd)
	securityCmd.AddCommand(securityDisableCmd)
	securityCmd.AddCommand(securityChangePasswordCmd)
	securityCmd.AddCommand(securitySavePasswordCmd)
	securityCmd.AddCommand(securityForgetPasswordCmd)
	rootCmd.AddCommand(securityCmd)

	// Add --yes flag to disable command
	securityDisableCmd.Flags().BoolVarP(&disableYesFlag, "yes", "y", false, "Skip confirmation prompt")
}

func runSecurityStatus(cmd *cobra.Command, args []string) {
	cfg, _ := config.Load()
	dbPath := db.GetDBPath()
	encryptedPath := dbPath + ".encrypted"

	fmt.Println("🔒 Database Security Status")

	// Check if encrypted file exists
	encryptedExists := fileExists(encryptedPath)
	plainExists := fileExists(dbPath)

	if encryptedExists {
		fmt.Println("Status:    ✅ Encrypted")
		fmt.Printf("File:      %s\n", encryptedPath)
		fmt.Println("Algorithm: AES-256-GCM with PBKDF2")
	} else if plainExists {
		fmt.Println("Status:    ⚠️  Not Encrypted (Plain Text)")
		fmt.Printf("File:      %s\n", dbPath)
		fmt.Println("\n⚠️  Your financial data is stored in plain text.")
		fmt.Println("   Run 'ung security enable' to encrypt it.")
	} else {
		fmt.Println("Status:    Database not created yet")
	}

	// Show keychain status
	fmt.Println("\n🔑 Password Storage:")
	if db.KeychainAvailable() {
		keychainName := db.GetKeychainPlatformName()
		if db.HasPasswordInKeychain() {
			fmt.Printf("   ✅ Password saved in %s\n", keychainName)
		} else {
			fmt.Printf("   ❌ No password saved in %s\n", keychainName)
			fmt.Println("      Run 'ung security save-password' to save it")
		}
	} else {
		fmt.Println("   ⚠️  OS keychain not available on this platform")
	}

	if encryptedExists && !db.HasPasswordInKeychain() {
		fmt.Println("\n💡 Run 'ung security save-password' to save your password")
	}

	if cfg.Security.EncryptDatabase {
		fmt.Println("\nConfig:    Encryption enabled in configuration")
	} else if encryptedExists {
		fmt.Println("\nConfig:    Encryption auto-detected (encrypted file exists)")
	}
}

func runSecurityEnable(cmd *cobra.Command, args []string) {
	cfg, _ := config.Load()
	dbPath := db.GetDBPath()
	encryptedPath := dbPath + ".encrypted"

	fmt.Println("🔒 Enable Database Encryption")

	// Check if already encrypted
	if fileExists(encryptedPath) {
		fmt.Println("✅ Database is already encrypted")
		return
	}

	// Check if database exists
	if !fileExists(dbPath) {
		// No database yet, just enable in config
		cfg.Security.EncryptDatabase = true
		if err := config.Save(cfg, false); err != nil {
			fmt.Printf("❌ Failed to save config: %v\n", err)
			return
		}
		fmt.Println("✅ Encryption enabled")
		fmt.Println("   Database will be encrypted when created")
		return
	}

	// Get password for encryption
	fmt.Print("Enter encryption password: ")
	password, err := readPassword()
	if err != nil {
		fmt.Printf("\n❌ Failed to read password: %v\n", err)
		return
	}

	fmt.Print("Confirm password: ")
	confirmPassword, err := readPassword()
	if err != nil {
		fmt.Printf("\n❌ Failed to read password: %v\n", err)
		return
	}

	if password != confirmPassword {
		fmt.Println("\n❌ Passwords don't match")
		return
	}

	// Encrypt the database
	fmt.Println("\n🔄 Encrypting database...")
	if err := db.EncryptDatabase(dbPath, encryptedPath, password); err != nil {
		fmt.Printf("❌ Failed to encrypt database: %v\n", err)
		return
	}

	// Update config
	cfg.Security.EncryptDatabase = true
	if err := config.Save(cfg, false); err != nil {
		fmt.Printf("❌ Failed to save config: %v\n", err)
		// Remove encrypted file since config update failed
		os.Remove(encryptedPath)
		return
	}

	// Remove plain text database
	if err := os.Remove(dbPath); err != nil {
		fmt.Printf("⚠️  Failed to remove plain text database: %v\n", err)
		fmt.Println("   Please remove it manually for security")
	}

	fmt.Println("✅ Database encrypted successfully")
	fmt.Printf("   Encrypted: %s\n", encryptedPath)

	// Offer to save password to keychain
	if db.KeychainAvailable() {
		fmt.Printf("\n💡 Would you like to save the password to %s? (yes/no): ", db.GetKeychainPlatformName())
		var response string
		fmt.Scanln(&response)
		if response == "yes" {
			if err := db.SavePasswordToKeychain(password); err != nil {
				fmt.Printf("⚠️  Failed to save to keychain: %v\n", err)
			} else {
				fmt.Printf("✅ Password saved to %s\n", db.GetKeychainPlatformName())
				fmt.Println("   Your password will be retrieved automatically from now on")
				return
			}
		}
	}

	fmt.Println("\n💡 To use the database:")
	if db.KeychainAvailable() {
		fmt.Println("   Run 'ung security save-password' to save your password securely")
	} else {
		fmt.Println("   Enter password when prompted")
	}
}

func runSecurityDisable(cmd *cobra.Command, args []string) {
	cfg, _ := config.Load()
	dbPath := db.GetDBPath()
	encryptedPath := dbPath + ".encrypted"

	fmt.Println("🔓 Disable Database Encryption")

	// Check if encrypted
	if !fileExists(encryptedPath) {
		fmt.Println("✅ Database is not encrypted")
		return
	}

	// Confirm action (skip if --yes flag provided)
	if !disableYesFlag {
		fmt.Println("⚠️  WARNING: This will decrypt your database and store it in plain text.")
		fmt.Print("Are you sure? (yes/no): ")
		var response string
		fmt.Scanln(&response)

		if response != "yes" {
			fmt.Println("Cancelled")
			return
		}
	}

	// Get password - try keychain first, then prompt
	password, err := db.GetDatabasePassword()
	if err != nil {
		fmt.Printf("❌ Failed to get password: %v\n", err)
		return
	}

	// Decrypt the database
	fmt.Println("\n🔄 Decrypting database...")
	if err := db.DecryptDatabase(encryptedPath, dbPath, password); err != nil {
		fmt.Printf("❌ Failed to decrypt database: %v\n", err)
		return
	}

	// Update config
	cfg.Security.EncryptDatabase = false
	if err := config.Save(cfg, false); err != nil {
		fmt.Printf("❌ Failed to save config: %v\n", err)
		return
	}

	// Remove encrypted database
	if err := os.Remove(encryptedPath); err != nil {
		fmt.Printf("⚠️  Failed to remove encrypted database: %v\n", err)
	}

	// Clean up keychain password since encryption is disabled
	if db.HasPasswordInKeychain() {
		if err := db.DeletePasswordFromKeychain(); err != nil {
			fmt.Printf("⚠️  Failed to remove password from keychain: %v\n", err)
		}
	}
	db.ClearPasswordCache()

	fmt.Println("✅ Database decrypted successfully")
	fmt.Printf("   Plain text: %s\n", dbPath)
	fmt.Println("\n⚠️  Your database is now stored in plain text")
}

func runSecurityChangePassword(cmd *cobra.Command, args []string) {
	dbPath := db.GetDBPath()
	encryptedPath := dbPath + ".encrypted"

	fmt.Println("🔑 Change Encryption Password")

	// Check if encrypted
	if !fileExists(encryptedPath) {
		fmt.Println("❌ Database is not encrypted")
		fmt.Println("   Run 'ung security enable' first")
		return
	}

	// Get current password
	fmt.Print("Enter current password: ")
	currentPassword, err := readPassword()
	if err != nil {
		fmt.Printf("\n❌ Failed to read password: %v\n", err)
		return
	}

	// Decrypt with current password
	fmt.Println("\n🔄 Verifying current password...")
	tempPath := dbPath + ".temp"
	if err := db.DecryptDatabase(encryptedPath, tempPath, currentPassword); err != nil {
		fmt.Printf("❌ Failed to decrypt: %v\n", err)
		fmt.Println("   (Wrong password?)")
		return
	}

	// Get new password
	fmt.Print("Enter new password: ")
	newPassword, err := readPassword()
	if err != nil {
		fmt.Printf("\n❌ Failed to read password: %v\n", err)
		os.Remove(tempPath)
		return
	}

	fmt.Print("Confirm new password: ")
	confirmPassword, err := readPassword()
	if err != nil {
		fmt.Printf("\n❌ Failed to read password: %v\n", err)
		os.Remove(tempPath)
		return
	}

	if newPassword != confirmPassword {
		fmt.Println("\n❌ Passwords don't match")
		os.Remove(tempPath)
		return
	}

	// Encrypt with new password
	fmt.Println("\n🔄 Re-encrypting with new password...")
	newEncryptedPath := encryptedPath + ".new"
	if err := db.EncryptDatabase(tempPath, newEncryptedPath, newPassword); err != nil {
		fmt.Printf("❌ Failed to encrypt: %v\n", err)
		os.Remove(tempPath)
		return
	}

	// Replace old encrypted file
	if err := os.Remove(encryptedPath); err != nil {
		fmt.Printf("❌ Failed to remove old encrypted file: %v\n", err)
		os.Remove(tempPath)
		os.Remove(newEncryptedPath)
		return
	}

	if err := os.Rename(newEncryptedPath, encryptedPath); err != nil {
		fmt.Printf("❌ Failed to replace encrypted file: %v\n", err)
		os.Remove(tempPath)
		return
	}

	// Clean up
	os.Remove(tempPath)

	fmt.Println("✅ Password changed successfully")

	// Update keychain if password was stored there
	if db.KeychainAvailable() && db.HasPasswordInKeychain() {
		if err := db.SavePasswordToKeychain(newPassword); err != nil {
			fmt.Printf("⚠️  Failed to update keychain: %v\n", err)
			fmt.Println("   Run 'ung security save-password' to save the new password")
		} else {
			fmt.Printf("✅ Updated password in %s\n", db.GetKeychainPlatformName())
		}
	} else if db.KeychainAvailable() {
		fmt.Printf("\n💡 Would you like to save the new password to %s? (yes/no): ", db.GetKeychainPlatformName())
		var response string
		fmt.Scanln(&response)
		if response == "yes" {
			if err := db.SavePasswordToKeychain(newPassword); err != nil {
				fmt.Printf("⚠️  Failed to save to keychain: %v\n", err)
			} else {
				fmt.Printf("✅ Password saved to %s\n", db.GetKeychainPlatformName())
			}
		}
	}
}

func readPassword() (string, error) {
	password, err := db.GetDatabasePassword()
	if err != nil {
		return "", err
	}
	// Clear cache so it prompts again next time
	db.ClearPasswordCache()
	return password, nil
}

func fileExists(path string) bool {
	_, err := os.Stat(path)
	return err == nil
}

func runSecuritySavePassword(cmd *cobra.Command, args []string) {
	fmt.Printf("🔑 Save Password to %s\n\n", db.GetKeychainPlatformName())

	// Check if keychain is available
	if !db.KeychainAvailable() {
		fmt.Println("❌ OS keychain is not available on this platform")
		fmt.Println("   You'll need to enter your password when prompted")
		return
	}

	// Check if database is encrypted
	dbPath := db.GetDBPath()
	encryptedPath := dbPath + ".encrypted"
	if !fileExists(encryptedPath) {
		fmt.Println("⚠️  Database is not encrypted")
		fmt.Println("   Enable encryption first with 'ung security enable'")
		return
	}

	// Check if password is already saved
	if db.HasPasswordInKeychain() {
		fmt.Println("A password is already saved in the keychain.")
		fmt.Print("Do you want to replace it? (yes/no): ")
		var response string
		fmt.Scanln(&response)
		if response != "yes" {
			fmt.Println("Cancelled")
			return
		}
	}

	// Get the password
	fmt.Print("Enter database password to save: ")
	password, err := readPassword()
	if err != nil {
		fmt.Printf("\n❌ Failed to read password: %v\n", err)
		return
	}

	// Verify the password by trying to decrypt
	fmt.Println("\n🔄 Verifying password...")
	tempPath := dbPath + ".verify"
	if err := db.DecryptDatabase(encryptedPath, tempPath, password); err != nil {
		fmt.Printf("❌ Invalid password: %v\n", err)
		return
	}
	os.Remove(tempPath)

	// Save to keychain
	if err := db.SavePasswordToKeychain(password); err != nil {
		fmt.Printf("❌ Failed to save password: %v\n", err)
		return
	}

	fmt.Printf("✅ Password saved to %s\n", db.GetKeychainPlatformName())
	fmt.Println("   Your password will be retrieved automatically from now on")
}

func runSecurityForgetPassword(cmd *cobra.Command, args []string) {
	fmt.Printf("🔑 Remove Password from %s\n\n", db.GetKeychainPlatformName())

	// Check if keychain is available
	if !db.KeychainAvailable() {
		fmt.Println("❌ OS keychain is not available on this platform")
		return
	}

	// Check if password is saved
	if !db.HasPasswordInKeychain() {
		fmt.Println("No password is saved in the keychain")
		return
	}

	// Confirm action
	fmt.Print("Are you sure you want to remove the saved password? (yes/no): ")
	var response string
	fmt.Scanln(&response)
	if response != "yes" {
		fmt.Println("Cancelled")
		return
	}

	// Delete from keychain
	if err := db.DeletePasswordFromKeychain(); err != nil {
		fmt.Printf("❌ Failed to remove password: %v\n", err)
		return
	}

	fmt.Printf("✅ Password removed from %s\n", db.GetKeychainPlatformName())
	fmt.Println("   You'll need to enter your password when prompted")
}
