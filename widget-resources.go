package go_sdl_widget

import (
	"bufio"
	"fmt"
	"os"
	"path"
	"path/filepath"
	"strconv"
	"strings"
	"sync"

	"github.com/veandco/go-sdl2/img"
	"github.com/veandco/go-sdl2/sdl"
	"github.com/veandco/go-sdl2/ttf"
)

type sdl_Resources struct {
	resourceDir        string
	fontFilename       string
	fontSize           int
	font               *ttf.Font
	textureCache       *SDL_TextureCache
	cacheLock          sync.Mutex
	colours            [][]*sdl.Color
	colourNames        map[string]*sdl.Color
	cursorInsertColour *sdl.Color
	cursorAppendColour *sdl.Color
	cursorSelectColour *sdl.Color
	selectCharsFwd     []byte
	selectCharsRev     []byte
}

type STATE_COLOUR uint
type STYLE_COLOUR uint

const (
	WIDGET_COLOUR_INDEX_ENABLED STATE_COLOUR = 0
	WIDGET_COLOUR_INDEX_DISABLE STATE_COLOUR = 1
	WIDGET_COLOUR_INDEX_FOCUS   STATE_COLOUR = 2
	WIDGET_COLOUR_INDEX_ERROR   STATE_COLOUR = 3
	WIDGET_COLOUR_INDEX_SIZE    int          = 4 // So we create an array the right size
	WIDGET_CURSOR_STATE         int          = 100
	WIDGET_SELECT_STATE         int          = 101

	WIDGET_COLOUR_STYLE_FG     STYLE_COLOUR = 0 // Section indexes
	WIDGET_COLOUR_STYLE_BG     STYLE_COLOUR = 1
	WIDGET_COLOUR_STYLE_BORDER STYLE_COLOUR = 2
	WIDGET_COLOUR_STYLE_SIZE   int          = 3 // So we create an array the right size

	pixel_BYTE_RED   = 0
	pixel_BYTE_GREEN = 1
	pixel_BYTE_BLUE  = 2
	pixel_BYTE_ALPHA = 3
)

var (
	sdlResourceInstanceLock = &sync.Mutex{}
	sdlResourceInstance     *sdl_Resources
	configMapState          = map[string]STATE_COLOUR{
		"enabled":  WIDGET_COLOUR_INDEX_ENABLED,
		"disabled": WIDGET_COLOUR_INDEX_DISABLE,
		"focused":  WIDGET_COLOUR_INDEX_FOCUS,
		"error":    WIDGET_COLOUR_INDEX_ERROR,
	}
	configMapStyle = map[string]STYLE_COLOUR{
		"fg":     WIDGET_COLOUR_STYLE_FG,
		"bg":     WIDGET_COLOUR_STYLE_BG,
		"border": WIDGET_COLOUR_STYLE_BORDER,
	}
)

