package go_sdl_widget

import (
	"fmt"
	"strings"
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
	text             string
	textLen          int
	history          []string
	cursor           int
	cursorAtEnd      bool
	cursorTimer      int
	selecteCharsFwd  []byte
	selecteCharsRev  []byte
	selectCharFrom   int
	selectCharToo    int
	ctrlKeyDown      bool
	indent           int32
	_invalid         bool
	leadin, leadout  int
	dragFrom, dragTo int32
	dragging         bool
	onChange         func(string, string, ENTRY_EVENT_TYPE) (string, error)
	screenData       *sdl_TextureCacheEntryRune
	screenDataLock   sync.Mutex
	keyPressLock     sync.Mutex
}

var _ SDL_Widget = (*SDL_Entry)(nil)        // Ensure SDL_Button 'is a' SDL_Widget
var _ SDL_CanSelectText = (*SDL_Entry)(nil) // Ensure SDL_Button 'is a' SDL_Widget

func NewSDLEntry(x, y, w, h, id int32, text string, style STATE_BITS, onChange func(string, string, ENTRY_EVENT_TYPE) (string, error)) *SDL_Entry {
	ent := &SDL_Entry{text: text, textLen: len(text), cursor: 0, cursorTimer: 0, leadin: 0, leadout: 0, ctrlKeyDown: false, _invalid: true, indent: 10, onChange: onChange}
	ent.ClearSelection()
	ent.SetSelecteCharsFwd(GetResourceInstance().GetSelectCharsFwd())
	ent.SetSelecteCharsRev(GetResourceInstance().GetSelectCharsRev())
	ent.SDL_WidgetBase = initBase(x, y, w, h, id, ent, 0, true, style, nil)
	return ent
}

func (b *SDL_Entry) SetSelecteCharsFwd(s string) {
	b.selecteCharsFwd = []byte(s)
}

func (b *SDL_Entry) GetSelecteCharsFwd() string {
	return string(b.selecteCharsFwd)
}

func (b *SDL_Entry) SetSelecteCharsRev(s string) {
	b.selecteCharsRev = []byte(s)
}

func (b *SDL_Entry) GetSelecteCharsRev() string {
	return string(b.selecteCharsRev)
}

func (b *SDL_Entry) String() string {
	return b.text
}

func (b *SDL_Entry) SetText(text string) {
	if b.text != text {
		b.screenDataLock.Lock()
		defer b.screenDataLock.Unlock()
		b.text = text
		b.textLen = len(text)
		b.ClearSelection()
		b.Invalid(true)
	}
}

func (b *SDL_Entry) findSelecteCharsFwd() uint {
	c := []byte(b.text)
	cur := b.cursor
	for i := cur; i < len(c); i++ {
		for _, sc := range b.selecteCharsFwd {
			if sc == c[i] {
				if i > 0 {
					return uint(i - 1)
				}
				return 0
			}
		}
	}
	return uint(len(c))
}

func (b *SDL_Entry) findSelecteCharsRev() uint {
	c := []byte(b.text)
	cur := b.cursor
	if cur >= len(c) {
		cur = len(c) - 1
	}
	for i := cur; i > 0; i-- {
		for _, sc := range b.selecteCharsRev {
			if sc == c[i] {
				if i < len(c) {
					return uint(i + 1)
				}
				return uint(i)
			}
		}
	}
	return 0
}

func (b *SDL_Entry) SetFocused(focus bool) {
	if b.onChange != nil {
		if focus {
			go b.onChange(b.text, b.text, ENTRY_EVENT_FOCUS)
		} else {
			go b.onChange(b.text, b.text, ENTRY_EVENT_UN_FOCUS)
		}
	}
	b.screenDataLock.Lock()
	defer b.screenDataLock.Unlock()
	b.SDL_WidgetBase.SetFocused(focus)
	b.ClearSelection()
	b.Invalid(true)
}

