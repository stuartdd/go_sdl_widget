package go_sdl_widget

import (
	"fmt"
	"math"
	"time"

	"github.com/veandco/go-sdl2/sdl"
	"github.com/veandco/go-sdl2/ttf"
)

type ALIGN_TEXT int
type ROTATE_SHAPE_90 int
type KBD_KEY_MODE int
type ENTRY_EVENT_TYPE int
type STATE_BITS uint16
type LOG_LEVEL int

const (
	LOG_LEVEL_ERROR LOG_LEVEL = iota
	LOG_LEVEL_WARN
	LOG_LEVEL_OK

	ALIGN_CENTER ALIGN_TEXT = iota
	ALIGN_LEFT
	ALIGN_RIGHT
	ALIGN_FIT

	ROTATE_0 ROTATE_SHAPE_90 = iota
	ROTATE_90
	ROTATE_180
	ROTATE_270

	ENTRY_EVENT_INSERT ENTRY_EVENT_TYPE = iota
	ENTRY_EVENT_DELETE
	ENTRY_EVENT_BS
	ENTRY_EVENT_FINISH
	ENTRY_EVENT_NONE
	ENTRY_EVENT_FOCUS
	ENTRY_EVENT_UN_FOCUS

	WIDGET_STYLE_DRAW_NONE          STATE_BITS = 0b0000000000000001
	WIDGET_STYLE_DRAW_BORDER        STATE_BITS = 0b0000000000000010
	WIDGET_STYLE_DRAW_BG            STATE_BITS = 0b0000000000001000
	WIDGET_STYLE_DRAW_BORDER_AND_BG STATE_BITS = WIDGET_STYLE_DRAW_BORDER | WIDGET_STYLE_DRAW_BG
	WIDGET_STATE_ENABLED            STATE_BITS = 0b0000000000010000
	WIDGET_STATE_VISIBLE            STATE_BITS = 0b0000000000100000
	WIDGET_STATE_NOT_FOCUSED        STATE_BITS = 0b0000000001000000
	WIDGET_STATE_NOT_ERROR          STATE_BITS = 0b0000000010000000
	WIDGET_STATE_NOT_CLICKED        STATE_BITS = 0b0000000100000000
	WIDGET_STATE_STA_BITS           STATE_BITS = 0b0000000111110000 // Clear state AND mask. Retains style.
	WIDGET_STATE_ENA_SET            STATE_BITS = 0b0000000100110000 // Enabled visible and not-clicked

	DEG_TO_RAD float64 = (math.Pi / 180)
)

var TEXTURE_CACHE_TEXT_PREF = "TxCaPr987"

type SDL_Widget interface {
	Draw(*sdl.Renderer, *ttf.Font) error
	Scale(float32)
	Click(*SDL_MouseData) bool
	SetOnClick(func(string, int32, int32, int32) bool) // Base
	Inside(int32, int32) (SDL_Widget, bool)            // Base
	GetRect() *sdl.Rect                                // Base
	SetRect(*sdl.Rect)                                 // Base
	SetWidgetId(int32)                                 // Base
	GetWidgetId() int32                                // Base
	SetVisible(bool)                                   // Base
	IsVisible() bool                                   // Base
	SetClicked(bool)                                   // Base
	IsClicked() bool                                   // Base
	SetEnabled(bool)                                   // Base
	IsEnabled() bool                                   // Base
	SetError(bool)                                     // Base
	IsError() bool                                     // Base
	SetPosition(int32, int32) bool                     // Base
	SetPositionRel(int32, int32) bool                  // Base
	GetPosition() (int32, int32)                       // Base
	SetSize(int32, int32) bool                         // Base
	GetSize() (int32, int32)                           // Base
	GetForeground() *sdl.Color                         // Base
	GetBackground() *sdl.Color                         // Base
	GetBorderColour() *sdl.Color                       // Base
	SetForeground(*sdl.Color)                          // Base
	SetBackground(*sdl.Color)                          // Base
	SetBorderColour(*sdl.Color)                        // Base
	SetDrawBackground(bool)                            // Base
	ShouldDrawBackground() bool                        // Base
	SetDrawBorder(bool)                                // Base
	ShouldDrawBorder() bool                            // Base
	Destroy()                                          // Base
	SetLog(func(LOG_LEVEL, string))
	Log(LOG_LEVEL, string)
	CanLog() bool
	CanFocus() bool
	KeyPress(c int, ctrl, down bool) bool
	SetFocused(bool)            // Base
	IsFocused() bool            // Base
	GetFocusColour() *sdl.Color // Base
	SetFocusColour(*sdl.Color)  // Base
	String() string
}

