package main

import (
	"log"
	"image/color"
	"math"
	"github.com/hajimehoshi/ebiten"
	)

const (
	WINDOW_WIDTH				= 600
	WINDOW_HEIGHT				= 600

	FRAME_WIDTH					= 600
	FRAME_HEIGHT				= 600

	PROJECTION_PLANE_WIDTH		= 20
	PROJECTION_PLANE_HEIGHT		= 20
	CAMERA_Z 					= -20

	EPSILON 					= 0.001
	MAX_INF						= 10000000
	MIN_INF						= -MAX_INF

	POINT_LIGHT_AMPLIFICATION	= 100000

	)
var (
	BACKGROUND_COLOR 	= RGB(255, 255, 255)
	VECTOR_INF			= Vector{MAX_INF, MAX_INF, MAX_INF}
	)

type Point struct {
	X, Y, Z float64
}

func (p Point) DistanceTo(other Point) float64 {
	return math.Sqrt(math.Pow(p.X - other.X, 2) + math.Pow(p.Y - other.Y, 2) + math.Pow(p.Z - other.Z, 2))
}

func (p Point) Add(v Vector) Point {
	return Point{p.X + v.X, p.Y + v.Y, p.Z + v.Z}
}

func Equal(num1, num2 float64) bool {
	return math.Abs(num1 - num2) <= EPSILON
}

type Vector struct {
	X, Y, Z float64
}

func (v Vector) Add(other Vector) Vector {
	return Vector{v.X + other.X, v.Y + other.Y, v.Z + other.Z}
}

func (v Vector) Mul(num float64) Vector {
	return Vector{v.X * num, v.Y * num, v.Z * num}
}

func (v Vector) ModuleSquare() float64 {
	return v.X * v.X + v.Y * v.Y + v.Z * v.Z
}

func (v Vector) Scalar(other Vector) float64 {
	return v.X * other.X + v.Y * other.Y + v.Z * other.Z
}

func VectorFromPoints(p1, p2 Point) Vector {
	return Vector{p2.X - p1.X, p2.Y - p1.Y, p2.Z - p1.Z}
}

type Object struct {
	Location	Point
	Intersect 	func (Point, Vector) (Vector, bool)
	Normal		func (Point) Vector
	PointIn 	func (Point) bool
	Color		color.RGBA
	Shine 		float64
}

func CreateSphere(center Point, radius float64, color color.RGBA, shine float64) Object {
	var obj = Object {Location: center, Color: color, Shine: shine}
	obj.Intersect = func (camera Point, v Vector) (Vector, bool) {
		/* O (center) - центр сферы, C (camera) - точка начала вектора луча V, P - точка пересечения вектора со сферой
		Параметризуем вектор CP как V*t с параметром t
		C + V * t = P
		CP = P - C = V * t
		|OP| = R
		|OC + CP| = R
		|OC + t*V| = R
		(OC + t*V)^2 = R^2
		OC^2 + t*2*V*OC + (t^2)*(V^2) = R^2
		(t^2) * |V|^2 + t * 2 * (V * OC) + (|OC|^2 - R^2) = 0

		a = |V|^2
		b = 2 * (V * OC)
		c = |OC|^2 - R^2
		D = b^2 - 4*a*c = 4 * (V * OC)^2 - 4 * (|V|^2) * (|OC|^2 - R^2)
		*/
		OC := VectorFromPoints(center, camera)
		Discriminant := 4 * math.Pow(v.Scalar(OC), 2) - 4 * v.ModuleSquare() * (OC.ModuleSquare() - math.Pow(radius, 2))
		if Discriminant < 0 {
			return Vector{}, false
		}
		// иначе Discriminant >= 0, берём минимальное t, большее единицы
		a := - 2 * v.Scalar(OC) - math.Sqrt(Discriminant)
		b := 2 * v.ModuleSquare()
		t := a / b

		if t < 1 {
			t += math.Sqrt(Discriminant) / v.ModuleSquare()
			if t < 1 {
				// Если t осталось меньше единицы, обе точки пересечения находятся за плоскостью проекции
				return Vector{}, false
			}
		}
		return v.Mul(t), true
	}
	obj.Normal = func (p Point) Vector {
		return VectorFromPoints(center, p)
	}
	obj.PointIn = func (p Point) bool {
		return center.DistanceTo(p) < radius
	}
	return obj
}
////////////////////////// COLOR //////////////////////////
func RGB(r, g, b uint8) color.RGBA {
	return color.RGBA{r, g, b, 0}
}
func ColorMul(c color.RGBA, num float64) color.RGBA {
	R := float64(c.R) * num
	if R > 255 { c.R = 255 } else { c.R = uint8(R) }
	G := float64(c.G) * num
	if G > 255 { c.G = 255 } else { c.G = uint8(G) }
	B := float64(c.B) * num
	if B > 255 { c.B = 255 } else { c.B = uint8(B) }
	return c
}
func ColorSum(x, y color.RGBA) color.RGBA {
	if x.R > 255 - y.R { x.R = 255 } else { x.R += y.R }
	if x.G > 255 - y.G { x.G = 255 } else { x.G += y.G }
	if x.B > 255 - y.B { x.B = 255 } else { x.B += y.B }
	return x
}

