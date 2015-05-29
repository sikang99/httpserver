//==================================================================================
// Image
// - https://github.com/golang-samples/image
// - https://www.socketloop.com/tutorials/golang-save-image-to-png-jpeg-or-gif-format
//==================================================================================

package streamimage

import (
	"bytes"
	"fmt"
	"image"
	"image/color"
	"image/draw"
	"image/gif"
	"image/jpeg"
	"image/png"
	"io"
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
// encode image to buffer in PNG, JPEG, GIF
//----------------------------------------------------------------------------------
func PutImageToBuffer(img image.Image, itype string, optnum int) ([]byte, error) {
	var err error

	buf := new(bytes.Buffer)
	err = EncodeImageByType(buf, img, itype, optnum)

	return buf.Bytes(), err
}

//----------------------------------------------------------------------------------
// encode image to file in PNG, JPEG, GIF
//----------------------------------------------------------------------------------
func PutImageToFile(img image.Image, fname string, optnum int) error {
	var err error

	out, err := os.Create(fname)
	if err != nil {
		log.Println(err)
		return err
	}
	defer out.Close()

	itype := filepath.Ext(fname)
	return EncodeImageByType(out, img, itype[1:], optnum)
}

//----------------------------------------------------------------------------------
// encode image to writer in PNG, JPEG, GIF
//----------------------------------------------------------------------------------
func EncodeImageByType(out io.Writer, img image.Image, itype string, optnum int) error {
	var err error

	switch itype {
	case "png":
		err = png.Encode(out, img)
		if err != nil {
			log.Println(err)
			return err
		}

	case "gif":
		var opt gif.Options
		opt.NumColors = optnum // 256, you can add more parameters if you want

		err = gif.Encode(out, img, &opt) // put num of colors to 256
		if err != nil {
			log.Println(err)
			return err
		}

	case "jpg", "jpeg":
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

	// fill the color
	for y := img.Rect.Min.Y; y < img.Rect.Max.Y; y++ {
		for x := img.Rect.Min.X; x < img.Rect.Max.X; x++ {
			img.Set(x, y, color.RGBA{0x88, 0xff, 0x88, 0xff})
		}
	}

	// fill the color
	draw.Draw(img, img.Bounds(), &image.Uniform{blue}, image.ZP, draw.Src)

	// draw a line
	for i := img.Bounds().Min.X; i < img.Bounds().Max.X; i++ {
		img.Set(i, img.Bounds().Max.Y/2, white) // to change a single pixel
	}

	return img
}

//----------------------------------------------------------------------------------
// generate gradient image
// - https://github.com/felixpalmer/go_images
//----------------------------------------------------------------------------------
func GenGradientImage(xz, yz int) image.Image {
	img := image.NewRGBA(image.Rect(0, 0, xz, yz))
	size := img.Bounds().Size()

	for x := 0; x < size.X; x++ {
		for y := 0; y < size.Y; y++ {
			color := color.RGBA{
				uint8(255 * x / size.X),
				uint8(255 * y / size.Y),
				55,
				255}
			img.Set(x, y, color)
		}
	}

	return img
}

//----------------------------------------------------------------------------------
// generate gradient image
// - https://github.com/felixpalmer/go_images
//----------------------------------------------------------------------------------
func GenSpiralImage(xz, yz int) image.Image {
	canvas := NewCanvas(image.Rect(0, 0, xz, yz))
	canvas.DrawGradient()

	// Draw a set of spirals randomly over the image
	rand.Seed(time.Now().UTC().UnixNano())
	for i := 0; i < 100; i++ {
		x := float64(xz) * rand.Float64()
		y := float64(yz) * rand.Float64()
		color := color.RGBA{uint8(rand.Intn(255)),
			uint8(rand.Intn(255)),
			uint8(rand.Intn(255)),
			255}

		canvas.DrawSpiral(color, Vector{x, y})
	}

	return canvas
}

//----------------------------------------------------------------------------------
// generate random image
//----------------------------------------------------------------------------------
func GenRandomImage(xz, yz int) image.Image {
	rand.Seed(time.Now().UTC().UnixNano())

	imgRect := image.Rect(0, 0, xz, yz)
	img := image.NewRGBA(imgRect)
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
func GenClockImage(clock_size int) image.Image {
	radius := float64(clock_size / 3)

	var colour color.RGBA

	circle := func(clock *image.RGBA) {
		for angle := float64(0); angle < 360; angle++ {
			radian_angle := math.Pi * 2 * angle / 360
			x := radius*math.Sin(radian_angle) + float64(clock_size/2)
			y := radius*math.Cos(radian_angle) + float64(clock_size/2)
			clock.Set(int(x), int(y), colour)
		}
	}
	hand := func(clock *image.RGBA, angle float64, length float64) {
		radian_angle := math.Pi * 2 * angle
		x_inc := math.Sin(radian_angle)
		y_inc := -math.Cos(radian_angle)
		for i := float64(0); i < length; i++ {
			x := i*x_inc + float64(clock_size/2)
			y := i*y_inc + float64(clock_size/2)
			clock.Set(int(x), int(y), colour)
		}
	}

	img := image.NewRGBA(image.Rect(0, 0, clock_size, clock_size))
	colour.A = 255
	circle(img)

	time := time.Now()
	colour.R, colour.G, colour.B = 255, 0, 0
	hand(img, (float64(time.Hour())+float64(time.Minute())/60)/12, radius*0.5) // hour hand

	colour.R, colour.G, colour.B = 0, 255, 0
	hand(img, (float64(time.Minute())+float64(time.Second())/60)/60, radius*0.6) // minute hand

	colour.R, colour.G, colour.B = 0, 0, 255
	hand(img, float64(time.Second())/60, radius*0.8) // Second hand

	return img
}

//----------------------------------------------------------------------------------
// generate fractal image
// - https://cyberroadie.wordpress.com/2012/04/28/go-fern-fractal/
//----------------------------------------------------------------------------------
func transformPoint(x, y, a, b, c, d, s float32) (float32, float32) {
	return ((a * x) + (b * y)), ((c * x) + (d * y) + s)
}

func transform(x float32, y float32) (float32, float32) {
	rnd := rand.Intn(101)
	switch {
	case rnd == 1:
		x, y = transformPoint(x, y, 0.0, 0.0, 0.0, 0.16, 0.0)
	case rnd <= 7:
		x, y = transformPoint(x, y, 0.2, -0.26, 0.23, 0.22, 0.0)
	case rnd <= 14:
		x, y = transformPoint(x, y, -0.15, 0.28, 0.26, 0.24, 0.44)
	case rnd <= 100:
		x, y = transformPoint(x, y, 0.85, 0.04, -0.04, 0.85, 1.6)
	}
	return x, y
}

func drawPoint(m *image.RGBA, x float32, y float32) {
	b := m.Bounds()
	height := float32(b.Max.Y)
	width := float32(b.Max.X)
	scale := float32(height / 11)
	y = (height - 25) - (scale * y)
	x = (width / 2) + (scale * x)
	m.Set(int(x), int(y), color.RGBA{0, 255, 0, 255})
}

func drawFern(m *image.RGBA, x float32, y float32, steps int) {
	if steps != 0 {
		x, y = transform(x, y)
		drawPoint(m, x, y)
		drawFern(m, x, y, steps-1)
	}
}

func GenFractalImage(xz, yz, optnum int) image.Image {
	img := image.NewRGBA(image.Rect(0, 0, xz, yz))

	blue := color.RGBA{0, 0, 0, 255}
	draw.Draw(img, img.Bounds(), &image.Uniform{blue}, image.ZP, draw.Src)
	drawFern(img, float32(xz), float32(yz), optnum)

	return img
}

// ---------------------------------E-----N-----D-----------------------------------
