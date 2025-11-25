# Media Assets

## Icons

### Extension Icon (icon.png)
The extension requires a 128x128 PNG icon. Convert the provided icon.svg to PNG:

```bash
# Using ImageMagick
convert -background none -size 128x128 icon.svg icon.png

# Or using Inkscape
inkscape --export-type=png --export-filename=icon.png --export-width=128 --export-height=128 icon.svg

# Or use an online converter
# Upload icon.svg to: https://cloudconvert.com/svg-to-png
```

### Icon Design
The icon features:
- Green-to-blue gradient background
- White invoice/document symbol
- Dollar sign for billing
- Small clock icon for time tracking
- Modern, professional appearance

## Tree View Icons

VSCode built-in icons are used throughout the extension via ThemeIcon:
- Invoices: `check`, `clock`, `warning` (status-based)
- Contracts: `file-code`
- Clients: `person`
- Expenses: `credit-card`
- Tracking: `clock`
- Actions: `refresh`, `add`, `play`, `debug-stop`

No custom icon files are needed for tree views.
