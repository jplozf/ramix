package main

// ----------------------------------------------------------------------------
// IMPORTS
// ----------------------------------------------------------------------------
import (
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/theme"
)

// ----------------------------------------------------------------------------
// TYPES
// ----------------------------------------------------------------------------
type compactTheme struct {
	fyne.Theme
}

// ----------------------------------------------------------------------------
// Size()
// ----------------------------------------------------------------------------
func (m compactTheme) Size(name fyne.ThemeSizeName) float32 {
	if name == theme.SizeNamePadding {
		return 1 // Reduces space between elements (often default 4)
	}
	if name == theme.SizeNameText {
		return 11 // Slightly reduces font size (default 14)
	}
	if name == theme.SizeNameInlineIcon {
		return 12 // Reduces size of icons in menus/buttons
	}
	if name == theme.SizeNameScrollBar {
		return 8 // Reduces size of scrollbars
	}
	if name == theme.SizeNameHeadingText {
		return 14 // Reduces size of headings (default 18)
	}
	if name == theme.SizeNameInputBorder {
		return 1 // Reduces thickness of input borders
	}
	if name == theme.SizeNameSeparatorThickness {
		return 1 // Reduces thickness of separators
	}
	return m.Theme.Size(name)
}
