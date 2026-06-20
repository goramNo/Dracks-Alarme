package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"sync"

	"github.com/gorilla/websocket"
)

type ClientInfo struct {
	Hostname string `json:"hostname"`
	User     string `json:"user"`
}

type Command struct {
	Type string `json:"type"`
	Data string `json:"data"`
}

type Client struct {
	Hostname string
	User     string
	Conn     *websocket.Conn
}

var (
	clients = make(map[*websocket.Conn]Client)
	mutex   sync.Mutex
)

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

func printClients() {

	fmt.Println("\n=== Clients connectés ===")

	if len(clients) == 0 {
		fmt.Println("Aucun client connecté")
		return
	}

	for _, c := range clients {
		fmt.Printf("- %s (%s)\n", c.Hostname, c.User)
	}
}

func handleWS(w http.ResponseWriter, r *http.Request) {

	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Println(err)
		return
	}

	fmt.Printf("\nNouveau client connecté : %s\n", r.RemoteAddr)

	defer func() {

		mutex.Lock()

		if client, ok := clients[conn]; ok {

			fmt.Printf(
				"\nClient déconnecté : %s (%s)\n",
				client.Hostname,
				client.User,
			)

			delete(clients, conn)
		}

		mutex.Unlock()

		printClients()

		conn.Close()
	}()

	for {

		_, message, err := conn.ReadMessage()

		if err != nil {
			break
		}

		var info ClientInfo

		err = json.Unmarshal(message, &info)

		if err != nil {
			continue
		}

		mutex.Lock()

		clients[conn] = Client{
			Hostname: info.Hostname,
			User:     info.User,
			Conn:     conn,
		}

		mutex.Unlock()

		fmt.Println("\n=== Client enregistré ===")
		fmt.Println("Hostname :", info.Hostname)
		fmt.Println("Utilisateur :", info.User)

		printClients()
	}
}

func sendNotificationAll(message string) {

	mutex.Lock()
	defer mutex.Unlock()

	cmd := Command{
		Type: "notification",
		Data: message,
	}

	data, _ := json.Marshal(cmd)

	for conn := range clients {

		err := conn.WriteMessage(
			websocket.TextMessage,
			data,
		)

		if err != nil {
			log.Println(err)
		}
	}

	fmt.Println("Notification envoyée :", message)
}

func notifyHandler(
	w http.ResponseWriter,
	r *http.Request,
) {

	message := r.URL.Query().Get("msg")

	if message == "" {
		message = "Bonjour depuis Dracks Alert"
	}

	sendNotificationAll(message)

	fmt.Fprintf(
		w,
		"Notification envoyée : %s",
		message,
	)
}

func clientsHandler(
	w http.ResponseWriter,
	r *http.Request,
) {

	mutex.Lock()
	defer mutex.Unlock()

	type ClientView struct {
		Hostname string `json:"hostname"`
		User     string `json:"user"`
	}

	var list []ClientView

	for _, c := range clients {

		list = append(list, ClientView{
			Hostname: c.Hostname,
			User:     c.User,
		})
	}

	w.Header().Set(
		"Content-Type",
		"application/json",
	)

	json.NewEncoder(w).Encode(list)
}

func main() {

	http.HandleFunc("/ws", handleWS)
	http.HandleFunc("/notify", notifyHandler)
	http.HandleFunc("/clients", clientsHandler)

	// Interface Web
	http.Handle(
		"/",
		http.FileServer(
			http.Dir("../web"),
		),
	)

	fmt.Println("===================================")
	fmt.Println("      DRACKS ALERT SERVER")
	fmt.Println("===================================")
	fmt.Println("Interface Web : http://localhost:8080")
	fmt.Println("WebSocket     : /ws")
	fmt.Println("Notify        : /notify")
	fmt.Println("Clients       : /clients")
	fmt.Println("En attente de connexions...")

	log.Fatal(
		http.ListenAndServe(
			":8080",
			nil,
		),
	)
}