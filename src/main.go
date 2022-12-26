package main

import (
	"bytes"
	"fmt"
	"image"

	"image/draw"
	_ "image/draw"
	_ "image/gif"
	_ "image/jpeg"
	_ "image/png"
	"main/src/util"
	"math"
	"runtime"
	"strings"
	"unsafe"

	gl "github.com/go-gl/gl/v4.6-core/gl"
	glfw "github.com/go-gl/glfw/v3.3/glfw"

	// not actually glm but I'll call it glm anyway
	glm "github.com/go-gl/mathgl/mgl32"
)

// X,Y, U,V
/*
var positions = []float32{
	-0.5, -0.5, 0.0, 0.0,
	 0.5, -0.5, 1.0, 0.0,
	 0.5,  0.5, 1.0, 1.0,
	-0.5,  0.5, 0.0, 1.0,
}

var indices = []uint32{
	0, 1, 2,
	2, 3, 0,
}
*/

// X,Y,Z, U,V
var positions = []float32{
	 1.0, -1.0, -1.0,   1.0, 0.0,
	 1.0, -1.0,  1.0,   0.0, 0.0,
	-1.0, -1.0,  1.0,   1.0, 1.0,
	-1.0, -1.0, -1.0,   1.0, 0.0,
	 1.0,  1.0, -1.0,   0.0, 0.0,
	 1.0,  1.0,  1.0,   1.0, 1.0,
	-1.0,  1.0,  1.0,   1.0, 0.0,
	-1.0,  1.0, -1.0,   0.0, 1.0,
}

var indices = []uint32{
	1, 2, 3,
	7, 6, 5,
	4, 5, 1,
	5, 6, 2,
	2, 6, 7,
	0, 3, 7,
	0, 1, 3,
	4, 7, 5,
	0, 4, 1,
	1, 5, 2,
	3, 2, 7,
	4, 0, 7,
}

