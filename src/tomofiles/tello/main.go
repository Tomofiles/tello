package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"net"
	"net/http"
	"os"
	"strings"

	"github.com/gorilla/websocket"
	"gobot.io/x/gobot"
	"gobot.io/x/gobot/platforms/dji/tello"
)

var connWSs []*websocket.Conn

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
}

func main() {
	stdin := bufio.NewScanner(os.Stdin)

	drone := tello.NewDriver("8889")

	workDrone := func() {
		go staticServer()
		go websocketServer()
		go udpServer()

		go func() {
			drone.SendCommand("command")
			for {
				fmt.Print("command? > ")
				if !stdin.Scan() {
					break
				}
				command := stdin.Text()
				err := drone.SendCommand(command)
				if err != nil {
					fmt.Println(err.Error())
				}
			}
		}()
	}

	robotDrone := gobot.NewRobot("tello",
		[]gobot.Connection{},
		[]gobot.Device{drone},
		workDrone,
	)

	robotDrone.Start()
}

func udpServer() {
	connUDP, err := net.ListenPacket("udp", ":8890")
	if err != nil {
		fmt.Println(err)
	}
	defer connUDP.Close()

	var buf [1500]byte
	for {
		n, _, err := connUDP.ReadFrom(buf[:])
		if err != nil {
			fmt.Println(err)
			break
		}
		for _, connWS := range connWSs {
			telemetry := parseTelemetry(string(buf[:n]))
			connWS.WriteMessage(websocket.TextMessage, telemetry)
		}
	}
}

func staticServer() {
	fs := http.FileServer(http.Dir("static"))
	http.Handle("/", fs)
	http.ListenAndServe(":8080", nil)
}

func websocketServer() {
	http.HandleFunc("/telemetry", func(w http.ResponseWriter, r *http.Request) {
		connWS, _ := upgrader.Upgrade(w, r, nil)
		connWSs = append(connWSs, connWS)
	})
}

func parseTelemetry(telemetry string) []byte {
	items := strings.Split(telemetry, ";")
	itemsMap := make(map[string]string)
	for _, item := range items {
		pair := strings.Split(item, ":")
		if pair[0] == "\r\n" {
			break
		}
		itemsMap[pair[0]] = pair[1]
	}
	jsonTelemetry, _ := json.Marshal(itemsMap)
	return jsonTelemetry
}
