//==================================================================================
// Author: Stoney Kang, sikang99@gmail.com, 2015
// Test for Stream Info
//==================================================================================

package streaminfo

import (
	"fmt"
	"log"
	"testing"
)

//----------------------------------------------------------------------------------
// test for stream request info
//----------------------------------------------------------------------------------
func TestStreamInfo(t *testing.T) {
	src := NewSource(4)
	fmt.Println(src)

	chn := NewChannel(5)
	fmt.Println(chn)
}

//----------------------------------------------------------------------------------
// test for stream request info
//----------------------------------------------------------------------------------
func TestStreamRequest(t *testing.T) {
	var err error

	uri := "https://localhost:8080/stream?channel=100&source=2&timeout=v#f"
	sreq, err := GetStreamRequestFromURI(uri)
	if err != nil {
		log.Fatalln(err)
	}
	fmt.Println(sreq)

	uri = "ws://localhost:8080/stream"
	sreq, err = GetStreamRequestFromURI(uri)
	if err != nil {
		log.Fatalln(err)
	}
	fmt.Println(sreq)
}

//----------------------------------------------------------------------------------
// test for uuid
//----------------------------------------------------------------------------------
func TestUUID(t *testing.T) {
	uid := StdGetUUID()
	fmt.Println(uid)

	uib := StdParseUUID(uid)
	fmt.Println(uib)

	uid = PseudoGetUUID()
	fmt.Println(uid)
}

// ---------------------------------E-----N-----D-----------------------------------
