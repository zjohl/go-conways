package main

import (
	"log"
	"runtime"

	"fmt"
	"os"
	"strings"

	"github.com/go-gl/gl/v4.1-core/gl"
	"github.com/go-gl/glfw/v3.2/glfw"
	"math/rand"
	"time"
)

const (
	width  = 500
	height = 500
	rows = 100
	columns = 100

	fps = 10

	vertexShaderSource = `
		#version 410
		in vec3 vp;
		void main() {
			gl_Position = vec4(vp, 1.0);
		}
	` + "\x00"

	fragmentShaderSource = `
		#version 410
		out vec4 frag_colour;
		void main() {
			frag_colour = vec4(1, 1, 1, 1.0);
		}
	` + "\x00"
)

var (
	square = []float32{
		-0.5, 0.5, 0,
		-0.5, -0.5, 0,
		0.5, -0.5, 0,

		-0.5, 0.5, 0,
		0.5, 0.5, 0,
		0.5, -0.5, 0,
	}
)

func main() {
	runtime.LockOSThread()

	window, err := initGlfw()
	if err != nil {
		log.Fatalln(err)
		os.Exit(1)
	}
	defer glfw.Terminate()

	program, err := initOpenGL()
	if err != nil {
		log.Fatalln(err)
		os.Exit(1)
	}

	cells := makeCells()
	for !window.ShouldClose() {
		t := time.Now()

		for x := range cells {
			for _, c := range cells[x] {
				c.checkState(cells)
			}
		}

		draw(cells, window, program)

		time.Sleep(time.Second/time.Duration(fps) - time.Since(t))
	}
	return
	os.Exit(0)
}

func initGlfw() (*glfw.Window, error) {
	if err := glfw.Init(); err != nil {
		return nil, err
	}

	glfw.WindowHint(glfw.Resizable, glfw.False)
	glfw.WindowHint(glfw.ContextVersionMajor, 4)
	glfw.WindowHint(glfw.ContextVersionMinor, 1)
	glfw.WindowHint(glfw.OpenGLProfile, glfw.OpenGLCoreProfile)
	glfw.WindowHint(glfw.OpenGLForwardCompatible, glfw.True)

	window, err := glfw.CreateWindow(width, height, "Conway's game of life", nil, nil)
	if err != nil {
		return nil, err
	}
	window.MakeContextCurrent()
	return window, nil
}

func initOpenGL() (uint32, error) {
	if err := gl.Init(); err != nil {
		return 0, err
	}
	version := gl.GoStr(gl.GetString(gl.VERSION))
	log.Println("OpenGL version", version)

	vertexShader, err := compileShader(vertexShaderSource, gl.VERTEX_SHADER)
	if err != nil {
		return 0, err
	}
	fragmentShader, err := compileShader(fragmentShaderSource, gl.FRAGMENT_SHADER)
	if err != nil {
		return 0, err
	}

	prog := gl.CreateProgram()
	gl.AttachShader(prog, vertexShader)
	gl.AttachShader(prog, fragmentShader)
	gl.LinkProgram(prog)
	return prog, nil
}

func draw(cells [][]*Cell, window *glfw.Window, program uint32) {
	gl.Clear(gl.COLOR_BUFFER_BIT | gl.DEPTH_BUFFER_BIT)
	gl.UseProgram(program)

	for _, row := range cells {
		for _, cell := range row {
			cell.draw()
		}
	}

	glfw.PollEvents()
	window.SwapBuffers()
}

// returns a vertex array from the points provided
func makeVao(points []float32) uint32 {
	var vbo uint32
	gl.GenBuffers(1, &vbo)
	gl.BindBuffer(gl.ARRAY_BUFFER, vbo)
	gl.BufferData(gl.ARRAY_BUFFER, 4*len(points), gl.Ptr(points), gl.STATIC_DRAW)

	var vao uint32
	gl.GenVertexArrays(1, &vao)
	gl.BindVertexArray(vao)
	gl.EnableVertexAttribArray(0)
	gl.BindBuffer(gl.ARRAY_BUFFER, vbo)
	gl.VertexAttribPointer(0, 3, gl.FLOAT, false, 0, nil)

	return vao
}

