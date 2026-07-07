package main

import (
	"image"
	"image/color"
	"image/png"
	"log"
	"math"
	"os"
	"path/filepath"
)

const size = 512

func main() {
	img := image.NewRGBA(image.Rect(0, 0, size, size))

	drawCircle(img, size/2, size/2, 230, color.RGBA{R: 32, G: 129, B: 246, A: 255})
	drawCircle(img, size/2-80, size/2-90, 68, color.RGBA{R: 106, G: 173, B: 255, A: 210})

	drawCup(img)

	if err := os.MkdirAll("assets", 0o755); err != nil {
		log.Fatalf("create assets directory: %v", err)
	}

	path := filepath.Join("assets", "Icon.png")
	file, err := os.Create(path)
	if err != nil {
		log.Fatalf("create icon: %v", err)
	}
	defer file.Close()

	if err := png.Encode(file, img); err != nil {
		log.Fatalf("encode icon: %v", err)
	}
}

func drawCircle(img *image.RGBA, cx, cy, radius int, c color.RGBA) {
	for y := cy - radius; y <= cy+radius; y++ {
		for x := cx - radius; x <= cx+radius; x++ {
			if x < 0 || y < 0 || x >= size || y >= size {
				continue
			}
			dx := float64(x - cx)
			dy := float64(y - cy)
			if math.Sqrt(dx*dx+dy*dy) <= float64(radius) {
				img.SetRGBA(x, y, c)
			}
		}
	}
}

func drawCup(img *image.RGBA) {
	outline := color.RGBA{R: 235, G: 246, B: 255, A: 255}
	glass := color.RGBA{R: 224, G: 242, B: 255, A: 235}
	water := color.RGBA{R: 78, G: 183, B: 255, A: 245}
	shadow := color.RGBA{R: 17, G: 80, B: 160, A: 80}

	for y := 156; y <= 388; y++ {
		progress := float64(y-156) / float64(388-156)
		left := int(172 + progress*34)
		right := int(340 - progress*34)
		for x := left; x <= right; x++ {
			switch {
			case x-left < 8 || right-x < 8 || y < 166 || y > 378:
				img.SetRGBA(x, y, outline)
			case y > 250:
				img.SetRGBA(x, y, water)
			default:
				img.SetRGBA(x, y, glass)
			}
		}
	}

	for y := 392; y <= 410; y++ {
		for x := 206; x <= 306; x++ {
			img.SetRGBA(x, y, shadow)
		}
	}

	for y := 136; y <= 158; y++ {
		for x := 164; x <= 348; x++ {
			img.SetRGBA(x, y, outline)
		}
	}
}
