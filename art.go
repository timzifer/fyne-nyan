package nyan

import (
	"image"
	"image/color"
	"math"
)

// The scene is drawn on a grid of "art pixels" (units). The bar height is
// divided into gridRows units, every unit becomes a square block of
// physical pixels so the artwork stays crisp at any scale.
const (
	gridRows   = 18
	bandHeight = 2  // height of one rainbow band in units
	minWave    = 4  // narrowest rainbow wave segment in units
	maxWave    = 10 // widest rainbow wave segment in units
	tailTuck   = 6  // how far the rainbow reaches under the cat in units
	catTop     = 2  // top row of the pop-tart body (before bobbing)
	starPitch  = 22 // average horizontal distance between stars in units
)

var (
	colorSpace = color.NRGBA{R: 0x0d, G: 0x2b, B: 0x5e, A: 0xff}
	colorStar  = color.NRGBA{R: 0xff, G: 0xff, B: 0xff, A: 0xff}

	rainbow = [...]color.NRGBA{
		{R: 0xff, G: 0x1a, B: 0x1a, A: 0xff},
		{R: 0xff, G: 0x9a, B: 0x1a, A: 0xff},
		{R: 0xff, G: 0xf0, B: 0x1a, A: 0xff},
		{R: 0x4c, G: 0xe6, B: 0x1a, A: 0xff},
		{R: 0x1a, G: 0xa3, B: 0xff, A: 0xff},
		{R: 0x75, G: 0x4c, B: 0xff, A: 0xff},
	}

	palette = map[byte]color.NRGBA{
		'k': {R: 0x14, G: 0x14, B: 0x1e, A: 0xff}, // outline
		't': {R: 0xf8, G: 0xc8, B: 0x8c, A: 0xff}, // crust
		'p': {R: 0xff, G: 0xa6, B: 0xe0, A: 0xff}, // frosting
		's': {R: 0xf0, G: 0x3c, B: 0x8c, A: 0xff}, // sprinkle
		'g': {R: 0x9e, G: 0xa0, B: 0xa8, A: 0xff}, // fur
		'w': {R: 0xff, G: 0xff, B: 0xff, A: 0xff}, // eye shine
		'r': {R: 0xff, G: 0x8c, B: 0x9e, A: 0xff}, // cheeks
	}
)

// head is drawn on top of the right part of the pop-tart.
var head = [...]string{
	".kk.......kk.",
	"kggk.....kggk",
	"kgggkkkkkgggk",
	"kgggggggggggk",
	"kggwkggggwkgk",
	"kggkkggkgkkgk",
	"krrgggggggrrk",
	"kggkgkkgkgggk",
	".kggggggggggk",
	"..kkkkkkkkkk.",
}

var sprinkles = [...][2]int{
	{3, 2}, {8, 2}, {12, 3}, {5, 4}, {10, 5}, {3, 6}, {7, 7}, {4, 9}, {9, 9}, {12, 8},
}

const (
	bodyLeft   = 4
	bodyWidth  = 17
	bodyHeight = 12
	headLeft   = 14
	headTop    = 4
	catWidth   = headLeft + 13 // 27 units
	catFrames  = 4
)

// catSprites holds the pre-rendered animation frames of the cat.
var catSprites = buildCatFrames()

type sprite struct {
	w, h int
	px   []byte // palette keys, 0 means transparent
}

func (s *sprite) set(x, y int, c byte) {
	if x >= 0 && y >= 0 && x < s.w && y < s.h {
		s.px[y*s.w+x] = c
	}
}

func (s *sprite) at(x, y int) byte {
	return s.px[y*s.w+x]
}

