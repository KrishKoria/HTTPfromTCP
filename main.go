package main

import (
	"io"
	"log"
	"os"
)

func main() {
	file, err := os.Open("messages.txt")
	if err != nil {
		log.Fatal(err)
	}
	defer file.Close()
	buffer := make([]byte, 8)
	for {
		bytesRead, err := file.Read(buffer)
		if err == io.EOF {
			break
		}
		if err != nil {
			log.Fatal(err)
		}
		log.Printf("read: %s\n", buffer[:bytesRead])	
	}
}