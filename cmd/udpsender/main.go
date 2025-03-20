package main

import (
	"bufio"
	"fmt"
	"log"
	"net"
	"os"
)

func main()  {
	updResolver, err := net.ResolveUDPAddr("udp", "localhost:42069")
	if err != nil {
		log.Fatalf("Failed to resolve UDP address: %v", err)
	}
	conn, err := net.DialUDP("udp", nil, updResolver)
	if err != nil {
		log.Fatalf("Failed to dial UDP: %v", err)
	}

	defer conn.Close()

	reader := bufio.NewReader(os.Stdin)
	for {
		fmt.Print("> ")
		line, err := reader.ReadString('\n')
		if err != nil {
			log.Fatalf("Failed to read input: %v", err)
			continue
		}
		_, err = conn.Write([]byte(line))
		if err != nil {
			log.Fatalf("Failed to write to UDP: %v", err)
			continue
		}
	}
}