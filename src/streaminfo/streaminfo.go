//==================================================================================
// Info
// - https://github.com/nu7hatch/gouuid - Go binding for libuuid
// - https://github.com/sony/sonyflake - A distributed unique ID generator inspired by Twitter's Snowflake
//==================================================================================

package streaminfo

import (
	"crypto/rand"
	"fmt"
	"log"
	"net/url"
	"strconv"
	"time"

	//"github.com/twinj/uuid"
	//"github.com/nu7hatch/gouuid"
	uuid "github.com/nu7hatch/gouuid"
)

const (
	NUM_DEF_TRACKS  = 4
	NUM_DEF_SOURCES = 2
)

//----------------------------------------------------------------------------------
type Uid int

type Track struct {
	Id   Uid
	Desc string
}

//----------------------------------------------------------------------------------
// string source information
//----------------------------------------------------------------------------------
func (trk *Track) String() string {
	//str := fmt.Sprintf("[Track]")
	str := fmt.Sprintf("\tId: %v", trk.Id)
	str += fmt.Sprintf("\tDesc: %s", trk.Desc)
	return str
}

//----------------------------------------------------------------------------------
type Source struct {
	Id     Uid
	Desc   string
	Time   time.Time
	Tracks []Track // media track such as audio, video, text, ...
}

//----------------------------------------------------------------------------------
// string source information
//----------------------------------------------------------------------------------
func (src *Source) String() string {
	//str := fmt.Sprintf("[Source]")
	str := fmt.Sprintf("\tId: %v", src.Id)
	str += fmt.Sprintf("\tTime: %v", src.Time)
	str += fmt.Sprintf("\tDesc: %s", src.Desc)
	return str
}

func NewSource(num int) *Source {
	return &Source{
		Time:   time.Now(),
		Desc:   "blank source",
		Tracks: make([]Track, num),
	}
}

func (src *Source) GetId() Uid {
	return src.Id
}

//----------------------------------------------------------------------------------
type Channel struct {
	Id   Uid
	Name string
	Desc string
	Time time.Time
	Srcs []Source
}

//----------------------------------------------------------------------------------
// string chanel information
//----------------------------------------------------------------------------------
func (chn *Channel) String() string {
	str := fmt.Sprintf("[Channel]")
	str += fmt.Sprintf("\tId: %d", chn.Id)
	str += fmt.Sprintf("\tName: %s", chn.Name)
	str += fmt.Sprintf("\tTime: %v", chn.Time)
	str += fmt.Sprintf("\tDesc: %s\n", chn.Desc)
	for i := range chn.Srcs {
		str += fmt.Sprintf("\t[%d] %s\n", i, &chn.Srcs[i])
	}
	return str
}

//----------------------------------------------------------------------------------
// make a new channel with the number of sources
//----------------------------------------------------------------------------------
func NewChannel(num int) *Channel {
	return &Channel{
		Time: time.Now(),
		Desc: "blank channel",
		Srcs: make([]Source, num),
	}
}

//----------------------------------------------------------------------------------
// get/set channel id
//----------------------------------------------------------------------------------
func (chn *Channel) GetId() Uid {
	return chn.Id
}

func (chn *Channel) SetId(id Uid) Uid {
	chn.Id = id
	return chn.Id
}

//----------------------------------------------------------------------------------
type StreamRequest struct {
	Channel   int
	Source    int
	StartTime time.Time // access time
	Who       string    // string for hostname:port
	Desc      string
}

//----------------------------------------------------------------------------------
// string stream request information
//----------------------------------------------------------------------------------
func (sr *StreamRequest) String() string {
	str := fmt.Sprintf("[Request]")
	str += fmt.Sprintf("\tChannel: %d", sr.Channel)
	str += fmt.Sprintf("\tSource: %d", sr.Source)
	str += fmt.Sprintf("\tWho: %s\n", sr.Who)
	str += fmt.Sprintf("\tStartAt: %s", sr.StartTime)
	str += fmt.Sprintf("\tDesciption: %s\n", sr.Desc)
	return str
}

//----------------------------------------------------------------------------------
// get information of stream request
//----------------------------------------------------------------------------------
func GetStreamRequestFrom(str string) (*StreamRequest, error) {
	var err error

	params, err := url.ParseQuery(str)
	if err != nil {
		log.Println(err)
		return nil, err
	}
	//fmt.Println(params)

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
		Channel:   chn,
		Source:    src,
		StartTime: time.Now(),
	}
	//fmt.Println(sreq)

	return sreq, err
}

//----------------------------------------------------------------------------------
// get information of stream request from URI
// - https://gobyexample.com/url-parsing
// - https://www.socketloop.com/tutorials/golang-get-uri-segments-by-number-and-assign-as-variable-example
//----------------------------------------------------------------------------------
func GetStreamRequestFromURI(uri string) (*StreamRequest, error) {
	var err error

	ureq, err := url.ParseRequestURI(uri)
	if err != nil {
		log.Println(err)
		return nil, err
	}

	sreq, err := GetStreamRequestFrom(ureq.RawQuery)
	if err != nil {
		log.Println(err)
		return nil, err
	}

	sreq.Desc = uri

	return sreq, err
}

//----------------------------------------------------------------------------------
// get a new  uuid
// - http://stackoverflow.com/questions/15130321/is-there-a-method-to-generate-a-uuid-with-go-language
//----------------------------------------------------------------------------------
func StdGetUUID() (uid string) {
	u4, err := uuid.NewV4()
	if err != nil {
		log.Println(err)
		return
	}
	uid = u4.String()
	return
}

func StdParseUUID(uid string) (b []byte) {
	u, err := uuid.ParseHex(uid)
	if err != nil {
		log.Println(err)
		return
	}
	copy(b, u[:])
	return
}

func PseudoGetUUID() (uid string) {
	b := make([]byte, 16)
	_, err := rand.Read(b)
	if err != nil {
		log.Println(err)
		return
	}
	uid = fmt.Sprintf("%x-%x-%x-%x-%x", b[0:4], b[4:6], b[6:8], b[8:10], b[10:])
	return
}

// ---------------------------------E-----N-----D-----------------------------------
