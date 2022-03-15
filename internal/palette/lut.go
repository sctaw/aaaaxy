// Copyright 2021 Google LLC
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package palette

import (
	"fmt"
	"image"
	"image/color"
	"math"
	"sort"
	"sync"
	"time"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/lucasb-eyer/go-colorful"

	"github.com/divVerent/aaaaxy/internal/flag"
	"github.com/divVerent/aaaaxy/internal/log"
	m "github.com/divVerent/aaaaxy/internal/math"
)

var (
	paletteColordist = flag.String("palette_colordist", "weighted", "color distance function to use; one of 'weighted', 'redmean', 'cielab', 'cieluv'")
)

type rgb [3]float64 // Range is from 0 to 1 in sRGB color space.

func (c rgb) String() string {
	n := c.toNRGBA()
	return fmt.Sprintf("#%02X%02X%02X", n.R, n.G, n.B)
}

func (c rgb) diff2(other rgb) float64 {
	switch *paletteColordist {
	case "weighted":
		dr := c[0] - other[0]
		dg := c[1] - other[1]
		db := c[2] - other[2]
		return 3*dr*dr + 4*dg*dg + 2*db*db
	case "redmean":
		dr := c[0] - other[0]
		dg := c[1] - other[1]
		db := c[2] - other[2]
		rr := (c[0] + other[0]) / 2
		return (2+rr)*dr*dr + 4*dg*dg + (2+255/256.0-rr)*db*db
	case "cielab":
		return math.Pow(c.toColorful().DistanceLab(other.toColorful()), 2)
	case "cieluv":
		return math.Pow(c.toColorful().DistanceLuv(other.toColorful()), 2)
	default:
		*paletteColordist = "weighted"
		return c.diff2(other)
	}
}

func (c rgb) toNRGBA() color.NRGBA {
	return color.NRGBA{
		R: uint8(c[0]*255 + 0.5),
		G: uint8(c[1]*255 + 0.5),
		B: uint8(c[2]*255 + 0.5),
		A: 255,
	}
}

func (c rgb) toColorful() colorful.Color {
	return colorful.Color{
		R: c[0],
		G: c[1],
		B: c[2],
	}
}

func (p *Palette) lookup(i int) rgb {
	u := p.colors[i]
	return rgb{
		float64(u>>16) / 255,
		float64((u>>8)&0xFF) / 255,
		float64(u&0xFF) / 255,
	}
}

// lookupNearest returns the palette color nearest to c.
func (p *Palette) lookupNearest(c rgb) int {
	bestI := 0
	bestS := c.diff2(p.lookup(0))
	for i := 1; i < p.size; i++ {
		s := c.diff2(p.lookup(i))
		if s < bestS {
			bestI, bestS = i, s
		}
	}
	return bestI
}

