package proto
import "image/color"

const (
	MAP_HEIGHT 			= 30
	MAP_WIDTH 			= 30
	
	START_HEALTH		= 100
	)

type Point struct {
	X, Y int
}

type Player struct {
	Name string
	Color color.RGBA
	Health int
	Coords Point
	Direction string
}

type State map[string]Player

type ClientMessage struct {
	Update map[string]string
	// Update may contain:
	// movement: up, down, left, right
	// rotation: up, down, left, right
}

type ServerMessage struct {
	State State
}
