package gui

import (
	_ "embed"
	"image/color"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/theme"
	"github.com/user/go-claude-code/internal/config"
)

//go:embed assets/CascadiaMono.ttf
var monoFont []byte

// Palette: near-black background, ANSI-inspired accents (FR-023). Per 003 the
// theme is a layered "Deep Goblue" dark scheme with a single LetsGO accent.
var (
	// AccentColor is the distinctive LetsGO accent (FR-024): the Go blue is a
	// visual tie-in to the product name. Used for the rail active mark, approval
	// bars, palette selection and mode highlights.
	AccentColor = color.NRGBA{R: 0x00, G: 0xAD, B: 0xD8, A: 0xFF}

	// colorTransparent is a fully transparent fill for rings and overlays.
	colorTransparent = color.NRGBA{R: 0, G: 0, B: 0, A: 0}

	// Deep Goblue levels (003): chrome/abyss, surface, raised, well.
	darkBackground  = color.NRGBA{R: 0x0A, G: 0x0E, B: 0x13, A: 0xFF} // abyss
	darkSurface     = color.NRGBA{R: 0x0F, G: 0x14, B: 0x1B, A: 0xFF} // panels
	darkRaised      = color.NRGBA{R: 0x16, G: 0x1D, B: 0x26, A: 0xFF} // cards
	darkWell        = color.NRGBA{R: 0x1E, G: 0x27, B: 0x33, A: 0xFF} // hover/rail bg
	darkBorder      = color.NRGBA{R: 0x24, G: 0x2D, B: 0x38, A: 0xFF}
	darkForeground  = color.NRGBA{R: 0xE6, G: 0xED, B: 0xF3, A: 0xFF}
	darkSecondary   = color.NRGBA{R: 0x9B, G: 0xA8, B: 0xB7, A: 0xFF}
	darkPlaceholder = color.NRGBA{R: 0x8A, G: 0x96, B: 0xA6, A: 0xFF}
	darkSelection   = color.NRGBA{R: 0x26, G: 0x4F, B: 0x78, A: 0xFF}
	darkDisabled    = color.NRGBA{R: 0x3F, G: 0x4A, B: 0x57, A: 0xFF}
	darkHover       = color.NRGBA{R: 0x1B, G: 0x2A, B: 0x36, A: 0xFF}
	darkFocus       = AccentColor
	darkPrimary     = color.NRGBA{R: 0x4F, G: 0xC1, B: 0xFF, A: 0xFF}
	darkSuccess     = color.NRGBA{R: 0x2E, G: 0xA0, B: 0x43, A: 0xFF}
	darkWarning     = color.NRGBA{R: 0xD2, G: 0x99, B: 0x22, A: 0xFF}
	darkError       = color.NRGBA{R: 0xF8, G: 0x51, B: 0x49, A: 0xFF}

	// Light variant (FR-007): same structure, lighter base.
	lightBackground  = color.NRGBA{R: 0xED, G: 0xF1, B: 0xF5, A: 0xFF}
	lightSurface     = color.NRGBA{R: 0xF5, G: 0xF7, B: 0xFA, A: 0xFF}
	lightRaised      = color.NRGBA{R: 0xFF, G: 0xFF, B: 0xFF, A: 0xFF}
	lightWell        = color.NRGBA{R: 0xE8, G: 0xEC, B: 0xF1, A: 0xFF}
	lightBorder      = color.NRGBA{R: 0xC9, G: 0xD2, B: 0xDC, A: 0xFF}
	lightForeground  = color.NRGBA{R: 0x1C, G: 0x2A, B: 0x3A, A: 0xFF}
	lightSecondary   = color.NRGBA{R: 0x44, G: 0x50, B: 0x60, A: 0xFF}
	lightPlaceholder = color.NRGBA{R: 0x8A, G: 0x94, B: 0xA0, A: 0xFF}
	lightSelection   = color.NRGBA{R: 0xBF, G: 0xDB, B: 0xFF, A: 0xFF}
	lightDisabled    = color.NRGBA{R: 0xB9, G: 0xC2, B: 0xCC, A: 0xFF}
	lightHover       = color.NRGBA{R: 0xE2, G: 0xE9, B: 0xF2, A: 0xFF}
	lightPrimary     = color.NRGBA{R: 0x00, G: 0x68, B: 0xA8, A: 0xFF}
	lightSuccess     = color.NRGBA{R: 0x1F, G: 0x7A, B: 0x37, A: 0xFF}
	lightWarning     = color.NRGBA{R: 0x9A, G: 0x67, B: 0x00, A: 0xFF}
	lightError       = color.NRGBA{R: 0xD1, G: 0x24, B: 0x2F, A: 0xFF}

	// Legacy ANSI accents still used by terminals/chips (kept for 002).
	colorCyan    = color.NRGBA{R: 0x6B, G: 0xE3, B: 0xE3, A: 0xFF}
	colorGreen   = color.NRGBA{R: 0x9E, G: 0xCE, B: 0x6A, A: 0xFF}
	colorYellow  = color.NRGBA{R: 0xE5, G: 0xC0, B: 0x7B, A: 0xFF}
	colorMagenta = color.NRGBA{R: 0xC6, G: 0x8E, B: 0xE6, A: 0xFF}
	colorOrange  = color.NRGBA{R: 0xF5, G: 0xA6, B: 0x5C, A: 0xFF}
)

