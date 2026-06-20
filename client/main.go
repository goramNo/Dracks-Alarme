package main

import (
	"encoding/json"
	"log"
	"os"
	"os/user"

	"github.com/gen2brain/dlgs"
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

func main() {

	conn, _, err := websocket.DefaultDialer.Dial("ws://localhost:8080/ws", nil)
	if err != nil {
		log.Fatal(err)
	}

	defer conn.Close()

	hostname, _ := os.Hostname()
	currentUser, _ := user.Current()

	info := ClientInfo{
		Hostname: hostname,
		User:     currentUser.Username,
	}

	data, _ := json.Marshal(info)

	err = conn.WriteMessage(websocket.TextMessage, data)
	if err != nil {
		log.Fatal(err)
	}

	log.Println("Informations envoyées au serveur")

	for {

		_, message, err := conn.ReadMessage()

		if err != nil {
			log.Println("Déconnecté du serveur")
			break
		}

		var cmd Command

		err = json.Unmarshal(message, &cmd)
		if err != nil {
			continue
		}

		switch cmd.Type {

		case "notification":

			log.Printf("Notification reçue : %s\n", cmd.Data)

			dlgs.Info(
				"Dracks Alert",
				cmd.Data,
			)
		}
	}
}