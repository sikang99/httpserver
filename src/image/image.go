//==================================================================================
// Image
// - https://www.socketloop.com/tutorials/golang-save-image-to-png-jpeg-or-gif-format
// - https://code.google.com/p/plotinum/wiki/Examples
//==================================================================================

package image

import (
	"fmt"
	"image"
	"image/color"
	"image/draw"
	"image/gif"
	"image/jpeg"
	"image/png"
	"log"
	"math"
	"math/rand"
	"os"
	"path/filepath"
	"time"
)

var (
	white color.Color = color.RGBA{255, 255, 255, 255}
	black color.Color = color.RGBA{0, 0, 0, 255}
	blue  color.Color = color.RGBA{0, 0, 255, 255}
)

//----------------------------------------------------------------------------------
// encode image to file in PNG, JPEG, PNG
//----------------------------------------------------------------------------------
func MakeImageFile(img image.Image, fname string, optnum int) error {
	var err error

	out, err := os.Create(fname)
	if err != nil {
		log.Println(err)
		return err
	}
	defer out.Close()

	itype := filepath.Ext(fname)

	switch itype {
	case ".png":
		err = png.Encode(out, img)
		if err != nil {
			log.Println(err)
			return err
		}

	case ".gif":
		var opt gif.Options
		opt.NumColors = optnum // 256, you can add more parameters if you want

		err = gif.Encode(out, img, &opt) // put num of colors to 256
		if err != nil {
			log.Println(err)
			return err
		}

	case ".jpg", ".jpeg":
		var opt jpeg.Options
		opt.Quality = optnum // 80

		err = jpeg.Encode(out, img, &opt) // put quality to 80%
		if err != nil {
			log.Println(err)
			return err
		}

	default:
		return fmt.Errorf("%s is unknown format", itype)
	}

	return err
}

//----------------------------------------------------------------------------------
// generate simple image
//----------------------------------------------------------------------------------
func GenSimpleImage(xz, yz int) image.Image {
	img := image.NewRGBA(image.Rect(0, 0, xz, yz))

	for y := img.Rect.Min.Y; y < img.Rect.Max.Y; y++ {
		for x := img.Rect.Min.X; x < img.Rect.Max.X; x++ {
			img.Set(x, y, color.RGBA{0x88, 0xff, 0x88, 0xff})
		}
	}

	return img
}

//----------------------------------------------------------------------------------
// generate random image
//----------------------------------------------------------------------------------
func GenRandomImage(xz, yz int) image.Image {
	rand.Seed(time.Now().UTC().UnixNano())

	imgRect := image.Rect(0, 0, xz, yz)
	img := image.NewGray(imgRect)
	draw.Draw(img, img.Bounds(), &image.Uniform{color.White}, image.ZP, draw.Src)

	for y := 0; y < yz; y += 10 {
		for x := 0; x < xz; x += 10 {
			fill := &image.Uniform{color.Black}
			if rand.Intn(10)%2 == 0 {
				fill = &image.Uniform{color.White}
			}
			draw.Draw(img, image.Rect(x, y, x+10, y+10), fill, image.ZP, draw.Src)
		}
	}

	return img
}

//----------------------------------------------------------------------------------
// generate clock face image
// - http://studygolang.com/articles/223
//----------------------------------------------------------------------------------
func GenClockImage(size int) image.Image {
	const clock_size = 400
	const radius = clock_size / 3

	var colour color.RGBA

	circle := func(clock *image.RGBA) {
		for angle := float64(0); angle < 360; angle++ {
			radian_angle := math.Pi * 2 * angle / 360
			x := radius*math.Sin(radian_angle) + clock_size/2
			y := radius*math.Cos(radian_angle) + clock_size/2
			clock.Set(int(x), int(y), colour)
		}
	}
	hand := func(clock *image.RGBA, angle float64, length float64) {
		radian_angle := math.Pi * 2 * angle
		x_inc := math.Sin(radian_angle)
		y_inc := -math.Cos(radian_angle)
		for i := float64(0); i < length; i++ {
			x := i*x_inc + clock_size/2
			y := i*y_inc + clock_size/2
			clock.Set(int(x), int(y), colour)
		}
	}

	clock := image.NewRGBA(image.Rect(0, 0, clock_size, clock_size))
	colour.A = 255
	circle(clock)

	time := time.Now()
	colour.R, colour.G, colour.B = 255, 0, 0
	hand(clock, (float64(time.Hour())+float64(time.Minute())/60)/12, radius*0.5) // hour hand

	colour.R, colour.G, colour.B = 0, 255, 0
	hand(clock, (float64(time.Minute())+float64(time.Second())/60)/60, radius*0.6) // minute hand

	colour.R, colour.G, colour.B = 0, 0, 255
	hand(clock, float64(time.Second())/60, radius*0.8) // Second hand

	return clock
}

// ---------------------------------E-----N-----D-----------------------------------
