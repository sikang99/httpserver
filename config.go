package main

import (
	"fmt"
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
		panic(err)
	}

	for i, iface := range list {
		fmt.Printf("%d name=%s %v\n", i, iface.Name, iface)
		addrs, err := iface.Addrs()
		if err != nil {
			panic(err)
		}
		for j, addr := range addrs {
			fmt.Printf(" %d %v\n", j, addr)
		}
	}
}