func GetResourceInstance() *sdl_Resources {
	if sdlResourceInstance == nil {
		sdlResourceInstanceLock.Lock()
		defer sdlResourceInstanceLock.Unlock()

		if sdlResourceInstance == nil {

			sdlResourceInstance = &sdl_Resources{
				textureCache: NewTextureCache(),
				colours:      make([][]*sdl.Color, WIDGET_COLOUR_INDEX_SIZE),
			}
			var sci STATE_COLOUR = WIDGET_COLOUR_INDEX_ENABLED
			for sci = 0; sci < STATE_COLOUR(WIDGET_COLOUR_INDEX_SIZE); sci++ {
				sdlResourceInstance.colours[sci] = make([]*sdl.Color, WIDGET_COLOUR_STYLE_SIZE)
			}
			sdlResourceInstance.colourNames = make(map[string]*sdl.Color)

			sdlResourceInstance.colours[WIDGET_COLOUR_INDEX_ENABLED][WIDGET_COLOUR_STYLE_FG] = &sdl.Color{R: 0, G: 255, B: 0, A: 255}
			sdlResourceInstance.colours[WIDGET_COLOUR_INDEX_ENABLED][WIDGET_COLOUR_STYLE_BG] = &sdl.Color{R: 0, G: 100, B: 0, A: 255}
			sdlResourceInstance.colours[WIDGET_COLOUR_INDEX_ENABLED][WIDGET_COLOUR_STYLE_BORDER] = &sdl.Color{R: 0, G: 255, B: 0, A: 255}

			sdlResourceInstance.colours[WIDGET_COLOUR_INDEX_DISABLE][WIDGET_COLOUR_STYLE_FG] = &sdl.Color{R: 0, G: 150, B: 0, A: 255}
			sdlResourceInstance.colours[WIDGET_COLOUR_INDEX_DISABLE][WIDGET_COLOUR_STYLE_BG] = &sdl.Color{R: 0, G: 100, B: 0, A: 255}
			sdlResourceInstance.colours[WIDGET_COLOUR_INDEX_DISABLE][WIDGET_COLOUR_STYLE_BORDER] = &sdl.Color{R: 0, G: 150, B: 0, A: 255}

			sdlResourceInstance.colours[WIDGET_COLOUR_INDEX_FOCUS][WIDGET_COLOUR_STYLE_FG] = &sdl.Color{R: 255, G: 255, B: 0, A: 255}
			sdlResourceInstance.colours[WIDGET_COLOUR_INDEX_FOCUS][WIDGET_COLOUR_STYLE_BG] = &sdl.Color{R: 150, G: 150, B: 0, A: 255}
			sdlResourceInstance.colours[WIDGET_COLOUR_INDEX_FOCUS][WIDGET_COLOUR_STYLE_BORDER] = &sdl.Color{R: 255, G: 255, B: 0, A: 255}

			sdlResourceInstance.colours[WIDGET_COLOUR_INDEX_ERROR][WIDGET_COLOUR_STYLE_FG] = &sdl.Color{R: 255, G: 0, B: 0, A: 255}
			sdlResourceInstance.colours[WIDGET_COLOUR_INDEX_ERROR][WIDGET_COLOUR_STYLE_BG] = &sdl.Color{R: 100, G: 0, B: 0, A: 255}
			sdlResourceInstance.colours[WIDGET_COLOUR_INDEX_ERROR][WIDGET_COLOUR_STYLE_BORDER] = &sdl.Color{R: 255, G: 0, B: 0, A: 255}

			sdlResourceInstance.cursorInsertColour = &sdl.Color{R: 255, G: 255, B: 255, A: 255}
			sdlResourceInstance.cursorAppendColour = &sdl.Color{R: 255, G: 0, B: 255, A: 255}
			sdlResourceInstance.cursorSelectColour = &sdl.Color{R: 100, G: 0, B: 100, A: 255}

			sdlResourceInstance.SetSelectCharsFwd("/.")
			sdlResourceInstance.SetSelectCharsRev("/")
			p, err := os.Getwd()
			if err != nil {
				sdlResourceInstance.resourceDir = "resources"
			} else {
				sdlResourceInstance.resourceDir = path.Join(p, "resources")
			}
			sdlResourceInstance.fontFilename = path.Join(sdlResourceInstance.resourceDir, "buttonFont.ttf")
			sdlResourceInstance.fontSize = 35
		}
	}

	return sdlResourceInstance
}

func (r *sdl_Resources) Config(filename string) error {
	inFile, err := os.Open(filename)
	if err != nil {
		return err
	}
	defer inFile.Close()
	var s string = ""
	scanner := bufio.NewScanner(inFile)
	for scanner.Scan() {
		s = scanner.Text()
		s = strings.TrimSpace(s)
		if !strings.HasPrefix(s, "#") && len(s) > 0 {
			l := strings.Split(s, "=")
			if len(l) == 2 {
				err := r.ConfigNameValue(l[0], l[1])
				if err != nil {
					return fmt.Errorf("line:'%s'. %s. Config File is '%s'", s, err.Error(), filename)
				}
			} else {
				return fmt.Errorf("config file line is invalid '%s'. Config File is '%s'", s, filename)
			}
		}
	}
	return nil
}

func (r *sdl_Resources) ConfigNameValue(n, v string) error {
	n = strings.TrimSpace(n)
	if len(n) > 0 {
		v = strings.TrimSpace(v)
		n = strings.TrimSpace(n)
		nl := strings.Split(n, ".")
		if len(nl) > 2 && nl[0] == "res" {
			err := r.setConfigValue(nl[1:], v)
			if err != nil {
				return err
			}
		} else {
			return fmt.Errorf("config name is invalid. Expecting prefix 'res.*.*' Found '%s'", n)
		}
	}
	return nil
}

