//==================================================================================
// Author: Stoney Kang, sikang99@gmail.com, 2015
// Image
// - https://www.socketloop.com/tutorials/golang-save-image-to-png-jpeg-or-gif-format
//==================================================================================

package streamimage

import (
	"fmt"
	"log"
	"testing"

	"github.com/nfnt/resize"

	sb "stoney/httpserver/src/streambase"
)

//----------------------------------------------------------------------------------
// string information for the slot
// - https://github.com/nfnt/resize
//----------------------------------------------------------------------------------
func TestMakeImageFiles(t *testing.T) {
	var err error

	//image := GenSimpleImage(1024, 768)
	//image := GenGradientImage(1024, 768)
	//image := GenSpiralImage(1080, 768)
	//image := GenRandomImageColor(1024, 768)
	//image := GenRandomImageBlock(1024, 768)
	//image := GenClockImage(1000)
	image := GenFractalImage(800, 800, 100000)

	err = PutImageToFile(image, "output.png", 0)
	if err != nil {
		log.Fatalln(err)
	}

	err = PutImageToFile(image, "output.jpg", 80)
	if err != nil {
		log.Fatalln(err)
	}

	err = PutImageToFile(image, "output.gif", 256)
	if err != nil {
		log.Fatalln(err)
	}

	err = PutImageToFile(image, "output.vid", 999)
	if err != nil {
		log.Println(err)
	}

	newImage := resize.Resize(400, 0, image, resize.Lanczos3)
	err = PutImageToFile(newImage, "output_resize.png", 0)
	if err != nil {
		log.Fatalln(err)
	}
}

//----------------------------------------------------------------------------------
// string information for the slot
//----------------------------------------------------------------------------------
func TestStreamImage(t *testing.T) {
	var err error

	image := GenFractalImage(800, 800, 100000)
	data, err := PutImageToBuffer(image, "jpg", 80)
	if err != nil {
		log.Fatalln(err)
	}

	fmt.Printf("%d KB\n", len(data)/sb.KBYTE)
}

// ---------------------------------E-----N-----D-----------------------------------
