package go_sdl_widget

import (
	"fmt"
	"math"

	"github.com/veandco/go-sdl2/sdl"
	"github.com/veandco/go-sdl2/ttf"
)

type ALIGN_TEXT int
type ROTATE_SHAPE_90 int
type KBD_KEY_MODE int
type TEXT_CHANGE_TYPE int
type STATE_BITS uint16

const (
	ALIGN_CENTER ALIGN_TEXT = iota
	ALIGN_LEFT
	ALIGN_RIGHT
	ALIGN_FIT

	ROTATE_0 ROTATE_SHAPE_90 = iota
	ROTATE_90
	ROTATE_180
	ROTATE_270

	TEXT_CHANGE_INSERT TEXT_CHANGE_TYPE = iota
	TEXT_CHANGE_DELETE
	TEXT_CHANGE_BS
	TEXT_CHANGE_FINISH
	TEXT_CHANGE_SELECTED
	TEXT_CHANGE_NONE

	WIDGET_STYLE_NONE          STATE_BITS = 0b0000000000000001
	WIDGET_STYLE_DRAW_BORDER   STATE_BITS = 0b0000000000000010
	WIDGET_STYLE_DRAW_BG       STATE_BITS = 0b0000000000001000
	WIDGET_STYLE_BORDER_AND_BG STATE_BITS = WIDGET_STYLE_DRAW_BORDER | WIDGET_STYLE_DRAW_BG
	WIDGET_STATE_ENABLED       STATE_BITS = 0b0000000000010000
	WIDGET_STATE_VISIBLE       STATE_BITS = 0b0000000000100000
	WIDGET_STATE_NOT_FOCUSED   STATE_BITS = 0b0000000001000000
	WIDGET_STATE_NOT_ERROR     STATE_BITS = 0b0000000010000000
	WIDGET_STATE_NOT_CLICKED   STATE_BITS = 0b0000000100000000
	WIDGET_STATE_STA_BITS      STATE_BITS = 0b0000000111110000 // Clear state AND mask. Retains style.
	WIDGET_STATE_ENA_BITS      STATE_BITS = 0b0000000000001111 // Clear style AND mask. Retains state.
	WIDGET_STATE_ENA_SET       STATE_BITS = 0b0000000100110000 // Enabled visible and not-clicked

	WIDGET_COLOR_FG     int = 0 // Section indexes
	WIDGET_COLOR_BG     int = 1
	WIDGET_COLOR_BORDER int = 2
	WIDGET_COLOR_ENTRY  int = 3
	WIDGET_COLOR_MAX    int = 4

	WIDGET_COLOUR_ENABLED = 0
	WIDGET_COLOUR_DISABLE = 1
	WIDGET_COLOUR_FOCUS   = 2
	WIDGET_COLOUR_ERROR   = 3

	DEG_TO_RAD float64 = (math.Pi / 180)
)

var TEXTURE_CACHE_TEXT_PREF = "TxCaPr987"

type SDL_Widget interface {
	Draw(*sdl.Renderer, *ttf.Font) error
	Scale(float32)
	Click(*SDL_MouseData) bool
	Inside(int32, int32) bool      // Base
	GetRect() *sdl.Rect            // Base
	SetWidgetId(int32)             // Base
	GetWidgetId() int32            // Base
	SetVisible(bool)               // Base
	IsVisible() bool               // Base
	SetEnabled(bool)               // Base
	IsEnabled() bool               // Base
	SetError(bool)                 // Base
	IsError() bool                 // Base
	SetFocused(bool)               // Base
	IsFocused() bool               // Base
	SetClicked(bool)               // Base
	IsClicked() bool               // Base
	SetPosition(int32, int32) bool // Base
	GetPosition() (int32, int32)   // Base
	SetSize(int32, int32) bool     // Base
	GetSize() (int32, int32)       // Base

	GetForeground() *sdl.Color   // Base
	GetBackground() *sdl.Color   // Base
	GetBorderColour() *sdl.Color // Base
	GetEntryColour() *sdl.Color  // Base

	SetForeground(*sdl.Color)   // Base
	SetBackground(*sdl.Color)   // Base
	SetBorderColour(*sdl.Color) // Base
	SetEntryColour(*sdl.Color)  // Base

	SetDrawBackground(bool)
	ShouldDrawBackground() bool
	SetDrawBorder(bool)
	ShouldDrawBorder() bool

	Destroy() // Base
}

