//==================================================================
// Author : Stoney Kang, sikang99@gmail.com, 2015
// Test for HTTP streaming
//==================================================================

package protohttp

import (
	"fmt"
	"testing"
)

//------------------------------------------------------------------
// test for information
//------------------------------------------------------------------
func TestInfo(t *testing.T) {
	ph := NewProtoHttp("localhost", "8080")
	fmt.Println(ph)
}

//------------------------------------------------------------------
// test for information
//------------------------------------------------------------------
