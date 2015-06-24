//==================================================================================
// Author: Stoney Kang, sikang99@gmail.com, 2015
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
	"time"

	//"github.com/twinj/uuid"
	//"github.com/nu7hatch/gouuid"
	uuid "github.com/nu7hatch/gouuid"
)

const (
	NUM_DEF_TRACKS  = 4
	NUM_DEF_SOURCES = 2

	ID_DEF_CHANNEL = "100"
	ID_DEF_SOURCE  = "110"
)

//==================================================================================
// track struc
//----------------------------------------------------------------------------------
type Track struct {
	Id   string
	Desc string
}

func NewTrack() *Track {
	nt := &Track{Id: "1", Desc: "blank track"}
	return nt
}

//----------------------------------------------------------------------------------
// string source information
//----------------------------------------------------------------------------------
func (trk *Track) BaseString() string {
	str := fmt.Sprintf("[Track]")
	str += fmt.Sprintf("\tId: %s", trk.Id)
	str += fmt.Sprintf("\tDesc: %s", trk.Desc)
	return str
}

func (trk *Track) String() string {
	str := fmt.Sprintf("%s", trk.BaseString())
	return str
}

func (trk *Track) GetId() string {
	return trk.Id
}

func (trk *Track) SetId(id int) string {
	trk.Id = fmt.Sprintf("%d", id+1)
	return trk.Id
}

//==================================================================================
// source struc
//----------------------------------------------------------------------------------
type Source struct {
	Id   string
	Desc string
	Time time.Time
	Trks []Track // media track such as audio, video, text, ...
}

func NewSource(num int) *Source {
	trks := make([]Track, num)
	for j := range trks {
		trks[j].Id = fmt.Sprintf("%d", 10+(j+1))
	}
	ns := &Source{
		Id:   "10",
		Time: time.Now(),
		Desc: "blank source",
		Trks: trks,
	}
	return ns
}

//----------------------------------------------------------------------------------
// string source information
//----------------------------------------------------------------------------------
func (src *Source) BaseString() string {
	str := fmt.Sprintf("[Source]")
	str += fmt.Sprintf("\tId: %s", src.Id)
	str += fmt.Sprintf("\tTime: %v", src.Time)
	str += fmt.Sprintf("\tDesc: %s", src.Desc)
	return str
}

func (src *Source) String() string {
	str := fmt.Sprintf("%s\n", src.BaseString())
	for i := range src.Trks {
		str += fmt.Sprintf("\t[%d] %s\n", i, src.Trks[i].String())
	}
	return str
}

func (src *Source) GetId() string {
	return src.Id
}

func (src *Source) SetId(id int) string {
	src.Id = fmt.Sprintf("%d%d", (id+1)*10)
	return src.Id
}

//==================================================================================
// channel struct
//----------------------------------------------------------------------------------
type Channel struct {
	Id     string
	Name   string
	Desc   string
	Time   time.Time
	Status int
	Use    int
	Srcs   []Source
}

func NewChannel(snum, tnum int) *Channel {
	srcs := make([]Source, snum)
	for i := range srcs {
		srcs[i].Id = fmt.Sprintf("%d%d", 1, i+1)
		srcs[i].Time = time.Now()
		trks := make([]Track, tnum)
		srcs[i].Trks = trks
		for j := range trks {
			trks[j].Id = fmt.Sprintf("%d%d%d", 1, i+1, j+1)
		}
	}

	nc := &Channel{
		Id:   "100",
		Time: time.Now(),
		Desc: "blank channel",
		Srcs: srcs,
	}

	return nc
}

//----------------------------------------------------------------------------------
// string chanel information
//----------------------------------------------------------------------------------
func (chn *Channel) BaseString() string {
	str := fmt.Sprintf("[Channel]")
	str += fmt.Sprintf("\tId: %s", chn.Id)
	str += fmt.Sprintf("\tName: %s", chn.Name)
	str += fmt.Sprintf("\tTime: %v", chn.Time)
	str += fmt.Sprintf("\tStatus: %v", chn.Status)
	str += fmt.Sprintf("\tUse: %v", chn.Use)
	str += fmt.Sprintf("\tDesc: %s", chn.Desc)
	return str
}

func (chn *Channel) String() string {
	str := fmt.Sprintf("%s\n", chn.BaseString())
	for i := range chn.Srcs {
		str += fmt.Sprintf("\t[%d] %s\n", i, chn.Srcs[i].BaseString())
		for j := range chn.Srcs[i].Trks {
			str += fmt.Sprintf("\t\t[%d] %s\n", j, chn.Srcs[i].Trks[j].String())
		}
	}

	return str
}

//----------------------------------------------------------------------------------
// get/set channel id
//----------------------------------------------------------------------------------
func (chn *Channel) GetId() string {
	return chn.Id
}

func (chn *Channel) SetId(id int) string {
	chn.Id = fmt.Sprintf("%d", (id+1)*100)
	return chn.Id
}

//==================================================================================
// stream request struc
//----------------------------------------------------------------------------------
type StreamRequest struct {
	Channel string
	Source  string
	Time    time.Time // access time
	Who     string    // string for hostname:port
	Desc    string
}

//----------------------------------------------------------------------------------
// string stream request information
//----------------------------------------------------------------------------------
func (sq *StreamRequest) String() string {
	str := fmt.Sprintf("[Request]")
	str += fmt.Sprintf("\tChannel: %v", sq.Channel)
	str += fmt.Sprintf("\tSource: %v", sq.Source)
	str += fmt.Sprintf("\tWho: %s\n", sq.Who)
	str += fmt.Sprintf("\tStartAt: %s", sq.Time)
	str += fmt.Sprintf("\tDesciption: %s\n", sq.Desc)
	return str
}

//----------------------------------------------------------------------------------
// get information of stream request
//----------------------------------------------------------------------------------
func GetStreamRequestFromQuery(str string) (*StreamRequest, error) {
	var err error

	params, err := url.ParseQuery(str)
	if err != nil {
		log.Println(err)
		return nil, err
	}
	//fmt.Println(params)

	// assign default values
	sreq := &StreamRequest{
		Channel: ID_DEF_CHANNEL,
		Source:  ID_DEF_SOURCE,
		Time:    time.Now(),
	}

	if len(params) > 0 {
		sreq.Channel = params["channel"][0]
		sreq.Source = params["source"][0]
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

	sreq, err := GetStreamRequestFromQuery(ureq.RawQuery)
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
