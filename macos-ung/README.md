# UNG macOS App

Native macOS application for UNG freelance management.

## Development

### Build & Run

```bash
# Open in Xcode
open ung.xcodeproj

# Or build from command line (Debug)
xcodebuild -project ung.xcodeproj -scheme ung -configuration Debug build
```

## Release Process

### Prerequisites

1. **Apple Developer Account**: You need an active Apple Developer Program membership
2. **Xcode**: Latest version installed
3. **App Store Connect API Key**:

   - Log in to [App Store Connect](https://appstoreconnect.apple.com)
   - Go to Users and Access > Keys
   - Create a new API key with App Manager or Admin role
   - Download the `.p8` key file
   - Note your Key ID and Issuer ID

4. **Team ID**: Find your Team ID in:

   - Xcode: Preferences > Accounts > select your Apple ID > View Details
   - Or in [Apple Developer Portal](https://developer.apple.com/account) > Membership

5. **Update ExportOptions.plist**:
   - Replace `YOUR_TEAM_ID` with your actual Team ID
   - Update provisioning profile name if different

### Building for Release

#### Using Makefile (Recommended)

```bash
# From project root directory

# Build in Release configuration
make macos-build

# Create archive for distribution
make macos-archive

# Export for App Store (requires ExportOptions.plist)
make macos-export

# Upload to App Store Connect (requires API credentials)
# Set environment variables first:
export API_KEY_ID="YOUR_KEY_ID"
export API_ISSUER_ID="YOUR_ISSUER_ID"
export APP_STORE_CONNECT_API_KEY_PATH="/path/to/AuthKey_XXXXX.p8"

make macos-upload

# Or run complete release process (archive + export + upload)
make macos-release
```

#### Manual Process

1. **Archive the App**:

   ```bash
   xcodebuild -project ung.xcodeproj \
     -scheme ung \
     -configuration Release \
     -archivePath dist/macos/UNG.xcarchive \
     archive
   ```

2. **Export for App Store**:

   ```bash
   xcodebuild -exportArchive \
     -archivePath dist/macos/UNG.xcarchive \
     -exportOptionsPlist ExportOptions.plist \
     -exportPath dist/macos/export
   ```

3. **Upload to App Store Connect**:

   **Option A: Using altool (deprecated but still works)**

   ```bash
   xcrun altool --upload-app \
     -f dist/macos/export/UNG.pkg \
     -t macos \
     --apiKey YOUR_KEY_ID \
     --apiIssuer YOUR_ISSUER_ID
   ```

   **Option B: Using notarytool (recommended)**

   ```bash
   # First, validate the app
   xcrun notarytool submit dist/macos/export/UNG.pkg \
     --key /path/to/AuthKey_XXXXX.p8 \
     --key-id YOUR_KEY_ID \
     --issuer YOUR_ISSUER_ID \
     --wait

   # Then upload using Transporter or altool
   ```

   **Option C: Using Xcode's Organizer**

   - Open Xcode
   - Window > Organizer
   - Select your archive
   - Click "Distribute App"
   - Follow the wizard

### Distribution Methods

#### 1. App Store Distribution

- Most users will discover your app here
- Automatic updates via App Store
- Requires Apple's review process (7-14 days)
- 30% revenue share for paid apps (first year)

**Steps**:

1. Archive and export as shown above
2. Upload to App Store Connect
3. Fill in app metadata (screenshots, description, etc.)
4. Submit for review
5. Once approved, release to App Store

#### 2. Direct Distribution (Notarized DMG)

For distributing outside the App Store:

```bash
# Build release version
xcodebuild -project ung.xcodeproj \
  -scheme ung \
  -configuration Release \
  -derivedDataPath build

# Create DMG
hdiutil create -volname "UNG" \
  -srcfolder build/Build/Products/Release/UNG.app \
  -ov -format UDZO \
  UNG.dmg

# Notarize the DMG
xcrun notarytool submit UNG.dmg \
  --key /path/to/AuthKey_XXXXX.p8 \
  --key-id YOUR_KEY_ID \
  --issuer YOUR_ISSUER_ID \
  --wait

# Staple the notarization ticket
xcrun stapler staple UNG.dmg
```

#### 3. Homebrew Cask

After creating a notarized DMG:

1. Upload DMG to GitHub releases
2. Calculate SHA256: `shasum -a 256 UNG.dmg`
3. Create a Homebrew cask (see example below)

**Example Cask** (`Casks/ung.rb`):

```ruby
cask "ung" do
  version "1.0.0"
  sha256 "abc123..."

  url "https://github.com/Andriiklymiuk/ung/releases/download/v#{version}/UNG.dmg"
  name "UNG"
  desc "Your Next Gig - Freelance billing and time tracking"
  homepage "https://github.com/Andriiklymiuk/ung"

  app "UNG.app"

  zap trash: [
    "~/Library/Application Support/com.ung.ung",
    "~/.ung",
  ]
end
```

### Version Management

Update version in Xcode:

1. Select project in navigator
2. Select target "ung"
3. General tab > Identity section
4. Update "Version" field
5. Update "Build" number

Or use `agvtool`:

```bash
# Set version
xcrun agvtool new-marketing-version 1.0.0

# Increment build number
xcrun agvtool next-version -all
```

### Troubleshooting

**"Unable to validate your application"**

- Ensure code signing is set up correctly in Xcode
- Check that your certificates are not expired
- Verify Team ID in ExportOptions.plist

**"Invalid provisioning profile"**

- Update provisioning profile name in ExportOptions.plist
- Or use automatic signing by removing `provisioningProfiles` key

**"API key not found"**

- Check environment variables are set correctly
- Verify the path to your .p8 key file
- Ensure Key ID and Issuer ID are correct

**Upload fails with "Invalid bundle"**

- Ensure you're building for Release configuration
- Check that all required entitlements are set
- Verify bundle identifier matches App Store Connect

### Resources

- [App Store Connect](https://appstoreconnect.apple.com)
- [Apple Developer Portal](https://developer.apple.com)
- [Notarizing macOS Software](https://developer.apple.com/documentation/security/notarizing_macos_software_before_distribution)
- [App Store Review Guidelines](https://developer.apple.com/app-store/review/guidelines/)
- [Xcode Build Settings Reference](https://developer.apple.com/documentation/xcode/build-settings-reference)

### Support

For issues with the release process:

1. Check Xcode logs in Organizer
2. Review App Store Connect feedback
3. Check Apple Developer forums
4. Open an issue on GitHub
