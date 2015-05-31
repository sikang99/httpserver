//==================================================================================
// Info
//==================================================================================

package streaminfo

import (
	"fmt"
	"log"
	"testing"
)

//----------------------------------------------------------------------------------
// test for stream info
//----------------------------------------------------------------------------------
func TestStreamInfo(t *testing.T) {
	var err error

	uri := "https://localhost:8080/stream?channel=100;200&source=2&timeout=v#f"

	chn, err := GetStreamInfoFromUrl(uri)
	if err != nil {
		log.Fatalln(err)
	}

	fmt.Println(chn)
}

// ---------------------------------E-----N-----D-----------------------------------
