package main

import (
	"bufio"
	"fmt"
	"image"
	"image/color"
	"image/draw"
	"io"
	"log"
	"math"
	"os"
	"time"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/ebitenutil"
	"github.com/hajimehoshi/ebiten/v2/inpututil"
	"github.com/hajimehoshi/ebiten/v2/vector"
	"samhza.com/place"
)

type Game struct {
	image  *image.RGBA
	eimage *ebiten.Image
	fd     *os.File
	f      *bufio.Reader
	keys   []ebiten.Key

	x float64
	y float64

	zoom  float64
	speed int

	paused  bool
	debug   bool
	changed bool

	time     int
	nexttime int

	updateCount int

	mouse          bool
	mouseX, mouseY int

	expandedright bool
	expandeddown  bool
}

var colors = [][]byte{
	{0x00, 0x00, 0x00},
	{0x00, 0x75, 0x6F},
	{0x00, 0x9E, 0xAA},
	{0x00, 0xA3, 0x68},
	{0x00, 0xCC, 0x78},
	{0x00, 0xCC, 0xC0},
	{0x24, 0x50, 0xA4},
	{0x36, 0x90, 0xEA},
	{0x49, 0x3A, 0xC1},
	{0x51, 0x52, 0x52},
	{0x51, 0xE9, 0xF4},
	{0x6A, 0x5C, 0xFF},
	{0x6D, 0x00, 0x1A},
	{0x6D, 0x48, 0x2F},
	{0x7E, 0xED, 0x56},
	{0x81, 0x1E, 0x9F},
	{0x89, 0x8D, 0x90},
	{0x94, 0xB3, 0xFF},
	{0x9C, 0x69, 0x26},
	{0xB4, 0x4A, 0xC0},
	{0xBE, 0x00, 0x39},
	{0xD4, 0xD7, 0xD9},
	{0xDE, 0x10, 0x7F},
	{0xE4, 0xAB, 0xFF},
	{0xFF, 0x38, 0x81},
	{0xFF, 0x45, 0x00},
	{0xFF, 0x99, 0xAA},
	{0xFF, 0xA8, 0x00},
	{0xFF, 0xB4, 0x70},
	{0xFF, 0xD6, 0x35},
	{0xFF, 0xF8, 0xB8},
	{0xFF, 0xFF, 0xFF},
}

func clampInt(v, min, max int) int {
	if v < min {
		return min
	}
	if v > max {
		return max
	}
	return v
}

func (g *Game) Update() error {
	if _, yoff := ebiten.Wheel(); yoff != 0 {
		if ebiten.IsMouseButtonPressed(ebiten.MouseButtonRight) {
			g.speed += int(yoff)
		} else {
			fac := math.Pow(1.2, float64(yoff))
			g.changeZoom(fac, true)
		}
	}
	if inpututil.IsMouseButtonJustPressed(ebiten.MouseButtonLeft) {
		g.mouseX, g.mouseY = ebiten.CursorPosition()
		g.mouse = true
	}
	if g.mouse {
		x, y := ebiten.CursorPosition()
		g.x += float64(x-g.mouseX) / g.zoom
		g.y += float64(y-g.mouseY) / g.zoom
		g.mouseX, g.mouseY = x, y
		if inpututil.IsMouseButtonJustReleased(ebiten.MouseButtonLeft) {
			g.mouse = false
		}
	}
	g.keys = inpututil.AppendJustPressedKeys(g.keys[:0])
	frametime := int(1000 / 60 * math.Pow(1.5, float64(g.speed)))
	for _, k := range g.keys {
		switch k {
		case ebiten.KeyR:
			g.reset()
		case ebiten.Key0:
			g.zoom = 1
			g.x = 0
			g.y = 0
		case ebiten.KeyPeriod:
			frametime *= 60 * 60
		case ebiten.KeyD:
			g.debug = !g.debug
		case ebiten.KeySpace:
			g.paused = !g.paused
		case ebiten.KeyEqual:
			g.changeZoom(1.2, false)
		case ebiten.KeyMinus:
			g.changeZoom(1/1.2, false)
		case ebiten.KeyBracketLeft:
			g.speed -= 1
		case ebiten.KeyBracketRight:
			g.speed += 1
		}
	}
	g.keys = inpututil.AppendPressedKeys(g.keys[:0])
	for _, k := range g.keys {
		switch k {
		case ebiten.KeyLeft:
			g.x += 10 / g.zoom
		case ebiten.KeyRight:
			g.x -= 10 / g.zoom
		case ebiten.KeyUp:
			g.y += 10 / g.zoom
		case ebiten.KeyDown:
			g.y -= 10 / g.zoom
		}
	}
	g.speed = clampInt(g.speed, 0, 25)
	if g.paused {
		return nil
	}
	start := g.time
	if start < g.nexttime {
		g.time += frametime
		return nil
	}
	g.nexttime = 0
	g.updateCount = 0
	for {
		buf, err := g.f.Peek(10)
		if err == io.EOF {
			return nil
		}
		if err != nil {
			return err
		}
		var chg place.Change
		chg.Decode(buf)
		if start == 0 {
			start = chg.Time
		}
		if chg.Time > start+frametime {
			g.time += frametime
			g.nexttime = chg.Time
			break
		}
		g.changed = true
		g.updateCount++
		g.time = chg.Time
		_, err = g.f.Discard(10)
		if err != nil {
			return err
		}
		if !g.expandedright && chg.X1 > 1000 {
			for i := 0; i < 1000; i++ {
				for j := 0; j < 1000; j++ {
					g.image.SetRGBA(i+1000, j, color.RGBA{255, 255, 255, 255})
				}
			}
			g.expandedright = true
		}
		if !g.expandeddown && chg.Y1 > 1000 {
			for i := 0; i < 2000; i++ {
				for j := 0; j < 1000; j++ {
					g.image.SetRGBA(i, j+1000, color.RGBA{255, 255, 255, 255})
				}
			}
			g.expandeddown = true
		}
		if chg.X2 == 0 {
			copy(g.image.Pix[chg.Y1*2000*4+chg.X1*4:], colors[chg.Color])
			continue
		}
		color := color.RGBA{
			colors[chg.Color][0],
			colors[chg.Color][1],
			colors[chg.Color][2],
			0xFF,
		}
		draw.Draw(g.image, image.Rect(chg.X1, chg.Y1, chg.X2, chg.Y2), &image.Uniform{color}, image.Point{}, draw.Src)
	}
	return nil
}

