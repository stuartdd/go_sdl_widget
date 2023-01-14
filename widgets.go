package go_sdl_widget

import (
	"fmt"
	"math"
	"math/rand"

	"github.com/veandco/go-sdl2/gfx"
	"github.com/veandco/go-sdl2/sdl"
	"github.com/veandco/go-sdl2/ttf"
)

/****************************************************************************************
* SDL_Shape code.
* Implements SDL_Widget cos it is one!
**/
type SDL_Shape struct {
	SDL_WidgetBase
	validRect *sdl.Rect // If the state of the shape has changed this should be nil.
	vxIn      []int16
	vyIn      []int16
	vxOut     []int16
	vyOut     []int16
}

var _ SDL_Widget = (*SDL_Shape)(nil) // Ensure SDL_Button 'is a' SDL_Widget

func NewSDLShape(x, y, w, h, id int32, style STATE_BITS, onClick func(string, int32, int32, int32) bool) *SDL_Shape {
	shape := &SDL_Shape{vxIn: make([]int16, 0), vyIn: make([]int16, 0), validRect: nil}
	shape.SDL_WidgetBase = initBase(x, y, w, h, id, shape, 0, false, style, onClick)
	return shape
}

func NewSDLShapeArrowRight(x, y, w, h, id int32, style STATE_BITS, onClick func(string, int32, int32, int32) bool) *SDL_Shape {
	sh := NewSDLShape(x, y, w, h, id, style, onClick)
	var halfH int32 = h / 2
	var qtr1H int32 = h / 4
	var thrd1W int32 = w / 6
	var thrd2W int32 = thrd1W * 4
	sh.Add(thrd1W, -qtr1H)
	sh.Add(thrd2W, -qtr1H)
	sh.Add(thrd2W, -halfH)
	sh.Add(w, 0)
	sh.Add(thrd2W, +halfH)
	sh.Add(thrd2W, +qtr1H)
	sh.Add(thrd1W, +qtr1H)
	return sh
}

func (s *SDL_Shape) SetPosition(x, y int32) bool {
	b := s.SDL_WidgetBase.SetPosition(x, y)
	if b {
		s.validRect = nil
	}
	return b
}

func (s *SDL_Shape) SetSize(w, h int32) bool {
	b := s.SDL_WidgetBase.SetSize(w, h)
	if b {
		s.validRect = nil
	}
	return b
}

func (s *SDL_Shape) Scale(sc float32) {
	s.SDL_WidgetBase.Scale(sc)
	for i := 0; i < len(s.vxIn); i++ {
		s.vxIn[i] = int16(float32(s.vxIn[i]) * sc)
		s.vyIn[i] = int16(float32(s.vyIn[i]) * sc)
	}
	s.validRect = nil
}

func (s *SDL_Shape) Add(x, y int32) {
	s.vxIn = append(s.vxIn, int16(x))
	s.vyIn = append(s.vyIn, int16(y))
	s.validRect = nil
}

func (s *SDL_Shape) Draw(renderer *sdl.Renderer, font *ttf.Font) error {
	if s.IsVisible() {
		s.GetRect() // Make sure we update the Out Arrays is the state of the shape was changed
		if s.ShouldDrawBackground() {
			gfx.FilledPolygonColor(renderer, s.vxOut, s.vyOut, *s.GetBackground())
		}
		if s.ShouldDrawBorder() {
			gfx.PolygonColor(renderer, s.vxOut, s.vyOut, *s.GetBorderColour())
		}
	}
	return nil
}

func (s *SDL_Shape) Rotate(angle int) {
	rad := float64(angle) * DEG_TO_RAD
	var px float64 = 0
	var py float64 = 0
	sinA := math.Sin(rad)
	cosA := math.Cos(rad)
	for i := 0; i < len(s.vxIn); i++ {
		px = float64(s.vxIn[i])
		py = float64(s.vyIn[i])
		s.vxIn[i] = int16(cosA*px - sinA*py)
		s.vyIn[i] = int16(sinA*px + cosA*py)
	}
	s.validRect = nil
}

