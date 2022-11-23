package go_sdl_widget

import (
	"fmt"
	"sync"

	"github.com/veandco/go-sdl2/sdl"
	"github.com/veandco/go-sdl2/ttf"
)

/****************************************************************************************
* SDL_Label code
* Implements SDL_Widget cos it is one!
* Implements SDL_TextWidget because it has text and uses the texture cache
**/
type SDL_Entry struct {
	SDL_WidgetBase
	text              string
	selectedTextFlags []bool
	selecteCharsFwd   []byte
	selecteCharsRev   []byte
	textLen           int
	history           []string
	cursor            int
	cursorAtEnd       bool
	cursorTimer       int
	ctrlKeyDown       bool
	textureCache      *SDL_TextureCache
	indent            int32
	_invalid          bool
	leadin, leadout   int
	dragFrom, dragTo  int32
	selFrom, selTo    int32
	dragging          bool
	onChange          func(string, string, TEXT_CHANGE_TYPE) (string, error)
	screenData        *sdl_TextureCacheEntryRune
	screenDataLock    sync.Mutex
	keyPressLock      sync.Mutex
}

var _ SDL_Widget = (*SDL_Entry)(nil)   // Ensure SDL_Button 'is a' SDL_Widget
var _ SDL_CanFocus = (*SDL_Entry)(nil) // Ensure SDL_Button 'is a' SDL_Widget

func NewSDLEntry(x, y, w, h, id int32, text string, style STATE_BITS, onChange func(string, string, TEXT_CHANGE_TYPE) (string, error)) *SDL_Entry {
	ent := &SDL_Entry{text: text, textLen: len(text), textureCache: nil, cursor: 0, cursorTimer: 0, leadin: 0, leadout: 0, ctrlKeyDown: false, _invalid: true, indent: 10, onChange: onChange}
	ent.ClearSelection()
	ent.SetSelecteCharsFwd(GetResourceInstance().GetSelecteCharsFwd())
	ent.SetSelecteCharsRev(GetResourceInstance().GetSelecteCharsRev())
	ent.SDL_WidgetBase = initBase(x, y, w, h, id, 0, style)
	return ent
}

func (b *SDL_Entry) SetTextureCache(tc *SDL_TextureCache) {
	b.textureCache = tc
}

func (b *SDL_Entry) GetTextureCache() *SDL_TextureCache {
	return b.textureCache
}

func (b *SDL_Entry) SetSelecteCharsFwd(s string) {
	b.selecteCharsFwd = []byte(s)
}

func (b *SDL_Entry) GetSelecteCharsFwd() string {
	return string(b.selecteCharsFwd)
}

func (b *SDL_Entry) findSelecteCharsFwd(c []byte) int {
	cur := b.cursor
	if cur >= len(c) && cur > 0 {
		cur = len(c) - 1
	}
	for i := b.cursor; i < len(c); i++ {
		for _, sc := range b.selecteCharsFwd {
			if sc == c[i] {
				return i
			}
		}
	}
	return len(c)
}

func (b *SDL_Entry) findSelecteCharsRev(c []byte) int {
	cur := b.cursor
	if cur >= len(c) && cur > 0 {
		cur = len(c) - 1
	}
	for i := cur; i >= 0; i-- {
		for _, sc := range b.selecteCharsRev {
			if sc == c[i] {
				if i > 0 {
					return i + 1
				}
				return i
			}
		}
	}
	return 0
}

func (b *SDL_Entry) SetSelecteCharsRev(s string) {
	b.selecteCharsRev = []byte(s)
}

func (b *SDL_Entry) GetSelecteCharsRev() string {
	return string(b.selecteCharsRev)
}

func (b *SDL_Entry) SetForeground(c *sdl.Color) {
	b.SDL_WidgetBase.SetForeground(c)
	b.Invalid(true)
}

func (b *SDL_Entry) pushHistory(val string) {
	if len(b.history) > 0 {
		if (b.history)[len(b.history)-1] == val {
			return
		}
	}
	b.history = append(b.history, val)
}

func (b *SDL_Entry) SetText(text string) {
	if b.text != text {
		b.setText(text)
	}
}

func (b *SDL_Entry) setText(text string) {
	b.keyPressLock.Lock()
	defer b.keyPressLock.Unlock()
	b.text = text
	b.textLen = len(b.text)
	b.ClearSelection()
	b.Invalid(true)
}