func (g *Game) changeZoom(fac float64, aboutMouse bool) {
	newzoom := fac * g.zoom
	if newzoom < 0.1 || newzoom > 10 {
		return
	}
	var x, y int
	if aboutMouse {
		x, y = ebiten.CursorPosition()
	} else {
		w, h := ebiten.WindowSize()
		x = w / 2
		y = h / 2
	}
	g.x -= (float64(x) / g.zoom) * (1 - 1/fac)
	g.y -= (float64(y) / g.zoom) * (1 - 1/fac)
	g.zoom = newzoom
}

func (g *Game) Draw(screen *ebiten.Image) {
	if g.changed {
		g.eimage.WritePixels(g.image.Pix)
		g.changed = false
	}
	opts := &ebiten.DrawImageOptions{}
	if g.zoom < 1 {
		opts.Filter = ebiten.FilterLinear
	}
	opts.GeoM.Scale(g.zoom, g.zoom)
	opts.GeoM.Translate(g.x*g.zoom, g.y*g.zoom)
	screen.DrawImage(g.eimage, opts)
	mx, my := ebiten.CursorPosition()
	x := -g.x + (float64(mx) / g.zoom)
	y := -g.y + (float64(my) / g.zoom)
	if g.debug {
		const format = "2006-01-02 15:04:05.000"
		vector.DrawFilledRect(screen, 0, 0, 160, 96, color.RGBA{0, 0, 0, 127}, false)
		ebitenutil.DebugPrint(screen, fmt.Sprintf(
			"%s\nZoom: %0.2f\nSpeed: %0.2fx\nPaused: %v\nChanges last tick: %d\nPosition: %4.0f %4.0f",
			time.UnixMilli(int64(g.time)+place.Epoch).Format(format),
			g.zoom,
			math.Pow(1.5, float64(g.speed)),
			g.paused,
			g.updateCount,
			math.Ceil(x), math.Ceil(y),
		))
	}
}

func (g *Game) reset() {
	g.time = 0
	g.nexttime = 0
	g.updateCount = 0
	g.expandedright = false
	g.expandeddown = false
	g.fd.Seek(0, 0)
	g.f.Reset(g.fd)
	g.image.Pix = make([]uint8, 2000*2000*4)
	for i := 0; i < 1000; i++ {
		for j := 0; j < 1000; j++ {
			copy(g.image.Pix[(i*2000+j)*4:], []byte{0xff, 0xff, 0xff, 0xff})
		}
	}
}

func (g *Game) Layout(oW, oH int) (int, int) {
	return oW, oH
}

func main() {
	ebiten.SetWindowSize(800, 600)
	ebiten.SetWindowTitle("r/place viewer")
	f, err := os.Open("squished")
	if err != nil {
		log.Fatalln(err)
	}
	g := &Game{
		image:  image.NewRGBA(image.Rect(0, 0, 2000, 2000)),
		eimage: ebiten.NewImage(2000, 2000),
		fd:     f,
		f:      bufio.NewReader(f),
		zoom:   1,
	}
	for i := 0; i < 1000; i++ {
		for j := 0; j < 1000; j++ {
			copy(g.image.Pix[(i*2000+j)*4:], []byte{0xff, 0xff, 0xff, 0xff})
		}
	}
	g.eimage.WritePixels(g.image.Pix)
	if err := ebiten.RunGame(g); err != nil {
		log.Fatal(err)
	}
}
