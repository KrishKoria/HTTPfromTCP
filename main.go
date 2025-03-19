package main

import (
	"io"
	"log"
	"os"
	"strings"
)

func main() {
	
	file, err := os.Open("messages.txt")
	if err != nil {
		log.Fatal(err)
	}
	
	defer file.Close()
	
	currentLine := ""
	for {
		
		buffer := make([]byte, 8)
		
		bytesRead, err := file.Read(buffer)
		
		if err != nil {
			log.Fatal(err)
			if err == io.EOF {
				break
			}
		}
		
		data := string(buffer[:bytesRead])
        
        parts := strings.Split(data, "\n")
        
        for i := range parts {
            if i < len(parts)-1 {
                currentLine += parts[i]
                log.Printf("read: %s", currentLine)
                currentLine = ""
            } else if strings.HasSuffix(data, "\n") {
                currentLine += parts[i]	
                log.Printf("read: %s", currentLine)
                currentLine = ""
            } else {
                currentLine += parts[i]
            }
        }	
	}

	if currentLine != "" {
        log.Printf("read: %s", currentLine)
    }
}