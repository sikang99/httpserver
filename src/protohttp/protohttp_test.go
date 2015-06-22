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
	ph1 := NewProtoHttp()
	fmt.Println(ph1)

	ph2 := NewProtoHttp("bstring", "Testing")
	fmt.Println(ph2)

	ph3 := NewProtoHttpWithPorts("8001", "8002", "8003")
	fmt.Println(ph3)
}

//------------------------------------------------------------------
// test for information
//------------------------------------------------------------------
func TestServer(t *testing.T) {

}