type SDL_CanFocus interface {
	SetFocused(bool) // Base
	IsFocused() bool // Base
	KeyPress(int, bool, bool) bool
	ClearSelection()
	GetSelectedText() string
}

type SDL_TextWidget interface {
	SetTextureCache(*SDL_TextureCache)
	GetTextureCache() *SDL_TextureCache
	SetText(text string)
	GetText() string
	GetWidgetId() int32
	IsEnabled() bool
	GetForeground() *sdl.Color
}

type SDL_ImageWidget interface {
	SetFrame(tf int32)
	GetFrame() int32
	NextFrame() int32
	GetFrameCount() int32
}

type SDL_TextureCacheWidget interface {
	SetTextureCache(*SDL_TextureCache)
	GetTextureCache() *SDL_TextureCache
}

type SDL_MouseData struct {
	x, y, draggingToX, draggingToY int32
	button                         uint8
	down                           bool
	dragged                        bool
	dragging                       bool
	widgetId                       int32
}

type SDL_WidgetBase struct {
	x, y, w, h   int32
	widgetId     int32
	deBounce     int
	background   *sdl.Color
	foreground   *sdl.Color
	borderColour *sdl.Color
	entryColour  *sdl.Color
	state        STATE_BITS
}

/****************************************************************************************
* Common (base) functions for ALL SDL_Widget instances
**/
func initBase(x, y, w, h, widgetId int32, deBounce int, style STATE_BITS) SDL_WidgetBase {
	return SDL_WidgetBase{
		x:        x,
		y:        y,
		w:        w,
		h:        h,
		widgetId: widgetId,

		deBounce:     deBounce,
		background:   nil,
		foreground:   nil,
		borderColour: nil,
		entryColour:  nil,
		state:        style | WIDGET_STATE_STA_BITS, // Clear state bits and set enabled, visible and notpressed. Leave style unchanged
	}
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

func (b *SDL_WidgetBase) GetPosition() (int32, int32) {
	return b.x, b.y
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

func (b *SDL_WidgetBase) GetRect() *sdl.Rect {
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

func (b *SDL_WidgetBase) SetFocused(v bool) {
	if b.IsEnabled() && v {
		b.state = b.state & ^WIDGET_STATE_NOT_FOCUSED
	} else {
		b.state = b.state | WIDGET_STATE_NOT_FOCUSED
	}
}

func (b *SDL_WidgetBase) IsFocused() bool {
	return (b.state & WIDGET_STATE_NOT_FOCUSED) == 0
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

func (b *SDL_WidgetBase) SetEntryColour(c *sdl.Color) {
	b.entryColour = c
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
	return GetResourceInstance().GetColour(getStateColourIndex(b.state), WIDGET_COLOR_FG)
}

func (b *SDL_WidgetBase) GetBackground() *sdl.Color {
	if b.background != nil {
		return b.background
	}
	return GetResourceInstance().GetColour(getStateColourIndex(b.state), WIDGET_COLOR_BG)
}

func (b *SDL_WidgetBase) GetBorderColour() *sdl.Color {
	if b.borderColour != nil {
		return b.borderColour
	}
	return GetResourceInstance().GetColour(getStateColourIndex(b.state), WIDGET_COLOR_BORDER)
}

func (b *SDL_WidgetBase) GetEntryColour() *sdl.Color {
	if b.entryColour != nil {
		return b.entryColour
	}
	return GetResourceInstance().GetColour(getStateColourIndex(b.state), WIDGET_COLOR_ENTRY)
}

func (b *SDL_WidgetBase) Inside(x, y int32) bool {
	if b.IsVisible() {
		return isInsideRect(x, y, b.GetRect())
	}
	return false
}

/****************************************************************************************
* Container for SDL_Widgets. A list of lists
**/
type SDL_WidgetGroup struct {
	wigetLists []*SDL_WidgetSubGroup
}

func NewWidgetGroup(font *ttf.Font) *SDL_WidgetGroup {
	return &SDL_WidgetGroup{wigetLists: make([]*SDL_WidgetSubGroup, 0)}
}

func (wg *SDL_WidgetGroup) NewWidgetSubGroup(font *ttf.Font, id int32) *SDL_WidgetSubGroup {
	if font == nil {
		font = GetResourceInstance().GetFont()
	}
	sg := &SDL_WidgetSubGroup{list: make([]*SDL_Widget, 0), font: font, id: id}
	wg.wigetLists = append(wg.wigetLists, sg)
	return sg
}

func (wg *SDL_WidgetGroup) AllWidgets() []*SDL_Widget {
	l := make([]*SDL_Widget, 0)
	for _, wList := range wg.wigetLists {
		l = append(l, wList.ListWidgets()...)
	}
	return l
}

func (wg *SDL_WidgetGroup) SetFocused(id int32) {
	for _, wList := range wg.wigetLists {
		wList.SetFocused(id)
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

func (wg *SDL_WidgetGroup) GetFocused() SDL_CanFocus {
	for _, wList := range wg.wigetLists {
		f := wList.GetFocused()
		if f != nil {
			return f
		}
	}
	return nil
}

func (wl *SDL_WidgetGroup) GetWidgetSubGroup(id int32) *SDL_WidgetSubGroup {
	for _, w := range wl.wigetLists {
		if (*w).GetId() == id {
			return w
		}
	}
	return nil
}

func (wl *SDL_WidgetGroup) GetWidget(id int) *SDL_Widget {
	for _, w := range wl.wigetLists {
		wi := w.GetWidget(id)
		if wi != nil {
			return wi
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

func (wg *SDL_WidgetGroup) Destroy() {
	for _, w := range wg.wigetLists {
		w.Destroy()
	}
	GetResourceInstance().GetTextureCache().Destroy()
}

func (wg *SDL_WidgetGroup) Draw(renderer *sdl.Renderer) {
	for _, w := range wg.wigetLists {
		w.Draw(renderer)
	}
}

func (wg *SDL_WidgetGroup) Inside(x, y int32) SDL_Widget {
	for _, wl := range wg.wigetLists {
		w := wl.Inside(x, y)
		if w != nil {
			return w
		}
	}
	return nil
}

/****************************************************************************************
* Container for SDL_Widget instances.
**/
type SDL_WidgetSubGroup struct {
	id   int32
	list []*SDL_Widget
	font *ttf.Font
}

func (wl *SDL_WidgetSubGroup) GetId() int32 {
	return wl.id
}

func (wl *SDL_WidgetSubGroup) Add(widget SDL_Widget) {
	wl.list = append(wl.list, &widget)
}

func (wl *SDL_WidgetSubGroup) Inside(x, y int32) SDL_Widget {
	for _, w := range wl.list {
		if (*w).Inside(x, y) {
			return (*w)
		}
	}
	return nil
}

func (wl *SDL_WidgetSubGroup) ListWidgets() []*SDL_Widget {
	return wl.list
}

func (wl *SDL_WidgetSubGroup) GetWidget(id int) *SDL_Widget {
	for _, w := range wl.list {
		if int((*w).GetWidgetId()) == id {
			return w
		}
	}
	return nil
}

func (wl *SDL_WidgetSubGroup) SetFocused(id int32) {
	for _, w := range wl.list {
		f, ok := (*w).(SDL_CanFocus)
		if ok {
			f.SetFocused((*w).GetWidgetId() == id)
		}
	}
}

func (wl *SDL_WidgetSubGroup) ClearFocus() {
	for _, w := range wl.list {
		f, ok := (*w).(SDL_CanFocus)
		if ok {
			f.SetFocused(false)
		}
	}
}

func (wl *SDL_WidgetSubGroup) ClearSelection() {
	for _, w := range wl.list {
		f, ok := (*w).(SDL_CanFocus)
		if ok {
			f.ClearSelection()
		}
	}
}
func (wl *SDL_WidgetSubGroup) GetFocused() SDL_CanFocus {
	for _, w := range wl.list {
		f, ok := (*w).(SDL_CanFocus)
		if ok {
			if f.IsFocused() {
				return f
			}
		}
	}
	return nil
}

func (wl *SDL_WidgetSubGroup) KeyPress(c int, ctrl, down bool) bool {
	for _, w := range wl.list {
		f, ok := (*w).(SDL_CanFocus)
		if ok {
			if f.IsFocused() {
				if f.KeyPress(c, ctrl, down) {
					return true
				}
			}
		}
	}
	return false
}

func (wl *SDL_WidgetSubGroup) ArrangeLR(xx, yy, padding int32) (int32, int32) {
	x := xx
	y := yy
	var width int32
	var w *SDL_Widget
	for _, w = range wl.list {
		if (*w).IsVisible() {
			(*w).SetPosition(x, y)
			width, _ = (*w).GetSize()
			x = x + width + padding
		}
	}
	return x, y
}

func (wl *SDL_WidgetSubGroup) ArrangeRL(xx, yy, padding int32) (int32, int32) {
	x := xx
	y := yy
	var width int32
	for _, w := range wl.list {
		if (*w).IsVisible() {
			width, _ = (*w).GetSize()
			(*w).SetPosition(x-width, y)
			x = (x - width) - padding
		}
	}
	return x, y
}

func (wl *SDL_WidgetSubGroup) SetEnable(e bool) {
	for _, w := range wl.list {
		(*w).SetEnabled(e)
	}
}

func (wl *SDL_WidgetSubGroup) SetVisible(e bool) {
	for _, w := range wl.list {
		(*w).SetVisible(e)
	}
}

func (wl *SDL_WidgetSubGroup) Draw(renderer *sdl.Renderer) {
	for _, w := range wl.list {
		(*w).Draw(renderer, wl.font)
	}
}

func (wl *SDL_WidgetSubGroup) SetFont(font *ttf.Font) {
	wl.font = font
}

func (wl *SDL_WidgetSubGroup) GetFont() *ttf.Font {
	return wl.font
}

func (wl *SDL_WidgetSubGroup) Scale(s float32) {
	for _, w := range wl.list {
		(*w).Scale(s)
	}
}

func (wl *SDL_WidgetSubGroup) Destroy() {
	for _, w := range wl.list {
		(*w).Destroy()
	}
}

func (md *SDL_MouseData) String() string {
	return fmt.Sprintf("x:%d y:%d Dx:%d Dy:%d ID:%d Btn:%d Down:%t Dragged:%t Dragging:%t", md.GetX(), md.GetY(), md.draggingToX, md.draggingToY, md.widgetId, md.button, md.IsDown(), md.IsDragged(), md.IsDragging())
}

func (md *SDL_MouseData) IsDown() bool {
	return md.down
}

func (md *SDL_MouseData) IsDragged() bool {
	return md.dragged
}

func (md *SDL_MouseData) IsDragging() bool {
	return md.dragging
}

func (md *SDL_MouseData) SetDragged(d bool) {
	md.dragging = false
	md.dragged = d
}

func (md *SDL_MouseData) SetDragging(d bool) {
	if d {
		md.dragging = true
	} else {
		md.draggingToX = 0
		md.draggingToY = 0
		md.dragging = false
		md.widgetId = 0
	}
	md.dragged = false
}

func (md *SDL_MouseData) GetWidgetId() int32 {
	return md.widgetId
}

func (md *SDL_MouseData) SetWidgetId(id int32) {
	if md.widgetId != id {
		md.draggingToX = 0
		md.draggingToY = 0
		md.dragging = false
		md.dragged = false
	}
	md.widgetId = id
}

func (md *SDL_MouseData) GetY() int32 {
	return md.y
}

func (md *SDL_MouseData) GetX() int32 {
	return md.x
}

func (md *SDL_MouseData) GetButtons() uint8 {
	return md.button
}

func (md *SDL_MouseData) SetButtons(b uint8) {
	md.button = b
}

func (md *SDL_MouseData) SetXY(x, y int32) {
	if md.dragging {
		md.draggingToX = x
		md.draggingToY = y
	} else {
		md.x = x
		md.y = y
	}
	md.down = true
	md.dragged = false
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

func getStateColourIndex(state STATE_BITS) int {
	if state&WIDGET_STATE_ENA_SET == WIDGET_STATE_ENA_SET {
		if state&WIDGET_STATE_NOT_ERROR == 0 {
			return WIDGET_COLOUR_ERROR
		}
		if state&WIDGET_STATE_NOT_FOCUSED == 0 {
			return WIDGET_COLOUR_FOCUS
		}
		return WIDGET_COLOUR_ENABLED
	}
	return WIDGET_COLOUR_DISABLE
}