func (r *sdl_Resources) setConfigValue(n []string, v string) error {
	n0 := strings.ToLower(n[0])
	n1 := strings.ToLower(n[1])

	i1, ok := configMapState[n0]
	if !ok {
		switch n0 {
		case "font":
			switch n1 {
			case "name":
				err := r.SetFontFilename(v)
				return err
			case "size":
				si, err := strconv.Atoi(v)
				if err != nil {
					return err
				}
				err = r.SetFontSize(si)
				return err
			default:
				return fmt.Errorf("invalid name. Expecting 'font.name, font.size' Found '%s'", n)
			}
		case "resource":
			switch n1 {
			case "dir":
				err := r.SetResourceDir(v)
				return err
			default:
				return fmt.Errorf("invalid name. Expecting 'resource.dir' Found '%s'", n)
			}
		case "colour", "color":
			c, err := parseColourString(v)
			if err != nil {
				return err
			}
			err = r.SetColourName(n[1], c)
			if err != nil {
				return err
			}
			return nil
		case "cursor":
			c, err := parseColourString(v)
			if err != nil {
				return err
			}
			switch n1 {
			case "insert":
				r.SetCursorInsertColour(c)
				return nil
			case "append":
				r.SetCursorAppendColour(c)
				return nil
			case "select":
				r.SetCursorSelectColour(c)
				return nil
			default:
				return fmt.Errorf("invalid name. Expecting 'cursor.insert, cursor.append, cursor.select' Found '%s'", n)
			}
		case "select":
			if len(v) < 1 {
				return fmt.Errorf("invalid value. Expecting string longer than 1 char")
			}
			switch n1 {
			case "forward":
				r.SetSelectCharsFwd(v)
				return nil
			case "backward":
				r.SetSelectCharsRev(v)
				return nil
			default:
				return fmt.Errorf("invalid name. Expecting 'select.forward, select.backward' Found '%s'", n)
			}
		default:
			return fmt.Errorf("invalid name. Expecting a name from %v or 'cursor or select' Found '%s'", configMapState, n0)
		}
	}
	i2, ok := configMapStyle[n1]
	if !ok {
		return fmt.Errorf("config name. Expecting styles name from %v Found '%s'", configMapStyle, n1)
	}
	c, err := parseColourString(v)
	if err != nil {
		return err
	}
	r.SetColour(i1, i2, c)
	return nil
}

func (r *sdl_Resources) Destroy() {
	if r.font != nil {
		r.font.Close()
		ttf.Quit()
	}
	r.GetTextureCache().Destroy()
}

func (r *sdl_Resources) LoadFont() error {
	info, err := os.Stat(r.fontFilename)
	if err != nil {
		return fmt.Errorf("failed to load Font '%s'. Error: %s", r.fontFilename, err.Error())
	}
	if info.IsDir() {
		return fmt.Errorf("failed to load Font. Path '%s' is a directory", r.fontFilename)
	}
	if r.font == nil {
		if err := ttf.Init(); err != nil {
			return fmt.Errorf("failed to init the ttf font system: %s", err)
		}
	}
	font, err := ttf.OpenFont(r.fontFilename, r.fontSize)
	if err != nil {
		return fmt.Errorf("failed to load the font from file: %s", err)
	}
	if r.font == nil {
		r.font = font
	}
	return nil
}

func (r *sdl_Resources) GetColour(stateIndex STATE_COLOUR, styleIndex STYLE_COLOUR) *sdl.Color {
	return r.colours[stateIndex][styleIndex]
}

func (r *sdl_Resources) SetColour(stateIndex STATE_COLOUR, styleIndex STYLE_COLOUR, c *sdl.Color) {
	r.colours[stateIndex][styleIndex] = c
}

func (r *sdl_Resources) SetColourName(name string, c *sdl.Color) error {
	if len(name) < 2 {
		return fmt.Errorf("failed to set colour name. Name '%s' must be longer than 1 char", name)
	}
	r.colourNames[name] = c
	return nil
}

func (r *sdl_Resources) GetColourName(name string, stateIndex STATE_COLOUR, styleIndex STYLE_COLOUR) *sdl.Color {
	c, ok := r.colourNames[name]
	if ok {
		return c
	}
	return r.GetColour(stateIndex, styleIndex)
}

