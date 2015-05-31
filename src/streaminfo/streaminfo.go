//==================================================================================
// Info
//==================================================================================

package streaminfo

import (
	"fmt"
	"log"
	"net/url"
	"strconv"
)

//----------------------------------------------------------------------------------
type Source struct {
	Id   int
	Desc string
}

type Channel struct {
	Id   int
	Name string
	Srcs []Source
	Desc string
}

//----------------------------------------------------------------------------------
// string chanel information
//----------------------------------------------------------------------------------
func (chn *Channel) String() string {
	str := fmt.Sprintf("Id: %d\t", chn.Id)
	str += fmt.Sprintf("Name: %s\t", chn.Name)
	str += fmt.Sprintf("Desc: %s\n", chn.Desc)
	return str
}

func (chn *Channel) GetId() int {
	return chn.Id
}

//----------------------------------------------------------------------------------
type StreamRequest struct {
	Channel int
	Source  int
	Addr    string
}

//----------------------------------------------------------------------------------
// string chanel information
//----------------------------------------------------------------------------------
func (sr *StreamRequest) String() string {
	str := fmt.Sprintf("Channel: %d\t", sr.Channel)
	str += fmt.Sprintf("Source: %d\t", sr.Source)
	str += fmt.Sprintf("Address: %s\n", sr.Addr)
	return str
}

func GetStreamRequest(str string) (*StreamRequest, error) {
	var err error

	params, err := url.ParseQuery(str)
	fmt.Println(params)

	chn, err := strconv.Atoi(params["channel"][0])
	if err != nil {
		log.Println(err)
		return nil, err
	}

	src, err := strconv.Atoi(params["source"][0])
	if err != nil {
		log.Println(err)
		return nil, err
	}

	sreq := &StreamRequest{
		Channel: chn,
		Source:  src,
	}

	return sreq, err
}

//----------------------------------------------------------------------------------
// get stream information
// - https://gobyexample.com/url-parsing
// - https://www.socketloop.com/tutorials/golang-get-uri-segments-by-number-and-assign-as-variable-example
//----------------------------------------------------------------------------------
func GetStreamInfoFromUrl(uri string) (*Channel, error) {
	var err error

	u, err := url.ParseRequestURI(uri)
	if err != nil {
		log.Println(err)
		return nil, err
	}

	fmt.Println(u.Scheme)
	fmt.Println(u.User)
	fmt.Println(u.RawQuery)
	fmt.Println(u.Fragment)

	params, _ := url.ParseQuery(u.RawQuery)
	fmt.Println(params)

	id, err := strconv.Atoi(params["channel"][0])
	if err != nil {
		log.Println(err)
		return nil, err
	}

	chn := &Channel{
		Id:   id,
		Desc: "Sample Channel",
	}

	return chn, err
}

// ---------------------------------E-----N-----D-----------------------------------