func (b *SDL_Entry) insertAtCursor(text string) string {
	if b.cursor < b.textLen {
		return fmt.Sprintf("%s%s%s", b.text[0:b.cursor], text, b.text[b.cursor:])
	} else {
		return fmt.Sprintf("%s%s", b.text, text)
	}
}

func (b *SDL_Entry) SetFocused(focus bool) {
	b.SDL_WidgetBase.SetFocused(focus)
	if !focus {
		b.ClearSelection()
	}
	b.Invalid(true)
}

func (b *SDL_Entry) KeyPress(c int, ctrl bool, down bool) bool {
	if b.IsEnabled() && b.IsFocused() {
		b.keyPressLock.Lock()
		defer b.keyPressLock.Unlock()
		oldValue := b.text
		newValue := b.text
		onChangeType := TEXT_CHANGE_NONE
		insertLen := 1
		saveHistory := true
		if ctrl {
			// if ctrl key then just remember its state (up or down) and return
			if c == sdl.K_LCTRL || c == sdl.K_RCTRL {
				b.ctrlKeyDown = down
				return true
			}
			// if the control key is down then it is a control sequence like CTRL-Z
			if b.ctrlKeyDown {
				b.ctrlKeyDown = false // Stop repeat keys - Ctrl key must be released and pressed again
				switch c {
				case sdl.K_z:
					if len(b.history) > 0 {
						newValue = (b.history)[len(b.history)-1]
						b.history = (b.history)[0 : len(b.history)-1]
						saveHistory = false
					}
				case sdl.K_c:
					sdl.SetClipboardText(b.GetSelectedText())
					return true
				case sdl.K_v:
					s, err := sdl.GetClipboardText()
					if err == nil {
						newValue = b.insertAtCursor(s)
						onChangeType = TEXT_CHANGE_INSERT
						insertLen = len(s)
					}
				}
			} else {
				if down {
					if c < 32 || c == 127 {
						switch c {
						case sdl.K_DELETE:
							if b.cursor < b.textLen {
								newValue = fmt.Sprintf("%s%s", oldValue[0:b.cursor], oldValue[b.cursor+1:])
								onChangeType = TEXT_CHANGE_DELETE
							}
						case sdl.K_BACKSPACE:
							if b.cursor > 0 {
								if b.cursor < b.textLen {
									newValue = fmt.Sprintf("%s%s", oldValue[0:b.cursor-1], oldValue[b.cursor:])
								} else {
									newValue = oldValue[0 : b.textLen-1]
								}
								onChangeType = TEXT_CHANGE_BS
							}
						case sdl.K_RETURN:
							if b.onChange != nil {
								b.onChange("", b.text, TEXT_CHANGE_FINISH)
							}
						default:
							return false
						}
					} else {
						switch c | 0x40000000 {
						case sdl.K_RIGHT:
							b.MoveCursor(1)
						case sdl.K_UP:
							b.SetCursor(99)
						case sdl.K_DOWN:
							b.SetCursor(0)
						case sdl.K_LEFT:
							b.MoveCursor(-1)
						default:
							return false
						}
					}
				} else {
					// If it is NOT down then we ignore an
					return false
				}
			}
			// If it is NOT a ctrl key or a control sequence then we only react on a DOWN

		} else {
			// not a control key. insert it at the cursor
			newValue = b.insertAtCursor(fmt.Sprintf("%c", c))
			onChangeType = TEXT_CHANGE_INSERT
		}
		if oldValue != newValue && b.onChange != nil {
			var err error
			newValue, err = b.onChange(oldValue, newValue, onChangeType)
			b.SetError(err != nil)
		}
		if newValue != oldValue {
			if saveHistory {
				b.pushHistory(oldValue)
			}
			b.text = newValue
			b.textLen = len(b.text)
			switch onChangeType {
			case TEXT_CHANGE_INSERT:
				b.MoveCursor(insertLen)
			case TEXT_CHANGE_BS:
				b.MoveCursor(-insertLen)
			}
			b.Invalid(true)
			b.ClearSelection()
			return true
		}
	}
	return false
}

func (b *SDL_Entry) SetCursor(i int) {
	b.screenDataLock.Lock()
	defer b.screenDataLock.Unlock()
	b.setCursorNoLock(i)
}