func (r *sdl_Resources) GetCursorInsertColour() *sdl.Color {
	if r.cursorInsertColour == nil {
		return &sdl.Color{R: 255, G: 255, B: 255, A: 255}
	}
	return r.cursorInsertColour
}

func (r *sdl_Resources) GetCursorAppendColour() *sdl.Color {
	if r.cursorAppendColour == nil {
		return &sdl.Color{R: 0, G: 255, B: 255, A: 255}
	}
	return r.cursorAppendColour
}

func (r *sdl_Resources) GetCursorSelectColour() *sdl.Color {
	if r.cursorSelectColour == nil {
		return &sdl.Color{R: 100, G: 0, B: 100, A: 255}
	}
	return r.cursorSelectColour
}

func (r *sdl_Resources) SetCursorInsertColour(c *sdl.Color) {
	r.cursorInsertColour = c
}

func (r *sdl_Resources) SetCursorAppendColour(c *sdl.Color) {
	r.cursorAppendColour = c
}

func (r *sdl_Resources) SetCursorSelectColour(c *sdl.Color) {
	r.cursorSelectColour = c
}

func (r *sdl_Resources) SetResourceDir(dir string) error {
	info, err := os.Stat(dir)
	if err != nil {
		p, err := os.Getwd()
		if err != nil {
			return fmt.Errorf("failed to set Resource Directrory. Path was '%s'. Error: %s", dir, err.Error())
		}
		dir = path.Join(p, dir)
		info, err = os.Stat(dir)
		if err != nil {
			return fmt.Errorf("failed to set Resource Directrory. Derived Path was '%s'. Error: %s", dir, err.Error())
		}
	}
	if info.IsDir() {
		r.resourceDir = dir
		return nil
	}
	return fmt.Errorf("failed to set Resource Directrory. Path '%s' is NOT a directory", dir)
}

func (r *sdl_Resources) GetResourceDir() string {
	return r.resourceDir
}

func (r *sdl_Resources) SetFontFilename(filename string) error {
	info, err := os.Stat(filename)
	if err != nil {
		fn := path.Join(r.resourceDir, filename)
		info, err = os.Stat(fn)
		if err != nil {
			return fmt.Errorf("failed to set Font File Name. Derived File was '%s'. Error: %s", fn, err.Error())
		}
		filename = fn
	}
	if !info.IsDir() {
		r.fontFilename = filename
		return nil
	}
	return fmt.Errorf("failed to set Font File Name. File '%s' is a directory", filename)
}

func (r *sdl_Resources) GetFontFilename() string {
	return r.fontFilename
}

func (r *sdl_Resources) GetFontSize() int {
	return r.fontSize
}

func (r *sdl_Resources) SetFontSize(size int) error {
	if size < 10 {
		return fmt.Errorf("failed to set Font Size. Must be 10 or above. Was set to '%d'", size)
	}
	r.fontSize = size
	return nil
}

func (r *sdl_Resources) GetFont() *ttf.Font {
	return r.font
}

func (b *sdl_Resources) SetSelectCharsFwd(s string) {
	b.selectCharsFwd = []byte(s)
}

func (b *sdl_Resources) GetSelectCharsFwd() string {
	return string(b.selectCharsFwd)
}

func (b *sdl_Resources) SetSelectCharsRev(s string) {
	b.selectCharsRev = []byte(s)
}

func (b *sdl_Resources) GetSelectCharsRev() string {
	return string(b.selectCharsRev)
}

func (r *sdl_Resources) GetTextureCache() *SDL_TextureCache {
	return r.textureCache
}

func (r *sdl_Resources) GetTextureForName(name string) (*sdl.Texture, int32, int32, error) {
	r.cacheLock.Lock()
	defer r.cacheLock.Unlock()
	tce := r.textureCache.textureMap[name]
	if tce == nil {
		return nil, 0, 0, fmt.Errorf("texture cache does not contain %s", name)
	}
	return tce.texture, tce.w, tce.h, nil
}

func (r *sdl_Resources) AddTexturesFromStringMap(renderer *sdl.Renderer, stringMap map[string]string, font *ttf.Font, colour *sdl.Color) error {
	r.cacheLock.Lock()
	defer r.cacheLock.Unlock()
	for name, text := range stringMap {
		tce, err := newTextureCacheEntryForString(renderer, text, font, colour)
		if err != nil {
			return err
		}
		r.textureCache.Add(name, tce)
	}
	return nil
}

