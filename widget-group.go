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
	return &SDL_WidgetGroup{font: font, wigetLists: make([]*SDL_WidgetSubGroup, 0)}
}

func (wg *SDL_WidgetGroup) NewWidgetSubGroup(x, y, w, h, id int32, style STATE_BITS) *SDL_WidgetSubGroup {
	wsg := NewWidgetSubGroup(x, y, w, h, id, wg.font, style)
	wg.wigetLists = append(wg.wigetLists, wsg)
	return wsg
}

func (wg *SDL_WidgetGroup) AllWidgets() []SDL_Widget {
	l := make([]SDL_Widget, 0)
	for _, wList := range wg.wigetLists {
		l = append(l, wList.ListWidgets()...)
	}
	return l
}

func (wg *SDL_WidgetGroup) SetFocusedId(id int32) {
	for _, wList := range wg.wigetLists {
		wList.SetFocusedId(id)
	}
}

func (wg *SDL_WidgetGroup) ClearFocus() {
	for _, wList := range wg.wigetLists {
		wList.ClearFocus()
	}
}

func (wg *SDL_WidgetGroup) ClearSelection() {
	for _, wList := range wg.wigetLists {
		wList.ClearSelection()
	}
}

func (wg *SDL_WidgetGroup) GetFocusedWidget() SDL_Widget {
	for _, wList := range wg.wigetLists {
		f := wList.GetFocusedWidget()
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
	for _, wList := range wg.wigetLists {
		if wList.KeyPress(c, ctrl, down) {
			return true
		}
	}
	return false
}

func (wg *SDL_WidgetGroup) Scale(s float32) {
	for _, w := range wg.wigetLists {
		w.Scale(s)
	}
}

func (wg *SDL_WidgetGroup) NextFrame() {
	for _, w := range wg.wigetLists {
		w.NextFrame()
	}
}

func (wg *SDL_WidgetGroup) Destroy() {
	for _, w := range wg.wigetLists {
		w.Destroy()
	}
	GetResourceInstance().GetTextureCache().Destroy()
}

func (wg *SDL_WidgetGroup) Draw(renderer *sdl.Renderer) {
	for _, w := range wg.wigetLists {
		w.Draw(renderer, wg.font)
	}
}

func (wg *SDL_WidgetGroup) InsideWidget(x, y int32) SDL_Widget {
	for _, wl := range wg.wigetLists {
		w := wl.InsideWidget(x, y)
		if w != nil {
			return w
		}
	}
	return nil
}
