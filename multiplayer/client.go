package main

import (
	"log"
	"fmt"
	"net"
	"io"
	"os"
	"encoding/json"
	"image/color"
	ebiten "github.com/hajimehoshi/ebiten/v2"
	ebitenutil "github.com/hajimehoshi/ebiten/v2/ebitenutil"
	inpututil "github.com/hajimehoshi/ebiten/v2/inpututil"
	"multiplayer/proto"
	)

const (
	SCREEN_HEIGHT				= 600
	SCREEN_WIDTH				= 600
	INFO_WIDTH					= 200
	CELL_HEIGHT					= SCREEN_HEIGHT / proto.MAP_HEIGHT
	CELL_WIDTH					= SCREEN_WIDTH / proto.MAP_WIDTH
	)

const (
	SERVER_ADDR = "localhost"
	SERVER_PORT = "1234"
	)
	
type Game struct{
	MainImage *ebiten.Image
	
	State proto.State
	StateUpdate map[string]string
	
	Conn net.Conn
	Enc *json.Encoder
	Dec *json.Decoder
}

func (g *Game) Update() error {
	// making update from client input
    if ebiten.IsKeyPressed(ebiten.KeyW) {
    	duration := inpututil.KeyPressDuration(ebiten.KeyW)
    	if duration % 5 == 0 {
			g.StateUpdate["movement"] = "up"
		}
	}
	if ebiten.IsKeyPressed(ebiten.KeyD) {
		g.StateUpdate["movement"] = "right"
	}
	if ebiten.IsKeyPressed(ebiten.KeyS) {
		g.StateUpdate["movement"] = "down"
	}
	if ebiten.IsKeyPressed(ebiten.KeyA) {
		g.StateUpdate["movement"] = "left"
	}
	
	// sending update to server
	if len(g.StateUpdate) != 0 {
		g.Enc.Encode(proto.ClientMessage{g.StateUpdate})
		log.Println("Sent update:", g.StateUpdate)
		g.StateUpdate = make(map[string]string)
	}

    return nil
}

func (g *Game) ReadFromServer() {
	for {
		// checking for state update from server
		var mes proto.ServerMessage
		err := g.Dec.Decode(&mes)
		if err != nil && err != io.EOF {
			log.Fatal(err)
		} else if mes.State != nil {
			// updating state
			g.State = mes.State
			log.Println("Got update:", mes)
		} else {
			log.Println("Got State = nil")
			fmt.Println("Server disconnected")
			os.Exit(1)
		}
	}
}

func (g *Game) Draw(screen *ebiten.Image) {
    g.MainImage.Fill(color.White)
	
	for _, player := range g.State {
		for i := 0; i < CELL_WIDTH; i++ {
			for j := 0; j < CELL_HEIGHT; j++ {
				g.MainImage.Set(player.Coords.X * CELL_WIDTH + i, player.Coords.Y * CELL_HEIGHT + j, player.Color)
			}
		}
	}
	
	op := &ebiten.DrawImageOptions{}
	op.GeoM.Translate(INFO_WIDTH, 0)
	screen.DrawImage(g.MainImage, op)
	
	var debug_string string
	for IP, player := range g.State {
		debug_string += fmt.Sprintf("%s (%s): Health = %d\n", player.Name, IP, player.Health)
	}
	ebitenutil.DebugPrint(screen, debug_string)
}

func (g *Game) Layout(outsideWidth, outsideHeight int) (int, int) {
    return SCREEN_WIDTH + INFO_WIDTH, SCREEN_HEIGHT
}

func main() {
	log.Println("Connecting to remote server...")
	conn, err := net.Dial("tcp", SERVER_ADDR + ":" + SERVER_PORT)
	if err != nil {
		log.Fatal(err)
	}
	defer conn.Close()
	log.Println("Connected!")
	
	var Name string
	fmt.Print("Enter ypur name: ")
	fmt.Scan(&Name)

    game := &Game{}
    game.MainImage = ebiten.NewImage(SCREEN_WIDTH, SCREEN_HEIGHT)
    game.StateUpdate = make(map[string]string)
    game.StateUpdate["name"] = Name
    game.Conn = conn
    game.Dec = json.NewDecoder(conn)	// for reading
	game.Enc = json.NewEncoder(conn)	// for writing
	
	go game.ReadFromServer()				// reading from server
	
    ebiten.SetWindowSize(SCREEN_WIDTH + INFO_WIDTH, SCREEN_HEIGHT)
    ebiten.SetWindowTitle("Squares")
    if err := ebiten.RunGame(game); err != nil {
        log.Fatal(err)
    }
}
	