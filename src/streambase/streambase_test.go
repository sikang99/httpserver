//==================================================================================
// Author: Stoney Kang, sikang99@gmail.com, 2015
// Test for stream buffer for multipart media
//==================================================================================

package streambase

import (
	"errors"
	"fmt"
	"log"
	"math/rand"
	"testing"
	"time"

	"github.com/cenkalti/backoff"
)

//----------------------------------------------------------------------------------
// test for status
//----------------------------------------------------------------------------------
func TestStatusValue(t *testing.T) {
	fmt.Println(STATUS_USING)
	fmt.Println(STATUS_IDLE)
}

//----------------------------------------------------------------------------------
// test for timestamp
//----------------------------------------------------------------------------------
func TestTimestamp(t *testing.T) {
	//tstamp, _ := time.Parse(time.RFC3339, strconv.Itoa(file.CreatedAt))
	//println(tstamp)

	ctime := time.Now().Unix()
	tstring := time.Unix(ctime, 0).Format(time.RFC3339)
	fmt.Printf("\tCurrent Timestamp: %v, %v\n", ctime, tstring)

	dsec := MakeTimestamp()
	sec := MakeTimestampSecond()
	msec := MakeTimestampMillisecond()
	nsec := MakeTimestampNanosecond()
	fmt.Printf("%v\n%v\n%v\n%v\n", dsec, sec, msec, nsec)
}

//----------------------------------------------------------------------------------
// test for timer and ticker
// - https://mmcgrana.github.io/2012/09/go-by-example-timers-and-tickers.html
//----------------------------------------------------------------------------------
func TestTimerAndTicker(t *testing.T) {

	timeChan := time.NewTimer(time.Second).C
	tickChan := time.NewTicker(time.Millisecond * 400).C
	doneChan := make(chan bool)

	go func() {
		time.Sleep(time.Second * 2)
		doneChan <- true
	}()

	for {
		select {
		case <-timeChan:
			fmt.Println("Timer expired")
		case <-tickChan:
			fmt.Println("Ticker ticked")
		case <-doneChan:
			fmt.Println("Done")
			return
		}
	}
}

//----------------------------------------------------------------------------------
// test for timer and ticker
// - https://github.com/cenkalti/backoff
//----------------------------------------------------------------------------------
func TestBackoff(t *testing.T) {
	rnd := rand.New(rand.NewSource(time.Now().UnixNano()))

	operation := func() error {
		rn := rnd.Intn(100)
		fmt.Println(rn)

		if rn < 80 {
			return errors.New("error")
		} else {
			return nil
		}
	}

	err := backoff.Retry(operation, backoff.NewExponentialBackOff())
	if err != nil {
		log.Println(err)
	}
}

//----------------------------------E-----N-----D-----------------------------------