func (r *sdl_Resources) AddTexturesFromFileMap(renderer *sdl.Renderer, fileNames map[string]string, colours ...*sdl.Color) error {
	r.cacheLock.Lock()
	defer r.cacheLock.Unlock()
	for name, fileName := range fileNames {
		fn := filepath.Join(r.GetResourceDir(), fileName)
		tce, err := newTextureCacheEntryForFile(renderer, fn, colours...)
		if err != nil {
			return fmt.Errorf("file '%s':%s", fileName, err.Error())
		}
		r.textureCache.Add(name, tce)
	}
	return nil
}

func (r *sdl_Resources) GetTextureListFromCachedRunes(text string, colour *sdl.Color) []*SDL_TextureCacheEntry {
	list := make([]*SDL_TextureCacheEntry, len(text))
	cid := GetColourId(colour)
	for i, c := range text {
		ec := r.textureCache.textureMap[fmt.Sprintf("|%c%d", c, cid)]
		if ec == nil {
			return nil
		}
		list[i] = ec
	}
	return list
}

func (r *sdl_Resources) GetScaledTextureListFromCachedRunesLinked(text string, colour *sdl.Color, offset, height int32) *sdl_TextureCacheEntryRune {
	var rootEnt *sdl_TextureCacheEntryRune
	var currentEnt *sdl_TextureCacheEntryRune
	var nextEnt *sdl_TextureCacheEntryRune
	var sw float32 = 0
	ofs := float32(offset)
	cid := GetColourId(colour)
	for i, c := range text {
		tce := r.textureCache.textureMap[fmt.Sprintf("|%c%d", c, cid)]
		if tce == nil {
			return rootEnt
		}
		sw = float32(height) * (float32(tce.w) / float32(tce.h))
		nextEnt = &sdl_TextureCacheEntryRune{pos: i, te: tce, offset: int32(ofs), width: int32(sw), next: nil}
		if rootEnt == nil {
			rootEnt = nextEnt
			currentEnt = nextEnt
		} else {
			currentEnt.next = nextEnt
			currentEnt = nextEnt
		}
		ofs = ofs + sw
	}
	return rootEnt
}

func (r *sdl_Resources) UpdateTextureCachedRunes(renderer *sdl.Renderer, font *ttf.Font, colour *sdl.Color, text string) error {
	r.cacheLock.Lock()
	defer r.cacheLock.Unlock()
	for _, c := range text {
		key := fmt.Sprintf("|%c%d", c, GetColourId(colour))
		ok := r.textureCache.Peek(key)
		if !ok {
			ec, err := newTextureCacheEntryForRune(renderer, c, font, colour)
			if err != nil {
				return err
			}
			r.textureCache.Add(key, ec)
		}
	}
	return nil
}

func (r *sdl_Resources) UpdateTextureFromString(renderer *sdl.Renderer, cacheKey, text string, font *ttf.Font, colour *sdl.Color) (*SDL_TextureCacheEntry, error) {
	r.cacheLock.Lock()
	defer r.cacheLock.Unlock()
	gtwe, ok := r.textureCache.textureMap[cacheKey]
	if ok {
		if gtwe.value == text {
			return gtwe, nil
		}
	}
	ctwe, err := newTextureCacheEntryForString(renderer, text, font, colour)
	if err != nil {
		return nil, err
	}
	if gtwe == nil {
		r.textureCache.Add(cacheKey, ctwe)
		return ctwe, nil
	}
	gtwe.Update(ctwe)
	return gtwe, nil
}

func GetColourId(c *sdl.Color) uint32 {
	return uint32(c.A) | uint32(c.R)<<8 | uint32(c.G)<<16 | uint32(c.B)<<24
}

/****************************************************************************************
* Texture cache for widgets that have textures to display.
* Textures are sdl resources and need to be Destroyed.
* SDL_WidgetList destroys all textures via the SDL_Widgets Destroy() function.
* SDL_WidgetGroup destroys all textures via SDL_WidgetsGroup Destroy() function
* USE:		widgetGroup := NewWidgetGroup()
*       	defer widgetGroup.Destroy()
**/
type SDL_TextureCache struct {
	textureMap map[string]*SDL_TextureCacheEntry
	in, out    int
}

