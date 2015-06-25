//==================================================================================
// Author: Stoney Kang, sikang99@gmail.com, 2015
// Test for Stream Info : channel, source, track
//==================================================================================

package streaminfo

import (
	"fmt"
	"log"
	"testing"

	sb "stoney/httpserver/src/streambase"
)

//----------------------------------------------------------------------------------
// test for stream request info
//----------------------------------------------------------------------------------
func TestStreamInfo(t *testing.T) {
	trk := NewTrack()
	fmt.Println(trk)

	src := NewSource(3) // (N) : N tracks
	fmt.Println(src)

	chn := NewChannel(4, 3) // (M,N) : M sources, N tracks
	fmt.Println(chn)

	fmt.Printf("\tCH(0): %s\n", chn.GetId())
	fmt.Printf("\tCH(0)/SC(3): %s\n", chn.Srcs[3].GetId())
	fmt.Printf("\tCH(0)/SC(3)/TK(2): %s\n", chn.Srcs[3].Trks[2].GetId())
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

	uri = "ws://localhost:8080/stream?track=888&what=xxxx"
	sreq, err = GetStreamRequestFromURI(uri)
	if err != nil {
		log.Fatalln(err)
	}
	fmt.Println(sreq)

	uri = "ws://localhost:8080/command?track=333&channel=300&source=3"
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
	uid := GetStdUUID()
	fmt.Println(uid)

	uid = GetStdUUID()
	fmt.Println(uid)

	uib := ParseStdUUID(uid)
	sb.HexDump(uib)

	uid = GetPseudoUUID()
	fmt.Println(uid)
}

// ---------------------------------E-----N-----D-----------------------------------
