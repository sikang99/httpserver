//=================================================================================
// Author: Stoney Kang, sikang99@gmail.com, 2015
// Package for handling protocol base information
//=================================================================================
package protobase

import (
	"fmt"
	"log"
	"testing"
	"time"

	sb "stoney/httpserver/src/streambase"
)

//---------------------------------------------------------------------------------
func init() {
	log.SetFlags(log.LstdFlags | log.Lshortfile)
}

//---------------------------------------------------------------------------------
// test for pkg info handling
//---------------------------------------------------------------------------------
func TestHandleInfo(t *testing.T) {
	pb := NewProtoBase()
	fmt.Println(pb)

	go pb.Run()
	fmt.Println(<-pb.Sign)

	time.Sleep(2 * time.Second)
	pb.Check()
	log.Println(<-pb.Sign)
	time.Sleep(2 * time.Second)
	pb.Stop()
	log.Println(<-pb.Sign)

	pb.Tomb.Go(pb.Run)
	fmt.Println(<-pb.Sign)

	time.Sleep(3 * time.Second)
	pb.Dye()

	fmt.Println(sb.FuncName())
	fmt.Println(sb.FName())
	sb.Trace()
}

//----------------------------------E-----N-----D----------------------------------