func (b *SDL_Entry) setCursorNoLock(i int) {
	if b.IsFocused() {
		li := b.leadin
		lo := b.leadout
		span := lo - li
		if b.CanLog() {
			b.Log(1, fmt.Sprintf("setC i:%d span:%d li:%d lo:%d end:%t\n", i, span, li, lo, b.cursorAtEnd))
		}
	}
}

func (b *SDL_Entry) MoveCursor(i int) {
	b.screenDataLock.Lock()
	defer b.screenDataLock.Unlock()
	b.setCursorNoLock(b.cursor + i)
}

func (b *SDL_Entry) SetEnabled(e bool) {
	if b.IsEnabled() != e {
		b.SDL_WidgetBase.SetEnabled(e)
		b.Invalid(true)
	}
}

func (b *SDL_Entry) GetText() string {
	return b.text
}

func (b *SDL_Entry) GetSelectedText() string {
	return b.screenData.GetSelectedText()
}

func (b *SDL_Entry) ClearSelection() {
	b.screenDataLock.Lock()
	defer b.screenDataLock.Unlock()
	b.screenData.SetAlSelected(false)
}

func (b *SDL_Entry) selectAtCursor(clicks int) bool {
	b.screenDataLock.Lock()
	defer b.screenDataLock.Unlock()
	switch clicks {
	case 0, 1:
		return false
	case 2:
		flags := make([]bool, len(b.text))
		c := []byte(b.text)
		from := b.findSelecteCharsRev(c)
		too := b.findSelecteCharsFwd(c)
		for i := from; i < too; i++ {
			flags[i] = true
		}
		b.selectedTextFlags = flags
	case 3:
		flags := make([]bool, len(b.text))
		for i := 0; i < len(b.selectedTextFlags); i++ {
			flags[i] = true
		}
		b.selectedTextFlags = flags
	}
	return true
}

func (b *SDL_Entry) Click(md *SDL_MouseData) bool {
	if b.IsEnabled() {
		b.keyPressLock.Lock()
		defer b.keyPressLock.Unlock()

		if md.IsDragging() {
			if !b.dragging {
				b.dragFrom = md.draggingToX
				b.dragTo = md.draggingToX
				b.dragging = true
			} else {
				b.dragTo = md.draggingToX
			}
			return true
		} else {
			b.dragging = false
		}

		if md.IsDragged() {
			go func() {
				b.screenDataLock.Lock()
				defer b.screenDataLock.Unlock()
				b.selTo = b.dragTo
				b.selFrom = b.dragFrom
				oldSel := b.screenData.GetSelectedText()
				sel := false

				if b.selTo < b.selFrom {
					temp := b.selFrom
					b.selFrom = b.selTo
					b.selTo = temp
				}
				linked := b.screenData
				if linked == nil {
					return
				}
				for linked != nil {
					if linked.Inside(b.selFrom) {
						b.selFrom = linked.offset
						sel = true
					}
					linked.selected = sel
					if linked.Inside(b.selTo) {
						b.selTo = linked.offset + linked.width
						sel = false
					}
					linked = linked.next
				}
				b.selTo = 0
				b.selFrom = 0

				newSel := b.screenData.GetSelectedText()
				if newSel != oldSel {
					s, _ := b.onChange(oldSel, newSel, TEXT_CHANGE_SELECTED)
					if s == oldSel {
						fmt.Println("OLDSEL")
						// To Do restore old selection
					}
				}
			}()
			return true
		}
		if md.GetClickCount() > 1 {
			return b.selectAtCursor(md.GetClickCount())
		}
		go func() {
			b.screenDataLock.Lock()
			defer b.screenDataLock.Unlock()
			linked := b.screenData
			if linked == nil {
				return
			}

			notFound := true
			for linked != nil {
				if linked.Inside(md.x) {
					b.setCursorNoLock(linked.pos)
					notFound = false
					break
				}
				linked = linked.next
			}
			if notFound {
				b.setCursorNoLock(b.textLen)
			}
		}()
	}
	return false
}

