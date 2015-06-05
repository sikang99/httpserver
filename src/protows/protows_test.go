//=================================================================================
// Author: Stoney Kang, sikang99@gmail.com, 2015
// Package for WebSocket Test
//=================================================================================

package protows

import (
	"fmt"
	"testing"
)

//---------------------------------------------------------------------------------
func TestProtoInfo(t *testing.T) {
	ws := NewProtoWs("www.google.com", "9000", "8443", "Testing")
	fmt.Println(ws)

	ws.Reset()
	fmt.Println(ws)

	ws.Clear()
	fmt.Println(ws)

	ws.SetAddr("www.facebook.com", "8080", "8443", "Redirect")
	fmt.Println(ws)
}

// ----------------------------------E-----N-----D---------------------------------
