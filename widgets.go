package go_sdl_widget

import (
	"fmt"
	"math"
	"math/rand"
	"time"

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
	onClick   func(SDL_Widget, int32, int32) bool
}

var _ SDL_Widget = (*SDL_Shape)(nil) // Ensure SDL_Button 'is a' SDL_Widget

func NewSDLShape(x, y, w, h, id int32, style STATE_BITS, onClick func(SDL_Widget, int32, int32) bool) *SDL_Shape {
	shape := &SDL_Shape{vxIn: make([]int16, 0), vyIn: make([]int16, 0), validRect: nil, onClick: onClick}
	shape.SDL_WidgetBase = initBase(x, y, w, h, id, 0, style)
	return shape
}

func NewSDLShapeArrowRight(x, y, w, h, id int32, style STATE_BITS, onClick func(SDL_Widget, int32, int32) bool) *SDL_Shape {
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

func (b *SDL_Shape) Click(md *SDL_MouseData) bool {
	if b.IsEnabled() && b.onClick != nil {
		if b.deBounce > 0 {
			b.SetClicked(true)
			defer func() {
				time.Sleep(time.Millisecond * time.Duration(b.deBounce))
				b.SetClicked(false)
			}()
		}
		return b.onClick(b, md.x, md.y)
	}
	return false
}

func (b *SDL_Shape) Destroy() {
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

func (s *SDL_Shape) Inside(x, y int32) bool {
	if s.IsVisible() {
		return isInsideRect(x, y, s.GetRect())
	}
	return false
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
	text         string
	cacheKey     string
	cacheInvalid bool
	align        ALIGN_TEXT
}

var _ SDL_Widget = (*SDL_Label)(nil)     // Ensure SDL_Button 'is a' SDL_Widget
var _ SDL_TextWidget = (*SDL_Label)(nil) // Ensure SDL_Button 'is a' SDL_TextWidget

func NewSDLLabel(x, y, w, h, id int32, text string, align ALIGN_TEXT, style STATE_BITS) *SDL_Label {
	but := &SDL_Label{text: text, cacheInvalid: true, align: align, cacheKey: fmt.Sprintf("label:%d:%d", id, rand.Intn(100))}
	but.SDL_WidgetBase = initBase(x, y, w, h, id, 0, style)
	return but
}

func (b *SDL_Label) SetForeground(c *sdl.Color) {
	if b.foreground != c {
		b.SDL_WidgetBase.SetForeground(c)
		b.cacheInvalid = true
	}
}

func (b *SDL_Label) SetText(text string) {
	if b.text != text {
		b.cacheInvalid = true
		b.text = text
	}
}

func (b *SDL_Label) SetEnabled(e bool) {
	if b.IsEnabled() != e {
		b.cacheInvalid = true
		b.SDL_WidgetBase.SetEnabled(e)
	}
}

func (b *SDL_Label) GetText() string {
	return b.text
}

func (b *SDL_Label) Click(md *SDL_MouseData) bool {
	return false
}

func (b *SDL_Label) Draw(renderer *sdl.Renderer, font *ttf.Font) error {
	if b.IsVisible() {
		ctwe, err := GetResourceInstance().UpdateTextureFromString(renderer, b.cacheKey, b.text, font, b.GetForeground())
		if err != nil {
			renderer.SetDrawColor(255, 0, 0, 255)
			renderer.DrawRect(&sdl.Rect{X: b.x, Y: b.y, W: b.w, H: b.h})
			return nil
		}
		if b.align == ALIGN_FIT {
			b.SetSize(ctwe.W, b.h)
		}

		bh := float32(b.h)
		th := int32(bh - (bh / 4))
		tw := int32(float32(ctwe.W) * (bh / float32(ctwe.H)))
		var tx int32

		switch b.align {
		case ALIGN_CENTER:
			tx = (b.w - tw) / 2
		case ALIGN_LEFT:
			tx = 10
		case ALIGN_RIGHT:
			tx = (b.x + b.w) - tw
		}
		ty := (b.h - th) / 2

		if b.ShouldDrawBackground() {
			bc := b.GetBackground()
			renderer.SetDrawColor(bc.R, bc.G, bc.B, bc.A)
			renderer.FillRect(&sdl.Rect{X: b.x, Y: b.y, W: b.w, H: b.h})
		}
		renderer.Copy(ctwe.Texture, nil, &sdl.Rect{X: b.x + tx, Y: b.y + ty, W: tw, H: th})
		if b.ShouldDrawBorder() {
			bc := b.GetBorderColour()
			renderer.SetDrawColor(bc.R, bc.G, bc.B, bc.A)
			renderer.DrawRect(&sdl.Rect{X: b.x + 1, Y: b.y + 1, W: b.w - 2, H: b.h - 2})
			renderer.DrawRect(&sdl.Rect{X: b.x + 2, Y: b.y + 2, W: b.w - 4, H: b.h - 4})
		}
	}
	return nil
}
func (b *SDL_Label) Destroy() {
	// Image cache takes care of all images!
}

/****************************************************************************************
* SDL_Button code
* Implements SDL_Widget cos it is one!
* Implements SDL_TextWidget because it has text and uses the texture cache
**/
type SDL_Button struct {
	SDL_WidgetBase
	text    string
	onClick func(SDL_Widget, int32, int32) bool
}

var _ SDL_Widget = (*SDL_Button)(nil)     // Ensure SDL_Button 'is a' SDL_Widget
var _ SDL_TextWidget = (*SDL_Button)(nil) // Ensure SDL_Button 'is a' SDL_TextWidget

func NewSDLButton(x, y, w, h, id int32, text string, style STATE_BITS, deBounce int, onClick func(SDL_Widget, int32, int32) bool) *SDL_Button {
	but := &SDL_Button{text: text, onClick: onClick}
	but.SDL_WidgetBase = initBase(x, y, w, h, id, deBounce, style)
	return but
}

func (b *SDL_Button) SetOnClick(f func(SDL_Widget, int32, int32) bool) {
	b.onClick = f
}

func (b *SDL_Button) SetText(text string) {
	b.text = text
}

func (b *SDL_Button) GetText() string {
	return b.text
}

func (b *SDL_Button) Click(md *SDL_MouseData) bool {
	if b.IsEnabled() && b.onClick != nil {
		if b.deBounce > 0 {
			b.SetClicked(true)
			fmt.Println("Clicked")
			defer func() {
				time.Sleep(time.Millisecond * time.Duration(b.deBounce))
				b.SetClicked(false)
				fmt.Println("un-Clicked")
			}()
		}
		return b.onClick(b, md.x, md.y)
	}
	return false
}

func (b *SDL_Button) Destroy() {

}

func (b *SDL_Button) Draw(renderer *sdl.Renderer, font *ttf.Font) error {
	if b.IsVisible() {
		cacheKey := fmt.Sprintf("%s.%s.%t", TEXTURE_CACHE_TEXT_PREF, b.text, b.IsEnabled() && !b.IsClicked())
		ctwe, err := GetResourceInstance().UpdateTextureFromString(renderer, cacheKey, b.text, font, b.GetForeground())
		if err != nil {
			renderer.SetDrawColor(255, 0, 0, 255)
			renderer.DrawRect(&sdl.Rect{X: b.x, Y: b.y, W: b.w, H: b.h})
			return nil
		}
		if b.ShouldDrawBackground() {
			bc := b.GetBackground()
			renderer.SetDrawColor(bc.R, bc.G, bc.B, bc.A)
			renderer.FillRect(&sdl.Rect{X: b.x, Y: b.y, W: b.w, H: b.h})
		}
		// Center the text inside the buttonj
		bh := float32(b.h)
		th := int32(bh - (bh / 4))
		tw := int32(float32(ctwe.W) * (bh / float32(ctwe.H)))
		tx := (b.w - tw) / 2
		ty := (b.h - th) / 2

		renderer.Copy(ctwe.Texture, nil, &sdl.Rect{X: b.x + tx, Y: b.y + ty, W: tw, H: th})
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
* SDL_Image code
* Implements SDL_Widget cos it is one!
* Implements SDL_TextWidget because it has text and uses the texture cache
**/
type SDL_Image struct {
	SDL_WidgetBase
	textureName string
	frame       int32
	frameCount  int32
	onClick     func(SDL_Widget, int32, int32) bool
}

var _ SDL_Widget = (*SDL_Image)(nil)      // Ensure SDL_Image 'is a' SDL_Widget
var _ SDL_ImageWidget = (*SDL_Image)(nil) // Ensure SDL_Image 'is a' SDL_ImageWidget

func NewSDLImage(x, y, w, h, id int32, textureName string, frame, frameCount int32, style STATE_BITS, deBounce int, onClick func(SDL_Widget, int32, int32) bool) *SDL_Image {
	but := &SDL_Image{textureName: textureName, frame: frame, frameCount: frameCount, onClick: onClick}
	but.SDL_WidgetBase = initBase(x, y, w, h, id, deBounce, style)
	return but
}

func (b *SDL_Image) SetFrame(tf int32) {
	if tf >= b.frameCount {
		tf = 0
	}
	b.frame = tf
}

func (b *SDL_Image) GetFrame() int32 {
	return b.frame
}

func (b *SDL_Image) NextFrame() int32 {
	b.frame++
	if b.frame >= b.frameCount {
		b.frame = 0
	}
	return b.frame
}

func (b *SDL_Image) GetFrameCount() int32 {
	return b.frameCount
}

func (b *SDL_Image) Click(md *SDL_MouseData) bool {
	if b.IsEnabled() && b.onClick != nil {
		if b.deBounce > 0 {
			b.SetClicked(true)
			defer func() {
				time.Sleep(time.Millisecond * time.Duration(b.deBounce))
				b.SetClicked(false)
			}()
		}
		return b.onClick(b, md.x, md.y)
	}
	return false
}

func (b *SDL_Image) Draw(renderer *sdl.Renderer, font *ttf.Font) error {
	if b.IsVisible() {
		borderRect := &sdl.Rect{X: b.x, Y: b.y, W: b.w, H: b.h}
		outRect := &sdl.Rect{X: b.x, Y: b.y, W: b.w, H: b.h}
		if b.ShouldDrawBorder() {
			outRect = widgetShrinkRect(outRect, 4)
		}
		if b.ShouldDrawBackground() {
			bc := b.GetBackground()
			renderer.SetDrawColor(bc.R, bc.G, bc.B, bc.A)
			renderer.FillRect(borderRect)
		}
		image, irw, _, err := GetResourceInstance().GetTextureForName(b.textureName)
		if err != nil {
			renderer.SetDrawColor(255, 0, 0, 255)
			renderer.DrawRect(&sdl.Rect{X: b.x, Y: b.y, W: 100, H: 100})
			return nil
		}
		// if bg != nil || fg != nil {
		// 	outRect = widgetShrinkRect(outRect, 4)
		// }
		if b.frameCount > 1 {
			w := (irw / b.frameCount)
			x := (w * b.frame)
			inRect := &sdl.Rect{X: x, Y: 0, W: w, H: outRect.H}
			outRect := &sdl.Rect{X: outRect.X, Y: outRect.Y, W: w, H: outRect.H}
			renderer.Copy(image, inRect, outRect)
		} else {
			renderer.Copy(image, nil, outRect)
		}
		if b.ShouldDrawBorder() {
			bc := b.GetBorderColour()
			renderer.SetDrawColor(bc.R, bc.G, bc.B, bc.A)
			renderer.DrawRect(&sdl.Rect{X: b.x + 1, Y: b.y + 1, W: b.w - 2, H: b.h - 2})
			renderer.DrawRect(&sdl.Rect{X: b.x + 2, Y: b.y + 2, W: b.w - 4, H: b.h - 4})
		}
	}
	return nil
}

func (b *SDL_Image) Destroy() {
	// Image cache takes care of all images!
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
	but.SDL_WidgetBase = initBase(x, y, w, h, id, 0, style)
	return but
}

func (b *SDL_Separator) Click(md *SDL_MouseData) bool {
	return false
}

func (b *SDL_Separator) Draw(renderer *sdl.Renderer, font *ttf.Font) error {
	if b.IsEnabled() {
		if b.ShouldDrawBackground() {
			bc := b.GetBackground()
			renderer.SetDrawColor(bc.R, bc.G, bc.B, bc.A)
			renderer.FillRect(&sdl.Rect{X: b.x, Y: b.y, W: b.w, H: b.h})
		}
		if b.ShouldDrawBorder() {
			bc := b.GetBorderColour()
			renderer.SetDrawColor(bc.R, bc.G, bc.B, bc.A)
			renderer.DrawRect(&sdl.Rect{X: b.x + 1, Y: b.y + 1, W: b.w - 2, H: b.h - 2})
			renderer.DrawRect(&sdl.Rect{X: b.x + 2, Y: b.y + 2, W: b.w - 4, H: b.h - 4})
		}
	}
	return nil
}

func (b *SDL_Separator) Destroy() {
	// Image cache takes care of all images!
}