func (s *SDL_Shape) GetRect() *sdl.Rect {
	if s.validRect == nil {
		count := len(s.vxIn)
		vxOut := make([]int16, count)
		vyOut := make([]int16, count)
		x := int16(s.x)
		y := int16(s.y)
		for i := 0; i < count; i++ {
			vxOut[i] = x + s.vxIn[i]
			vyOut[i] = y + s.vyIn[i]
		}

		var minx int16 = math.MaxInt16
		var miny int16 = math.MaxInt16
		var maxx int16 = math.MinInt16
		var maxy int16 = math.MinInt16
		for i := 0; i < count; i++ {
			if vxOut[i] < minx {
				minx = vxOut[i]
			}
			if vxOut[i] > maxx {
				maxx = vxOut[i]
			}
			if vyOut[i] < miny {
				miny = vyOut[i]
			}
			if vyOut[i] > maxy {
				maxy = vyOut[i]
			}
		}
		s.vxOut = vxOut
		s.vyOut = vyOut
		s.validRect = &sdl.Rect{X: int32(minx), Y: int32(miny), W: int32(maxx - minx), H: int32(maxy - miny)}
	}
	return s.validRect
}

/****************************************************************************************
* SDL_Label code
* Implements SDL_Widget cos it is one!
* Implements SDL_TextWidget because it has text and uses the texture cache
**/
type SDL_Label struct {
	SDL_WidgetBase
	text     string
	cacheKey string
	align    ALIGN_TEXT
}

var _ SDL_TextWidget = (*SDL_Label)(nil) // Ensure SDL_Button 'is a' SDL_TextWidget
var _ SDL_Widget = (*SDL_Label)(nil)     // Ensure SDL_Button 'is a' SDL_Widget

func NewSDLLabel(x, y, w, h, id int32, text string, align ALIGN_TEXT, style STATE_BITS) *SDL_Label {
	but := &SDL_Label{text: text, align: align, cacheKey: fmt.Sprintf("label:%d:%d", id, rand.Intn(100))}
	but.SDL_WidgetBase = initBase(x, y, w, h, id, but, 0, false, style, nil)
	return but
}

func (b *SDL_Label) SetText(text string) {
	if b.text != text {
		b.text = text
	}
}

func (b *SDL_Label) GetText() string {
	return b.text
}

func (b *SDL_Label) Draw(renderer *sdl.Renderer, font *ttf.Font) error {
	if b.IsVisible() {
		cachedTexture, err := GetResourceInstance().UpdateTextureFromString(renderer, b.cacheKey, b.text, font, b.GetForeground())
		if err != nil {
			renderer.SetDrawColor(255, 0, 0, 255)
			renderer.DrawRect(&sdl.Rect{X: b.x, Y: b.y, W: b.w, H: b.h})
			return nil
		}
		if b.align == ALIGN_FIT {
			b.SetSize(cachedTexture.w, b.h)
		}
		tm := (b.h / 9)
		rw, sw, sh := cachedTexture.ScaledWidthHeight(b.h-(2*tm), b.w)

		al := b.align
		if sw > (b.w - 10) {
			al = ALIGN_LEFT
			sw = b.w - 10
		}
		var tx int32

		switch al {
		case ALIGN_CENTER:
			tx = (b.w - sw) / 2
		case ALIGN_LEFT:
			tx = 0
		case ALIGN_RIGHT:
			tx = (b.x + b.w) - sw
		}

		if b.ShouldDrawBackground() {
			bc := b.GetBackground()
			renderer.SetDrawColor(bc.R, bc.G, bc.B, bc.A)
			renderer.FillRect(&sdl.Rect{X: b.x, Y: b.y, W: b.w, H: b.h})
		}
		renderer.Copy(
			cachedTexture.texture,
			&sdl.Rect{X: 0, Y: 0, W: rw, H: cachedTexture.h},
			&sdl.Rect{X: b.x + tx, Y: b.y + tm, W: sw, H: sh})
		if b.ShouldDrawBorder() {
			bc := b.GetBorderColour()
			renderer.SetDrawColor(bc.R, bc.G, bc.B, bc.A)
			renderer.DrawRect(&sdl.Rect{X: b.x + 1, Y: b.y + 1, W: b.w - 2, H: b.h - 2})
			renderer.DrawRect(&sdl.Rect{X: b.x + 2, Y: b.y + 2, W: b.w - 4, H: b.h - 4})
		}
	}
	return nil
}

/****************************************************************************************
* SDL_Button code
* Implements SDL_Widget cos it is one!
* Implements SDL_TextWidget because it has text and uses the texture cache
**/
type SDL_Button struct {
	SDL_WidgetBase
	text            string
	backgroundImage string
}

var _ SDL_TextWidget = (*SDL_Button)(nil) // Ensure SDL_Button 'is a' SDL_TextWidget
var _ SDL_Widget = (*SDL_Button)(nil)     // Ensure SDL_Button 'is a' SDL_Widget