func (p *Palette) ToLUT(img *ebiten.Image) (int, int) {
	defer func(t0 time.Time) {
		dt := time.Since(t0)
		log.Infof("building palette LUT took %v", dt)
	}(time.Now())
	bounds := img.Bounds()
	w := bounds.Max.X - bounds.Min.X
	h := bounds.Max.Y - bounds.Min.Y
	lutSize := int(math.Cbrt(float64(w) * float64(h)))
	var perRow, heightNeeded, widthNeeded int
	for {
		perRow = w / lutSize
		widthNeeded = perRow * lutSize
		rows := (lutSize + perRow - 1) / perRow
		heightNeeded = rows * lutSize
		if heightNeeded <= h {
			break
		}
		lutSize--
	}

	// Note: creating a temp image, and copying to that, so this does not invoke
	// thread synchronization as writing to an ebiten.Image would.
	rect := image.Rectangle{
		Min: bounds.Min,
		Max: image.Point{
			X: bounds.Min.X + widthNeeded,
			Y: bounds.Min.Y + heightNeeded,
		},
	}

	pix := make([]uint8, heightNeeded*widthNeeded*4)

	var wg sync.WaitGroup
	// TODO(divVerent): Also compute for each pixel the distance to the next color when adding or subtracting to all of r,g,b.
	// Use this to compute a dynamic Bayer scale.
	// At points, Bayer scale should be the MIN of the distances to next colors.
	// Elsewhere, Bayer scale ideally should be those values interpolated.
	// What can we practically get?
	// Store that data in the alpha channel.
	for y := 0; y < heightNeeded; y++ {
		wg.Add(1)
		go func(y int) {
			g := y % lutSize
			gFloat := (float64(g) + 0.5) / float64(lutSize)
			bY := (y / lutSize) * perRow
			for x := 0; x < widthNeeded; x++ {
				r := x % lutSize
				rFloat := (float64(r) + 0.5) / float64(lutSize)
				b := bY + x/lutSize
				bFloat := (float64(b) + 0.5) / float64(lutSize)
				c := rgb{rFloat, gFloat, bFloat}
				i := p.lookupNearest(c)
				cNew := p.lookup(i)
				rgba := cNew.toNRGBA()
				o := (y*widthNeeded + x) * 4
				pix[o] = rgba.R
				pix[o+1] = rgba.G
				pix[o+2] = rgba.B
				pix[o+3] = 255
			}
			wg.Done()
		}(y)
	}
	wg.Wait()

	// For each protected palette index, find its ideal bayer scale.
	scales := make([]int, p.protected)
	for i := 0; i < p.protected; i++ {
		wg.Add(1)
		go func(i int) {
			c := p.lookup(i).toNRGBA()
			scale := 1
		FoundScale:
			for scale < 256 {
				for d := -1; d <= 1; d += 2 {
					rr := int(c.R) + scale*d
					gg := int(c.G) + scale*d
					bb := int(c.B) + scale*d
					r := rr * lutSize / 255
					g := gg * lutSize / 255
					b := bb * lutSize / 255
					if r < 0 {
						r = 0
					}
					if r >= lutSize {
						r = lutSize - 1
					}
					if g < 0 {
						g = 0
					}
					if g >= lutSize {
						g = lutSize - 1
					}
					if b < 0 {
						b = 0
					}
					if b >= lutSize {
						b = lutSize - 1
					}
					x := r + lutSize*(b%perRow)
					y := g + lutSize*(b/perRow)
					o := (y*widthNeeded + x) * 4
					if pix[o] != c.R || pix[o+1] != c.G || pix[o+2] != c.B {
						break FoundScale
					}
				}
				scale++
			}
			scale--
			// Make all scales one LUT entry lower.
			// This fixes pathological gradients due to a roundoff error
			// in the color right next to a palette color.
			scale -= (255 + lutSize - 1) / lutSize
			if scale < 0 {
				scale = 0
			}
			scales[i] = scale
			wg.Done()
		}(i)
	}
	wg.Wait()

	// Set alpha channel to best Bayer scale for each pixel.
	for i := 0; i < p.protected; i++ {
		c := p.lookup(i).toNRGBA()
		rr := int(c.R)
		gg := int(c.G)
		bb := int(c.B)
		r := rr * lutSize / 255
		g := gg * lutSize / 255
		b := bb * lutSize / 255
		if r >= lutSize {
			r = lutSize - 1
		}
		if g >= lutSize {
			g = lutSize - 1
		}
		if b >= lutSize {
			b = lutSize - 1
		}
		x := r + lutSize*(b%perRow)
		y := g + lutSize*(b/perRow)
		o := (y*widthNeeded + x) * 4
		pix[o+3] = uint8(scales[i])
	}
	for y := 0; y < heightNeeded; y++ {
		wg.Add(1)
		go func(y int) {
			g := y % lutSize
			gFloat := (float64(g) + 0.5) / float64(lutSize)
			bY := (y / lutSize) * perRow
			for x := 0; x < widthNeeded; x++ {
				o := (y*widthNeeded + x) * 4
				if pix[o+3] != 255 {
					continue
				}
				r := x % lutSize
				rFloat := (float64(r) + 0.5) / float64(lutSize)
				b := bY + x/lutSize
				bFloat := (float64(b) + 0.5) / float64(lutSize)
				c := rgb{rFloat, gFloat, bFloat}
				sum, weight := 0.0, 0.0
				for i, scale := range scales {
					c2 := p.lookup(i)
					f := 1 / c.diff2(c2)
					sum += f * float64(scale)
					weight += f
				}
				scale := m.Rint(sum / weight)
				pix[o+3] = uint8(scale)
			}
			wg.Done()
		}(y)
	}
	wg.Wait()

	img.SubImage(rect).(*ebiten.Image).ReplacePixels(pix)

	return lutSize, perRow
}

func sizeBayer(size int) (sizeSquare int, scale, offset float64) {
	sizeSquare = size * size
	bits := 0
	if size > 1 {
		bits = math.Ilogb(float64(size-1)) + 1
	}
	sizeCeil := 1 << bits
	sizeCeilSquare := sizeCeil * sizeCeil
	// Map to [-1..1] _inclusive_ ranges.
	// Not _perfect_, but way nicer to work with.
	if sizeCeilSquare > 1 {
		scale = 2.0 / float64(sizeCeilSquare-1)
	}
	offset = -float64(sizeCeilSquare-1) / 2.0
	return
}

func sizeHalftone(size int) (sizeSquare int, scale, offset float64) {
	sizeSquare = size * size
	// Map to [-1..1] _inclusive_ ranges.
	// Not _perfect_, but way nicer to work with.
	if sizeSquare > 1 {
		scale = 2.0 / float64(sizeSquare-1)
	}
	offset = -float64(sizeSquare-1) / 2.0
	return
}

func clamp(a, mi, ma float64) float64 {
	if a < mi {
		return mi
	}
	if a > ma {
		return ma
	}
	return a
}

// BayerPattern computes the Bayer pattern for this palette.
func (p *Palette) BayerPattern(size int) []float32 {
	sizeSquare, scale, offset := sizeBayer(size)
	bayern := make([]float32, sizeSquare)
	for i := range bayern {
		x := i % size
		y := i / size
		z := x ^ y
		b := 0
		for bit := 1; bit < size; bit *= 2 {
			b *= 4
			if y&bit != 0 {
				b += 1
			}
			if z&bit != 0 {
				b += 2
			}
		}
		bayern[i] = float32((float64(b) + offset) * scale)
	}
	return bayern
}

// HalftonePattern computes the Halftone pattern for this palette.
func (p *Palette) HalftonePattern(size int) []float32 {
	sizeSquare, scale, offset := sizeHalftone(size)
	type index struct {
		i     int
		order float64
	}
	weighted := make([]index, sizeSquare)
	cX := float64(size)*0.5 + 0.03
	cY := float64(size)*0.5 + 0.04
	for i := range weighted {
		x := i % size
		y := i / size
		d := math.Hypot(float64(x)-cX, float64(y)-cY)
		weighted[i] = index{i, d}
	}
	sort.Slice(weighted, func(i, j int) bool {
		return weighted[i].order < weighted[j].order
	})
	bayern := make([]float32, sizeSquare)
	for b, idx := range weighted {
		bayern[idx.i] = float32((float64(b) + offset) * scale)
	}
	return bayern
}