func buildCatFrames() [catFrames]*sprite {
	var frames [catFrames]*sprite
	bob := [catFrames]int{0, 0, 1, 1}
	tail := [catFrames]int{5, 6, 7, 6}
	legs := [catFrames]int{0, 1, 0, -1}

	for f := range frames {
		s := &sprite{w: catWidth, h: catTop + bodyHeight + 4}
		s.px = make([]byte, s.w*s.h)

		// legs first, the bobbing body covers their upper half
		for _, x := range [...]int{bodyLeft + 1, bodyLeft + 5, bodyLeft + 10, bodyLeft + 14} {
			top := catTop + bodyHeight - 1
			for dy := 0; dy < 4; dy++ {
				lx := x + legs[f]
				c := byte('g')
				if dy == 3 {
					c = 'k'
				}
				s.set(lx, top+dy, 'k')
				s.set(lx+1, top+dy, c)
				s.set(lx+2, top+dy, 'k')
			}
		}

		top := catTop + bob[f]

		// tail
		ty := tail[f] - 1 + bob[f]
		for x := 0; x <= bodyLeft; x++ {
			s.set(x, ty, 'k')
			s.set(x, ty+1, 'g')
			s.set(x, ty+2, 'k')
		}
		s.set(0, ty+1, 'k')

		// pop-tart
		for y := 0; y < bodyHeight; y++ {
			for x := 0; x < bodyWidth; x++ {
				c := byte('p')
				switch {
				case y == 0 || y == bodyHeight-1 || x == 0 || x == bodyWidth-1:
					c = 'k'
				case y == 1 || y == bodyHeight-2 || x == 1 || x == bodyWidth-2:
					c = 't'
				}
				s.set(bodyLeft+x, top+y, c)
			}
		}
		for _, sp := range sprinkles {
			s.set(bodyLeft+sp[0], top+sp[1], 's')
		}

		// head
		for y, row := range head {
			for x := 0; x < len(row); x++ {
				if row[x] != '.' {
					s.set(headLeft+x, top+headTop+y-catTop, row[x])
				}
			}
		}

		frames[f] = s
	}
	return frames
}

// scene describes everything needed to paint one frame.
type scene struct {
	ratio    float64 // determinate: progress from 0 to 1
	infinite bool    // indeterminate: cat position is derived from travel
	travel   float64 // indeterminate: distance flown so far in units
	tick     int     // animation frame counter
	radius   float32 // corner radius in physical pixels
}

// grid returns the size of one unit in pixels and the number of columns.
func grid(w, h int) (unit, cols int) {
	unit = max(1, h/gridRows)
	return unit, (w + unit - 1) / unit
}

// catX returns the left edge of the cat in units.
func (sc scene) catX(cols int) int {
	if sc.infinite {
		// the flight starts with the whole cat visible at the left edge
		cycle := cols + catWidth
		return (int(sc.travel)+catWidth)%cycle - catWidth
	}
	r := math.Min(1, math.Max(0, sc.ratio))
	return int(math.Round(r * float64(max(0, cols-catWidth))))
}

// paint renders the scene into a new image of w×h physical pixels.
func paint(w, h int, sc scene) *image.NRGBA {
	img := image.NewNRGBA(image.Rect(0, 0, w, h))
	if w <= 0 || h <= 0 {
		return img
	}

	unit, cols := grid(w, h)
	offY := (h - unit*gridRows) / 2

	fill(img, 0, 0, w, h, colorSpace)

	block := func(ux, uy int, c color.NRGBA) {
		fill(img, ux*unit, offY+uy*unit, unit, unit, c)
	}

	// stars drift to the left and twinkle
	// one star per stripe of starPitch columns keeps them evenly spread
	stars := cols/starPitch + 1
	period := stars * starPitch
	for i := 0; i < stars; i++ {
		hx := hash(uint32(i)*2 + 1)
		hy := hash(uint32(i)*2 + 2)
		x := i*starPitch + int(hx%(starPitch-4)) - sc.tick*2
		x = ((x%period)+period)%period - 2
		y := 2 + int(hy%(gridRows-4))
		drawStar(block, x, y, (sc.tick+int(hx>>8))%6)
	}

	catX := sc.catX(cols)
	frame := sc.tick % catFrames

	if !sc.infinite {
		drawRainbow(img, unit, offY, 0, catX+bodyLeft+tailTuck, 0, sc.tick)
		drawCat(block, catX, frame)
		roundCorners(img, sc.radius)
		return img
	}

	// indeterminate: a finite flag follows the cat and fades out. The
	// previous lap is drawn as well so the flag leaves on the right while
	// the cat comes back in on the left.
	length := max(catWidth, cols/3)
	cycle := cols + catWidth
	laps := []int{catX}
	if int(sc.travel)+catWidth >= cycle {
		laps = append(laps, catX+cycle)
	}
	for _, x := range laps {
		end := x + bodyLeft + tailTuck
		drawRainbow(img, unit, offY, end-length, end, length*2/5, sc.tick)
		drawCat(block, x, frame)
	}
	roundCorners(img, sc.radius)
	return img
}

// drawRainbow paints the trail from column start to end. The wave is
// anchored at end, so it travels along with the cat. The first fade
// columns blend into the background.
func drawRainbow(img *image.NRGBA, unit, offY, start, end, fade, tick int) {
	seg, segLeft := 0, waveWidth(0)
	for x := end - 1; x >= start; x-- {
		if segLeft == 0 {
			seg++
			segLeft = waveWidth(seg)
		}
		segLeft--

		alpha := 1.0
		if d := x - start; d < fade {
			alpha = float64(d+1) / float64(fade+1)
		}
		if x < 0 {
			continue
		}

		dy := (seg + tick/2) % 2
		for b, c := range rainbow {
			// sparkles travel with the wave and change every few frames
			if hash(uint32(end-x)*7919+uint32(b)*131+uint32(tick/3))%61 == 0 {
				c = mix(c, colorStar, 0.55)
			}
			for k := 0; k < bandHeight; k++ {
				blend(img, x*unit, offY+(catTop+dy+b*bandHeight+k)*unit, unit, unit, c, alpha)
			}
		}
	}
}

