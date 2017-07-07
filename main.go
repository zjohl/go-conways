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
	width   = 500
	height  = 500
	rows    = 50
	columns = 50

	fps = 10
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
				c.Update(cells)
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

	vertexShader, err := compileShader(vertexShaderSource(), gl.VERTEX_SHADER)
	if err != nil {
		return 0, err
	}
	fragmentShader, err := compileShader(fragmentShaderSource(), gl.FRAGMENT_SHADER)
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
			cell.Draw()
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
	for x := 0; x < rows; x++ {
		for y := 0; y < columns; y++ {
			cell := newCell(x, y)

			cell.alive = rand.Float64() < threshold
			cell.aliveNext = cell.alive

			cells[x] = append(cells[x], cell)
		}
	}
	return cells
}

func newCell(x, y int) *Cell {
	var points []float32
	width := 1.0 / float32(columns)
	height := 1.0 / float32(rows)

	for _, vertex := range square {
		points = append(points, normalize(vertex.x, width, position(float32(x), width)))
		points = append(points, normalize(vertex.y, height, position(float32(y), height)))
		points = append(points, vertex.z)
	}

	return &Cell{
		drawable: makeVao(points),

		x: x,
		y: y,
	}
}

func position(coord, size float32) float32 {
	return coord * size
}

func normalize(coord, size, position float32) float32 {
	if coord < 0 {
		return (position * 2) - 1
	} else {
		return ((position + size) * 2) - 1
	}
}
