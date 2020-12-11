package main

import (
	"flag"
	"fmt"
	"log"
	"net/http" 
	
	"golang.org/x/net/websocket"
)

// reference for how to use Go
// https://medium.com/@johnshenk77/create-a-simple-chat-application-in-go-using-websocket-d2cb387db836

// Message type
type Message struct {
	From string `json:"from"`
	Text string `json:"text"`
}

// holds clients through a map, and channels to send the message, client additions, and removals
type hub struct {
	clients				map[string]*websocket.Conn
	addClientChan		chan *websocket.Conn
	removeClientChan	chan *websocket.Conn
	broadcastChan		chan Message
}

// consistent port between server and client, flag variable
var (
	port = flag.String("port", "9000", "port used for ws connection")
)

func server(port string) error {

	// Routes sole path `/` to handler
	// every client that connects will call this handler
	// it calls hub.run then listens for messages from that client in a for loop

	h := newHub()
	mux := http.NewServeMux()
	mux.Handle("/", websocket.Handler(func(ws *websocket.Conn) {
		handler(ws, h)
	}))

	s := http.Server {
		Addr: ":" + port, 
		Handler: mux,
	}

	return s.ListenAndServe()
}

func handler(ws *websocket.Conn, h *hub) {
	go h.run()

	h.addClientChan <- ws

	for {
		var m Message
		err := websocket.JSON.Receive(ws, &m)
		if err != nil {
			h.broadcastChan <- Message{"server", err.Error()}
			h.removeClient(ws)
			return
		}
		h.broadcastChan <- m
	}
}

// constructor for hub object
func newHub() *hub {
	return &hub {
		clients: 			make(map[string]*websocket.Conn),
		addClientChan: 		make(chan *websocket.Conn),
		removeClientChan: 	make(chan *websocket.Conn),
		broadcastChan: 		make(chan Message),
	}
}

// create the hubs run() method which is called via goroutine in the handler
// listens to all of the hubs various channels via go's idiomatic for-select, 
// calling the appropritate method for each incoming channel

func (h *hub) run() {
	for {
		select {
		case conn := <- h.addClientChan: 
			h.addClient(conn)
		case conn := <- h.removeClientChan : 
			h.removeClient(conn)
		case m := <- h.broadcastChan:
			h.broadcastMessage(m)
		}
	}
}

// onto creating those three methods

func (h *hub) removeClient(conn *websocket.Conn) {
	delete(h.clients, conn.LocalAddr().String())
}

func (h *hub) addClient(conn *websocket.Conn) {
	h.clients[conn.RemoteAddr().String()] = conn
}

func (h *hub) broadcastMessage(m Message) {
	for _, conn := range h.clients {
		err := websocket.JSON.Send(conn, m)
		if err != nil {
			fmt.Println("Error broadcasting message: ", err)
			return
		}
	}
}

func main() {
	flag.Parse()
	log.Fatal(server(*port))
}
