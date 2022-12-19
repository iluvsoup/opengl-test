package main

import (
	"bytes"
	"fmt"
	"image"
	_ "image/gif"
	_ "image/jpeg"
	_ "image/png"
	"reflect"
	"runtime"
	"strings"
	"unsafe"

	gl "github.com/go-gl/gl/v4.6-core/gl"
	glfw "github.com/go-gl/glfw/v3.3/glfw"
)

var positions = []float32{
	-0.5, -0.5,
	0.5, -0.5,
	0.5, 0.5,
	-0.5, 0.5,
}

var indices = []uint32{
	0, 1, 2,
	2, 3, 0,
}

func sizeofuint() int {
	return int(reflect.TypeOf((*uint)(nil)).Elem().Size())
}

func sizeofint() int {
	return int(reflect.TypeOf((*int)(nil)).Elem().Size())
}

func sizeoffloat32() int {
	return 4
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

	gl.UseProgram(program)
	defer gl.DeleteProgram(program)

	var vbo uint32
	gl.GenBuffers(1, &vbo)
	gl.BindBuffer(gl.ARRAY_BUFFER, vbo)
	gl.BufferData(gl.ARRAY_BUFFER, len(positions)*sizeoffloat32(), gl.Ptr(positions), gl.STATIC_DRAW)

	var vao uint32
	gl.GenVertexArrays(1, &vao)
	gl.BindVertexArray(vao)
	gl.BindBuffer(gl.ARRAY_BUFFER, vbo)

	gl.EnableVertexAttribArray(0)
	gl.VertexAttribPointer(0, 2, gl.FLOAT, false, 0, nil)

	var ibo uint32
	gl.GenBuffers(1, &ibo)
	gl.BindBuffer(gl.ELEMENT_ARRAY_BUFFER, ibo)
	gl.BufferData(gl.ELEMENT_ARRAY_BUFFER, len(indices)*sizeofuint(), gl.Ptr(indices), gl.STATIC_DRAW)

	for !window.ShouldClose() {
		gl.Clear(gl.COLOR_BUFFER_BIT | gl.DEPTH_BUFFER_BIT)

		gl.DrawElements(gl.TRIANGLES, 6, gl.UNSIGNED_INT, nil)

		glfw.PollEvents()
		window.SwapBuffers()
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

func throwNotification(notificationMessage string) {
	fmt.Println("\x1b[1;34m" + notificationMessage + "\x1b[0m")
}

func throwWarning(warningMessage string) {
	fmt.Println("\x1b[1;33m" + warningMessage + "\x1b[0m")
}

func throwError(errorMessage error) {
	fmt.Print("\n")
	panic(
		"\x1b[1;31m" + fmt.Sprint(errorMessage),
	)
}

func keyCallback(window *glfw.Window, key glfw.Key, scancode int, action glfw.Action, mods glfw.ModifierKey) {
	if key == glfw.KeyQ && action == glfw.Press {
		fmt.Println("HI")
	}
}

func messageCallback(source uint32, gltype uint32, id uint32, severity uint32, length int32, message string, userparam unsafe.Pointer) {
	var messageType string
	switch gltype {
		case gl.DEBUG_TYPE_ERROR:							  messageType = "Error"
		case gl.DEBUG_TYPE_DEPRECATED_BEHAVIOR: messageType = "Deprecated behavior"
		case gl.DEBUG_TYPE_UNDEFINED_BEHAVIOR:  messageType = "Undefined behavior"
		case gl.DEBUG_TYPE_PORTABILITY: 				messageType = "Portability issue"
		case gl.DEBUG_TYPE_PERFORMANCE: 				messageType = "Performance issue"
		case gl.DEBUG_TYPE_MARKER: 							messageType = "Marker"
		case gl.DEBUG_TYPE_OTHER: 							messageType = "Other"
	}

	var messageSource string
	switch source {
	case gl.DEBUG_SOURCE_API: 						messageSource = "OpenGL API"
	case gl.DEBUG_SOURCE_WINDOW_SYSTEM: 	messageSource = "Window-system API"
	case gl.DEBUG_SOURCE_SHADER_COMPILER: messageSource = "Shader compiler"
	case gl.DEBUG_SOURCE_THIRD_PARTY: 		messageSource = "Application associated with OpenGL"
	case gl.DEBUG_SOURCE_APPLICATION: 		messageSource = "The user of this application"
	case gl.DEBUG_SOURCE_OTHER: 					messageSource = "Other"
	}

	newMessage := messageType + ": " + message + "\n	Source: " + messageSource
	if severity == gl.DEBUG_SEVERITY_HIGH {
		throwError(fmt.Errorf(newMessage))
	} else if severity == gl.DEBUG_SEVERITY_MEDIUM {
		throwWarning(newMessage)
	} else if severity == gl.DEBUG_SEVERITY_LOW {
		throwNotification(newMessage)
	} else if severity == gl.DEBUG_SEVERITY_NOTIFICATION {
		fmt.Println(newMessage)
	}
}

func initOpenGL(vertexShaderSource string, fragmentShaderSource string) uint32 {
	if err := gl.Init(); err != nil {
		throwError(err)
	}

	version := gl.GoStr(gl.GetString(gl.VERSION))
	fmt.Println("\x1b[1m" + version + "\x1b[0m")

	gl.Enable(gl.DEBUG_OUTPUT)
	gl.DebugMessageCallback(messageCallback, nil)

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
