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
	text              string
	selectedTextFlags []bool
	selecteCharsFwd   []byte
	selecteCharsRev   []byte
	textLen           int
	history           []string
	cursor            int
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
	screenData        []*rune_ScaledScreenData
	screenDataReady   sync.Mutex
	keyPressLock      sync.Mutex
}

type rune_ScaledScreenData struct {
	runePos int
	runeX   int32
	runeW   int32
	sel     bool
	te      *SDL_TextureCacheEntry
}

func (r *rune_ScaledScreenData) inside(x int32) bool {
	return x > r.runeX && x < (r.runeX+r.runeW)
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
	if b.IsFocused() {
		if i < 0 {
			i = 0
		}
		if i > b.textLen {
			i = b.textLen
		}
		b.cursor = i
	}
}

func (b *SDL_Entry) MoveCursor(i int) {
	b.SetCursor(b.cursor + i)
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
	return b.getSelectedTextFromFlags(b.selectedTextFlags)
}

func (b *SDL_Entry) ClearSelection() {
	b.screenDataReady.Lock()
	defer b.screenDataReady.Unlock()
	b.selFrom = 0
	b.selTo = 0
	b.selectedTextFlags = make([]bool, len(b.text))
}

func (b *SDL_Entry) selectAtCursor(clicks int) bool {
	b.screenDataReady.Lock()
	defer b.screenDataReady.Unlock()
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

func (b *SDL_Entry) getSelectedTextFromFlags(sels []bool) string {
	var sb strings.Builder
	for i, sel := range sels {
		if sel {
			sb.WriteString(b.text[i : i+1])
		}
	}
	return sb.String()
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
				b.screenDataReady.Lock()
				defer b.screenDataReady.Unlock()
				list := b.screenData
				if list == nil {
					return
				}
				b.selTo = b.dragTo
				b.selFrom = b.dragFrom
				oldSel := b.getSelectedTextFromFlags(b.selectedTextFlags)
				sel := false
				newSels := make([]bool, len(list))

				if b.selTo < b.selFrom {
					temp := b.selFrom
					b.selFrom = b.selTo
					b.selTo = temp
				}

				for i, sd := range list {
					if sd.inside(b.selFrom) {
						b.selFrom = sd.runeX
						sel = true
						sd.sel = true
					}
					sd.sel = sel
					newSels[i] = sel
					if sd.inside(b.selTo) {
						b.selTo = sd.runeX + sd.runeW
						sel = false
					}
				}
				b.selTo = 0
				b.selFrom = 0

				newSel := b.getSelectedTextFromFlags(newSels)
				if newSel != oldSel {
					s, _ := b.onChange(oldSel, newSel, TEXT_CHANGE_SELECTED)
					if s == newSel {
						b.selectedTextFlags = newSels
					}
				}
			}()
			return true
		}
		if md.GetClickCount() > 1 {
			return b.selectAtCursor(md.GetClickCount())
		}
		go func() {
			b.screenDataReady.Lock()
			defer b.screenDataReady.Unlock()
			list := b.screenData
			if list == nil {
				return
			}
			// Position the cursor!
			for i, sd := range list {
				if sd.runeX > md.x {
					b.SetCursor(i)
					return
				}
			}
			b.SetCursor(b.leadout)
		}()
	}
	return false
}

func (b *SDL_Entry) Draw(renderer *sdl.Renderer, font *ttf.Font) error {
	if b.IsVisible() {
		b.screenDataReady.Lock()
		defer b.screenDataReady.Unlock()

		var err error
		var ec *SDL_TextureCacheEntry
		if b._invalid {
			b.Invalid(false)
			err = GetResourceInstance().UpdateTextureCachedRunes(renderer, font, b.GetForeground(), b.text)
			if err != nil {
				renderer.SetDrawColor(255, 0, 0, 255)
				renderer.DrawRect(&sdl.Rect{X: b.x, Y: b.y, W: b.w, H: b.h})
				return nil
			}

		}

		// *******************************************************
		// Find the number of chars thet can be displayed 'cc'
		tx := b.x + b.indent
		cc := 0
		if b.leadin < 0 { // Ensure leadin is not negative
			b.leadin = 0
		}
		b.leadout = b.textLen
		list := GetResourceInstance().GetTextureListFromCachedRunes(b.text, b.GetForeground()) // Get the textures and their widths
		if list == nil {
			return nil
		}
		// work out how many chars will fit in the rectangle
		for pos := b.leadin; pos < b.textLen; pos++ {
			ec = list[pos]
			tx = tx + ec.W
			if tx >= b.x+b.w {
				break
			}
			cc++
		}

		if b.leadin > 0 && cc == 0 {
			b.leadin--
			cc++
		}

		if b.leadin > 0 && b.leadin == b.cursor {
			b.leadin--
		}
		b.leadout = b.leadin + cc
		if b.cursor > b.leadout {
			b.leadin = b.cursor - cc
			b.leadout = b.leadin + cc
		}
		if b.leadout > b.textLen {
			b.leadout = b.textLen
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
		inset := float32(b.h) / 4
		th := float32(b.h) - inset
		ty := (float32(b.h) - th) / 2

		tx = b.x + int32(b.indent)
		cursorNotVisible := true
		paintCursor := b.IsEnabled() && b.IsFocused() && (sdl.GetTicks64()%1000) > 300
		//
		//
		// Copy each (scaled) char image to the renderer
		//
		sdPos := 0
		sd := make([]*rune_ScaledScreenData, b.leadout-b.leadin)
		for pos := b.leadin; pos < b.leadout; pos++ {
			ec := list[pos]
			aspect := float32(ec.W) / float32(ec.H)
			tw := th * aspect
			sd[sdPos] = &rune_ScaledScreenData{te: ec, runePos: pos, runeX: tx, runeW: int32(tw)}
			if pos < len(b.selectedTextFlags) {
				if b.selectedTextFlags[pos] {
					renderer.SetDrawColor(100, 100, 100, 25)
					renderer.FillRect(&sdl.Rect{X: tx, Y: b.y + int32(ty), W: int32(tw), H: int32(th)})
				}
			}
			renderer.Copy(ec.Texture, nil, &sdl.Rect{X: tx, Y: b.y + int32(ty), W: int32(tw), H: int32(th)})
			if paintCursor {
				if pos == b.cursor {
					renderer.SetDrawColor(255, 255, 255, 255)
					renderer.FillRect(&sdl.Rect{X: tx, Y: b.y, W: 2, H: b.h})
					cursorNotVisible = false
				}
			}
			tx = tx + int32(tw)
			sdPos++
		}
		b.screenData = sd
		if cursorNotVisible && paintCursor {
			renderer.SetDrawColor(255, 255, 255, 255)
			renderer.FillRect(&sdl.Rect{X: tx, Y: b.y, W: 2, H: b.h})
		}
		if b.ShouldDrawBackground() {
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
