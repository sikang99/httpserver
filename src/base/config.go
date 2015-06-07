package base

import (
	"fmt"
	"log"
	"net"
	"net/http"
	"os"
	"time"
)

type Config struct {
	Port       string `json:"port"`
	SecurePort string `json:"secure_port"`
	Http2Port  string `json:"http2_port"`
	LogFile    string `json:"log_file"`
	Desciption string `json:"description"`
}

func ShowNetInterfaces() {
	list, err := net.Interfaces()
	if err != nil {
		log.Fatal(err)
	}

	for i, iface := range list {
		fmt.Printf("%d %s %v\n", i, iface.Name, iface)
		addrs, err := iface.Addrs()
		if err != nil {
			log.Fatal(err)
		}
		for j, addr := range addrs {
			fmt.Printf("\t%d %v\n", j, addr)
		}
	}
}

func openLogFile(logfile string) {
	if logfile != "" {
		lf, err := os.OpenFile(logfile, os.O_WRONLY|os.O_CREATE|os.O_APPEND, 0640)

		if err != nil {
			log.Fatal("OpenLogfile: os.OpenFile:", err)
		}

		log.SetOutput(lf)
	}
}

// http://stackoverflow.com/questions/30552447/how-to-set-which-ip-to-use-for-a-http-request
// Create a transport like http.DefaultTransport, but with a specified localAddr
func setConfig() {
	transport := &http.Transport{
		Proxy: http.ProxyFromEnvironment,
		Dial: (&net.Dialer{
			Timeout:   30 * time.Second,
			KeepAlive: 30 * time.Second,
			//LocalAddr: localAddr,
		}).Dial,
		TLSHandshakeTimeout: 10 * time.Second,
	}

	client := &http.Client{
		Transport: transport,
	}

	fmt.Println(client)
}
