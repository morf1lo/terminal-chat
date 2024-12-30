package main

import (
	"bufio"
	"encoding/json"
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

	SYSTEM_CHAT_NAME_COLOR = color.BgHiBlue
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
		color.BgMagenta,
		color.BgHiBlack,
		color.BgHiRed,
		color.BgHiCyan,
		color.BgHiGreen,
		color.BgHiMagenta,
	}
)

type Connection struct {
	roomID int
	conn   net.Conn
}

type Message struct {
	Ok              bool
	Author          string
	Text            string
	AuthorNameColor color.Attribute
}

func main() {
	server, err := net.Listen(SERVER_TYPE, SERVER_HOST + ":" + SERVER_PORT)
	if err != nil {
		log.Fatal(err)
	}
	defer server.Close()

	log.Println("Server started")

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
		msgJSON, err := json.Marshal(&Message{Ok: false, Author: "🔧 System", Text: "Invalid room ID", AuthorNameColor: SYSTEM_CHAT_NAME_COLOR})
		if err != nil {
			log.Fatal(err)
		}
		broadcastPrivateMessage(append(msgJSON, '\n'), conn)
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
			return
		}

		msg = strings.TrimSpace(msg)
		log.Printf("received: %s\n", msg)
		msgJSON, err := json.Marshal(&Message{Ok: true, Author: name, Text: msg, AuthorNameColor: clientNameColor})
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
				log.Printf("error writing to connection: %s", err)
			}
			return
		}
	}
}

func broadcastPrivateMessage(msg []byte, conn net.Conn) {
	connMutex.Lock()
	defer connMutex.Unlock()

	if _, err := conn.Write(msg); err != nil {
		log.Printf("error private writing to connection: %s", err.Error())
	}
}

func removeConnection(conn net.Conn) {
	connMutex.Lock()
	defer connMutex.Unlock()

	for i, c := range connections {
		if c.conn == conn {
			connections = append(connections[:i], connections[i+1:]...)
			return
		}
	}
}