func NewSDLButton(x, y, w, h, id int32, text string, style STATE_BITS, deBounce int, onClick func(string, int32, int32, int32) bool) *SDL_Button {
	but := &SDL_Button{text: text, backgroundImage: ""}
	but.SDL_WidgetBase = initBase(x, y, w, h, id, but, deBounce, false, style, onClick)
	return but
}

func (b *SDL_Button) SetText(text string) {
	b.text = text
}

func (b *SDL_Button) GetText() string {
	return b.text
}

func (b *SDL_Button) SetBackgroundImage(backgroundImage string) {
	b.backgroundImage = backgroundImage
}

func (b *SDL_Button) GetBackgroundImage() string {
	return b.backgroundImage
}

func (b *SDL_Button) Draw(renderer *sdl.Renderer, font *ttf.Font) error {
	if b.IsVisible() {
		var image *sdl.Texture
		var imagew, imageh int32
		var err error
		if b.backgroundImage != "" {
			tn := b.backgroundImage
			if !b.IsEnabled() {
				tn = tn + ".dis"
			}
			image, imagew, imageh, err = GetResourceInstance().GetTextureForName(tn)
			if err != nil {
				image, imagew, imageh, err = GetResourceInstance().GetTextureForName(b.backgroundImage)
				if err != nil {
					renderer.SetDrawColor(255, 0, 0, 255)
					renderer.DrawRect(&sdl.Rect{X: b.x, Y: b.y, W: ifZeroUseN(b.w, 200), H: ifZeroUseN(b.h, 200)})
					return nil
				}
			}
			if b.w <= 0 {
				b.w = imagew
			}
			if b.h <= 0 {
				b.h = imageh
			}
		}
		cacheKey := fmt.Sprintf("%s.%s.%t", TEXTURE_CACHE_TEXT_PREF, b.text, b.IsEnabled() && !b.IsClicked())
		ctwe, err := GetResourceInstance().UpdateTextureFromString(renderer, cacheKey, b.text, font, b.GetForeground())
		if err != nil {
			renderer.SetDrawColor(255, 0, 0, 255)
			renderer.DrawRect(&sdl.Rect{X: b.x, Y: b.y, W: ifZeroUseN(b.w, 200), H: ifZeroUseN(b.h, 200)})
			return nil
		}
		if b.w <= 0 {
			b.w = ctwe.w
		}
		if b.h <= 0 {
			b.h = ctwe.h
		}
		if b.ShouldDrawBackground() {
			bc := b.GetBackground()
			renderer.SetDrawColor(bc.R, bc.G, bc.B, bc.A)
			renderer.FillRect(&sdl.Rect{X: b.x, Y: b.y, W: b.w, H: b.h})
		}
		if image != nil {
			renderer.Copy(image, nil, &sdl.Rect{X: b.x, Y: b.y, W: b.w, H: b.h})
		}
		// Center the text inside the scaled button
		bh := float32(b.h)
		th := int32(bh - (bh / 4))
		tw := int32(float32(ctwe.w) * (bh / float32(ctwe.h)))
		tx := (b.w - tw) / 2
		ty := (b.h - th) / 2
		renderer.Copy(ctwe.texture, nil, &sdl.Rect{X: b.x + tx, Y: b.y + ty, W: tw, H: th})
		if b.ShouldDrawBorder() && b.backgroundImage == "" {
			bc := b.GetBorderColour()
			renderer.SetDrawColor(bc.R, bc.G, bc.B, bc.A)
			renderer.DrawRect(&sdl.Rect{X: b.x + 1, Y: b.y + 1, W: b.w - 2, H: b.h - 2})
			renderer.DrawRect(&sdl.Rect{X: b.x + 2, Y: b.y + 2, W: b.w - 4, H: b.h - 4})
		}
	}
	return nil
}

/****************************************************************************************
* SDL_Image code
* Implements SDL_Widget cos it is one!
* Implements SDL_TextWidget because it has text and uses the texture cache
**/
type SDL_Image struct {
	SDL_WidgetBase
	textureName string
	frame       int32
	frameCount  int32
}

var _ SDL_ImageWidget = (*SDL_Image)(nil) // Ensure SDL_Image 'is a' SDL_ImageWidget
var _ SDL_Widget = (*SDL_Image)(nil)      // Ensure SDL_Image 'is a' SDL_Widget

func NewSDLImage(x, y, w, h, id int32, textureName string, frame, frameCount int32, style STATE_BITS, deBounce int, onClick func(string, int32, int32, int32) bool) *SDL_Image {
	but := &SDL_Image{textureName: textureName, frame: frame, frameCount: frameCount}
	but.SDL_WidgetBase = initBase(x, y, w, h, id, but, deBounce, false, style, onClick)
	return but
}