////////////////////////// LIGHT //////////////////////////

type Light interface {
	Power() float64
	Illumination(Object, Point, Vector) float64
}

type AmbientLight float64
func (l AmbientLight) Power() float64 {
	return float64(l)
}
func (l AmbientLight) Illumination(obj Object, p Point, view Vector) float64 {
	return 1
}

type PointLight struct {
	StandardPower float64
	Location Point
}
func (l PointLight) Power() float64 {
	return l.StandardPower
}
func (l PointLight) Illumination(obj Object, p Point, view Vector) float64 {
	return StandardIllumination(obj, p, view, VectorFromPoints(l.Location, p))
}

type DirectLight struct {
	StandardPower float64
	Direction Vector
}
func (l DirectLight) Power() float64 {
	return l.StandardPower
}
func (l DirectLight) Illumination(obj Object, p Point, view Vector) float64 {
	return StandardIllumination(obj, p, view, l.Direction)
}

func StandardIllumination(obj Object, p Point, view, light Vector) float64 {
	var result float64
	// 1. Диффузность
	/* N (normal) - нормаль к поверхности в точке P, L - расположение источника света
	PL * N = |PL| * |N| * cos(PL, N)
	cos(PL, N) = (PL * N) / sqrt(|PL|^2 * |N|^2)
	*/
	normal := obj.Normal(p)
	normal = normal.Mul(1 / math.Sqrt(normal.ModuleSquare()))
	scalar := -light.Scalar(normal)
	if scalar < 0 {
		scalar = 0
	}
	result += scalar / math.Sqrt(light.ModuleSquare() * normal.ModuleSquare())
	log.Print("before: ", result)
	// 2. Зеркальность
	/* Нужно найти угол между лучом отражения R и лучом взгляда V, зная луч падения L и нормаль N
	Нормализуем нормаль, тогда R = 2*N*(N*(-L)) + L */
	reflection := normal.Mul(2 * scalar).Add(light)
	scalar = -reflection.Scalar(view)
	if scalar < 0 {
		scalar = 0
	}
	result += math.Pow(scalar / math.Sqrt(reflection.ModuleSquare() * view.ModuleSquare()), obj.Shine)
	log.Println("after:", result)
	return result
}


type World struct {
	Objects []Object
	Lights []Light
	TotalLightPower float64
	Frame [FRAME_HEIGHT][FRAME_WIDTH]color.RGBA
	FrameTopLeftPoint Point
	FrameDownVector Vector
	FrameRightVector Vector
	CameraLocation Point
}

