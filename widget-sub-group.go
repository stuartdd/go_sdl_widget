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
	base, temp           *SDL_LinkedWidget
	countBase, countTemp int
	font                 *ttf.Font
}

var _ SDL_Widget = (*SDL_WidgetSubGroup)(nil)    // Ensure SDL_Button 'is a' SDL_Widget
var _ SDL_Container = (*SDL_WidgetSubGroup)(nil) // Ensure SDL_Button 'is a' SDL_Widget

func NewWidgetSubGroup(x, y, w, h, id int32, font *ttf.Font, style STATE_BITS) *SDL_WidgetSubGroup {
	if font == nil {
		font = GetResourceInstance().GetFont()
	}
	sg := &SDL_WidgetSubGroup{font: font, base: nil, temp: nil, countBase: 0}
	sg.SDL_WidgetBase = initBase(x, y, w, h, id, sg, 0, false, style, nil)
	return sg
}

func (wl *SDL_WidgetSubGroup) Draw(renderer *sdl.Renderer, f *ttf.Font) error {
	if wl.IsVisible() {
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

func (wl *SDL_WidgetSubGroup) Inside(x, y int32) (SDL_Widget, bool) {
	if wl.IsVisible() {
		linkedWidget := wl.base
		for linkedWidget != nil {
			ww := linkedWidget.widget
			wc, isContainer := ww.(SDL_Container)
			if isContainer {
				www, found := wc.Inside(x, y)
				if found {
					return www, true
				}
			} else {
				if isInsideRect(x, y, ww.GetRect()) {
					return ww, true
				}
			}
			linkedWidget = linkedWidget.next
		}
	}
	return nil, false
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
	min := wl.GetSmallestRect(nil)
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
		wl.countBase = 1
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
		wl.countBase = c
	}
	return widget
}

func (wl *SDL_WidgetSubGroup) addToTemp(widget SDL_Widget) SDL_Widget {
	if widget == nil {
		return nil
	}
	if wl.temp == nil {
		wl.temp = &SDL_LinkedWidget{widget: widget, next: nil}
		wl.countTemp = 1
	} else {
		c := 1
		w := wl.temp
		for w != nil {
			c++
			if w.next == nil {
				w.next = &SDL_LinkedWidget{widget: widget, next: nil}
				break
			}
			w = w.next
		}
		wl.countTemp = c
	}
	return widget
}

func (wl *SDL_WidgetSubGroup) swapTemp() {
	v := wl.IsVisible()
	wl.SetVisible(false)
	defer wl.SetVisible(v)
	w := wl.base
	wl.base = wl.temp
	wl.countBase = wl.countTemp
	wl.temp = nil
	wl.countTemp = 0
	for w != nil {
		w.widget.Destroy()
		w = w.next
	}
}

func (wl *SDL_WidgetSubGroup) GetSmallestRect(within *sdl.Rect) *sdl.Rect {
	r := sdl.Rect{X: wl.x, Y: wl.y, W: 0, H: 0}
	w := wl.base
	for w != nil {
		r = r.Union(w.widget.GetRect())
		w = w.next
	}
	if within != nil {
		max := (within.X + within.W)
		if (r.X + r.W) > max {
			r.W = max - r.X
		}
		max = (within.Y + within.H)
		if (r.Y + r.H) > max {
			r.H = max - r.Y
		}
	}
	return &r
}

func (wl *SDL_WidgetSubGroup) ListWidgets() []SDL_Widget {
	list := make([]SDL_Widget, wl.countBase)
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
		wc, ok := w.widget.(SDL_Container)
		if ok {
			wc.NextFrame()
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

func (wl *SDL_WidgetSubGroup) ArrangeGrid(rect *sdl.Rect, padding, rowHeight int32, colsW []int32) (int32, int32) {
	if wl.IsVisible() && wl.base != nil {
		sr := wl.GetSmallestRect(rect)
		wl.SetRect(sr)
		x := rect.X
		y := rect.Y
		maxW := x + rect.W
		//	maxH := y + rect.H
		var width int32
		c := 0
		linkedW := wl.base
		for linkedW != nil && linkedW.widget != nil {
			ww := linkedW.widget
			if ww.IsVisible() {
				width, _ = ww.GetSize()
				ww.SetPosition(x, y)
				if c == len(colsW)-1 {
					ww.SetSize(maxW-x, rowHeight)
				} else {
					ww.SetSize(colsW[c], rowHeight)
				}
				x = (x + width)
				c++
				if c >= len(colsW) {
					c = 0
					x = rect.X
					y = y + rowHeight + padding - 3
				}
			}
			linkedW = linkedW.next
		}

		return x, y
	}
	return 0, 0
}
