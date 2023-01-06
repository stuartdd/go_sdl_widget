package go_sdl_widget

import (
	"github.com/veandco/go-sdl2/sdl"
	"github.com/veandco/go-sdl2/ttf"
)

/****************************************************************************************
* Container for SDL_Widgets. A list of lists
**/
type SDL_WidgetGroup struct {
	wigetLists []*SDL_WidgetSubGroup
	font       *ttf.Font
}

func NewWidgetGroup(font *ttf.Font) *SDL_WidgetGroup {
	if font == nil {
		font = GetResourceInstance().GetFont()
	}
	return &SDL_WidgetGroup{font: font, wigetLists: make([]*SDL_WidgetSubGroup, 0)}
}

func (wg *SDL_WidgetGroup) NewWidgetSubGroup(x, y, w, h, id int32, style STATE_BITS) *SDL_WidgetSubGroup {
	wsg := NewWidgetSubGroup(x, y, w, h, id, wg.font, style)
	wg.wigetLists = append(wg.wigetLists, wsg)
	return wsg
}

func (wg *SDL_WidgetGroup) AllWidgets() []SDL_Widget {
	l := make([]SDL_Widget, 0)
	for _, wl := range wg.wigetLists {
		l = append(l, wl.ListWidgets()...)
	}
	return l
}

func (wg *SDL_WidgetGroup) SetFocusedId(id int32) {
	for _, wl := range wg.wigetLists {
		wl.SetFocusedId(id)
	}
}

func (wg *SDL_WidgetGroup) ClearFocus() {
	for _, wl := range wg.wigetLists {
		wl.ClearFocus()
	}
}

func (wg *SDL_WidgetGroup) ClearSelection() {
	for _, wl := range wg.wigetLists {
		wl.ClearSelection()
	}
}

func (wg *SDL_WidgetGroup) GetFocusedWidget() SDL_Widget {
	for _, wl := range wg.wigetLists {
		f := wl.GetFocusedWidget()
		if f != nil {
			return f
		}
	}
	return nil
}

func (wl *SDL_WidgetGroup) GetWidgetSubGroup(id int32) *SDL_WidgetSubGroup {
	for _, w := range wl.wigetLists {
		if (*w).GetWidgetId() == id {
			return w
		}
	}
	return nil
}

func (wl *SDL_WidgetGroup) GetWidgetWithId(id int32) SDL_Widget {
	for _, wl := range wl.wigetLists {
		w := wl.GetWidgetWithId(id)
		if w != nil {
			return w
		}
	}
	return nil
}

func (wg *SDL_WidgetGroup) AllSubGroups() []*SDL_WidgetSubGroup {
	l := make([]*SDL_WidgetSubGroup, 0)
	l = append(l, wg.wigetLists...)
	return l
}

func (wg *SDL_WidgetGroup) KeyPress(c int, ctrl, down bool) bool {
	for _, wl := range wg.wigetLists {
		if wl.IsEnabled() {
			if wl.KeyPress(c, ctrl, down) {
				return true
			}
		}
	}
	return false
}

func (wg *SDL_WidgetGroup) Scale(s float32) {
	for _, wl := range wg.wigetLists {
		wl.Scale(s)
	}
}

func (wg *SDL_WidgetGroup) NextFrame() {
	for _, wl := range wg.wigetLists {
		wl.NextFrame()
	}
}

func (wg *SDL_WidgetGroup) Destroy() {
	for _, w := range wg.wigetLists {
		w.Destroy()
	}
}

func (wg *SDL_WidgetGroup) Draw(renderer *sdl.Renderer) {
	for _, wl := range wg.wigetLists {
		if wl.IsVisible() {
			wl.Draw(renderer, wg.font)
		}
	}
}

func (wg *SDL_WidgetGroup) InsideWidget(x, y int32) SDL_Widget {
	for _, wl := range wg.wigetLists {
		if wl.IsEnabled() {
			w, ok := wl.Inside(x, y)
			if ok && w != nil {
				return w
			}
		}
	}
	return nil
}