func main() {
	runtime.LockOSThread() // This is because GLFW has to run on the same thread it was initialized on

	window := initGlfw(600, 800, "test")
	defer glfw.Terminate()

	cat, err := Asset("assets/cat.png")
	if err != nil {
		util.ThrowError(err)
	}

	setIcon(window, cat)


	vertexShaderSource, err := Asset("assets/vertex.glsl")
	if err != nil {
		util.ThrowError(err)
	}

	fragmentShaderSource, err := Asset("assets/frag.glsl")
	if err != nil {
		util.ThrowError(err)
	}

	program, err := initOpenGL(
		string(vertexShaderSource)+"\x00", // \x00 is required because openGL
		string(fragmentShaderSource)+"\x00",
	)

	defer gl.DeleteProgram(program)

	if err != nil {
		util.ThrowError(err)
	}

	gl.UseProgram(program)


	// vertex buffer
	var vbo uint32
	gl.GenBuffers(1, &vbo)
	defer gl.DeleteBuffers(1, &vbo)
	
	gl.BindBuffer(gl.ARRAY_BUFFER, vbo)
	gl.BufferData(gl.ARRAY_BUFFER, len(positions)*util.Sizeoffloat32(), gl.Ptr(positions), gl.STATIC_DRAW)


	// vertex array
	var vao uint32
	gl.GenVertexArrays(1, &vao)
	defer gl.DeleteVertexArrays(1, &vao)
	
	gl.BindVertexArray(vao)

	gl.EnableVertexAttribArray(0)
	gl.VertexAttribPointerWithOffset(0, 3, gl.FLOAT, false, int32(5*util.Sizeoffloat32()), 0)

	gl.EnableVertexAttribArray(1)
	gl.VertexAttribPointerWithOffset(1, 2, gl.FLOAT, false, int32(5*util.Sizeoffloat32()), uintptr(3*util.Sizeoffloat32()))


	// index buffer
	var ibo uint32
	gl.GenBuffers(1, &ibo)
	defer gl.DeleteBuffers(1, &ibo)
	
	gl.BindBuffer(gl.ELEMENT_ARRAY_BUFFER, ibo)
	gl.BufferData(gl.ELEMENT_ARRAY_BUFFER, len(indices)*util.Sizeofuint32(), gl.Ptr(indices), gl.STATIC_DRAW)


	// texture
	catBytes, err := Asset("assets/morgana.jpg")
	if err != nil {
		util.ThrowError(err)
	}

	reader := bytes.NewReader(catBytes)
	catImage, _, err := image.Decode(reader)
	if err != nil {
		util.ThrowError(err)
	}

	catTexture := image.NewRGBA(catImage.Bounds())
	// stolen
	draw.Draw(catTexture, catTexture.Bounds(), catImage, image.Point{0, 0}, draw.Src)
	flippedPixels := flipImage(catTexture)

	var texture uint32
	gl.GenTextures(1, &texture)
	defer gl.DeleteTextures(1, &texture)
	
	gl.BindTexture(gl.TEXTURE_2D, texture)
	
	gl.TexParameteri(gl.TEXTURE_2D, gl.TEXTURE_MIN_FILTER, gl.LINEAR)
	gl.TexParameteri(gl.TEXTURE_2D, gl.TEXTURE_MAG_FILTER, gl.LINEAR)
	gl.TexParameteri(gl.TEXTURE_2D, gl.TEXTURE_WRAP_S, gl.CLAMP_TO_EDGE)
	gl.TexParameteri(gl.TEXTURE_2D, gl.TEXTURE_WRAP_T, gl.CLAMP_TO_EDGE)

	gl.TexImage2D(
		gl.TEXTURE_2D,
		0,
		gl.RGBA,
		int32(catTexture.Rect.Size().X),
		int32(catTexture.Rect.Size().Y),
		0,
		gl.RGBA,
		gl.UNSIGNED_BYTE,
		gl.Ptr(flippedPixels),
	)

	gl.ActiveTexture(gl.TEXTURE0)

	// probably does nothing because go has a garbage collector but it helps me sleep at night
	catBytes = nil
	catImage = nil
	catTexture = nil

	
	// bind texture to texture slot 0
	textureLocation := uniformLocation("u_Texture", &program)
	gl.Uniform1i(textureLocation, 0)

	
	colorLocation := uniformLocation("u_Color", &program)
	mvpLocation := uniformLocation("u_MVP", &program)

	var previousTime float64
	var fpsUpdatePreviousTime float64

	var r float32 = 0.0; var r_phase bool = true;
	var g float32 = 0.0; var g_phase bool = true;
	var b float32 = 0.0; var b_phase bool = true;

	var angle float32 = 0.0;

	for !window.ShouldClose() {
		// FPS
		currentTime := glfw.GetTime()
		deltaTime := currentTime - previousTime
		timeSinceFPSUpdate := currentTime - fpsUpdatePreviousTime

		if timeSinceFPSUpdate >= 1 {
			fps := int(math.Round(1 / deltaTime))
			window.SetTitle(fmt.Sprintf("%d FPS", fps))

			fpsUpdatePreviousTime = currentTime
		}
		
		previousTime = currentTime

		gl.Clear(gl.COLOR_BUFFER_BIT | gl.DEPTH_BUFFER_BIT)

		screenX, screenY := window.GetSize()
		aspectRatio := float32(screenX) / float32(screenY)
		
		projection := glm.Perspective(glm.DegToRad(90), aspectRatio, 0.01, 1000.0)
		view := glm.LookAtV(glm.Vec3{3,3,3}, glm.Vec3{0,0,0}, glm.Vec3{0,1,0})
		model := glm.HomogRotate3D(angle, glm.Vec3{0, 1, 0})
		
		mvp := projection.Mul4(view).Mul4(model)
		gl.UniformMatrix4fv(mvpLocation, 1, false, &mvp[0])

		gl.Uniform4f(colorLocation, r, g, b, 1.0)
		gl.DrawElements(gl.TRIANGLES, int32(len(indices)), gl.UNSIGNED_INT, nil)
		// gl.DrawArrays(gl.POINTS, 0, int32(len(positions) / 3))

		glfw.PollEvents()
		window.SwapBuffers()

		deltaTime32 := float32(deltaTime)

		angle += deltaTime32

		// "mimimimimimi ugly code"
		// stfu it's compact
		if (r_phase == true) { r += deltaTime32 * 0.50 } else { r -= deltaTime32 * 0.50 }
		if (g_phase == true) { g += deltaTime32 * 0.25 } else { g -= deltaTime32 * 0.25 }
		if (b_phase == true) { b += deltaTime32 * 0.10 } else { b -= deltaTime32 * 0.10 }

		if (r >= 1) { r_phase = false } else if (r <= 0) { r_phase = true }
		if (g >= 1) { g_phase = false } else if (g <= 0) { g_phase = true }
		if (b >= 1) { b_phase = false } else if (b <= 0) { b_phase = true }
	}
}

func flipImage(image *image.RGBA) []uint8 {
	pixels := make([]uint8, len(image.Pix))
	imageX := image.Rect.Size().X
	imageY := image.Rect.Size().Y

	for y := 0; y < imageY; y++ {	
		for x := 0; x < imageX; x++ {
			index := (y * imageX + x) * 4
			flippedIndex := ((imageY - y - 1) * imageX + x) * 4

			pixels[index + 0] = image.Pix[flippedIndex + 0]
			pixels[index + 1] = image.Pix[flippedIndex + 1]
			pixels[index + 2] = image.Pix[flippedIndex + 2]
			pixels[index + 3] = image.Pix[flippedIndex + 3]
		}
	}

	return pixels
}