func (b *SDL_Entry) KeyPress(c int, ctrl bool, down bool) bool {
	if b.IsEnabled() && b.IsFocused() {
		b.keyPressLock.Lock()
		defer b.keyPressLock.Unlock()
		oldValue := b.text
		newValue := b.text
		onChangeType := ENTRY_EVENT_NONE
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
						onChangeType = ENTRY_EVENT_INSERT
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
								onChangeType = ENTRY_EVENT_DELETE
							}
						case sdl.K_BACKSPACE:
							if b.cursor > 0 {
								if b.cursor < b.textLen {
									newValue = fmt.Sprintf("%s%s", oldValue[0:b.cursor-1], oldValue[b.cursor:])
								} else {
									newValue = oldValue[0 : b.textLen-1]
								}
								onChangeType = ENTRY_EVENT_BS
							}
						case sdl.K_RETURN:
							if b.onChange != nil {
								b.onChange("", b.text, ENTRY_EVENT_FINISH)
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
			onChangeType = ENTRY_EVENT_INSERT
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
			b.SetText(newValue)
			switch onChangeType {
			case ENTRY_EVENT_INSERT:
				b.MoveCursor(insertLen)
			case ENTRY_EVENT_BS:
				b.MoveCursor(-insertLen)
			}
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

func (b *SDL_Entry) pushHistory(val string) {
	if len(b.history) > 0 {
		if (b.history)[len(b.history)-1] == val {
			return
		}
	}
	b.history = append(b.history, val)
}

func (b *SDL_Entry) setCursorNoLock(i int) {
	if b.IsFocused() {
		if i < 0 {
			i = 0
		}
		if i >= b.textLen {
			i = b.textLen
			b.cursorAtEnd = true
		} else {
			b.cursorAtEnd = false
		}

		li := b.leadin
		lo := b.leadout
		span := lo - li

		if i > lo && !b.cursorAtEnd {
			li = li + 1
		}
		if i < li {
			li = li - 1
		}
		b.leadin = li
		b.leadout = li + span
		b.cursor = i
	}
}

func (b *SDL_Entry) MoveCursor(i int) {
	b.screenDataLock.Lock()
	defer b.screenDataLock.Unlock()
	b.setCursorNoLock(b.cursor + i)
}

func (b *SDL_Entry) GetText() string {
	return b.text
}

func (b *SDL_Entry) GetSelectedText() string {
	if b.selectCharFrom < 0 || b.selectCharToo < 0 {
		return ""
	}
	if b.selectCharFrom > b.selectCharToo {
		return ""
	}
	var sb strings.Builder
	for i, c := range b.text {
		if i >= b.selectCharFrom && i <= b.selectCharToo {
			sb.WriteRune(c)
		}
	}
	return sb.String()
}

func (b *SDL_Entry) SetSelectedTextBounds(from, too uint) error {
	if from > too || too > uint(b.textLen) {
		b.ClearSelection()
		return fmt.Errorf("invalid selection from:%d, to:%d", from, too)
	}
	b.selectCharFrom = int(from)
	b.selectCharToo = int(too)
	return nil
}

func (b *SDL_Entry) ClearSelection() {
	b.selectCharFrom = -1
	b.selectCharToo = -1
}

func (b *SDL_Entry) selectAtCursor(clicks int) bool {
	switch clicks {
	case 0, 1:
		return false
	case 2:
		from := b.findSelecteCharsRev()
		too := b.findSelecteCharsFwd()
		b.SetSelectedTextBounds(uint(from), uint(too))
	case 3:
		b.SetSelectedTextBounds(0, uint(b.textLen-1))
	}
	return true
}

func (b *SDL_Entry) insertAtCursor(text string) string {
	if b.cursor < b.textLen {
		return fmt.Sprintf("%s%s%s", b.text[0:b.cursor], text, b.text[b.cursor:])
	} else {
		return fmt.Sprintf("%s%s", b.text, text)
	}
}

func (b *SDL_Entry) Click(md *SDL_MouseData) bool {
	if b.IsEnabled() {
		b.keyPressLock.Lock()
		defer b.keyPressLock.Unlock()

		if md.GetClickCount() > 1 {
			return b.selectAtCursor(md.GetClickCount())
		}

		/*
			Is currently dragging so get dragged from (start of draggings)
			Set the flag
		*/
		if md.IsDragging() {
			if !b.dragging {
				b.dragFrom = md.GetDraggingX()
				b.dragTo = b.dragFrom
				b.dragging = true
			} else {
				b.dragTo = md.GetDraggingX()
			}
			return true
		} else {
			// If not dragging then clear flag
			b.dragging = false
		}

		/*
			Done dragging so work out what is to be selected
		*/
		if md.IsDragged() {
			go func() {
				b.screenDataLock.Lock()
				defer b.screenDataLock.Unlock()
				selTo := b.dragTo
				selFrom := b.dragFrom

				if selTo < selFrom {
					temp := selFrom
					selFrom = selTo
					selTo = temp
				}

				linked := b.screenData
				if linked == nil {
					return
				}

				/*
					Go through each image and check from and too are inside.
				*/
				sel := false
				for linked != nil {
					if linked.Inside(selFrom) && !sel {
						// Start selecting
						b.selectCharFrom = linked.pos
						b.selectCharToo = linked.pos
						sel = true
					}
					if linked.Inside(selTo) && sel {
						// End selecting
						b.selectCharToo = linked.pos
						sel = false
					}
					linked = linked.next
				}
			}()
			return true
		}

		/*
			Clicked on widget so work out where to set the cursor
		*/
		go func() {
			b.screenDataLock.Lock()
			defer b.screenDataLock.Unlock()
			linked := b.screenData
			if linked == nil {
				return
			}
			notFound := true
			for linked != nil {
				if linked.Inside(md.GetX()) {
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
		th := b.h - int32(float32(b.h)/6)
		ty := (b.h - th) / 2

		if b._invalid {
			b.Invalid(false)
			fg := b.GetForeground()
			err = GetResourceInstance().UpdateTextureCachedRunes(renderer, font, fg, b.text)
			if err != nil {
				renderer.SetDrawColor(255, 0, 0, 255)
				renderer.DrawRect(&sdl.Rect{X: b.x, Y: b.y, W: b.w, H: b.h})
				return nil
			}
			sd := GetResourceInstance().GetScaledTextureListFromCachedRunesLinked(b.text, fg, tx, th)
			if sd == nil {
				if err != nil {
					renderer.SetDrawColor(255, 0, 0, 255)
					renderer.DrawRect(&sdl.Rect{X: b.x, Y: b.y, W: b.w, H: b.h})
					return nil
				}
			}
			b.screenData = sd
		}

		screenData := b.screenData
		// *******************************************************
		// Find the number of chars thet can be displayed 'cc'
		if b.leadin >= b.textLen { // Ensure leadin is not past the end
			b.leadin = b.textLen - 1
		}
		if b.leadin < 0 { // Ensure leadin is not negative
			b.leadin = 0
		}

		screenData.SetAllVisible(false)
		leadIn := screenData.Indexed(b.leadin)
		if leadIn == nil {
			return nil
		}

		//*********************************************************
		if b.ShouldDrawBackground() {
			bc := b.GetBackground()
			renderer.SetDrawColor(bc.R, bc.G, bc.B, bc.A)
			renderer.FillRect(&sdl.Rect{X: b.x, Y: b.y, W: b.w, H: b.h})
		}
		//
		// If selectiing with the mouse then draw the background
		//
		if b.dragging && b.IsFocused() {
			c := GetResourceInstance().GetCursorSelectColour()
			renderer.SetDrawColor(c.R, c.G, c.B, c.A)
			if b.dragFrom > b.dragTo {
				renderer.FillRect(&sdl.Rect{X: b.dragTo, Y: b.y + 1, W: b.dragFrom - b.dragTo, H: b.h - 2})
			} else {
				renderer.FillRect(&sdl.Rect{X: b.dragFrom, Y: b.y + 1, W: b.dragTo - b.dragFrom, H: b.h - 2})
			}
		}

		paintCursor := b.IsEnabled() && b.IsFocused() && (sdl.GetTicks64()%1000) > 300

		//
		// Scale the text to fit the height but keep the aspect ration the same so we know the width of each char
		//   Need to use floats to prevent rounding
		//
		var rect *sdl.Rect
		tx = b.x + b.indent
		max := b.x + b.w
		last := 0
		disp := leadIn
		for disp != nil && tx+disp.width < max {
			last = disp.pos
			tw := disp.width
			rect = &sdl.Rect{X: tx, Y: b.y + ty, W: tw, H: th}
			disp.SetVisible(true)
			if disp.pos >= b.selectCharFrom && disp.pos <= b.selectCharToo {
				c := GetResourceInstance().GetCursorSelectColour()
				renderer.SetDrawColor(c.R, c.G, c.B, c.A)
				renderer.FillRect(rect)
			}
			renderer.Copy(disp.te.texture, nil, rect)
			if !b.cursorAtEnd && paintCursor && disp.pos == b.cursor {
				c := GetResourceInstance().GetCursorInsertColour()
				renderer.SetDrawColor(c.R, c.G, c.B, c.A)
				renderer.FillRect(&sdl.Rect{X: tx, Y: b.y, W: 5, H: b.h})
			}
			tx = tx + tw
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
