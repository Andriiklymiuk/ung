# VSCode Extension Publishing Guide

Quick reference for publishing the UNG VSCode extension to the marketplace.

## 🚀 One-Time Setup

### 1. Create Azure DevOps Account
- Go to: https://dev.azure.com
- Sign in with Microsoft/GitHub account

### 2. Get Personal Access Token (PAT)
1. Click User Settings → Personal Access Tokens
2. New Token
3. Settings:
   - **Name**: `VSCode Marketplace - UNG`
   - **Organization**: All accessible organizations
   - **Scopes**: Select `Marketplace` → Check `Manage`
4. Create and **SAVE THE TOKEN** (shown only once!)

### 3. Create Publisher
1. Go to: https://marketplace.visualstudio.com/manage
2. **Create Publisher**
3. Settings:
   - **ID**: `andriiklymiuk` (must match package.json)
   - **Display Name**: Your name or company
   - **Upload logo** (optional but recommended)

### 4. Install Publishing Tool

```bash
npm install -g @vscode/vsce
```

## 📦 Publishing Process

### Manual Publishing

```bash
cd vscode-ung

# 1. Install dependencies
npm install

# 2. Build TypeScript
npm run compile

# 3. Test the extension
# Press F5 in VSCode to launch Extension Development Host

# 4. Update version in package.json (if needed)
# Example: "version": "1.0.1"

# 5. Login (first time only)
vsce login andriiklymiuk
# Enter your PAT when prompted

# 6. Package (optional - creates .vsix file for testing)
vsce package

# 7. Publish
vsce publish

# Or publish with automatic version bump:
vsce publish patch  # 1.0.0 → 1.0.1
vsce publish minor  # 1.0.0 → 1.1.0
vsce publish major  # 1.0.0 → 2.0.0
```

### Automated Publishing with GitHub Actions

Already configured in `.github/workflows/publish-vscode.yml`

**To publish:**

```bash
# Update version in package.json
# Commit changes

# Create and push version tag
git tag vscode-v1.0.1
git push origin vscode-v1.0.1

# GitHub Actions will automatically publish
```

**Required GitHub Secret:**
- Name: `GH_PAT`
- Value: Your Azure DevOps PAT

Add it: GitHub repo → Settings → Secrets → Actions → New repository secret

## 🔍 Pre-Publishing Checklist

Before every release:

- [ ] Update `version` in `package.json`
- [ ] Update `CHANGELOG.md` with changes
- [ ] Test extension locally (F5)
- [ ] Run `npm run compile` without errors
- [ ] Update README.md if features changed
- [ ] Check icon.png exists in media/
- [ ] Review package.json metadata:
  - [ ] `displayName` is user-friendly
  - [ ] `description` is clear and concise
  - [ ] `keywords` are relevant
  - [ ] `repository` URL is correct
  - [ ] `bugs` URL is correct
  - [ ] `license` is specified

## 📋 Version Numbering

Follow [Semantic Versioning](https://semver.org/):

- **MAJOR** (1.0.0 → 2.0.0): Breaking changes
- **MINOR** (1.0.0 → 1.1.0): New features, backward compatible
- **PATCH** (1.0.0 → 1.0.1): Bug fixes, backward compatible

## 🎯 Marketplace Listing

### Current Metadata

```json
{
  "name": "ung",
  "displayName": "UNG - Billing & Time Tracking",
  "description": "Comprehensive billing, invoicing, and time tracking extension for the UNG CLI tool",
  "version": "1.0.0",
  "publisher": "andriiklymiuk",
  "categories": ["Other", "Finance", "Productivity"],
  "keywords": ["billing", "invoicing", "time tracking", "freelance", "accounting"]
}
```

### Improve Discoverability

Add to README.md:
- Screenshots/GIFs of features
- Feature highlights with icons
- Quick start guide
- Video demo (optional)

## 🔧 Troubleshooting

### "Publisher not found"
```bash
# Make sure publisher ID in package.json matches
# your marketplace publisher ID
vsce login andriiklymiuk
```

### "Extension validation failed"
```bash
# Check compilation errors
npm run compile

# Validate package
vsce package
```

### "PAT token expired"
```bash
# Generate new PAT from Azure DevOps
# Login again
vsce login andriiklymiuk
```

### "Missing dependencies"
```bash
# Install all dependencies
npm install

# Check package.json for missing entries
```

## 📊 Post-Publishing

After publishing:

1. **Check Marketplace Listing**
   - Go to: https://marketplace.visualstudio.com/items?itemName=andriiklymiuk.ung
   - Verify description, screenshots, and version

2. **Test Installation**
   ```bash
   # In VSCode
   # Extensions → Search "UNG"
   # Install and test
   ```

3. **Monitor Stats**
   - Go to: https://marketplace.visualstudio.com/manage/publishers/andriiklymiuk
   - View installs, ratings, reviews

4. **Announce Release**
   - GitHub Release notes
   - Social media (if applicable)
   - Documentation updates

## 📚 Resources

- [VSCode Publishing Guide](https://code.visualstudio.com/api/working-with-extensions/publishing-extension)
- [Extension Manifest Reference](https://code.visualstudio.com/api/references/extension-manifest)
- [Marketplace Publisher Portal](https://marketplace.visualstudio.com/manage)
- [Azure DevOps](https://dev.azure.com)

## 🔐 Security Notes

- **Never commit PAT tokens** to git
- Store PAT in password manager
- Use GitHub Secrets for CI/CD
- Regenerate PAT if compromised
- Set PAT expiration (max 1 year)

---

**Quick Commands Reference:**

```bash
# Compile
npm run compile

# Package
vsce package

# Publish patch
vsce publish patch

# Publish with specific version
vsce publish 1.2.3

# Login
vsce login andriiklymiuk

# Show info
vsce show andriiklymiuk.ung
```
