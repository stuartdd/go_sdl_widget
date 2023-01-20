package go_sdl_widget

import (
	"fmt"
	"testing"

	"github.com/veandco/go-sdl2/sdl"
)

func TestWidgetCharCache(t *testing.T) {
	cols := make([]*sdl.Color, 0)
	cols = append(cols, &sdl.Color{R: 255, G: 0, B: 0, A: 255})
	cols = append(cols, &sdl.Color{R: 0, G: 255, B: 0, A: 255})
	cols = append(cols, &sdl.Color{R: 0, G: 0, B: 255, A: 255})
	cols = append(cols, &sdl.Color{R: 0, G: 0, B: 0, A: 255})
	cols = append(cols, &sdl.Color{R: 0, G: 0, B: 0, A: 0})
	for _, c := range cols {
		i := GetColourId(c)
		fmt.Printf("%10.0X %d\n", i, i)
	}
}

func TestWidgetStateColourIndex(t *testing.T) {
	w := NewSDLEntry(0, 0, 0, 0, 99, "HI", WIDGET_STYLE_DRAW_BORDER_AND_BG, nil)
	assertStateColour(t, "Initial 1", w.getResourceColourStateIndex(), WIDGET_COLOUR_INDEX_ENABLED)
	w.SetEnabled(false)
	assertStateColour(t, "Disabled", w.getResourceColourStateIndex(), WIDGET_COLOUR_INDEX_DISABLE)
	w.SetFocused(true)
	assertStateColour(t, "Disabled + Focused", w.getResourceColourStateIndex(), WIDGET_COLOUR_INDEX_DISABLE)
	w.SetError(true)
	assertStateColour(t, "Disabled + Focused + Error", w.getResourceColourStateIndex(), WIDGET_COLOUR_INDEX_DISABLE)

	w = NewSDLEntry(0, 0, 0, 0, 99, "HI", WIDGET_STYLE_DRAW_BORDER_AND_BG, nil)
	assertStateColour(t, "Initial 2", w.getResourceColourStateIndex(), WIDGET_COLOUR_INDEX_ENABLED)
	w.SetFocused(true)
	assertStateColour(t, "Focused", w.getResourceColourStateIndex(), WIDGET_COLOUR_INDEX_FOCUS)
	w.SetError(true)
	assertStateColour(t, "Focused + Error", w.getResourceColourStateIndex(), WIDGET_COLOUR_INDEX_ERROR)

	w = NewSDLEntry(0, 0, 0, 0, 99, "HI", WIDGET_STYLE_DRAW_BORDER_AND_BG, nil)
	assertStateColour(t, "Initial 3", w.getResourceColourStateIndex(), WIDGET_COLOUR_INDEX_ENABLED)
	w.SetError(true)
	assertStateColour(t, "Error", w.getResourceColourStateIndex(), WIDGET_COLOUR_INDEX_ERROR)
	w.SetFocused(true)
	assertStateColour(t, "Error", w.getResourceColourStateIndex(), WIDGET_COLOUR_INDEX_ERROR)

}