func (b *SDL_Entry) Draw(renderer *sdl.Renderer, font *ttf.Font) error {
	if b.IsVisible() {
		b.screenDataLock.Lock()
		defer b.screenDataLock.Unlock()

		var err error

		tx := b.x + b.indent
		th := int32(float32(b.h) - float32(b.h)/4)
		ty := int32((float32(b.h - th)) / 2)

		if b._invalid {
			b.Invalid(false)
			err = GetResourceInstance().UpdateTextureCachedRunes(renderer, font, b.GetForeground(), b.text)
			if err != nil {
				renderer.SetDrawColor(255, 0, 0, 255)
				renderer.DrawRect(&sdl.Rect{X: b.x, Y: b.y, W: b.w, H: b.h})
				return nil
			}
			sd := GetResourceInstance().GetScaledTextureListFromCachedRunesLinked(b.text, b.GetForeground(), tx, int32(th))
			if sd == nil {
				if err != nil {
					renderer.SetDrawColor(255, 0, 0, 255)
					renderer.DrawRect(&sdl.Rect{X: b.x, Y: b.y, W: b.w, H: b.h})
					return nil
				}
			}
			b.screenData = sd
		}

		linked := b.screenData
		// *******************************************************
		// Find the number of chars thet can be displayed 'cc'
		if b.leadin >= b.textLen { // Ensure leadin is not past the end
			b.leadin = b.textLen - 1
		}
		if b.leadin < 0 { // Ensure leadin is not negative
			b.leadin = 0
		}

		linked.SetAllVisible(false)
		leadIn := linked.Indexed(b.leadin)
		if leadIn == nil {
			return nil
		}

		//*********************************************************
		if b.ShouldDrawBackground() {
			bc := b.GetBackground()
			renderer.SetDrawColor(bc.R, bc.G, bc.B, bc.A)
			renderer.FillRect(&sdl.Rect{X: b.x, Y: b.y, W: b.w, H: b.h})
		}
		if b.dragging && b.IsFocused() {
			renderer.SetDrawColor(100, 100, 0, 25)
			if b.dragFrom > b.dragTo {
				renderer.FillRect(&sdl.Rect{X: b.dragTo, Y: b.y + 1, W: b.dragFrom - b.dragTo, H: b.h - 2})
			} else {
				renderer.FillRect(&sdl.Rect{X: b.dragFrom, Y: b.y + 1, W: b.dragTo - b.dragFrom, H: b.h - 2})
			}
		}
		//
		// Scale the text to fit the height but keep the aspect ration the same so we know the width of each char
		//   Need to use floats to prevent rounding
		//

		paintCursor := b.IsEnabled() && b.IsFocused() && (sdl.GetTicks64()%1000) > 300

		var rect *sdl.Rect

		tx = b.x + b.indent

		max := b.x + b.w
		last := 0
		disp := leadIn
		for disp != nil && tx+disp.width < max {
			last = disp.pos
			tw := float32(disp.width) * (float32(b.h) / float32(th))
			rect = &sdl.Rect{X: tx, Y: b.y + ty, W: int32(tw), H: th}
			disp.SetVisible(true)
			if disp.selected {
				renderer.SetDrawColor(100, 100, 100, 25)
				renderer.FillRect(rect)
			}
			renderer.Copy(disp.te.Texture, nil, rect)
			if !b.cursorAtEnd && paintCursor && disp.pos == b.cursor {
				c := GetResourceInstance().GetCursorInsertColour()
				renderer.SetDrawColor(c.R, c.G, c.B, c.A)
				renderer.FillRect(&sdl.Rect{X: tx, Y: b.y, W: 5, H: b.h})
			}
			tx = tx + int32(tw)
			disp = disp.next
		}
		b.leadout = last
		if b.cursorAtEnd && paintCursor && tx < max {
			c := GetResourceInstance().GetCursorAppendColour()
			renderer.SetDrawColor(c.R, c.G, c.B, c.A)
			renderer.FillRect(&sdl.Rect{X: tx, Y: b.y, W: 5, H: b.h})
		}
		if b.ShouldDrawBorder() {
			bc := b.GetBorderColour()
			renderer.SetDrawColor(bc.R, bc.G, bc.B, bc.A)
			renderer.DrawRect(&sdl.Rect{X: b.x + 1, Y: b.y + 1, W: b.w - 2, H: b.h - 2})
		}
	}
	return nil
}
func (b *SDL_Entry) Destroy() {
	// Image cache takes care of all images!
}

func (b *SDL_Entry) Invalid(yes bool) {
	b._invalid = yes
}
