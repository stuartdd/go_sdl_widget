package go_sdl_widget

import (
	"github.com/veandco/go-sdl2/sdl"
	"github.com/veandco/go-sdl2/ttf"
)

/****************************************************************************************
* Container for SDL_Widget instances.
**/
type SDL_WidgetSubGroup struct {
	SDL_WidgetBase
	base  *SDL_LinkedWidget
	count int
	font  *ttf.Font
}

var _ SDL_Widget = (*SDL_WidgetSubGroup)(nil) // Ensure SDL_Button 'is a' SDL_Widget

func NewWidgetSubGroup(x, y, w, h, id int32, font *ttf.Font, style STATE_BITS) *SDL_WidgetSubGroup {
	if font == nil {
		font = GetResourceInstance().GetFont()
	}
	sg := &SDL_WidgetSubGroup{font: font, base: nil, count: 0}
	sg.SDL_WidgetBase = initBase(x, y, w, h, id, 0, false, style, nil)
	return sg
}

func (wl *SDL_WidgetSubGroup) Draw(renderer *sdl.Renderer, f *ttf.Font) error {
	if wl.IsEnabled() {
		if wl.ShouldDrawBackground() {
			bc := wl.GetBackground()
			renderer.SetDrawColor(bc.R, bc.G, bc.B, bc.A)
			renderer.FillRect(&sdl.Rect{X: wl.x, Y: wl.y, W: wl.w, H: wl.h})
		}
		if f == nil {
			f = wl.GetFont()
		}
		w := wl.base
		for w != nil {
			err := w.widget.Draw(renderer, f)
			if err != nil {
				return err
			}
			w = w.next
		}
		if wl.ShouldDrawBorder() {
			bc := wl.GetBorderColour()
			renderer.SetDrawColor(bc.R, bc.G, bc.B, bc.A)
			renderer.DrawRect(&sdl.Rect{X: wl.x + 1, Y: wl.y + 1, W: wl.w - 2, H: wl.h - 2})
			renderer.DrawRect(&sdl.Rect{X: wl.x + 2, Y: wl.y + 2, W: wl.w - 4, H: wl.h - 4})
		}
	}
	return nil
}

func (wl *SDL_WidgetSubGroup) Click(md *SDL_MouseData) bool {
	if wl.IsEnabled() {
		w := wl.base
		for w != nil {
			if w.widget.Click(md) {
				return true
			}
			w = w.next
		}
	}
	return false
}

func (wl *SDL_WidgetSubGroup) KeyPress(c int, ctrl, down bool) bool {
	if wl.IsEnabled() {
		w := wl.base
		for w != nil {
			if w.widget.CanFocus() && w.widget.IsFocused() {
				if w.widget.KeyPress(c, ctrl, down) {
					return true
				}
			}
			w = w.next
		}
	}
	return false
}

func (wl *SDL_WidgetSubGroup) Scale(s float32) {
	wl.SDL_WidgetBase.Scale(s)
	w := wl.base
	for w != nil {
		w.widget.Scale(s)
		w = w.next
	}
}

func (wl *SDL_WidgetSubGroup) Destroy() {
	wl.SDL_WidgetBase.Destroy()
	w := wl.base
	for w != nil {
		w.widget.Destroy()
		w = w.next
	}
}

func (wl *SDL_WidgetSubGroup) Inside(x, y int32) bool {
	if wl.IsVisible() {
		return isInsideRect(x, y, wl.SDL_WidgetBase.GetRect())
	}
	return false
}

func (wl *SDL_WidgetSubGroup) SetPositionRel(x, y int32) bool {
	if x == 0 && y == 0 {
		return false
	}
	wl.SDL_WidgetBase.SetPositionRel(x, y)
	w := wl.base
	for w != nil {
		w.widget.SetPositionRel(x, y)
		w = w.next
	}
	return true
}

func (wl *SDL_WidgetSubGroup) SetPosition(x, y int32) bool {
	min := wl.GetSmallestRect()
	dx := x - min.X
	dy := y - min.Y
	wl.SetPositionRel(dx, dy)
	return true
}

// ------------------------------------------------------------
// Sub Group Specific (not part of SDL_Widget interface)
// ------------------------------------------------------------
func (wl *SDL_WidgetSubGroup) SetFont(font *ttf.Font) {
	wl.font = font
}

