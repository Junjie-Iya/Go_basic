package main

import (
	"bufio"
	"fmt"
	"net"
	"os"
	"io"
	"chat_app/message"
	"encoding/json"
)

func main() {
	conn, err := net.Dial("tcp", "localhost:8080")
	if err != nil {
		fmt.Println("Connection error:", err)
		return
	}
	defer conn.Close()

	reader := bufio.NewReader(conn)
	decoder := json.NewDecoder(reader)

	// 
	go func() {
		for {
			var msg message.Message
    		err := decoder.Decode(&msg)
			if err != nil {
				if err == io.EOF {
					break
				}
				fmt.Println("Decode error:", err)
				return
			}
			fmt.Print(msg.MakeString())
		}
	}()

	// send
	scanner := bufio.NewScanner(os.Stdin)
	for scanner.Scan() {
		text := scanner.Text()
		conn.Write([]byte(text + "\n"))
	}
}