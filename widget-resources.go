package go_sdl_widget

import (
	"fmt"
	"path/filepath"
	"sync"

	"github.com/veandco/go-sdl2/img"
	"github.com/veandco/go-sdl2/sdl"
	"github.com/veandco/go-sdl2/ttf"
)

type sdl_Resources struct {
	font               *ttf.Font
	textureCache       *SDL_TextureCache
	cacheLock          sync.Mutex
	colours            [][]*sdl.Color
	cursorInsertColour *sdl.Color
	cursorAppendColour *sdl.Color
	selecteCharsFwd    []byte
	selecteCharsRev    []byte
}

var sdlResourceInstanceLock = &sync.Mutex{}
var sdlResourceInstance *sdl_Resources

func GetResourceInstance() *sdl_Resources {
	if sdlResourceInstance == nil {
		sdlResourceInstanceLock.Lock()
		defer sdlResourceInstanceLock.Unlock()
		if sdlResourceInstance == nil {
			sdlResourceInstance = &sdl_Resources{
				textureCache: NewTextureCache(),
				colours:      make([][]*sdl.Color, 4),
			}
			for i := 0; i < 4; i++ {
				sdlResourceInstance.colours[i] = make([]*sdl.Color, WIDGET_COLOR_MAX)
			}
			sdlResourceInstance.colours[WIDGET_COLOUR_ENABLED][WIDGET_COLOR_FG] = &sdl.Color{R: 0, G: 255, B: 0, A: 255}
			sdlResourceInstance.colours[WIDGET_COLOUR_ENABLED][WIDGET_COLOR_BG] = &sdl.Color{R: 0, G: 100, B: 0, A: 255}
			sdlResourceInstance.colours[WIDGET_COLOUR_ENABLED][WIDGET_COLOR_BORDER] = &sdl.Color{R: 0, G: 255, B: 0, A: 255}
			sdlResourceInstance.colours[WIDGET_COLOUR_ENABLED][WIDGET_COLOR_ENTRY] = &sdl.Color{R: 0, G: 0, B: 255, A: 255}

			sdlResourceInstance.colours[WIDGET_COLOUR_DISABLE][WIDGET_COLOR_FG] = &sdl.Color{R: 0, G: 100, B: 0, A: 255}
			sdlResourceInstance.colours[WIDGET_COLOUR_DISABLE][WIDGET_COLOR_BG] = &sdl.Color{R: 0, G: 50, B: 0, A: 255}
			sdlResourceInstance.colours[WIDGET_COLOUR_DISABLE][WIDGET_COLOR_BORDER] = &sdl.Color{R: 0, G: 150, B: 0, A: 255}
			sdlResourceInstance.colours[WIDGET_COLOUR_DISABLE][WIDGET_COLOR_ENTRY] = &sdl.Color{R: 0, G: 0, B: 150, A: 255}

			sdlResourceInstance.colours[WIDGET_COLOUR_FOCUS][WIDGET_COLOR_FG] = &sdl.Color{R: 0, G: 255, B: 0, A: 255}
			sdlResourceInstance.colours[WIDGET_COLOUR_FOCUS][WIDGET_COLOR_BG] = &sdl.Color{R: 0, G: 0, B: 150, A: 255}
			sdlResourceInstance.colours[WIDGET_COLOUR_FOCUS][WIDGET_COLOR_BORDER] = &sdl.Color{R: 255, G: 0, B: 0, A: 255}
			sdlResourceInstance.colours[WIDGET_COLOUR_FOCUS][WIDGET_COLOR_ENTRY] = &sdl.Color{R: 0, G: 0, B: 255, A: 255}

			sdlResourceInstance.colours[WIDGET_COLOUR_ERROR][WIDGET_COLOR_FG] = &sdl.Color{R: 0, G: 255, B: 0, A: 255}
			sdlResourceInstance.colours[WIDGET_COLOUR_ERROR][WIDGET_COLOR_BG] = &sdl.Color{R: 150, G: 0, B: 0, A: 255}
			sdlResourceInstance.colours[WIDGET_COLOUR_ERROR][WIDGET_COLOR_BORDER] = &sdl.Color{R: 255, G: 0, B: 0, A: 255}
			sdlResourceInstance.colours[WIDGET_COLOUR_ERROR][WIDGET_COLOR_ENTRY] = &sdl.Color{R: 0, G: 0, B: 255, A: 255}

			sdlResourceInstance.cursorInsertColour = &sdl.Color{R: 255, G: 255, B: 255, A: 255}
			sdlResourceInstance.cursorAppendColour = &sdl.Color{R: 255, G: 0, B: 255, A: 255}

			sdlResourceInstance.SetSelecteCharsFwd("/.")
			sdlResourceInstance.SetSelecteCharsRev("/")
		}
	}

	return sdlResourceInstance
}

func (r *sdl_Resources) SetFont(font *ttf.Font) {
	r.font = font
}

func (r *sdl_Resources) SetTextureCache(textureCache *SDL_TextureCache) {
	r.textureCache = textureCache
}

func (r *sdl_Resources) SetColour(stateIndex, colorIndex int, c *sdl.Color) {
	r.colours[stateIndex][colorIndex] = c
}

func (b *sdl_Resources) GetCursorInsertColour() *sdl.Color {
	return b.cursorInsertColour
}

func (b *sdl_Resources) SetCursorInsertColour(c *sdl.Color) {
	b.cursorInsertColour = c
}

func (b *sdl_Resources) GetCursorAppendColour() *sdl.Color {
	return b.cursorAppendColour
}

func (b *sdl_Resources) SetCursorAppendColour(c *sdl.Color) {
	b.cursorAppendColour = c
}

func (r *sdl_Resources) GetFont() *ttf.Font {
	return r.font
}