func compileShader(source string, shaderType uint32) (uint32, error) {
	shader := gl.CreateShader(shaderType)

	csources, free := gl.Strs(source)
	gl.ShaderSource(shader, 1, csources, nil)
	free()
	gl.CompileShader(shader)

	var status int32
	gl.GetShaderiv(shader, gl.COMPILE_STATUS, &status)
	if status == gl.FALSE {
		var logLength int32
		gl.GetShaderiv(shader, gl.INFO_LOG_LENGTH, &logLength)

		log := strings.Repeat("\x00", int(logLength+1))
		gl.GetShaderInfoLog(shader, logLength, nil, gl.Str(log))

		return 0, fmt.Errorf("failed to compile %v: %v", source, log)
	}
	return shader, nil
}

func makeCells() [][]*Cell {
	threshold := .15
	rand.Seed(time.Now().UnixNano())

	cells := make([][]*Cell, rows, rows)
	for x := 0; x < rows; x ++ {
		for y := 0; y < columns; y++ {
			cell := newCell(x,y)

			cell.alive = rand.Float64() < threshold
			cell.aliveNext = cell.alive

			cells[x] = append(cells[x], cell)
		}
	}
	return cells
}

func newCell(x, y int) *Cell {
	points := make([]float32, len(square), len(square))
	copy(points, square)
	for i := 0; i < len(points); i ++ {
		var position float32
		var size float32
		switch i % 3 {
		case 0:
			size = 1.0 / float32(columns)
			position = float32(x) * size
		case 1:
			size = 1.0 / float32(rows)
			position = float32(y) * size
		default:
			continue
		}

		if points[i] < 0 {
			points[i] = (position * 2) -1
		} else {
			points[i] = ((position + size) * 2) -1
		}
	}

	return &Cell {
		drawable: makeVao(points),

		x: x,
		y: y,
	}
}

type Cell struct {
	drawable uint32

	alive     bool
	aliveNext bool

	x int
	y int
}

func (c *Cell) draw() {
	if !c.alive {
		return
	}

	gl.BindVertexArray(c.drawable)
	gl.DrawArrays(gl.TRIANGLES, 0, int32(len(square)/3))
}

func (c *Cell) checkState(Cells [][]*Cell) {
	c.alive = c.aliveNext
	c.aliveNext = c.alive

	liveCount := c.liveNeighbors(Cells)
	if c.alive {
		// Cells with fewer than two live neighbors dies
		if liveCount < 2 {
			c.aliveNext = false
		}

		// A live Cell with 2 or 3 live neighbors lives
		if liveCount == 2 || liveCount == 3 {
			c.aliveNext = true
		}

		// A live Cell with more than 3 neighbors dies
		if liveCount > 3 {
			c.aliveNext = false
		}
	} else {
		if liveCount == 3 {
			c.aliveNext = true
		}
	}
}

func (c *Cell) liveNeighbors(Cells [][]*Cell) int {
	var liveCount int
	add := func(x, y int) {
		// If we're at an edge, check the other side of the board.
		if x == len(Cells) {
			x = 0
		} else if x == -1 {
			x = len(Cells) - 1
		}
		if y == len(Cells[x]) {
			y = 0
		} else if y == -1 {
			y = len(Cells[x]) - 1
		}

		if Cells[x][y].alive {
			liveCount++
		}
	}

	add(c.x-1, c.y)   // To the left
	add(c.x+1, c.y)   // To the right
	add(c.x, c.y+1)   // up
	add(c.x, c.y-1)   // down
	add(c.x-1, c.y+1) // top-left
	add(c.x+1, c.y+1) // top-right
	add(c.x-1, c.y-1) // bottom-left
	add(c.x+1, c.y-1) // bottom-right

	return liveCount
}