//==================================================================================
// Image
// - https://www.socketloop.com/tutorials/golang-save-image-to-png-jpeg-or-gif-format
//==================================================================================

package image

import (
	"log"
	"testing"
)

//----------------------------------------------------------------------------------
// string information for the slot
//----------------------------------------------------------------------------------
func TestRandomImage(t *testing.T) {
	var err error

	//img := GenRandomImage(1080, 768)
	img := GenSimpleImage(1080, 768)

	err = MakeImageFile(img, "output.png", 0)
	if err != nil {
		log.Fatalln(err)
	}

	err = MakeImageFile(img, "output.jpg", 80)
	if err != nil {
		log.Fatalln(err)
	}

	err = MakeImageFile(img, "output.gif", 256)
	if err != nil {
		log.Fatalln(err)
	}

	err = MakeImageFile(img, "output.vid", 999)
	if err != nil {
		log.Println(err)
	}

	clock := GenClockImage(800)

	err = MakeImageFile(clock, "clock.jpg", 90)
	if err != nil {
		log.Println(err)
	}
}

// ---------------------------------E-----N-----D-----------------------------------
