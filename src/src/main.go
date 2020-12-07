package main

import (
	"log"
	"net/http"

	"github.com/gorilla/websocket"
)

// source: 
// https://scotch.io/bar-talk/build-a-realtime-chat-server-with-go-and-websockets
// This is my first go project, I will actually leave up the "hello" and "greetings" folder so that you can see me going through the 
// "installing go" tutorial. I plan on "beefing up" the code from scotch.io, but the core of the code is from there and I do not take credit for it

var clients = make(map[*websocket.Conn]bool)
var broadcast = make(chan Message)

// upgraders are objects with methods for taking normal HTTP connections and upgrading them to websockets
var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

type Message struct {
	Email string `json:"email"`
	Username string `json:"username"`
	Message string `json:"message"`
}

func handleConnections(w http.ResponseWriter, r *http.Request) {
	// upgrade the initial get request to a websocket
	// if there is an error we log it but do not exit
	ws, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Fatal(err)
	}

	// close the connection when the function returns
	defer ws.Close()

	// register the new client by adding it to the clients map
	clients[ws] = true

	// infinite loop that waits for a new message to be written to the websocket, 
	// unerializes it from json to a message, then puts it into the broadcast channel
	// handleMessages then takes it and sends it to everyone else connected
	// if there is an error then we say the client has disconnected 
	// and remove them from the clients map and we dont try to send messages to that client or read theirs

	// HTTP route handler functions are goroutines. This lets the http server to handle multiple incoming connections
	// without having to wait for another to finish

	for {
		var msg Message

		// read message from json and map it to a message
		err := ws.ReadJSON(&msg)
		if err != nil {
			log.Printf("error: %v", err)
			delete(clients, ws)
			break
		}

		// send the newly recieved message to broadcast
		broadcast <- msg
	}
}

func handleMessages() {
	// loop that continuously reads from the broadcast channel and then relays the messageto 
	// all clients over their respective websocket connection. if thre is an error then close the 
	// connection and remove it from the clients map
	
	// grab the next message from the broadcast channel
	msg := <- broadcast

	// send it out to every client that is currently connected
	for client := range clients {
		err := client.WriteJSON(msg)
		if err != nil {
			log.Printf("error: %v", err)
			client.Close()
			delete(clients, client)
		}
	}
}

func main() {
	// create a staic fileserver and tie that to the "/" route so that users can view index.html and assets
	fs := http.FileServer(http.Dir("../public"))
	http.Handle("/", fs)

	// you start a websocket when you go to send a message
	// that message gets sent everywhere through handlemessges


	// make a /ws route to handle any requests for instantiating a websocket
	http.HandleFunc("/ws", handleConnections)

	// start a goroutine called handleMessages. its is a concurrent process that runs simultaneously. it will take messages from the 
	// broadcast channel and pass them to clients over their respective websocket connections
	go handleMessages()

	log.Println("Http server started on :8000")
	err := http.ListenAndServe(":8000", nil)
	if err != nil {
		log.Fatal("ListenAndServe: ", err)
	}
}