type SDL_CanSelectText interface {
	ClearSelection()
	GetSelectedText() string
}

type SDL_TextWidget interface {
	SetText(text string)
	GetText() string
}

type SDL_Container interface {
	Add(SDL_Widget) SDL_Widget
	ListWidgets() []SDL_Widget
	GetWidgetWithId(int32) SDL_Widget
	SetFocusedId(int32)
	GetFocusedWidget() SDL_Widget
	ClearFocus()
	Inside(int32, int32) (SDL_Widget, bool)
	NextFrame()
}

type SDL_ImageWidget interface {
	SetFrame(tf int32)
	GetFrame() int32
	NextFrame() int32
	GetFrameCount() int32
}

type SDL_LinkedWidget struct {
	widget SDL_Widget
	next   *SDL_LinkedWidget
}

type SDL_WidgetBase struct {
	x, y, w, h   int32
	widgetId     int32
	instance     SDL_Widget
	deBounce     int
	onClick      func(string, int32, int32, int32) bool
	background   *sdl.Color
	foreground   *sdl.Color
	borderColour *sdl.Color
	focusColour  *sdl.Color
	state        STATE_BITS
	canfocus     bool
	log          func(LOG_LEVEL, string)
}

/****************************************************************************************
* Common (base) functions for ALL SDL_Widget instances
**/
func initBase(x, y, w, h, widgetId int32, instance SDL_Widget, deBounce int, canfocus bool, style STATE_BITS, onClick func(string, int32, int32, int32) bool) SDL_WidgetBase {
	return SDL_WidgetBase{
		x:            x,
		y:            y,
		w:            w,
		h:            h,
		widgetId:     widgetId,
		instance:     instance,
		canfocus:     canfocus,
		deBounce:     deBounce,
		onClick:      onClick,
		background:   nil,
		foreground:   nil,
		borderColour: nil,
		focusColour:  nil,
		state:        style | WIDGET_STATE_STA_BITS, // Clear state bits and set enabled, visible and notpressed. Leave style unchanged
	}
}

func (b *SDL_WidgetBase) String() string {
	if b.instance != nil {
		tw, isTextWidget := b.instance.(SDL_TextWidget)
		if isTextWidget {
			return tw.GetText()
		}
	}
	return fmt.Sprintf("ID:%d", b.widgetId)
}

func (b *SDL_WidgetBase) Click(md *SDL_MouseData) bool {
	if b.IsEnabled() && b.onClick != nil {
		if b.deBounce > 0 {
			b.SetClicked(true)
			defer func() {
				time.Sleep(time.Millisecond * time.Duration(b.deBounce))
				b.SetClicked(false)
			}()
		}
		return b.onClick(b.String(), b.widgetId, md.x, md.y)
	}
	return false
}

func (b *SDL_WidgetBase) SetOnClick(f func(string, int32, int32, int32) bool) {
	b.onClick = f
}

func (b *SDL_WidgetBase) KeyPress(c int, ctrl bool, down bool) bool {
	return false
}

func (b *SDL_WidgetBase) Destroy() {
}

func (b *SDL_WidgetBase) GetWidgetId() int32 {
	return b.widgetId
}

func (b *SDL_WidgetBase) SetWidgetId(widgetId int32) {
	b.widgetId = widgetId
}