func (w *World) SetStandardSettings() {
	w.FrameTopLeftPoint = Point{-PROJECTION_PLANE_WIDTH / 2, PROJECTION_PLANE_HEIGHT / 2, 0}
	w.FrameDownVector = Vector{0, -PROJECTION_PLANE_HEIGHT, 0}
	w.FrameRightVector = Vector{PROJECTION_PLANE_WIDTH, 0, 0}
	w.CameraLocation = Point{0, 0, CAMERA_Z}
}

func (w *World) FramePixelLocation(i, j int) Point {
	down := w.FrameDownVector.Mul(float64(i) / float64(FRAME_HEIGHT))
	right := w.FrameRightVector.Mul(float64(j) / float64(FRAME_WIDTH))
	return w.FrameTopLeftPoint.Add(down).Add(right)
}

func (w *World) ComputeFrame() {
	for i := 0; i < FRAME_HEIGHT; i++ {
		for j := 0; j < FRAME_WIDTH; j++ {
			w.Frame[i][j] = w.ComputeFramePixelColor(i, j)
		}
	}
}

func (w *World) ComputeFramePixelColor(i, j int) color.RGBA {
	var (
		PixelLocation Point = w.FramePixelLocation(i, j)
		RayVector = VectorFromPoints(w.CameraLocation, PixelLocation)
		vector Vector
		exists, found bool
		min_vector Vector
		object Object
		)
	for _, obj := range w.Objects {
		vector, exists = obj.Intersect(w.CameraLocation, RayVector)
		if exists && (vector.ModuleSquare() < min_vector.ModuleSquare() || min_vector == (Vector{})) {
			min_vector = vector
			object = obj
			found = true
		}
	}
	if !found {
		return BACKGROUND_COLOR
	}

	// Computing lighting
	var lighting float64
	for _, l := range w.Lights {
		lighting += l.Power() * l.Illumination(object, PixelLocation, RayVector)
	}
	return ColorMul(object.Color, lighting / w.TotalLightPower)
}

func (w *World) AddObject(obj Object) {
	w.Objects = append(w.Objects, obj)
}

func (w *World) AddLight(l Light) {
	w.Lights = append(w.Lights, l)
	w.TotalLightPower += l.Power()
}

func TranslatePixelToEbiten(i, j int) (x, y int) {
	return j, i
}

type Game struct {
	w World
}

func (g *Game) Update(screen *ebiten.Image) error {
    g.w.ComputeFrame()
    return nil
}

func (g *Game) Draw(screen *ebiten.Image) {
    for i := 0; i < FRAME_HEIGHT; i++ {
		for j := 0; j < FRAME_WIDTH; j++ {
			x, y := TranslatePixelToEbiten(i, j)
			screen.Set(x, y, g.w.Frame[i][j])
		}
	}
}

func (g *Game) Layout(outsideWidth, outsideHeight int) (screenWidth, screenHeight int) {
    return FRAME_WIDTH, FRAME_HEIGHT
}

func main() {
	var w World
	w.SetStandardSettings()
	w.AddObject(CreateSphere(Point{-8, 0, 15}, 3, RGB(200, 0, 0), 0))
	w.AddObject(CreateSphere(Point{10, -4, 10}, 7, RGB(0, 200, 0), 1))
	w.AddObject(CreateSphere(Point{0, -7, 2}, 2, RGB(0, 0, 200), 500))
	w.AddObject(CreateSphere(Point{-8, -4, 15}, 5, RGB(200, 200, 0), 10))
	w.AddObject(CreateSphere(Point{0, -30, 30}, 30, RGB(200, 0, 200), 1000))

	w.AddLight(AmbientLight(4))
	//w.AddLight(PointLight{8, Point{0, 00, 0}})
	w.AddLight(DirectLight{8, Vector{1, -1,  1}})
    game := &Game{w}
    ebiten.SetWindowSize(WINDOW_WIDTH, WINDOW_HEIGHT)
    ebiten.SetWindowTitle("3D")
    if err := ebiten.RunGame(game); err != nil {
        log.Fatal(err)
    }
}
