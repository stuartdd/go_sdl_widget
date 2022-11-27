package go_sdl_widget

import (
	"testing"
)

func TestSubGroups(t *testing.T) {
	sub := &SDL_WidgetSubGroup{font: nil, id: 0, base: nil, count: 0}
	if sub.count != 0 {
		t.Error("Count should be zero")
	}
	sub.Add(NewSDLSeparator(10, 10, 10, 10, 999, WIDGET_STYLE_BORDER_AND_BG))
	if sub.count != 1 {
		t.Error("Count should be 1")
	}
	sub.Add(NewSDLSeparator(10, 10, 10, 10, 888, WIDGET_STYLE_BORDER_AND_BG))
	if sub.count != 2 {
		t.Error("Count should be 2")
	}
	sub.Add(NewSDLSeparator(10, 10, 10, 10, 777, WIDGET_STYLE_BORDER_AND_BG))
	if sub.count != 3 {
		t.Error("Count should be 3")
	}

	sub.SetVisible(false)
	l := sub.ListWidgets()
	for i, w := range l {
		if (*w).IsVisible() {
			t.Errorf("Should not be visible %d:%d", i, (*w).GetWidgetId())
		}
	}
	sub.SetVisible(true)
	for i, w := range l {
		if !(*w).IsVisible() {
			t.Errorf("Should be visible %d:%d", i, (*w).GetWidgetId())
		}
	}
}
