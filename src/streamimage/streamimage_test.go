//==================================================================================
// Image
// - https://www.socketloop.com/tutorials/golang-save-image-to-png-jpeg-or-gif-format
//==================================================================================

package streamimage

import (
	"log"
	"testing"
)

//----------------------------------------------------------------------------------
// string information for the slot
//----------------------------------------------------------------------------------
func TestMakeImageFiles(t *testing.T) {
	var err error

	//img := GenSimpleImage(1024, 768)
	//img := GenGradientImage(1024, 768)
	img := GenSpiralImage(1080, 768)
	//img := GenRandomImage(1024, 768)
	//img := GenClockImage(1000)
	//img := GenFractalImage(800, 800, 100000)

	err = PutImageToFile(img, "output.png", 0)
	if err != nil {
		log.Fatalln(err)
	}

	err = PutImageToFile(img, "output.jpg", 80)
	if err != nil {
		log.Fatalln(err)
	}

	err = PutImageToFile(img, "output.gif", 256)
	if err != nil {
		log.Fatalln(err)
	}

	err = PutImageToFile(img, "output.vid", 999)
	if err != nil {
		log.Println(err)
	}
}

// ---------------------------------E-----N-----D-----------------------------------
