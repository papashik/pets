package main

import (
	"log"
	"encoding/json"
	"net"
	"multiplayer/proto"
	"math/rand"
	"image/color"
	//"io"
	"time"
	)

const (
	PORT			= "1234"
	)
	
var State proto.State = make(proto.State)
var StateChanged bool = false
var Connections map[string]net.Conn = make(map[string]net.Conn)

func StateChange() {
	StateChanged = true
}

func GetRandomEmptyCell() proto.Point {
	for {
		found := true
		point := proto.Point{rand.Intn(proto.MAP_WIDTH), rand.Intn(proto.MAP_HEIGHT)}
		for _, player := range State {
			if player.Coords == point {
				found = false
				break
			}
		}
		if found { return point }
	}
}

func NewPlayer() proto.Player {
	var player proto.Player
	player.Color = color.RGBA{uint8(rand.Int()), uint8(rand.Int()), uint8(rand.Int()), 255}
	player.Health = proto.START_HEALTH
	player.Coords = GetRandomEmptyCell()
	return player
}

func handle(conn net.Conn) {
	defer conn.Close()
	IP := conn.RemoteAddr().String()
	log.Println("New connection from", IP)
		
	Connections[IP] = conn
	State[IP] = NewPlayer()
	StateChanged = true
	
	dec := json.NewDecoder(conn)
	for {
		var mes proto.ClientMessage
		err := dec.Decode(&mes)
		if err != nil {
			log.Println(IP, "error:", err)
			log.Println("Connection with", IP, "closed")
			delete(Connections, IP)
			delete(State, IP)
			StateChanged = true
			return				// connection closed
		}

		if mes.Update != nil {
			log.Println("New update from", IP, "->", mes.Update)
			player := State[IP]
			switch mes.Update["movement"] {
				case "up":
					player.Coords.Y--
				case "right":
					player.Coords.X++
				case "down":
					player.Coords.Y++
				case "left":
					player.Coords.X--
			}
			if name, ok := mes.Update["name"]; ok { player.Name = name }
			State[IP] = player
			StateChanged = true
		}
	}
}

func listen(ln net.Listener) {
	for {
		conn, err := ln.Accept()
		if err != nil {
			log.Fatal(err)
		}
		go handle(conn)
	}
}
	

func main() {
	ln, err := net.Listen("tcp", ":" + PORT)
	if err != nil {
		log.Fatal(err)
	}
	log.Println("Listener started at", PORT)
	
	go listen(ln)
	
	for {
		time.Sleep(10 * time.Millisecond)
		if !StateChanged { continue }
		/*
		byte_state, err := json.Marshal(proto.ServerMessage{State})
		if err != nil {
			log.Fatal(err)
		}
		
		for _, conn := range Connections {
			conn.Write(byte_state)
		}
		log.Println("Sent", string(byte_state))*/
		for _, conn := range Connections {
			enc := json.NewEncoder(conn)
			err := enc.Encode(proto.ServerMessage{State})
			if err != nil {
				log.Fatal(err)
			}
		}
		
		
		StateChanged = false
	}
}