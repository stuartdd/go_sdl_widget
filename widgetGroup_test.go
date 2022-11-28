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
		if w.IsVisible() {
			t.Errorf("Should not be visible %d:%d", i, w.GetWidgetId())
		}
	}
	sub.SetVisible(true)
	for i, w := range l {
		if !w.IsVisible() {
			t.Errorf("Should be visible %d:%d", i, w.GetWidgetId())
		}
	}

	w7 := sub.GetWidget(777)
	if w7.GetWidgetId() != 777 {
		t.Errorf("Should be 777 not %d", w7.GetWidgetId())
	}
	w9 := sub.GetWidget(999)
	if w9.GetWidgetId() != 999 {
		t.Errorf("Should be 999 not %d", w9.GetWidgetId())
	}
	w8 := sub.GetWidget(888)
	if w8.GetWidgetId() != 888 {
		t.Errorf("Should be 888 not %d", w8.GetWidgetId())
	}
	w := sub.GetWidget(0)
	if w != nil {
		t.Errorf("Should not find %d", w.GetWidgetId())
	}

	sub.Scale(0.5)
	wi, hi := w8.GetSize()
	if wi != 5 {
		t.Errorf("id 888 wi should be 5 not %d", wi)
	}
	if hi != 5 {
		t.Errorf("id 888 hi should be 5 not %d", hi)
	}
	x, y := w8.GetPosition()
	if x != 5 {
		t.Errorf("id 888 x should be 5 not %d", x)
	}
	if y != 5 {
		t.Errorf("id 888 y should be 5 not %d", y)
	}

	wi, hi = w9.GetSize()
	if wi != 5 {
		t.Errorf("id 999 wi should be 5 not %d", wi)
	}
	if hi != 5 {
		t.Errorf("id 999 hi should be 5 not %d", hi)
	}
	x, y = w9.GetPosition()
	if x != 5 {
		t.Errorf("id 999 x should be 5 not %d", x)
	}
	if y != 5 {
		t.Errorf("id 999 y should be 5 not %d", y)
	}

}
