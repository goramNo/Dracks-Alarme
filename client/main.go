package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net"
	"net/url"
	"os"
	"os/user"
	"time"

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

func getHostname() string {
	name, err := os.Hostname()
	if err != nil {
		return "UNKNOWN"
	}
	return name
}

func getUsername() string {
	u, err := user.Current()
	if err != nil {
		return "UNKNOWN"
	}
	return u.Username
}

func connect() {

	u := url.URL{
		Scheme: "ws",
		Host:   "187.124.39.231:8080",
		Path:   "/ws",
	}

	for {

		conn, _, err := websocket.DefaultDialer.Dial(
			u.String(),
			nil,
		)

		if err != nil {
			fmt.Println("Connexion impossible, nouvelle tentative dans 5 secondes...")
			time.Sleep(5 * time.Second)
			continue
		}

		fmt.Println("Connecté au serveur")

		info := ClientInfo{
			Hostname: getHostname(),
			User:     getUsername(),
		}

		data, _ := json.Marshal(info)

		err = conn.WriteMessage(
			websocket.TextMessage,
			data,
		)

		if err != nil {
			conn.Close()
			time.Sleep(5 * time.Second)
			continue
		}

		for {

			_, message, err := conn.ReadMessage()

			if err != nil {
				fmt.Println("Connexion perdue")
				conn.Close()
				break
			}

			var cmd Command

			err = json.Unmarshal(message, &cmd)

			if err != nil {
				continue
			}

			switch cmd.Type {

			case "notification":

				fmt.Println()
				fmt.Println("==============")
				fmt.Println("NOTIFICATION")
				fmt.Println(cmd.Data)
				fmt.Println("==============")
				fmt.Println()

			case "ping":

				conn.WriteMessage(
					websocket.TextMessage,
					[]byte(`{"type":"pong"}`),
				)
			}
		}

		time.Sleep(5 * time.Second)
	}
}

func main() {

	fmt.Println("Client Dracks Alert démarré")
	fmt.Println("Connexion au serveur...")

	_, err := net.LookupHost("187.124.39.231")

	if err != nil {
		log.Println(err)
	}

	connect()
}