func NewTextureCache() *SDL_TextureCache {
	return &SDL_TextureCache{textureMap: make(map[string]*SDL_TextureCacheEntry), in: 0, out: 0}
}

func (tc *SDL_TextureCache) String() string {
	return fmt.Sprintf("TextureCache in:%d out:%d", tc.in, tc.out)
}

func (tc *SDL_TextureCache) Peek(name string) bool {
	_, ok := tc.textureMap[name]
	return ok
}

func (tc *SDL_TextureCache) Add(name string, tceIn *SDL_TextureCacheEntry) {
	tce := tc.textureMap[name]
	if tce != nil {
		tc.out = tc.out + tce.Destroy()
	}
	tc.textureMap[name] = tceIn
	tc.in++
}

func (tc *SDL_TextureCache) Remove(name string, tceIn *SDL_TextureCacheEntry) {
	tce := tc.textureMap[name]
	if tce != nil {
		tc.out = tc.out + tce.Destroy()
	}
	tc.textureMap[name] = nil
}

func (tc *SDL_TextureCache) Destroy() {
	for n, tce := range tc.textureMap {
		if tce != nil {
			tc.out = tc.out + tce.Destroy()
		}
		tc.textureMap[n] = nil
	}
}

/****************************************************************************************
* Texture cache Entry used to hold ALL textures in the SDL_TextureCache
**/
type SDL_TextureCacheEntry struct {
	texture *sdl.Texture
	value   string
	w, h    int32
}

func (tce *SDL_TextureCacheEntry) Destroy() int {
	if tce.texture != nil {
		tce.texture.Destroy()
		return 1
	}
	return 0
}

func (tce *SDL_TextureCacheEntry) Update(with *SDL_TextureCacheEntry) {
	tce.Destroy()
	tce.texture = with.texture
	tce.w = with.w
	tce.h = with.h
	tce.value = with.value
}

func (tce *SDL_TextureCacheEntry) ScaledWidthHeight(txtH, clientW int32) (int32, int32, int32) {
	w1 := int32(float32(tce.w) * (float32(txtH) / float32(tce.h)))
	if clientW <= w1 {
		return int32(float32(tce.w) * (float32(clientW) / float32(w1))), w1, txtH
	}
	return tce.w, w1, txtH
}

func newTextureCacheEntryForFile(renderer *sdl.Renderer, fileName string, colours ...*sdl.Color) (*SDL_TextureCacheEntry, error) {
	surface, err := img.Load(fileName)
	if err != nil {
		return nil, err
	}
	defer surface.Free()
	if len(colours) > 0 {
		if len(colours)%2 == 1 {
			colours = append(colours, sdlResourceInstance.GetColour(WIDGET_COLOUR_INDEX_ENABLED, WIDGET_COLOUR_STYLE_FG))
		}
		surface.Lock()
		pixels := surface.Pixels()
		bpp := surface.BytesPerPixel()
		if bpp != 4 {
			return nil, fmt.Errorf("bytes per pixel for image '%s' must be 4", fileName)
		}
		for i := 0; i < len(colours); i = i + 2 {
			updateSurfacePixels(pixels, bpp, colours[i+0], colours[i+1])
		}

		surface.Unlock()
	}
	clip := surface.ClipRect
	// Dont destroy the texture. Call Destroy on the SDL_Widgets object to destroy ALL cached textures
	txt, err := renderer.CreateTextureFromSurface(surface)
	if err != nil {
		return nil, err
	}
	return &SDL_TextureCacheEntry{texture: txt, value: fileName, w: clip.W, h: clip.H}, nil
}

func newTextureCacheEntryForRune(renderer *sdl.Renderer, char rune, font *ttf.Font, colour *sdl.Color) (*SDL_TextureCacheEntry, error) {
	if colour == nil {
		colour = &sdl.Color{R: 255, G: 255, B: 255, A: 255}
	}
	surface, err := font.RenderUTF8Blended(string(char), *colour)
	if err != nil {
		return nil, err
	}
	defer surface.Free()

	clip := surface.ClipRect
	// Dont destroy the texture. Call Destroy on the SDL_Widgets object to destroy ALL cached textures
	txt, err := renderer.CreateTextureFromSurface(surface)
	if err != nil {
		return nil, err
	}
	return &SDL_TextureCacheEntry{texture: txt, value: string(char), w: clip.W, h: clip.H}, nil
}

