package go_sdl_widget

import (
	"testing"
)

func TestJsonDataRename(t *testing.T) {
	w := NewSDLSeparator(0, 0, 0, 0, 99, WIDGET_STYLE_BORDER_1|WIDGET_STYLE_DRAW_BG)
	assertStateBools(t, "Initial state", w, true, true, true, true, true)

	w.SetNotError(false)
	assertStateBools(t, "Not Error false", w, true, true, true, false, true)
	w.SetNotError(false)
	assertStateBools(t, "Not Error false", w, true, true, true, false, true)
	w.SetNotError(true)
	assertStateBools(t, "Not Error true", w, true, true, true, true, true)
	w.SetNotError(true)
	assertStateBools(t, "Not Error true", w, true, true, true, true, true)

	w.SetNotFocused(false)
	assertStateBools(t, "Not Focused false", w, true, false, true, true, true)
	w.SetNotFocused(false)
	assertStateBools(t, "Not Focused false", w, true, false, true, true, true)
	w.SetNotFocused(true)
	assertStateBools(t, "Not Focused true", w, true, true, true, true, true)
	w.SetNotFocused(true)
	assertStateBools(t, "Not Focused true", w, true, true, true, true, true)

	w.SetVisible(false)
	assertStateBools(t, "Visible true", w, false, true, true, true, false)
	w.SetVisible(false)
	assertStateBools(t, "Visible true", w, false, true, true, true, false)
	w.SetVisible(true)
	assertStateBools(t, "Visible false", w, true, true, true, true, true)
	w.SetVisible(true)
	assertStateBools(t, "Visible false", w, true, true, true, true, true)

	w.SetNotClicked(false)
	assertStateBools(t, "Not Clicked false", w, false, true, false, true, true)
	w.SetNotClicked(false)
	assertStateBools(t, "Not Clicked false", w, false, true, false, true, true)
	w.SetNotClicked(true)
	assertStateBools(t, "Not Clicked true", w, true, true, true, true, true)
	w.SetNotClicked(true)
	assertStateBools(t, "Not Clicked true", w, true, true, true, true, true)

	w.SetEnabled(false)
	assertStateBools(t, "Enabled false", w, false, true, true, true, true)
	w.SetEnabled(false)
	assertStateBools(t, "Enabled false", w, false, true, true, true, true)
	w.SetEnabled(true)
	assertStateBools(t, "Enabled true", w, true, true, true, true, true)
	w.SetEnabled(true)
	assertStateBools(t, "Enabled true", w, true, true, true, true, true)

	assertBool(t, "Draw BG Initial", "ShouldDrawBackground", w.ShouldDrawBackground(), true)
	w.SetDrawBackground(false)
	assertBool(t, "Draw BG false", "ShouldDrawBackground", w.ShouldDrawBackground(), false)
	w.SetDrawBackground(false)
	assertBool(t, "Draw BG false", "ShouldDrawBackground", w.ShouldDrawBackground(), false)
	w.SetDrawBackground(true)
	assertBool(t, "Draw BG true", "ShouldDrawBackground", w.ShouldDrawBackground(), true)
	w.SetDrawBackground(true)
	assertBool(t, "Draw BG true", "ShouldDrawBackground", w.ShouldDrawBackground(), true)

	assertBool(t, "Draw Border Initial", "ShouldDrawBackground", w.ShouldDrawBorder(), true)
	w.SetDrawBorder(false)
	assertBool(t, "Draw Border false", "ShouldDrawBackground", w.ShouldDrawBorder(), false)
	w.SetDrawBorder(false)
	assertBool(t, "Draw Border false", "ShouldDrawBackground", w.ShouldDrawBorder(), false)
	w.SetDrawBorder(true)
	assertBool(t, "Draw Border true", "ShouldDrawBackground", w.ShouldDrawBorder(), true)
	w.SetDrawBorder(true)
	assertBool(t, "Draw Border true", "ShouldDrawBackground", w.ShouldDrawBorder(), true)

}

func assertStateBools(t *testing.T, message string, w SDL_Widget, ena, nfoc, ncli, nerr, vis bool) {
	assertBool(t, message, "IsEnabled", w.IsEnabled(), ena)
	assertBool(t, message, "IsNotFocused", w.IsNotFocused(), nfoc)
	assertBool(t, message, "IsNotClicked", w.IsNotClicked(), ncli)
	assertBool(t, message, "IsNotError", w.IsNotError(), nerr)
	assertBool(t, message, "IsVisible", w.IsVisible(), vis)
}

func assertBool(t *testing.T, message1, message2 string, val, expected bool) {
	if val != expected {
		t.Errorf("%s: %s Actual %t Expected %t", message1, message2, val, expected)
	}
}