func (b *sdl_Resources) SetSelecteCharsFwd(s string) {
	b.selecteCharsFwd = []byte(s)
}

func (b *sdl_Resources) GetSelecteCharsFwd() string {
	return string(b.selecteCharsFwd)
}

func (b *sdl_Resources) SetSelecteCharsRev(s string) {
	b.selecteCharsRev = []byte(s)
}

func (b *sdl_Resources) GetSelecteCharsRev() string {
	return string(b.selecteCharsRev)
}
func (r *sdl_Resources) GetTextureCache() *SDL_TextureCache {
	return r.textureCache
}

func (r *sdl_Resources) GetColour(stateIndex STATE_COLOUR, colorIndex STYLE_COLOUR) *sdl.Color {
	return r.colours[stateIndex][colorIndex]
}

func (r *sdl_Resources) GetTextureForName(name string) (*sdl.Texture, int32, int32, error) {
	r.cacheLock.Lock()
	defer r.cacheLock.Unlock()
	tce := r.textureCache._textureMap[name]
	if tce == nil {
		return nil, 0, 0, fmt.Errorf("texture cache does not contain %s", name)
	}
	return tce.Texture, tce.W, tce.H, nil
}

func (r *sdl_Resources) AddTexturesFromStringMap(renderer *sdl.Renderer, textMap map[string]string, font *ttf.Font, colour *sdl.Color) error {
	r.cacheLock.Lock()
	defer r.cacheLock.Unlock()
	for name, text := range textMap {
		tce, err := newTextureCacheEntryForString(renderer, text, font, colour)
		if err != nil {
			return err
		}
		r.textureCache.Add(name, tce)
	}
	return nil
}

func (r *sdl_Resources) AddTexturesFromFileMap(renderer *sdl.Renderer, applicationDataPath string, fileNames map[string]string) error {
	r.cacheLock.Lock()
	defer r.cacheLock.Unlock()
	for name, fileName := range fileNames {
		var fn string
		if applicationDataPath == "" {
			fn = fileName
		} else {
			fn = filepath.Join(applicationDataPath, fileName)
		}
		tce, err := newTextureCacheEntryForFile(renderer, fn)
		if err != nil {
			return fmt.Errorf("file '%s':%s", fileName, err.Error())
		}
		r.textureCache.Add(name, tce)
	}
	return nil
}

func (r *sdl_Resources) GetTextureListFromCachedRunes(text string, colour *sdl.Color) []*SDL_TextureCacheEntry {
	list := make([]*SDL_TextureCacheEntry, len(text))
	for i, c := range text {
		ec := r.textureCache._textureMap[fmt.Sprintf("|%c%d", c, GetColourId(colour))]
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

	for i, c := range text {
		tce := r.textureCache._textureMap[fmt.Sprintf("|%c%d", c, GetColourId(colour))]
		if tce == nil {
			return rootEnt
		}
		sw = float32(height) * (float32(tce.W) / float32(tce.H))
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
	gtwe, ok := r.textureCache._textureMap[cacheKey]
	if ok {
		if gtwe.value == text {
			return gtwe, nil
		}
	}
	ctwe, err := newTextureCacheEntryForString(renderer, text, font, colour)
	if err != nil {
		return nil, err
	}
	r.textureCache.Add(cacheKey, ctwe)
	return ctwe, nil
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
	_textureMap map[string]*SDL_TextureCacheEntry
	in, out     int
}

func NewTextureCache() *SDL_TextureCache {
	return &SDL_TextureCache{_textureMap: make(map[string]*SDL_TextureCacheEntry), in: 0, out: 0}
}

func (tc *SDL_TextureCache) String() string {
	return fmt.Sprintf("TextureCache in:%d out:%d", tc.in, tc.out)
}

func (tc *SDL_TextureCache) Peek(name string) bool {
	_, ok := tc._textureMap[name]
	return ok
}

func (tc *SDL_TextureCache) Add(name string, tceIn *SDL_TextureCacheEntry) {
	tce := tc._textureMap[name]
	if tce != nil {
		tc.out = tc.out + tce.Destroy()
	}
	tc._textureMap[name] = tceIn
	tc.in++
}

func (tc *SDL_TextureCache) Remove(name string, tceIn *SDL_TextureCacheEntry) {
	tce := tc._textureMap[name]
	if tce != nil {
		tc.out = tc.out + tce.Destroy()
	}
	tc._textureMap[name] = nil
}

func (tc *SDL_TextureCache) Destroy() {
	for n, tce := range tc._textureMap {
		if tce != nil {
			tc.out = tc.out + tce.Destroy()
		}
		tc._textureMap[n] = nil
	}
}

/****************************************************************************************
* Texture cache Entry used to hold ALL textures in the SDL_TextureCache
**/
type SDL_TextureCacheEntry struct {
	Texture *sdl.Texture
	value   string
	W, H    int32
}

func (tce *SDL_TextureCacheEntry) Destroy() int {
	if tce.Texture != nil {
		tce.Texture.Destroy()
		tce.value = ""
		return 1
	}
	return 0
}

func newTextureCacheEntryForFile(renderer *sdl.Renderer, fileName string) (*SDL_TextureCacheEntry, error) {
	texture, err := img.LoadTexture(renderer, fileName)
	if err != nil {
		return nil, err
	}
	_, _, t3, t4, err := texture.Query()
	if err != nil {
		return nil, err
	}
	return &SDL_TextureCacheEntry{Texture: texture, W: t3, H: t4, value: fileName}, nil
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
	return &SDL_TextureCacheEntry{Texture: txt, value: string(char), W: clip.W, H: clip.H}, nil
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
	return &SDL_TextureCacheEntry{Texture: txt, W: clip.W, H: clip.H, value: text}, nil
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
