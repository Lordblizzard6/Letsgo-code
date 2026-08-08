package gui

import (
	"image/color"
	"math"
	"testing"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/theme"
)

// contrastRatio computes WCAG 2.1 contrast between two colors.
func contrastRatio(fg, bg color.Color) float64 {
	l1 := relativeLuminance(fg)
	l2 := relativeLuminance(bg)
	if l2 > l1 {
		l1, l2 = l2, l1
	}
	return (l1 + 0.05) / (l2 + 0.05)
}

func relativeLuminance(c color.Color) float64 {
	r, g, b, _ := c.RGBA()
	return 0.2126*lin(r) + 0.7152*lin(g) + 0.0722*lin(b)
}

func lin(v uint32) float64 {
	f := float64(v>>8) / 255.0
	if f <= 0.04045 {
		return f / 12.92
	}
	return math.Pow((f+0.055)/1.055, 2.4)
}

func TestThemeTokensThreeLevelsDark(t *testing.T) {
	th := &terminalTheme{}
	bg := th.Color(theme.ColorNameBackground, theme.VariantDark)
	rail := th.Color(ThemeColorNameRail, theme.VariantDark)
	surface := th.Color(ThemeColorNameSurface, theme.VariantDark)
	raised := th.Color(ThemeColorNameRaised, theme.VariantDark)

	if sameColor(bg, rail) || sameColor(bg, surface) || sameColor(rail, surface) {
		t.Error("dark background/rail/surface must be distinct levels")
	}
	if sameColor(surface, raised) {
		t.Error("dark surface and raised must be distinct")
	}
}

func TestThemeTokensThreeLevelsLight(t *testing.T) {
	th := &terminalTheme{}
	bg := th.Color(theme.ColorNameBackground, theme.VariantLight)
	rail := th.Color(ThemeColorNameRail, theme.VariantLight)
	surface := th.Color(ThemeColorNameSurface, theme.VariantLight)
	raised := th.Color(ThemeColorNameRaised, theme.VariantLight)

	if sameColor(bg, rail) || sameColor(bg, surface) {
		t.Error("light background/rail/surface must be distinct")
	}
	if sameColor(surface, raised) {
		t.Error("light surface and raised must be distinct")
	}
}

func TestThemeAccentDistinctFromPrimary(t *testing.T) {
	th := &terminalTheme{}
	accent := th.Color(accentThemeColor, theme.VariantDark)
	primary := th.Color(theme.ColorNamePrimary, theme.VariantDark)
	if sameColor(accent, primary) {
		t.Error("accent (0,173,216) must differ from primary (79,193,255)")
	}
}

func TestThemeFontSplit(t *testing.T) {
	th := &terminalTheme{}
	mono := th.Font(fyne.TextStyle{Monospace: true})
	sans := th.Font(fyne.TextStyle{})
	if mono == nil || sans == nil {
		t.Fatal("both fonts must resolve")
	}
	if mono.Name() == sans.Name() {
		t.Errorf("monospace %q should differ from sans %q", mono.Name(), sans.Name())
	}
}

// TestPaletteContrast enforces WCAG AA (US4, T028): text pairs need >=4.5:1
// and non-text (accents, status) at least 3:1, in both variants.
func TestPaletteContrast(t *testing.T) {
	th := &terminalTheme{}
	pairs := []struct {
		name        string
		fg          color.Color
		bg          color.Color
		minRatio    float64
		variantText string
	}{
		{"dark fg/bg", darkForeground, darkBackground, 4.5, "text"},
		{"dark fg/raised", darkForeground, darkRaised, 4.5, "text"},
		{"dark secondary/raised", darkSecondary, darkRaised, 4.5, "text"},
		{"dark placeholder/surface", darkPlaceholder, darkSurface, 4.5, "text"},
		{"light fg/bg", lightForeground, lightBackground, 4.5, "text"},
		{"light fg/raised", lightForeground, lightRaised, 4.5, "text"},
		{"light placeholder/raised", lightPlaceholder, lightRaised, 4.5, "text"},
		{"dark accent/bg (non-text)", AccentColor, darkBackground, 3.0, "non-text"},
		{"light accent/bg (non-text)", color.NRGBA{R: 0x00, G: 0x87, B: 0xA9, A: 0xFF}, lightBackground, 3.0, "non-text"},
		{"light primary/bg", lightPrimary, lightBackground, 3.0, "non-text"},
		{"dark success/surface", darkSuccess, darkSurface, 3.0, "non-text"},
		{"light success/surface", lightSuccess, lightSurface, 3.0, "non-text"},
		{"accent/raised (rail mark)", AccentColor, darkRaised, 3.0, "non-text"},
	}
	_ = th
	for _, p := range pairs {
		got := contrastRatio(p.fg, p.bg)
		if got < p.minRatio {
			t.Errorf("%s: contrast %.2f:1 below %s AA %.1f:1", p.name, got, p.variantText, p.minRatio)
		}
	}

	// The accent theme color tokens must also pass (accentThemeColor);
	// resolve through a light-forced theme like the app does.
	lightTh := &terminalTheme{forcedVariant: theme.VariantLight}
	accentLight := lightTh.Color(accentThemeColor, theme.VariantLight)
	if r := contrastRatio(accentLight, lightBackground); r < 3.0 {
		t.Errorf("accentThemeColor light on light bg: %.2f:1 below 3:1", r)
	}
}

func sameColor(a, b interface {
	RGBA() (uint32, uint32, uint32, uint32)
}) bool {
	ar, ag, ab, aa := a.RGBA()
	br, bg, bb, ba := b.RGBA()
	return ar == br && ag == bg && ab == bb && aa == ba
}
