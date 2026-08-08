package gui

import (
	"testing"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/theme"
)

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

func sameColor(a, b interface {
	RGBA() (uint32, uint32, uint32, uint32)
}) bool {
	ar, ag, ab, aa := a.RGBA()
	br, bg, bb, ba := b.RGBA()
	return ar == br && ag == bg && ab == bb && aa == ba
}