func newTextureCacheEntryForString(renderer *sdl.Renderer, text string, font *ttf.Font, colour *sdl.Color) (*SDL_TextureCacheEntry, error) {
	if colour == nil {
		colour = &sdl.Color{R: 255, G: 255, B: 255, A: 255}
	}
	surface, err := font.RenderUTF8Blended(text, *colour)
	if err != nil {
		return nil, err
	}
	defer surface.Free()
	clip := surface.ClipRect
	// Dont destroy the texture. Call Destroy on the SDL_Widgets object to destroy ALL cached textures
	txt, err := renderer.CreateTextureFromSurface(surface)
	if err != nil {
		return nil, err
	}
	return &SDL_TextureCacheEntry{texture: txt, w: clip.W, h: clip.H, value: text}, nil
}

func updateSurfacePixels(pixels []byte, bpp int, frm, too *sdl.Color) {
	pc := len(pixels)
	b1 := frm.B
	g1 := frm.G
	r1 := frm.R
	b2 := too.B
	g2 := too.G
	r2 := too.R
	a2 := too.A
	for p := 0; p < pc; p = p + bpp {
		if pixels[p+pixel_BYTE_BLUE] == b1 && pixels[p+pixel_BYTE_GREEN] == g1 && pixels[p+pixel_BYTE_RED] == r1 {
			pixels[p+pixel_BYTE_BLUE] = b2
			pixels[p+pixel_BYTE_GREEN] = g2
			pixels[p+pixel_BYTE_RED] = r2
			pixels[p+pixel_BYTE_ALPHA] = a2
		}
	}
}

/****************************************************************************************
* TextureCacheEntryRune holds the state of a character in an entry field.
*   te is a SDL_TextureCacheEntry holding the image in a specific colour
*   pos is the position in the entry text
*   offset is the absolute x position on screen.
*   width if the width of the char after it is acaled. This is different to tx.W
* This is a linked list in entry text order.
* A ne list is created if the entry text is changed
**/

type sdl_TextureCacheEntryRune struct {
	pos           int // Position in the string this list represents
	offset, width int32
	visible       bool
	te            *SDL_TextureCacheEntry     // The texture data for the char at pos in the string this list represents
	next          *sdl_TextureCacheEntryRune // The next in the string this list represents. Nil at the end
}

func (er *sdl_TextureCacheEntryRune) String() string {
	return fmt.Sprintf("TCER: pos:%d '%s' ofs:%d eidth:%d", er.pos, er.te.value, er.offset, er.width)
}

func (er *sdl_TextureCacheEntryRune) SetVisible(s bool) {
	er.visible = s
}

func (er *sdl_TextureCacheEntryRune) IsVisible() bool {
	return er.visible

}

func (er *sdl_TextureCacheEntryRune) SetAllVisible(s bool) {
	n := er
	for n != nil {
		n.visible = s
		n = n.next
	}
}

func (er *sdl_TextureCacheEntryRune) Inside(x int32) bool {
	return (x >= er.offset) && (x <= er.offset+er.width)
}

func (er *sdl_TextureCacheEntryRune) Count() int {
	c := 0
	n := er
	for n != nil {
		n = n.next
		c++
	}
	return c
}

func (er *sdl_TextureCacheEntryRune) Indexed(index int) *sdl_TextureCacheEntryRune {
	if index == 0 {
		return er
	}
	n := er
	for n != nil {
		index--
		if index == 0 {
			return n
		}
		n = n.next
	}
	return nil
}

func parseColourString(cs string) (*sdl.Color, error) {
	vs := strings.Split(cs, ",")
	if len(vs) < 3 {
		return nil, fmt.Errorf("invalid colour. Expecting 3 or 4 byte values (for R,G,B,A) e.g. 255,255,255. found '%s'", cs)
	}
	vi := make([]uint8, 4)
	for i := 0; i < 4; i++ {
		vi[i] = 255
	}
	for i, s := range vs {
		si, err := strconv.Atoi(s)
		if err != nil || si < 0 || si > 255 {
			return nil, fmt.Errorf("invalid colour byte. Expecting int value from 0 to 255. Found '%s'", s)
		}
		vi[i] = uint8(si)
	}
	return &sdl.Color{R: vi[0], G: vi[1], B: vi[2], A: vi[3]}, nil
}