// accentThemeColor is the private theme color name resolving to AccentColor
// (FR-024). It is used by widgets that need the LetsGO accent as a label
// color while keeping the theme the single source of color truth.
const accentThemeColor fyne.ThemeColorName = "letsgo-accent"

// Typographic ramp (US3, T026): captions, body, large, title, heading and
// monospace sizes as theme tokens; headers and titles use these constants,
// never literal sizes.
const (
	TextSizeCaption float32 = 12
	TextSizeBody    float32 = 14
	TextSizeLarge   float32 = 16
	TextSizeTitle   float32 = 20
	TextSizeHeading float32 = 24
	TextSizeMono    float32 = 13
)

// cardBorderColor resolves the 1px card border token for the active variant
// (US3, T025 common card frame).
func cardBorderColor() color.Color {
	t := &terminalTheme{}
	if config.AppConfig.ThemeVariant == "light" {
		t.forcedVariant = theme.VariantLight
	}
	return t.Color(theme.ColorNameInputBorder, t.variant())
}

// Shell theme tokens (003): three background levels + rail surface.
const (
	ThemeColorNameRail    fyne.ThemeColorName = "letsgo-rail"
	ThemeColorNameSurface fyne.ThemeColorName = "letsgo-surface"
	ThemeColorNameRaised  fyne.ThemeColorName = "letsgo-raised"
)

var _ fyne.Theme = (*terminalTheme)(nil)

// terminalTheme renders the Deep Goblue palette. variant overrides the system
// variant when non-zero so config.theme ("dark"/"light") controls FR-007 even
// if the OS reports a different preference.
type terminalTheme struct {
	forcedVariant fyne.ThemeVariant
}

func (t *terminalTheme) variant() fyne.ThemeVariant {
	if t.forcedVariant == 0 {
		return theme.VariantDark
	}
	return t.forcedVariant
}

func (t *terminalTheme) Color(name fyne.ThemeColorName, _ fyne.ThemeVariant) color.Color {
	light := t.variant() == theme.VariantLight
	switch name {
	case accentThemeColor:
		if light {
			return color.NRGBA{R: 0x00, G: 0x87, B: 0xA9, A: 0xFF}
		}
		return AccentColor
	case theme.ColorNameBackground:
		if light {
			return lightBackground
		}
		return darkBackground
	case ThemeColorNameRail, theme.ColorNameHeaderBackground:
		if light {
			return lightWell
		}
		return darkWell
	case theme.ColorNameButton, theme.ColorNameMenuBackground, theme.ColorNameInputBackground:
		if light {
			return lightSurface
		}
		return darkSurface
	case ThemeColorNameSurface:
		if light {
			return lightSurface
		}
		return darkSurface
	case theme.ColorNameOverlayBackground, ThemeColorNameRaised:
		if light {
			return lightRaised
		}
		return darkRaised
	case theme.ColorNameDisabledButton:
		if light {
			return lightWell
		}
		return color.NRGBA{R: 0x14, G: 0x1A, B: 0x22, A: 0xFF}
	case theme.ColorNamePrimary:
		if light {
			return lightPrimary
		}
		return darkPrimary
	case theme.ColorNameForeground, theme.ColorNameForegroundOnPrimary:
		if light {
			return lightForeground
		}
		return darkForeground
	case theme.ColorNameDisabled:
		if light {
			return lightDisabled
		}
		return darkDisabled
	case theme.ColorNamePlaceHolder:
		if light {
			return lightPlaceholder
		}
		return darkPlaceholder
	case theme.ColorNameSeparator, theme.ColorNameShadow, theme.ColorNameInputBorder:
		if light {
			return lightBorder
		}
		return darkBorder
	case theme.ColorNameSelection:
		if light {
			return lightSelection
		}
		return darkSelection
	case theme.ColorNameFocus, theme.ColorNamePressed:
		if light {
			return lightPrimary
		}
		return darkFocus
	case theme.ColorNameHover:
		if light {
			return lightHover
		}
		return darkHover
	case theme.ColorNameError:
		if light {
			return lightError
		}
		return darkError
	case theme.ColorNameSuccess:
		if light {
			return lightSuccess
		}
		return darkSuccess
	case theme.ColorNameWarning:
		if light {
			return lightWarning
		}
		return darkWarning
	default:
		return theme.DefaultTheme().Color(name, t.variant())
	}
}

// Font returns the UI sans face for regular text and the embedded Cascadia
// Mono for monospace content (status line, code, tool payloads, composer).
func (t *terminalTheme) Font(style fyne.TextStyle) fyne.Resource {
	if style.Monospace {
		return fyne.NewStaticResource("CascadiaMono.ttf", monoFont)
	}
	return theme.DefaultTheme().Font(style)
}

func (t *terminalTheme) Icon(name fyne.ThemeIconName) fyne.Resource {
	return theme.DefaultTheme().Icon(name)
}

func (t *terminalTheme) Size(name fyne.ThemeSizeName) float32 {
	// Typographic ramp (US3, T026): map Fyne's named sizes to the LetsGO
	// scale so headers/titles use tokens instead of literal sizes.
	switch name {
	case theme.SizeNameCaptionText:
		return TextSizeCaption
	case theme.SizeNameHeadingText:
		return TextSizeHeading
	case theme.SizeNameSubHeadingText:
		return TextSizeLarge
	case theme.SizeNameText:
		return TextSizeBody
	}
	return theme.DefaultTheme().Size(name)
}
