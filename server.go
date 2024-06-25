package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"log"
	"math/rand"
	"net"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/fatih/color"
)

const (
	SERVER_HOST = "localhost"
	SERVER_PORT = "8090"
	SERVER_TYPE = "tcp"
)

var (
	connections = make([]*Connection, 0)
	connMutex   sync.Mutex
	
	allColors = []color.Attribute{
		color.BgBlack,
		color.BgRed,
		color.BgBlue,
		color.BgCyan,
		color.BgGreen,
		color.BgYellow,
		color.BgMagenta,
	}
)

type Connection struct {
	roomID int
	conn   net.Conn
}

type Message struct {
	Author          string
	Text            string
	AuthorNameColor color.Attribute
}

func main() {
	fmt.Println("Server running")
	server, err := net.Listen(SERVER_TYPE, SERVER_HOST + ":" + SERVER_PORT)
	if err != nil {
		log.Fatal(err)
	}
	defer server.Close()

	fmt.Println("Server started")

	for {
		conn, err := server.Accept()
		if err != nil {
			log.Fatal(err)
		}

		go handleClient(conn)
	}
}

func handleClient(conn net.Conn) {
	defer func()  {
		conn.Close()
		removeConnection(conn)
	}()

	reader := bufio.NewReader(conn)

	name, err := reader.ReadString('\n')
	if err != nil {
		log.Fatal(err)
	}
	name = strings.TrimSpace(name)

	roomIDString, err := reader.ReadString('\n')
	if err != nil {
		log.Fatal(err)
	}

	roomIDString = strings.TrimSpace(roomIDString)
	roomID, err := strconv.Atoi(roomIDString)
	if err != nil {
		fmt.Printf("invalid room ID: %s\n", err.Error())
		return
	}

	s := rand.NewSource(time.Now().Unix())
	r := rand.New(s)
	clientNameColor := allColors[r.Intn(len(allColors))]

	connMutex.Lock()
	connections = append(connections, &Connection{roomID: roomID, conn: conn})
	connMutex.Unlock()

	for {
		msg, err := reader.ReadString('\n')
		if err != nil {
			if err.Error() == "EOF" {
				break
			} else {
				log.Fatal(err)
			}
		}
	
		msg = strings.TrimSpace(msg)
		fmt.Printf("received: %s\n", msg)
		msgJSON, err := json.Marshal(&Message{Author: name, Text: msg, AuthorNameColor: clientNameColor})
		if err != nil {
			log.Fatal(err)
		}
		broadcastMessage(append(msgJSON, '\n'), &Connection{roomID: roomID, conn: conn})
	}
}

func broadcastMessage(msg []byte, sender *Connection) {
	connMutex.Lock()
	defer connMutex.Unlock()

	for _, c := range connections {
		if c.roomID == sender.roomID {
			if _, err := c.conn.Write(msg); err != nil {
				log.Printf("Error writing to connection: %s", err)
			}
		}
	}
}

func removeConnection(conn net.Conn) {
	connMutex.Lock()
	defer connMutex.Unlock()

	for i, c := range connections {
		if c.conn == conn {
			connections = append(connections[:i], connections[i+1:]...)
			break
		}
	}
}
