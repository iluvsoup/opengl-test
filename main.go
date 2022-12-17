package main

import (
	"bytes"
	"fmt"
	"image"
	_ "image/gif"
	_ "image/jpeg"
	_ "image/png"
	"runtime"
	"strings"

	gl "github.com/go-gl/gl/v4.6-core/gl"
	glfw "github.com/go-gl/glfw/v3.3/glfw"
)

var test = []float32{
	0, 0.5, 0,
	-0.5, -0.5, 0,
	0.5, -0.5, 0,
}

func main() {
	runtime.LockOSThread() // This is because GLFW has to run on the same thread it was initialized on

	window := initGlfw(800, 600, "test")
	setIcon(window, "assets/cat.png")

	defer glfw.Terminate()

	vertexShaderSource, err := Asset("assets/vertex.glsl")
	if err != nil {
		throwError(err)
	}

	fragmentShaderSource, err := Asset("assets/frag.glsl")
	if err != nil {
		throwError(err)
	}

	program := initOpenGL(
		string(vertexShaderSource)+"\x00", // \x00 is required because openGL
		string(fragmentShaderSource)+"\x00",
	)

	triangle := makeVao(test)

	for !window.ShouldClose() {
		draw(triangle, window, program)
	}
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

		gl.DeleteShader(shader)

		return 0, fmt.Errorf("Failed to compile shader\n%v: %v", source, log)
	}

	return shader, nil
}

func throwWarning(warningMessage string) {
	fmt.Println("warning:\x1b[1;33m", warningMessage, "\x1b[0m")
}

func throwError(errorMessage error) {
	fmt.Print("\n")
	panic(
		fmt.Sprint("\x1b[1;31m", errorMessage),
	)
}

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

func draw(vao uint32, window *glfw.Window, program uint32) {
	gl.Clear(gl.COLOR_BUFFER_BIT | gl.DEPTH_BUFFER_BIT)
	gl.UseProgram(program)

	gl.BindVertexArray(vao)
	gl.DrawArrays(gl.TRIANGLES, 0, int32(len(test)/3))

	glfw.PollEvents()
	window.SwapBuffers()
}

func keyCallback(window *glfw.Window, key glfw.Key, scancode int, action glfw.Action, mods glfw.ModifierKey) {
	if key == glfw.KeyQ && action == glfw.Press {
		fmt.Println("HI")
	}
}

func initOpenGL(vertexShaderSource string, fragmentShaderSource string) uint32 {
	if err := gl.Init(); err != nil {
		throwError(err)
	}

	version := gl.GoStr(gl.GetString(gl.VERSION))
	fmt.Println(version)

	vertexShader, err := compileShader(vertexShaderSource, gl.VERTEX_SHADER)
	if err != nil {
		throwError(err)
	}

	fragmentShader, err := compileShader(fragmentShaderSource, gl.FRAGMENT_SHADER)
	if err != nil {
		throwError(err)
	}

	program := gl.CreateProgram()
	gl.AttachShader(program, vertexShader)
	gl.AttachShader(program, fragmentShader)

	gl.LinkProgram(program)
	gl.ValidateProgram(program)

	gl.DeleteShader(vertexShader)
	gl.DeleteShader(fragmentShader)

	return program
}

func initGlfw(width, height int, name string) *glfw.Window {
	if err := glfw.Init(); err != nil {
		panic(err)
	}

	// openGL v4.6
	glfw.WindowHint(glfw.ContextVersionMajor, 4)
	glfw.WindowHint(glfw.ContextVersionMinor, 6)

	glfw.WindowHint(glfw.OpenGLProfile, glfw.OpenGLCoreProfile)
	glfw.WindowHint(glfw.OpenGLForwardCompatible, glfw.True)

	window, err := glfw.CreateWindow(width, height, name, nil, nil)
	if err != nil {
		throwError(err)
	}

	window.MakeContextCurrent()
	window.SetKeyCallback(keyCallback)

	return window
}

func setIcon(window *glfw.Window, filename string) {
	imageBytes, err := Asset(filename)
	if err != nil {
		throwError(err)
	}

	reader := bytes.NewReader(imageBytes)

	// defer reader.Close()

	png, _, err := image.Decode(reader)
	if err != nil {
		throwError(err)
	}

	// Dunno why it takes an array tbh
	var images = []image.Image{png}
	window.SetIcon(images)
}