func TestWidgetBaseState(t *testing.T) {
	w := NewSDLEntry(0, 0, 0, 0, 99, "HI", WIDGET_STYLE_DRAW_BORDER_AND_BG, nil)
	assertStateBools(t, "Initial state", w, true, false, false, false, true)

	w.SetError(true)
	assertStateBools(t, "1 Not Error false", w, true, false, false, true, true)
	w.SetError(true)
	assertStateBools(t, "2 Not Error false", w, true, false, false, true, true)
	w.SetError(false)
	assertStateBools(t, "3 Not Error true", w, true, false, false, false, true)
	w.SetError(false)
	assertStateBools(t, "4 Not Error true", w, true, false, false, false, true)

	w.SetFocused(true)
	assertStateBools(t, "5 Not Focused false", w, true, true, false, false, true)
	w.SetFocused(true)
	assertStateBools(t, "6 Not Focused false", w, true, true, false, false, true)
	w.SetFocused(false)
	assertStateBools(t, "7 Not Focused true", w, true, false, false, false, true)
	w.SetFocused(false)
	assertStateBools(t, "8 Not Focused true", w, true, false, false, false, true)

	w.SetVisible(false)
	assertStateBools(t, "9 Visible true", w, false, false, false, false, false)
	w.SetVisible(false)
	assertStateBools(t, "10 Visible true", w, false, false, false, false, false)
	w.SetVisible(true)
	assertStateBools(t, "11 Visible false", w, true, false, false, false, true)
	w.SetVisible(true)
	assertStateBools(t, "12 Visible false", w, true, false, false, false, true)

	w.SetClicked(true)
	assertStateBools(t, "13 Not Clicked false", w, false, false, true, false, true)
	w.SetClicked(true)
	assertStateBools(t, "14 Not Clicked false", w, false, false, true, false, true)
	w.SetClicked(false)
	assertStateBools(t, "15 Not Clicked true", w, true, false, false, false, true)
	w.SetClicked(false)
	assertStateBools(t, "16 Not Clicked true", w, true, false, false, false, true)

	w.SetEnabled(false)
	assertStateBools(t, "17 Enabled false", w, false, false, false, false, true)
	w.SetEnabled(false)
	assertStateBools(t, "18 Enabled false", w, false, false, false, false, true)
	w.SetEnabled(true)
	assertStateBools(t, "19 Enabled true", w, true, false, false, false, true)
	w.SetEnabled(true)
	assertStateBools(t, "20 Enabled true", w, true, false, false, false, true)

	assertBool(t, "Draw BG Initial", "ShouldDrawBackground", w.ShouldDrawBackground(), true)
	w.SetDrawBackground(false)
	assertBool(t, "Draw BG false", "ShouldDrawBackground", w.ShouldDrawBackground(), false)
	w.SetDrawBackground(false)
	assertBool(t, "Draw BG false", "ShouldDrawBackground", w.ShouldDrawBackground(), false)
	w.SetDrawBackground(true)
	assertBool(t, "Draw BG true", "ShouldDrawBackground", w.ShouldDrawBackground(), true)
	w.SetDrawBackground(true)
	assertBool(t, "Draw BG true", "ShouldDrawBackground", w.ShouldDrawBackground(), true)

	assertBool(t, "Draw Border Initial", "ShouldDrawBorder", w.ShouldDrawBorder(), true)
	w.SetDrawBorder(false)
	assertBool(t, "Draw Border false", "ShouldDrawBorder", w.ShouldDrawBorder(), false)
	w.SetDrawBorder(false)
	assertBool(t, "Draw Border false", "ShouldDrawBorder", w.ShouldDrawBorder(), false)
	w.SetDrawBorder(true)
	assertBool(t, "Draw Border true", "ShouldDrawBorder", w.ShouldDrawBorder(), true)
	w.SetDrawBorder(true)
	assertBool(t, "Draw Border true", "ShouldDrawBorder", w.ShouldDrawBorder(), true)

}

func TestWidgetButtonInitial(t *testing.T) {
	w := NewSDLButton(0, 0, 0, 0, 99, "Button", WIDGET_STYLE_DRAW_BORDER_AND_BG, 0, nil)
	assertStateBools(t, "NewSDLButton initial", w, true, false, false, false, true)
	assertBool(t, "NewSDLButton initial", "ShouldDrawBackground", w.ShouldDrawBackground(), true)
	assertBool(t, "NewSDLButton initial", "ShouldDrawBorder", w.ShouldDrawBorder(), true)
}

func assertStateBools(t *testing.T, message string, w SDL_Widget, ena, foc, cli, err, vis bool) {
	assertBool(t, message, "IsEnabled", w.IsEnabled(), ena)
	assertBool(t, message, "IsFocused", w.IsFocused(), foc)
	assertBool(t, message, "IsClicked", w.IsClicked(), cli)
	assertBool(t, message, "IsError", w.IsError(), err)
	assertBool(t, message, "IsVisible", w.IsVisible(), vis)

}

func assertBool(t *testing.T, message1, message2 string, val, expected bool) {
	if val != expected {
		t.Errorf("%s: %s Actual %t Expected %t", message1, message2, val, expected)
	}
}
func assertInt(t *testing.T, message1 string, val, expected int) {
	if val != expected {
		t.Errorf("%s: Actual %d Expected %d", message1, val, expected)
	}
}
func assertStateColour(t *testing.T, message1 string, val, expected STATE_COLOUR) {
	if val != expected {
		t.Errorf("%s: Actual %d Expected %d", message1, val, expected)
	}
}