func (im *SDL_Image) String() string {
	return im.textureName
}

func (im *SDL_Image) SetFrame(tf int32) {
	if tf >= im.frameCount {
		tf = 0
	}
	im.frame = tf
}

func (im *SDL_Image) GetFrame() int32 {
	return im.frame
}

func (im *SDL_Image) NextFrame() int32 {
	im.frame++
	if im.frame >= im.frameCount {
		im.frame = 0
	}
	return im.frame
}

func (im *SDL_Image) GetFrameCount() int32 {
	return im.frameCount
}

func (im *SDL_Image) Draw(renderer *sdl.Renderer, font *ttf.Font) error {
	if im.IsVisible() {
		tn := im.textureName
		if !im.IsEnabled() {
			tn = tn + ".dis"
		}
		image, irw, irh, err := GetResourceInstance().GetTextureForName(tn)
		if err != nil {
			image, irw, irh, err = GetResourceInstance().GetTextureForName(im.textureName)
			if err != nil {
				renderer.SetDrawColor(255, 0, 0, 255)
				renderer.DrawRect(&sdl.Rect{X: im.x, Y: im.y, W: ifZeroUseN(im.w, 200), H: ifZeroUseN(im.h, 200)})
				return nil
			}
		}
		if im.w <= 0 {
			im.w = irw
		}
		if im.h <= 0 {
			im.h = irh
		}
		outRect := &sdl.Rect{X: im.x, Y: im.y, W: im.w, H: im.h}
		borderRect := &sdl.Rect{X: im.x, Y: im.y, W: im.w, H: im.h}
		if im.ShouldDrawBorder() {
			outRect = widgetShrinkRect(outRect, 4)
		}
		if im.ShouldDrawBackground() {
			bc := im.GetBackground()
			renderer.SetDrawColor(bc.R, bc.G, bc.B, bc.A)
			renderer.FillRect(borderRect)
		}
		if im.frameCount > 1 {
			w := (irw / im.frameCount)
			x := (w * im.frame)
			inRect := &sdl.Rect{X: x, Y: 0, W: w, H: outRect.H}
			outRect := &sdl.Rect{X: outRect.X, Y: outRect.Y, W: w, H: outRect.H}
			renderer.Copy(image, inRect, outRect)
		} else {
			renderer.Copy(image, nil, outRect)
		}
		if im.ShouldDrawBorder() {
			bc := im.GetBorderColour()
			renderer.SetDrawColor(bc.R, bc.G, bc.B, bc.A)
			renderer.DrawRect(&sdl.Rect{X: im.x + 1, Y: im.y + 1, W: im.w - 2, H: im.h - 2})
			renderer.DrawRect(&sdl.Rect{X: im.x + 2, Y: im.y + 2, W: im.w - 4, H: im.h - 4})
		}
	}
	return nil
}

/****************************************************************************************
* SDL_Image code
* Implements SDL_Widget cos it is one!
* Implements SDL_TextWidget because it has text and uses the texture cache
**/
type SDL_Separator struct {
	SDL_WidgetBase
}

var _ SDL_Widget = (*SDL_Separator)(nil) // Ensure SDL_Button 'is a' SDL_Widget

func NewSDLSeparator(x, y, w, h, id int32, style STATE_BITS) *SDL_Separator {
	but := &SDL_Separator{}
	but.SDL_WidgetBase = initBase(x, y, w, h, id, but, 0, false, style, nil)
	return but
}

func (sep *SDL_Separator) Draw(renderer *sdl.Renderer, font *ttf.Font) error {
	if sep.IsEnabled() {
		if sep.ShouldDrawBackground() {
			bc := sep.GetBackground()
			renderer.SetDrawColor(bc.R, bc.G, bc.B, bc.A)
			renderer.FillRect(&sdl.Rect{X: sep.x, Y: sep.y, W: sep.w, H: sep.h})
		}
		if sep.ShouldDrawBorder() {
			bc := sep.GetBorderColour()
			renderer.SetDrawColor(bc.R, bc.G, bc.B, bc.A)
			renderer.DrawRect(&sdl.Rect{X: sep.x + 1, Y: sep.y + 1, W: sep.w - 2, H: sep.h - 2})
			renderer.DrawRect(&sdl.Rect{X: sep.x + 2, Y: sep.y + 2, W: sep.w - 4, H: sep.h - 4})
		}
	}
	return nil
}

func ifZeroUseN(i, n int32) int32 {
	if i == 0 {
		return n
	}
	return n
}
