package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"log"
	"net"
	"os"
	"os/signal"
	"syscall"

	"github.com/fatih/color"
)

const (
	SERVER_HOST = "localhost"
	SERVER_PORT = "8090"
	SERVER_TYPE = "tcp"
)

type Message struct {
	Author          string
	Text            string
	AuthorNameColor color.Attribute
}

func main() {
	conn, err := net.Dial(SERVER_TYPE, SERVER_HOST + ":" + SERVER_PORT)
	if err != nil {
		log.Fatal(err)
	}
	defer conn.Close()

	reader := bufio.NewReader(os.Stdin)

	fmt.Print("Enter your name: ")
	name, err := reader.ReadString('\n')
	if err != nil {
		log.Fatalf("invalid name: %s\n", err.Error())
	}

	if _, err := conn.Write([]byte(name)); err != nil {
		log.Fatal(err)
	}

	fmt.Print("Enter room ID: ")
	roomID, err := reader.ReadString('\n')
	if err != nil {
		log.Fatalf("invalid room ID: %s\n", err.Error())
	}

	if _, err := conn.Write([]byte(roomID)); err != nil {
		log.Fatal(err)
	}

	fmt.Println("To exit: Ctrl + C")

	go receiveMessages(conn)

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

	go func ()  {
		for {
			data, err := reader.ReadString('\n')
			if err != nil {
				log.Fatal(err)
			}

			fmt.Printf("\033[1A\033[K")
	
			if _, err := conn.Write([]byte(data)); err != nil {
				log.Fatal(err)
			}
		}
	}()

	<-quit

	if err := conn.Close(); err != nil {
		log.Fatal(err)
	}
}

func receiveMessages(conn net.Conn) {
	reader := bufio.NewReader(conn)
	for {
		msgJSON, err := reader.ReadString('\n')
		if err != nil {
			log.Fatal(err)
		}
		var msg Message
		if err := json.Unmarshal([]byte(msgJSON), &msg); err != nil {
			log.Fatal(err)
		}

		out := color.New(msg.AuthorNameColor, color.FgWhite).Add(color.Bold)
		out.Printf("%s:", msg.Author)
		fmt.Printf(" %s\n", msg.Text)
	}
}
