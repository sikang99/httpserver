package main

import (
	"fmt"
	"log"
	"net"
)

type Config struct {
	Port       string `json:"port"`
	SecurePort string `json:"secure_port"`
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
