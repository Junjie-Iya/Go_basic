package main

import (
	"fmt"
	"net"
	"sync"
	"bufio"
	"strings"
	"chat_app/message"
)

type ClientInfo struct {
	id string
}

type CommandHandler func(conn net.Conn, args string)

var command_handlers = map[string] CommandHandler {
	"/users": handleUsers,
	"/nick": handleNickname,
	"/msg": handleMsg,
}

var (
	history_message []message.Message
	max_save_nums int  = 10
)

func addHistoryMessage(msg message.Message) {
	// mu.Lock()
	// defer mu.Unlock()
	history_message = append(history_message, msg)
	if len(history_message) > max_save_nums {
		history_message = history_message[1:]
	}
}

func handleUsers(conn net.Conn, args string) {
	mu.Lock()
	defer mu.Unlock()
	list := "Current users:\n"
	for _, v := range clients {
		list += "- " + v.id + "\n"
	}
	json_msg := message.MakeSystemMessage(list)
	json_data_array := message.MessageToJson(json_msg)
	conn.Write(json_data_array)
}

func handleNickname(conn net.Conn, args string) {
	mu.Lock()
	defer mu.Unlock()
	
	new_nickname := strings.TrimSpace(strings.Fields(args)[0])
	self_msg := "Changed nickname to: " + new_nickname
	broadcast_msg := clients[conn].id + " changed nickname to: " + new_nickname
	clients[conn] = ClientInfo{id: new_nickname}

	json_msg := message.MakeSystemMessage(self_msg + "\n")
	json_data_array := message.MessageToJson(json_msg)
    conn.Write(json_data_array)
	for c := range clients {
		if conn != c {
			json_msg = message.MakeSystemMessage(broadcast_msg + "\n")
			json_data_array = message.MessageToJson(json_msg)
			c.Write(json_data_array)
		}
	}
}

func handleMsg(conn net.Conn, args string) {
	mu.Lock()
	defer mu.Unlock()
	usr := strings.Fields(args)[0]
	msg := strings.TrimPrefix(args, usr)
	found := false
	for c, v := range clients {
		if v.id == usr {
			json_msg := message.MakePrivateUserMessage(msg, clients[conn].id, v.id)
			json_data_array := message.MessageToJson(json_msg)
			c.Write([]byte(json_data_array))
			found = true
			break
		}
	}
	if !found {
		json_msg := message.MakeSystemMessage("No Such User, Type /users To check\n")
		json_data_array := message.MessageToJson(json_msg)
		conn.Write([]byte(json_data_array))
	}
}

var (
	clients = make(map[net.Conn] ClientInfo)
	mu sync.Mutex
)

func handleConnection(conn net.Conn) {
    defer conn.Close()
	json_msg := message.MakeSystemMessage("Please Enter Nickname: \n")
	json_data_array := message.MessageToJson(json_msg)
    conn.Write(json_data_array)
	reader := bufio.NewReader(conn)

	nickname, err := reader.ReadString('\n')
	if err != nil {
		fmt.Println("Read nickname error:", err)
		return
	}
	nickname = strings.TrimSpace(nickname)
    

	mu.Lock()
	clients[conn] = ClientInfo{id: nickname}
	// welcome msg
	for c := range clients {
		json_msg = message.MakeSystemMessage("Welcome [" + nickname + "] to Chat Server!\n")
		json_data_array = message.MessageToJson(json_msg)
		c.Write(json_data_array)
	}
	// add history broadcast messages
	for _, m := range history_message {
		json_data_array = message.MessageToJson(m)
		conn.Write(json_data_array)
	}
	if len(history_message) > 1 {
		json_msg = message.MakeSystemMessage("--- History Message Till Here ---\n")
		json_data_array = message.MessageToJson(json_msg)
		conn.Write(json_data_array)
	}
	mu.Unlock()

	// main loop 
	for {
		msg, err := reader.ReadString('\n')
		if (len(strings.TrimSpace(msg)) == 0) {
			continue
		}
		if err != nil {
			// fmt.Println("Read error:", err)
			mu.Lock()
			for c := range clients {
				if conn != c {
					json_msg = message.MakeSystemMessage(fmt.Sprintf("---%s has left the chat --- \n", clients[conn].id))
					json_data_array = message.MessageToJson(json_msg)
					c.Write(json_data_array)
				}
			}
			delete(clients, conn)
			mu.Unlock()
			return
		}

		// deal with msg starts with "/"
		if strings.HasPrefix(msg, "/") {
			cmd := strings.Fields(msg)[0]
			args := strings.TrimPrefix(msg, cmd)
			if handler, ok := command_handlers[cmd]; ok {
				handler(conn, args)
			} else {
				json_msg = message.MakeSystemMessage("Unknown Command\n")
				json_data_array = message.MessageToJson(json_msg)
				conn.Write(json_data_array)
			}
			continue
		}

		mu.Lock()
		for c := range clients {
			if conn != c {
				json_msg = message.MakeGeneralUserMessage(msg, clients[conn].id)
				json_data_array = message.MessageToJson(json_msg)
				c.Write(json_data_array)
				addHistoryMessage(json_msg)
			}
		}
		mu.Unlock()
	}
}

func main() {
	listener, err := net.Listen("tcp", ":8080")
	if err != nil {
		fmt.Println("Error: ", err)
		return
	}
	defer listener.Close()

	fmt.Println("Listening on port 8080")
	for {
		conn, err := listener.Accept()
		if err != nil {
			fmt.Println("Connection error: ", err)
			continue
		}
		fmt.Println("Client Connected")

		go handleConnection(conn)
	}
}