func uniformLocation(name string, program *uint32) int32 {
	location := gl.GetUniformLocation(*program, gl.Str(name + "\x00"))
	if location == -1 {
		util.ThrowWarning("Could not find location of uniform: " + name)
	}
	return location
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

		var shaderTypeString string

		switch shaderType {
		case gl.VERTEX_SHADER:
			shaderTypeString = "vertex"
		case gl.FRAGMENT_SHADER:
			shaderTypeString = "fragment"
		}

		return 0, fmt.Errorf("Failed to compile %s shader: %v", shaderTypeString, log)
	}
	
	return shader, nil
}

func keyCallback(window *glfw.Window, key glfw.Key, scancode int, action glfw.Action, mods glfw.ModifierKey) {
	if key == glfw.KeyQ && action == glfw.Press {
		fmt.Println("HI")
	}
}

func resizeCallback(window *glfw.Window, width int, height int) {
	gl.Viewport(0, 0, int32(width), int32(height))
	gl.Clear(gl.COLOR_BUFFER_BIT | gl.DEPTH_BUFFER_BIT)
	
	window.SwapBuffers()
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
		util.ThrowError(fmt.Errorf(newMessage))
	} else if severity == gl.DEBUG_SEVERITY_MEDIUM {
		util.ThrowWarning(newMessage)
	} else if severity == gl.DEBUG_SEVERITY_LOW {
		util.ThrowNotification(newMessage)
	} else if severity == gl.DEBUG_SEVERITY_NOTIFICATION {
		fmt.Println(newMessage)
	}
}

func initOpenGL(vertexShaderSource string, fragmentShaderSource string) (uint32, error) {
	if err := gl.Init(); err != nil {
		util.ThrowError(err)
	}

	version := gl.GoStr(gl.GetString(gl.VERSION))
	fmt.Println("\x1b[1m" + version + "\x1b[0m")

	gl.Enable(gl.DEBUG_OUTPUT)
	gl.DebugMessageCallback(messageCallback, nil)

	gl.Enable(gl.DEPTH_TEST)
	gl.DepthFunc(gl.LESS)

	vertexShader, err := compileShader(vertexShaderSource, gl.VERTEX_SHADER)
	if err != nil {
		util.ThrowError(err)
	}

	fragmentShader, err := compileShader(fragmentShaderSource, gl.FRAGMENT_SHADER)
	if err != nil {
		util.ThrowError(err)
	}

	program := gl.CreateProgram()
	gl.AttachShader(program, vertexShader)
	gl.AttachShader(program, fragmentShader)

	gl.LinkProgram(program)
	gl.ValidateProgram(program)

	var status int32
	gl.GetProgramiv(program, gl.LINK_STATUS, &status)
	if status == gl.FALSE {
		var logLength int32
		gl.GetProgramiv(program, gl.INFO_LOG_LENGTH, &logLength)

		log := strings.Repeat("\x00", int(logLength+1))
		gl.GetProgramInfoLog(program, logLength, nil, gl.Str(log))

		return 0, fmt.Errorf("failed to link program: %v", log)
	}

	gl.DeleteShader(vertexShader)
	gl.DeleteShader(fragmentShader)

	return program, nil
}

func initGlfw(width, height int, name string) *glfw.Window {
	if err := glfw.Init(); err != nil {
		util.ThrowError(err)
	}

	// opengl v4.6
	glfw.WindowHint(glfw.ContextVersionMajor, 4)
	glfw.WindowHint(glfw.ContextVersionMinor, 6)

	glfw.WindowHint(glfw.OpenGLProfile, glfw.OpenGLCoreProfile)
	glfw.WindowHint(glfw.OpenGLForwardCompatible, glfw.True)

	window, err := glfw.CreateWindow(width, height, name, nil, nil)
	if err != nil {
		util.ThrowError(err)
	}

	window.MakeContextCurrent()
	glfw.SwapInterval(1)

	window.SetKeyCallback(keyCallback)
	window.SetSizeCallback(resizeCallback)

	return window
}

func setIcon(window *glfw.Window, imageBytes []byte) {
	reader := bytes.NewReader(imageBytes)

	// defer reader.Close()

	png, _, err := image.Decode(reader)
	if err != nil {
		util.ThrowError(err)
	}

	// Dunno why it takes an array tbh
	var images = []image.Image{png}
	window.SetIcon(images)
}