func (wl *SDL_WidgetSubGroup) GetFont() *ttf.Font {
	if wl.font == nil {
		return GetResourceInstance().GetFont()
	}
	return wl.font
}

func (wl *SDL_WidgetSubGroup) Add(widget SDL_Widget) SDL_Widget {
	if widget == nil {
		return nil
	}
	if wl.base == nil {
		wl.base = &SDL_LinkedWidget{widget: widget, next: nil}
		wl.count = 1
	} else {
		c := 1
		w := wl.base
		for w != nil {
			c++
			if w.next == nil {
				w.next = &SDL_LinkedWidget{widget: widget, next: nil}
				break
			}
			w = w.next
		}
		wl.count = c
	}
	return widget
}

func (wl *SDL_WidgetSubGroup) GetSmallestRect() *sdl.Rect {
	r := sdl.Rect{X: 0, Y: 0, W: 0, H: 0}
	w := wl.base
	for w != nil {
		r = r.Union(w.widget.GetRect())
		w = w.next
	}
	return &r
}

func (wl *SDL_WidgetSubGroup) ListWidgets() []SDL_Widget {
	list := make([]SDL_Widget, wl.count)
	i := 0
	w := wl.base
	for w != nil {
		list[i] = w.widget
		w = w.next
		i++
	}
	return list
}

func (wl *SDL_WidgetSubGroup) GetWidgetWithId(id int32) SDL_Widget {
	w := wl.base
	for w != nil {
		if w.widget.GetWidgetId() == id {
			return w.widget
		}
		w = w.next
	}
	return nil
}

func (wl *SDL_WidgetSubGroup) InsideWidget(x, y int32) SDL_Widget {
	w := wl.base
	for w != nil {
		if w.widget.Inside(x, y) {
			return w.widget
		}
		w = w.next
	}
	return nil
}

func (wl *SDL_WidgetSubGroup) SetFocusedId(id int32) {
	w := wl.base
	for w != nil {
		if w.widget.CanFocus() {
			w.widget.SetFocused(w.widget.GetWidgetId() == id)
		}
		w = w.next
	}
}

func (wl *SDL_WidgetSubGroup) ClearFocus() {
	w := wl.base
	for w != nil {
		if w.widget.CanFocus() {
			w.widget.SetFocused(false)
		}
		w = w.next
	}
}

func (wl *SDL_WidgetSubGroup) RemoveAllWidgets() {
	w := wl.base
	for w != nil {
		w.widget.Destroy()
		w = w.next
	}
	wl.base = nil
}

func (wl *SDL_WidgetSubGroup) ClearSelection() {
	w := wl.base
	for w != nil {
		f, ok := w.widget.(SDL_CanSelectText)
		if ok {
			f.ClearSelection()
		}
		w = w.next
	}
}

func (wl *SDL_WidgetSubGroup) NextFrame() {
	w := wl.base
	for w != nil {
		iw, ok := w.widget.(SDL_ImageWidget)
		if ok {
			iw.NextFrame()
		}
		w = w.next
	}
}

func (wl *SDL_WidgetSubGroup) GetFocusedWidget() SDL_Widget {
	w := wl.base
	for w != nil {
		if w.widget.CanFocus() && w.widget.IsFocused() {
			return w.widget
		}
		w = w.next
	}
	return nil
}

func (wl *SDL_WidgetSubGroup) ArrangeLR(xx, yy, padding int32) (int32, int32) {
	x := xx
	y := yy
	var width int32
	w := wl.base
	for w != nil {
		ww := w.widget
		if ww.IsVisible() {
			ww.SetPosition(x, y)
			width, _ = ww.GetSize()
			x = x + width + padding
		}
		w = w.next
	}
	return x, y
}

func (wl *SDL_WidgetSubGroup) ArrangeRL(xx, yy, padding int32) (int32, int32) {
	x := xx
	y := yy
	var width int32

	w := wl.base
	for w != nil {
		ww := w.widget
		if ww.IsVisible() {
			width, _ = ww.GetSize()
			ww.SetPosition(x-width, y)
			x = (x - width) - padding
		}
		w = w.next
	}
	return x, y
}
