package imagegenerator

import (
	"fmt"
	"image"
	"image/color"
	"image/png"
	"os"
	"path/filepath"
)

func CurrentExePath() string {
	dir, _ := filepath.Abs(filepath.Dir(os.Args[0]))
	return dir
}

func fillRect(im *image.RGBA, col color.RGBA, xx, yy, ww, hh int) {
	for x := xx; x < xx+ww; x++ {
		for y := yy; y < yy+hh; y++ {
			im.SetRGBA(x, y, col)
		}
	}
}

func drawGazerIcon(im *image.RGBA) {
	w := im.Rect.Max.X
	m := w / 50
	if m < 1 {
		m = 1
	}
	o := w/2 + m
	s := w/2 - m

	col1 := color.RGBA{R: 1, G: 159, B: 228, A: 255} // 1,159,228,255
	col2 := color.RGBA{R: 21, G: 188, B: 80, A: 255} // 21,188,80,255

	//fillRect(im, color.RGBA{R: 255, G: 255, B: 255, A: 255}, 0, 0, w, w)

	fillRect(im, col1, 0, 0, s, s)
	fillRect(im, col2, o, 0, s, s)
	fillRect(im, col1, 0, o, s, s)
	fillRect(im, col1, o, o, s, s)
}

func genPng(path string, w, h int) {
	fmt.Println("generating png", path)
	var im *image.RGBA
	im = image.NewRGBA(image.Rect(0, 0, w, h))
	drawGazerIcon(im)
	out, err := os.Create(path)
	if err != nil {
		panic(err)
	}
	png.Encode(out, im)
}

func genIOSIcons(path string) {
	genPng(path+"Icon-App-20x20@1x.png", 20, 20)
	genPng(path+"Icon-App-20x20@2x.png", 40, 40)
	genPng(path+"Icon-App-20x20@3x.png", 60, 60)

	genPng(path+"Icon-App-29x29@1x.png", 29, 29)
	genPng(path+"Icon-App-29x29@2x.png", 58, 58)
	genPng(path+"Icon-App-29x29@3x.png", 87, 87)

	genPng(path+"Icon-App-40x40@1x.png", 40, 40)
	genPng(path+"Icon-App-40x40@2x.png", 80, 80)
	genPng(path+"Icon-App-40x40@3x.png", 120, 120)

	//genPng(path+"Icon-App-60x60@1x.png", 60, 60)
	genPng(path+"Icon-App-60x60@2x.png", 120, 120)
	genPng(path+"Icon-App-60x60@3x.png", 180, 180)

	genPng(path+"Icon-App-76x76@1x.png", 76, 76)
	genPng(path+"Icon-App-76x76@2x.png", 152, 152)
	//genPng(path+"Icon-App-60x60@3x.png", 180, 180)

	genPng(path+"Icon-App-83.5x83.5@2x.png", 167, 167)
	genPng(path+"Icon-App-1024x1024@1x.png", 1024, 1024)
}

func Generate() {
	path := CurrentExePath() + "/resources/gen/"
	os.MkdirAll(path, 0777)
	genIOSIcons(path)
}
