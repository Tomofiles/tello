package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"net"
	"net/http"
	"os"
	"os/signal"
	"strings"

	"github.com/gorilla/websocket"
)

var connWSs []*websocket.Conn

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
}

func main() {
	fmt.Println("hello.")

	stdin := bufio.NewScanner(os.Stdin)

	go udpClient(stdin)
	go staticServer()
	go websocketServer()
	go udpServer()

	quit := make(chan os.Signal)
	signal.Notify(quit, os.Interrupt)
	for {
		sig := <-quit
		if sig == os.Interrupt {
			fmt.Println("bye.")
			return
		}
	}
}

func udpClient(stdin *bufio.Scanner) {
	reqAddr, err := net.ResolveUDPAddr("udp", "192.168.10.1:8889")
	if err != nil {
		fmt.Println(err)
		return
	}
	respPort, err := net.ResolveUDPAddr("udp", ":8889")
	if err != nil {
		fmt.Println(err)
		return
	}
	cmdConn, err := net.DialUDP("udp", respPort, reqAddr)
	if err != nil {
		fmt.Println(err)
		return
	}
	defer cmdConn.Close()

	sendCommand(cmdConn, "command")
	for {
		fmt.Print("command? > ")
		if !stdin.Scan() {
			break
		}
		command := stdin.Text()
		response, err := sendCommand(cmdConn, command)
		if err != nil {
			fmt.Println(err.Error())
		}
		fmt.Printf("response < %s\n", response)
	}
	return
}

func sendCommand(cmdConn *net.UDPConn, command string) (string, error) {
	n, err := cmdConn.Write([]byte(command))
	if err != nil {
		return "", err
	}

	recvBuf := make([]byte, 1024)

	n, err = cmdConn.Read(recvBuf)
	if err != nil {
		return "", err
	}

	return string(recvBuf[:n]), nil
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
