package go_sdl_widget

import (
	"io/ioutil"
	"path"

	"github.com/veandco/go-sdl2/sdl"
	"github.com/veandco/go-sdl2/ttf"
)

type FILE_LIST_RESPONSE_CODE int

const (
	FILE_LIST_CLOSE_CANCEL = iota
	FILE_LIST_FILE_SELECT
	FILE_LIST_PATH_UP
	FILE_LIST_PATH_SELECT
)

type SDL_FileList struct {
	SDL_WidgetSubGroup
	currentPath  string
	selectedFile string
	onClose      func(string, FILE_LIST_RESPONSE_CODE, int32) bool
	ret          FILE_LIST_RESPONSE_CODE
}

func NewFileList(x, y, bh, id int32, currentPath string, font *ttf.Font, style STATE_BITS, onClose func(string, FILE_LIST_RESPONSE_CODE, int32) bool) (*SDL_FileList, error) {
	list, err := ioutil.ReadDir(currentPath)
	if err != nil {
		return nil, err
	}
	fl := &SDL_FileList{currentPath: currentPath, selectedFile: "", ret: 0, onClose: onClose}
	fl.SDL_WidgetSubGroup = SDL_WidgetSubGroup{font: font, base: nil, count: 0}
	fl.SDL_WidgetSubGroup.SDL_WidgetBase = initBase(x, y, 1000, bh, id, &fl.SDL_WidgetSubGroup, 0, false, style, nil)

	fl.Add(NewSDLButton(0, y, 1000, bh, id, "Cancel", WIDGET_STYLE_BORDER_AND_BG, 10, func(s string, i1, i2, i3 int32) bool {
		if fl.onClose != nil {
			if onClose(fl.selectedFile, FILE_LIST_CLOSE_CANCEL, id) {
				fl.Close(FILE_LIST_CLOSE_CANCEL)
			}
		} else {
			fl.Close(FILE_LIST_CLOSE_CANCEL)
		}
		return true
	}))

	y = y + bh

	for i, f := range list {
		la := NewSDLLabel(0, y, 1000, bh, id+int32(i+1), f.Name(), ALIGN_LEFT, WIDGET_STYLE_DRAW_BG)
		la.SetOnClick(func(s string, id, i2, i3 int32) bool {
			fl.selectedFile = (path.Join(fl.currentPath, s))
			if fl.onClose != nil {
				if onClose(fl.selectedFile, FILE_LIST_FILE_SELECT, id) {
					fl.Close(FILE_LIST_FILE_SELECT)
				}
			} else {
				fl.Close(FILE_LIST_FILE_SELECT)
			}
			return true
		})
		fl.Add(la)
		y = y + bh
	}
	return fl, nil
}

func (fl *SDL_FileList) Show(viewPort sdl.Rect) {
	fl.SetVisible(true)
}

func (fl *SDL_FileList) Close(r FILE_LIST_RESPONSE_CODE) {
	fl.ret = r
	fl.SetVisible(false)
}
