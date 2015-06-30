//=================================================================================
// Author: Stoney Kang, sikang99@gmail.com, 2015
// Package for handling protocol base information
// - http://blog.labix.org/2011/10/09/death-of-goroutines-under-control
//==================================================================================

package protobase

import (
	"fmt"
	"time"

	tomb "gopkg.in/tomb.v2"

	sb "stoney/httpserver/src/streambase"
	si "stoney/httpserver/src/streaminfo"
)

//---------------------------------------------------------------------------
type ProtoBase struct {
	// mandatory part
	Boundary string
	Status   int // run state
	// optional part
	Id       string
	URI      string
	Scheme   string
	User     string
	Password string
	Host     string
	Port     string
	Desc     string
	Tomb     tomb.Tomb
	Sign     chan string // signaling for state control
}

//---------------------------------------------------------------------------
// new ProtoBase struct
//---------------------------------------------------------------------------
func NewProtoBase() *ProtoBase {
	pb := &ProtoBase{
		Id:     si.GetNewId(),
		Status: sb.STATUS_IDLE,
		Sign:   make(chan string),
	}

	return pb
}

//---------------------------------------------------------------------------
// string ProtoBase information
//---------------------------------------------------------------------------
func (pb *ProtoBase) String() string {
	str := fmt.Sprintf("\tId: %s", pb.Id)
	str += fmt.Sprintf("\tStatus: %s(%d)", sb.StatusText[pb.Status], pb.Status)
	str += fmt.Sprintf("\tDesc: %s", pb.Desc)
	/*
		str += fmt.Sprintf("\tScheme: %s", pb.Scheme)
		str += fmt.Sprintf("\tUser: %s", pb.User)
		str += fmt.Sprintf("\tPassword: %s", pb.Password)
		str += fmt.Sprintf("\tHost: %s", pb.Host)
		str += fmt.Sprintf("\tPort: %s", pb.Port)
	*/
	return str
}

//---------------------------------------------------------------------------
// check status of struct
//---------------------------------------------------------------------------
func (pb *ProtoBase) IsRun() bool {
	if pb.Status == sb.STATUS_RUN {
		return true
	} else {
		return false
	}
}

func (pb *ProtoBase) Reset() {
	pb.Status = sb.STATUS_IDLE
}

//---------------------------------------------------------------------------
// set status
//---------------------------------------------------------------------------
func (pb *ProtoBase) SetStatusRun() error {
	if pb.Status == sb.STATUS_IDLE {
		pb.Status = sb.STATUS_RUN
		return nil
	} else {
		return sb.ErrStatus
	}
}

func (pb *ProtoBase) SetStatusIdle() error {
	if pb.Status == sb.STATUS_RUN {
		pb.Status = sb.STATUS_IDLE
		return nil
	} else {
		return sb.ErrStatus
	}
}

func (pb *ProtoBase) SetStatusClose() error {
	if pb.Status == sb.STATUS_RUN {
		pb.Status = sb.STATUS_CLOSE
		return nil
	} else {
		return sb.ErrStatus
	}
}

//---------------------------------------------------------------------------
// test function for tomb package
//---------------------------------------------------------------------------
func (pb *ProtoBase) Run() error {
	var err error

	pb.Sign <- si.GetStdUUID()

	for i := 0; i < 100; i++ {
		select {
		case str := <-pb.Sign:
			if str == "stop" {
				pb.Sign <- "stopped"
				fmt.Printf("%d %s -> %s\n", i, str, "stopped")
				return err
			} else {
				pb.Sign <- "alive"
				fmt.Printf("%d %s -> %s\n", i, str, "alive")
			}
		case <-pb.Tomb.Dying():
			fmt.Printf("%d %s -> %s\n", i, "dying", "dyed")
			return err
		default:
			fmt.Println(i, "none")
		}

		time.Sleep(time.Second)
	}

	return err
}

func (pb *ProtoBase) Dye() error {
	var err error

	pb.Tomb.Kill(nil)
	pb.Tomb.Wait()

	return err
}

func (pb *ProtoBase) Stop() error {
	var err error
	pb.Sign <- "stop"
	return err
}

func (pb *ProtoBase) Check() error {
	var err error
	pb.Sign <- "check"
	return err
}

// ---------------------------------E-----N-----D--------------------------------
