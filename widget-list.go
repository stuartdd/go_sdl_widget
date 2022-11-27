package go_sdl_widget

import (
	"github.com/veandco/go-sdl2/sdl"
	"github.com/veandco/go-sdl2/ttf"
)

type SDL_List struct {
	x, y, w, h   int32
	widgetId     int32
	state        STATE_BITS
	base         *SDL_LinkedWidget
	background   *sdl.Color
	borderColour *sdl.Color
}

var _ SDL_WidgetInList = (*SDL_List)(nil) // Ensure SDL_Button 'is a' SDL_Widget

func NewSDLList(x, y, w, h, widgetId int32, style STATE_BITS) *SDL_List {
	return &SDL_List{
		x:            x,
		y:            y,
		w:            w,
		h:            h,
		widgetId:     widgetId,
		background:   nil,
		borderColour: nil,
		state:        style | WIDGET_STATE_STA_BITS, // Clear state bits and set enabled, visible and notpressed. Leave style unchanged
	}
}

func (b *SDL_List) GetWidgetId() int32 {
	return b.widgetId
}

func (b *SDL_List) SetWidgetId(widgetId int32) {
	b.widgetId = widgetId
}

func (b *SDL_List) SetVisible(v bool) {
	if v {
		b.state = b.state | WIDGET_STATE_VISIBLE
	} else {
		b.state = b.state & ^WIDGET_STATE_VISIBLE
	}
}

func (b *SDL_List) SetEnabled(e bool) {
	if e {
		b.state = b.state | WIDGET_STATE_ENABLED
	} else {
		b.state = b.state & ^WIDGET_STATE_ENABLED
	}
}

func (b *SDL_List) IsEnabled() bool {
	return (b.state & WIDGET_STATE_ENA_SET) == WIDGET_STATE_ENA_SET
}

func (b *SDL_List) IsVisible() bool {
	return (b.state & WIDGET_STATE_VISIBLE) == WIDGET_STATE_VISIBLE
}

func (b *SDL_List) ShouldDrawBackground() bool {
	return (b.state & WIDGET_STYLE_DRAW_BG) == WIDGET_STYLE_DRAW_BG
}

func (b *SDL_List) ShouldDrawBorder() bool {
	return (b.state & WIDGET_STYLE_DRAW_BORDER) == WIDGET_STYLE_DRAW_BORDER
}

func (b *SDL_List) SetDrawBorder(e bool) {
	if e {
		b.state = b.state | WIDGET_STYLE_DRAW_BORDER
	} else {
		b.state = b.state & ^WIDGET_STYLE_DRAW_BORDER
	}
}

func (b *SDL_List) SetDrawBackground(e bool) {
	if e {
		b.state = b.state | WIDGET_STYLE_DRAW_BG
	} else {
		b.state = b.state & ^WIDGET_STYLE_DRAW_BG
	}
}
func (b *SDL_List) Inside(int32, int32) bool {
	return false
} // Base

func (b *SDL_List) GetBorderColour() *sdl.Color {
	if b.borderColour != nil {
		return b.borderColour
	}
	return GetResourceInstance().GetColour(getStateColourIndex(b.state), WIDGET_COLOR_BORDER)
}

func (b *SDL_List) GetBackground() *sdl.Color {
	if b.background != nil {
		return b.background
	}
	return GetResourceInstance().GetColour(getStateColourIndex(b.state), WIDGET_COLOR_BG)
}

func (b *SDL_List) Click(md *SDL_MouseData) bool {
	return false
}

func (b *SDL_List) Count() int {
	c := 0
	n := b.base
	for n != nil {
		n = n.next
		c++
	}
	return c
}

func (b *SDL_List) GetSize() (int32, int32) {
	return b.w, b.h
}

func (b *SDL_List) SetSize(w, h int32) bool {
	changed := false
	if w > 0 && b.w != w {
		b.w = w
		changed = true
	}
	if h > 0 && b.h != h {
		b.h = h
		changed = true
	}
	return changed
}

func (b *SDL_List) SetPosition(x, y int32) bool {
	if b.x != x || b.y != y {
		b.x = x
		b.y = y
		return true
	}
	return false
}

func (b *SDL_List) GetPosition() (int32, int32) {
	return b.x, b.y
}

func (b *SDL_List) Scale(s float32) {
	b.w = int32(float32(b.w) * s)
	b.h = int32(float32(b.h) * s)
	b.x = int32(float32(b.x) * s)
	b.y = int32(float32(b.y) * s)
}

func (b *SDL_List) Draw(renderer *sdl.Renderer, font *ttf.Font) error {
	if b.IsEnabled() {
		if b.ShouldDrawBackground() {
			bc := b.GetBorderColour()
			renderer.SetDrawColor(bc.R, bc.G, bc.B, bc.A)
			renderer.FillRect(&sdl.Rect{X: b.x, Y: b.y, W: b.w, H: b.h})
		}
		if b.ShouldDrawBorder() {
			bc := b.GetBorderColour()
			renderer.SetDrawColor(bc.R, bc.G, bc.B, bc.A)
			renderer.DrawRect(&sdl.Rect{X: b.x + 1, Y: b.y + 1, W: b.w - 2, H: b.h - 2})
			renderer.DrawRect(&sdl.Rect{X: b.x + 2, Y: b.y + 2, W: b.w - 4, H: b.h - 4})
		}
	}
	return nil
}

func (b *SDL_List) Destroy() {
	// Image cache takes care of all images!
}
