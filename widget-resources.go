package go_sdl_widget

import (
	"sync"

	"github.com/veandco/go-sdl2/sdl"
	"github.com/veandco/go-sdl2/ttf"
)

type sdl_Resources struct {
	font             *ttf.Font
	textureCache     *SDL_TextureCache
	backgroundColor  *sdl.Color
	foregroundColor  *sdl.Color
	borderColor      *sdl.Color
	borderFocusColor *sdl.Color
}

var sdlResourceInstanceLock = &sync.Mutex{}
var sdlResourceInstance *sdl_Resources

func GetResourceInstance() *sdl_Resources {
	if sdlResourceInstance == nil {
		sdlResourceInstanceLock.Lock()
		defer sdlResourceInstanceLock.Unlock()
		if sdlResourceInstance == nil {
			sdlResourceInstance = &sdl_Resources{}
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
func (r *sdl_Resources) SetBackgroundColor(backgroundColor *sdl.Color) {
	r.backgroundColor = backgroundColor
}
func (r *sdl_Resources) SetForegroundColor(foregroundColor *sdl.Color) {
	r.foregroundColor = foregroundColor
}
func (r *sdl_Resources) SetBorderColor(borderColor *sdl.Color) {
	r.borderColor = borderColor
}
func (r *sdl_Resources) SetBorderFocusColor(borderFocusColor *sdl.Color) {
	r.borderFocusColor = borderFocusColor
}

func (r *sdl_Resources) GetFont() *ttf.Font {
	return r.font
}
func (r *sdl_Resources) GetTextureCache() *SDL_TextureCache {
	return r.textureCache
}
func (r *sdl_Resources) GetBackgroundColor() *sdl.Color {
	return r.backgroundColor
}
func (r *sdl_Resources) GetForegroundColor() *sdl.Color {
	return r.foregroundColor
}
func (r *sdl_Resources) GetBorderColor() *sdl.Color {
	return r.borderColor
}
func (r *sdl_Resources) GetBorderFocusColor() *sdl.Color {
	return r.borderFocusColor
}
