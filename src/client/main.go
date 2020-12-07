package main

import (
	"bufio"
	"flag"
	"fmt"
	"log"
	"math/rand"
	"os"
	"time"

	"golang.org/x/net/websocket"
)

// source
// https://medium.com/@johnshenk77/create-a-simple-chat-application-in-go-using-websocket-d2cb387db836

type Message struct {
	From string `json:"from"`
	Text string `json:"text"`
}

var (
	port = flag.String("port", "9000", "port used for ws connection")
)

// connect function returning a pointer to websocket.Conn
// this is where you would add TLS on the client to keep eavesdroppers
// out of the chat

func connect(ip string) (*websocket.Conn, error) {
	return websocket.Dial(fmt.Sprintf("ws://localhost:%s", *port), "", ip)
}

// if we are running it locally we have to differentiate the clients and cant use localhost as the 3rd parameter
// (the orgin) to websocket.Dial() since every client will be localhost. mockedIP() function creates a faux IP as a string

func mockedIP() string {
	var arr [4]int
	for i := 0; i < 4; i++ {
		rand.Seed(time.Now().UnixNano())
		arr[i] = rand.Intn(256)
	}
	return fmt.Sprintf("http://%d.%d.%d.%d", arr[0], arr[1], arr[2], arr[3])
}

// put the rest of client login within main
func main() {
	flag.Parse()


	ip := mockedIP()
	// connect
	ws, err := connect(ip)
	if err != nil {
		log.Fatal(err)
	}
	defer ws.Close()

	// recieve
	var m Message
	go func() {
		for {
			err := websocket.JSON.Receive(ws, &m)
			if err != nil {
				fmt.Println("Error receiving message: ", err.Error())
				break
			}
			fmt.Println("From: ", m.From , " : ", m.Text)
		}
	}()

	// send
	scanner := bufio.NewScanner(os.Stdin)
	for scanner.Scan() {
		text := scanner.Text()
		if text == "" {
			continue
		}
		m := Message{
			From: ip,
			Text: text,
		}
		err = websocket.JSON.Send(ws, m)
		if err != nil {
			fmt.Println("Error sending message: ", err.Error())
			break
		}
	}
}