// waveWidth gives each wave segment its own width, so the rainbow does
// not look like a repeating pattern.
func waveWidth(seg int) int {
	return minWave + int(hash(uint32(seg)+0x9e37)%(maxWave-minWave+1))
}

func drawCat(block func(x, y int, c color.NRGBA), x0, frame int) {
	cat := catSprites[frame]
	for y := 0; y < cat.h; y++ {
		for x := 0; x < cat.w; x++ {
			if c := cat.at(x, y); c != 0 {
				block(x0+x, y, palette[c])
			}
		}
	}
}

func mix(a, b color.NRGBA, t float64) color.NRGBA {
	l := func(x, y uint8) uint8 { return uint8(float64(x) + (float64(y)-float64(x))*t + 0.5) }
	return color.NRGBA{R: l(a.R, b.R), G: l(a.G, b.G), B: l(a.B, b.B), A: 0xff}
}

// blend paints c with the given opacity over an opaque image region.
func blend(img *image.NRGBA, x, y, w, h int, c color.NRGBA, alpha float64) {
	if alpha >= 1 {
		fill(img, x, y, w, h, c)
		return
	}
	r := image.Rect(x, y, x+w, y+h).Intersect(img.Rect)
	for py := r.Min.Y; py < r.Max.Y; py++ {
		for px := r.Min.X; px < r.Max.X; px++ {
			img.SetNRGBA(px, py, mix(img.NRGBAAt(px, py), c, alpha))
		}
	}
}

func drawStar(block func(x, y int, c color.NRGBA), x, y, phase int) {
	switch phase {
	case 0:
		block(x, y, colorStar)
	case 1, 5:
		block(x, y, colorStar)
		for _, d := range [...][2]int{{-1, 0}, {1, 0}, {0, -1}, {0, 1}} {
			block(x+d[0], y+d[1], colorStar)
		}
	case 2, 4:
		for _, d := range [...][2]int{{-2, 0}, {-1, 0}, {1, 0}, {2, 0}, {0, -2}, {0, -1}, {0, 1}, {0, 2}} {
			block(x+d[0], y+d[1], colorStar)
		}
	case 3:
		for _, d := range [...][2]int{{-2, 0}, {2, 0}, {0, -2}, {0, 2}, {-1, -1}, {1, -1}, {-1, 1}, {1, 1}} {
			block(x+d[0], y+d[1], colorStar)
		}
	}
}

func fill(img *image.NRGBA, x, y, w, h int, c color.NRGBA) {
	r := image.Rect(x, y, x+w, y+h).Intersect(img.Rect)
	for py := r.Min.Y; py < r.Max.Y; py++ {
		i := img.PixOffset(r.Min.X, py)
		for px := r.Min.X; px < r.Max.X; px++ {
			img.Pix[i], img.Pix[i+1], img.Pix[i+2], img.Pix[i+3] = c.R, c.G, c.B, c.A
			i += 4
		}
	}
}

// roundCorners fades out the pixels outside a rounded rectangle, with a
// one pixel anti-aliased edge.
func roundCorners(img *image.NRGBA, radius float32) {
	w, h := img.Rect.Dx(), img.Rect.Dy()
	r := float64(min(radius, float32(min(w, h))/2))
	if r <= 0 {
		return
	}
	ri := int(math.Ceil(r))
	for y := 0; y < ri; y++ {
		for x := 0; x < ri; x++ {
			dx, dy := r-(float64(x)+0.5), r-(float64(y)+0.5)
			cover := math.Min(1, math.Max(0, r-math.Hypot(dx, dy)+0.5))
			if cover >= 1 {
				continue
			}
			for _, p := range [...][2]int{{x, y}, {w - 1 - x, y}, {x, h - 1 - y}, {w - 1 - x, h - 1 - y}} {
				i := img.PixOffset(p[0], p[1])
				img.Pix[i+3] = uint8(float64(img.Pix[i+3]) * cover)
			}
		}
	}
}

// hash is a small integer hash used to scatter the stars deterministically.
func hash(x uint32) uint32 {
	x ^= x >> 16
	x *= 0x7feb352d
	x ^= x >> 15
	x *= 0x846ca68b
	x ^= x >> 16
	return x
}
