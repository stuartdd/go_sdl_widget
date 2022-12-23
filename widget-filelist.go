package go_sdl_widget

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"

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
	filter       func(bool, string) bool
	onSelect     func(string, FILE_LIST_RESPONSE_CODE, int32) bool
	ret          FILE_LIST_RESPONSE_CODE
	rowHeight    int32
}

func NewFileList(x, y, rh, id int32, currentPath string, font *ttf.Font, style STATE_BITS, onSelect func(string, FILE_LIST_RESPONSE_CODE, int32) bool, filter func(bool, string) bool) (*SDL_FileList, error) {
	stat, err := os.Stat(currentPath)
	if err != nil {
		currentPath, err = os.Getwd()
		if err != nil {
			currentPath = os.TempDir()
		}
	} else {
		currentPath, _ = filepath.Abs(stat.Name())
	}
	fl := &SDL_FileList{currentPath: currentPath, selectedFile: "", rowHeight: rh, ret: 0, onSelect: onSelect, filter: filter}
	if fl.filter == nil {
		fl.filter = func(b bool, s string) bool {
			return true
		}
	}
	if fl.onSelect == nil {
		fl.onSelect = func(s string, f FILE_LIST_RESPONSE_CODE, i int32) bool {
			return true
		}
	}
	fl.SDL_WidgetSubGroup = SDL_WidgetSubGroup{font: font, base: nil, countBase: 0, temp: nil, countTemp: 0}
	fl.SDL_WidgetSubGroup.SDL_WidgetBase = initBase(x, y, 1000, rh, id, &fl.SDL_WidgetSubGroup, 0, false, style, nil)
	fl.Reload(currentPath)
	return fl, nil
}

func (fl *SDL_FileList) Reload(currentPath string) {
	err := fl.populateTemp(currentPath)
	if err == nil {
		fl.currentPath = currentPath
		fl.swapTemp()
	}
}

func (fl *SDL_FileList) populateTemp(currentPath string) error {
	abs_fname, err := filepath.Abs(currentPath)
	if err != nil {
		return err
	}
	list, err := ioutil.ReadDir(currentPath)
	if err != nil {
		return err
	}
	if len(list) == 0 {
		return fmt.Errorf("no files found for path %s", currentPath)
	}
	wid := fl.widgetId + 1
	x := fl.x
	y := fl.y
	w := fl.w
	h := fl.rowHeight

	fl.addToTemp(NewSDLButton(x, y, w, h, wid, "Cancel", WIDGET_STYLE_DRAW_BORDER, 10, func(s string, id, mousex, mousey int32) bool {
		fl.Close(FILE_LIST_CLOSE_CANCEL)
		return true
	}))

	var lab *SDL_Label
	lab = NewSDLLabel(x, y, w, h, wid, abs_fname, ALIGN_LEFT, WIDGET_STYLE_DRAW_BORDER_AND_BG)
	fl.addToTemp(lab)
	y = y + h

	if filepath.Dir(abs_fname) != abs_fname {
		lab = NewSDLLabel(x, y, w, h, wid, "D:..", ALIGN_LEFT, WIDGET_STYLE_DRAW_BG)
		lab.SetOnClick(func(s string, id, mouseX, mouseY int32) bool {
			if err == nil {
				fl.onSelect(filepath.Dir(abs_fname), FILE_LIST_PATH_SELECT, id)
			}
			return true
		})
		fl.addToTemp(lab)
	}

	for _, fil := range list {
		if fil.IsDir() && fl.filter(true, fil.Name()) {
			y = y + h
			wid++
			lab = NewSDLLabel(x, y, w, h, wid, fmt.Sprintf("D:%s", fil.Name()), ALIGN_LEFT, WIDGET_STYLE_DRAW_BG)
			lab.SetOnClick(func(s string, id, mouseX, mouseY int32) bool {
				fil := filepath.Join(fl.currentPath, s[2:])
				fl.onSelect(fil, FILE_LIST_PATH_SELECT, id)
				return true
			})
			fl.addToTemp(lab)
		}
	}

	for _, fil := range list {
		if !fil.IsDir() && fl.filter(false, fil.Name()) {
			y = y + h
			wid++
			lab = NewSDLLabel(x, y, w, h, wid, fmt.Sprintf("F:%s", fil.Name()), ALIGN_LEFT, WIDGET_STYLE_DRAW_BG)
			lab.SetOnClick(func(s string, id, mouseX, mouseY int32) bool {
				fil := filepath.Clean(filepath.Join(fl.currentPath, s[2:]))
				cur, err := os.Getwd()
				if err != nil {
					if fl.CanLog() {
						fl.Log(1, err.Error())
					}
					return false
				}
				fil, err = filepath.Rel(cur, fil)
				if err != nil {
					if fl.CanLog() {
						fl.Log(1, err.Error())
					}
					return false
				}
				if fl.onSelect(fil, FILE_LIST_FILE_SELECT, id) {
					fl.selectedFile = fil
					fl.Close(FILE_LIST_FILE_SELECT)
				}
				return true
			})
			fl.addToTemp(lab)
		}
	}
	return nil
}

func (fl *SDL_FileList) Show(viewPort sdl.Rect) {
	fl.SDL_WidgetSubGroup.SetVisible(true)
}

func (fl *SDL_FileList) Close(r FILE_LIST_RESPONSE_CODE) {
	fl.ret = r
	fl.SDL_WidgetSubGroup.SetVisible(false)
}