func (b *SDL_WidgetBase) SetPosition(x, y int32) bool {
	if b.x != x || b.y != y {
		b.x = x
		b.y = y
		return true
	}
	return false
}

func (b *SDL_WidgetBase) SetPositionRel(x, y int32) bool {
	if x == 0 && y == 0 {
		return false
	}
	b.x = b.x + x
	b.y = b.y + y
	return true
}

func (b *SDL_WidgetBase) GetPosition() (int32, int32) {
	return b.x, b.y
}

func (b *SDL_WidgetBase) SetLog(f func(LOG_LEVEL, string)) {
	b.log = f
}

func (b *SDL_WidgetBase) CanLog() bool {
	return b.log != nil
}

func (b *SDL_WidgetBase) Log(id LOG_LEVEL, s string) {
	if b.log != nil {
		b.log(id, s)
	}
}

func (b *SDL_WidgetBase) SetSize(w, h int32) bool {
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

func (b *SDL_WidgetBase) SetRect(r *sdl.Rect) {
	b.SetPosition(r.X, r.Y)
	b.SetSize(r.W, r.H)
}

func (b *SDL_WidgetBase) GetRect() *sdl.Rect {
	if b.IsVisible() {
		x := b.x
		if b.w < 0 {
			x = b.x + b.w
		}
		y := b.y
		if b.h < 0 {
			y = b.y + b.h
		}
		return &sdl.Rect{X: x, Y: y, W: b.w, H: b.h}
	}
	return &sdl.Rect{X: b.x, Y: b.y, W: 0, H: 0}
}

func (b *SDL_WidgetBase) GetSize() (int32, int32) {
	return b.w, b.h
}

func (b *SDL_WidgetBase) Scale(s float32) {
	b.w = int32(float32(b.w) * s)
	b.h = int32(float32(b.h) * s)
	b.x = int32(float32(b.x) * s)
	b.y = int32(float32(b.y) * s)
}

func (b *SDL_WidgetBase) SetVisible(v bool) {
	if v {
		b.state = b.state | WIDGET_STATE_VISIBLE
	} else {
		b.state = b.state & ^WIDGET_STATE_VISIBLE
	}
}

func (b *SDL_WidgetBase) IsVisible() bool {
	return (b.state & WIDGET_STATE_VISIBLE) == WIDGET_STATE_VISIBLE
}

func (b *SDL_WidgetBase) SetError(v bool) {
	if v {
		b.state = b.state & ^WIDGET_STATE_NOT_ERROR
	} else {
		b.state = b.state | WIDGET_STATE_NOT_ERROR
	}
}

func (b *SDL_WidgetBase) IsError() bool {
	return (b.state & WIDGET_STATE_NOT_ERROR) == 0
}

func (b *SDL_WidgetBase) SetClicked(v bool) {
	if v {
		b.state = b.state & ^WIDGET_STATE_NOT_CLICKED
	} else {
		b.state = b.state | WIDGET_STATE_NOT_CLICKED
	}
}

func (b *SDL_WidgetBase) IsClicked() bool {
	return (b.state & WIDGET_STATE_NOT_CLICKED) == 0
}

func (b *SDL_WidgetBase) CanFocus() bool {
	return b.canfocus
}

func (b *SDL_WidgetBase) SetFocused(v bool) {
	if b.IsEnabled() && b.CanFocus() && v {
		b.state = b.state & ^WIDGET_STATE_NOT_FOCUSED
	} else {
		b.state = b.state | WIDGET_STATE_NOT_FOCUSED
	}
}

func (b *SDL_WidgetBase) IsFocused() bool {
	if b.IsEnabled() && b.CanFocus() {
		return (b.state & WIDGET_STATE_NOT_FOCUSED) == 0
	}
	return false
}

func (b *SDL_WidgetBase) SetEnabled(e bool) {
	if e {
		b.state = b.state | WIDGET_STATE_ENABLED
	} else {
		b.state = b.state & ^WIDGET_STATE_ENABLED
	}
}

func (b *SDL_WidgetBase) IsEnabled() bool {
	return (b.state & WIDGET_STATE_ENA_SET) == WIDGET_STATE_ENA_SET
}

func (b *SDL_WidgetBase) SetDeBounce(db int) {
	b.deBounce = db
}

func (b *SDL_WidgetBase) GetDebounce() int {
	return b.deBounce
}

func (b *SDL_WidgetBase) SetForeground(c *sdl.Color) {
	b.foreground = c
}

func (b *SDL_WidgetBase) SetBackground(c *sdl.Color) {
	b.background = c
}

func (b *SDL_WidgetBase) SetBorderColour(c *sdl.Color) {
	b.borderColour = c
}

func (b *SDL_WidgetBase) SetFocusColour(c *sdl.Color) {
	b.focusColour = c
}

func (b *SDL_WidgetBase) SetDrawBackground(e bool) {
	if e {
		b.state = b.state | WIDGET_STYLE_DRAW_BG
	} else {
		b.state = b.state & ^WIDGET_STYLE_DRAW_BG
	}
}

func (b *SDL_WidgetBase) ShouldDrawBackground() bool {
	return (b.state & WIDGET_STYLE_DRAW_BG) == WIDGET_STYLE_DRAW_BG
}

func (b *SDL_WidgetBase) SetDrawBorder(e bool) {
	if e {
		b.state = b.state | WIDGET_STYLE_DRAW_BORDER
	} else {
		b.state = b.state & ^WIDGET_STYLE_DRAW_BORDER
	}
}

func (b *SDL_WidgetBase) ShouldDrawBorder() bool {
	return (b.state & WIDGET_STYLE_DRAW_BORDER) == WIDGET_STYLE_DRAW_BORDER
}

func (b *SDL_WidgetBase) GetForeground() *sdl.Color {
	if b.foreground != nil {
		return b.foreground
	}
	return GetResourceInstance().GetColour(getStateColourIndex(b.state), WIDGET_COLOUR_STYLE_FG)
}

func (b *SDL_WidgetBase) GetBackground() *sdl.Color {
	if b.background != nil {
		return b.background
	}
	return GetResourceInstance().GetColour(getStateColourIndex(b.state), WIDGET_COLOUR_STYLE_BG)
}

func (b *SDL_WidgetBase) GetBorderColour() *sdl.Color {
	if b.borderColour != nil {
		return b.borderColour
	}
	return GetResourceInstance().GetColour(getStateColourIndex(b.state), WIDGET_COLOUR_STYLE_BORDER)
}

func (b *SDL_WidgetBase) GetFocusColour() *sdl.Color {
	if b.focusColour != nil {
		return b.focusColour
	}
	return GetResourceInstance().GetColour(getStateColourIndex(b.state), WIDGET_COLOUR_STYLE_ENTRY)
}

func (b *SDL_WidgetBase) Inside(x, y int32) (SDL_Widget, bool) {
	if b.IsVisible() && isInsideRect(x, y, b.GetRect()) {
		return b.instance, true
	}
	return nil, false
}

/****************************************************************************************
* SDL_MouseData
* Used to pass mouse activity to the widgets
*
**/
type SDL_MouseData struct {
	x, y, draggingX, draggingY int32
	button                     uint8
	down                       bool
	clickCount                 int
	dragged                    bool
	dragging                   bool
	widgetId                   int32
}

func (md *SDL_MouseData) String() string {
	return fmt.Sprintf("x:%d y:%d Dx:%d Dy:%d ID:%d Btn:%d Down:%t CCount:%d Dragged:%t Dragging:%t", md.GetX(), md.GetY(), md.draggingX, md.draggingY, md.widgetId, md.button, md.IsDown(), md.GetClickCount(), md.IsDragged(), md.IsDragging())
}

func (md *SDL_MouseData) ActionStartDragging(me *sdl.MouseMotionEvent) *SDL_MouseData {
	md.dragging = true
	md.dragged = false
	md.clickCount = 1
	md.setXY(me.X, me.Y)
	return md
}

func (md *SDL_MouseData) ActionNotDragging(me *sdl.MouseMotionEvent) *SDL_MouseData {
	md.draggingX = 0
	md.draggingY = 0
	md.dragging = false
	md.widgetId = 0
	md.clickCount = 1
	md.setXY(me.X, me.Y)
	return md
}

func (md *SDL_MouseData) ActionMouseDown(me *sdl.MouseButtonEvent, id int32) *SDL_MouseData {
	if md.widgetId != id {
		md.draggingX = 0
		md.draggingY = 0
		md.dragging = false
		md.dragged = false
	}
	md.widgetId = id
	md.button = me.Button
	md.clickCount = int(me.Clicks)
	md.setXY(me.X, me.Y)
	return md
}

func (md *SDL_MouseData) ActionStopDragging(me *sdl.MouseButtonEvent) *SDL_MouseData {
	md.dragging = false
	md.dragged = true
	md.setXY(me.X, me.Y)
	return md
}

func (md *SDL_MouseData) ActionReset(me *sdl.MouseButtonEvent) *SDL_MouseData {
	md.dragging = false
	md.dragged = false
	md.draggingX = 0
	md.draggingY = 0
	md.widgetId = 0
	md.setXY(me.X, me.Y)
	return md
}

func (md *SDL_MouseData) IsDown() bool {
	return md.down
}

func (md *SDL_MouseData) GetClickCount() int {
	return md.clickCount
}

func (md *SDL_MouseData) IsDragged() bool {
	return md.dragged
}

func (md *SDL_MouseData) IsDragging() bool {
	return md.dragging
}

func (md *SDL_MouseData) GetWidgetId() int32 {
	return md.widgetId
}

func (md *SDL_MouseData) GetY() int32 {
	return md.y
}

func (md *SDL_MouseData) GetX() int32 {
	return md.x
}

func (md *SDL_MouseData) GetDraggingY() int32 {
	return md.draggingY
}

func (md *SDL_MouseData) GetDraggingX() int32 {
	return md.draggingX
}

func (md *SDL_MouseData) GetButtons() uint8 {
	return md.button
}

func (md *SDL_MouseData) setXY(x, y int32) {
	if md.dragging {
		md.draggingX = x
		md.draggingY = y
	} else {
		md.x = x
		md.y = y
	}
}

/****************************************************************************************
* Utilities
* getCachedTextWidgetEntry Returns cached texture data
*
* widgetColourDim takes a colour and returns a dimmer by same colour. Used for disabled widget text
* widgetColourBright takes a colour and returns a brighter by same colour. Used for Widget Borders
*
**/

func widgetShrinkRect(in *sdl.Rect, by int32) *sdl.Rect {
	if in == nil {
		return nil
	}
	return &sdl.Rect{X: in.X + by, Y: in.Y + by, W: in.W - (by * 2), H: in.H - (by * 2)}
}

func isInsideRect(x, y int32, r *sdl.Rect) bool {
	if x < r.X {
		return false
	}
	if y < r.Y {
		return false
	}
	if x > (r.X + r.W) {
		return false
	}
	if y > (r.Y + r.H) {
		return false
	}
	return true
}

func getStateColourIndex(state STATE_BITS) STATE_COLOUR {
	if state&WIDGET_STATE_ENA_SET == WIDGET_STATE_ENA_SET {
		if state&WIDGET_STATE_NOT_ERROR == 0 {
			return WIDGET_COLOUR_STATE_ERROR
		}
		if state&WIDGET_STATE_NOT_FOCUSED == 0 {
			return WIDGET_COLOUR_STATE_FOCUS
		}
		return WIDGET_COLOUR_STATE_ENABLED
	}
	return WIDGET_COLOUR_STATE_DISABLE
}
