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

	obj, err := Asset("assets/burger.obj")
	if err != nil {
		util.ThrowError(err)
	}

	vertices := parseObj(&obj)

	// vertex buffer
	var vbo uint32
	gl.GenBuffers(1, &vbo)
	defer gl.DeleteBuffers(1, &vbo)
	
	gl.BindBuffer(gl.ARRAY_BUFFER, vbo)
	gl.BufferData(gl.ARRAY_BUFFER, len(vertices)*util.Sizeoffloat32(), gl.Ptr(vertices), gl.STATIC_DRAW)


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
	/*
	var ibo uint32
	gl.GenBuffers(1, &ibo)
	defer gl.DeleteBuffers(1, &ibo)
	
	gl.BindBuffer(gl.ELEMENT_ARRAY_BUFFER, ibo)
	gl.BufferData(gl.ELEMENT_ARRAY_BUFFER, len(indices)*util.Sizeofuint32(), gl.Ptr(indices), gl.STATIC_DRAW)
	*/

	// texture
	catBytes, err := Asset("assets/texture.png")
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
	
	gl.TexParameteri(gl.TEXTURE_2D, gl.TEXTURE_MIN_FILTER, gl.NEAREST)
	gl.TexParameteri(gl.TEXTURE_2D, gl.TEXTURE_MAG_FILTER, gl.NEAREST)
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
	
	mvpLocation := uniformLocation("u_MVP", &program)
	burgerCountLocation := uniformLocation("u_BurgerCount", &program)
	radiusLocation := uniformLocation("u_Radius", &program)

	var burgerCount int32 = 25;
	var radius float32 = 10;

	var previousTime float64
	var fpsUpdatePreviousTime float64

	var angle float32 = 0.0;
	
	var previousCursorX float64

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
		
		cursorX, cursorY := window.GetCursorPos()
		
		normalizedCursorX := cursorX / float64(screenX / 2.0) - 1
		normalizedCursorY := cursorY / float64(screenY / 2.0) - 1

		cursorDistanceFromCenter := math.Sqrt(math.Pow(normalizedCursorX, 2) + math.Pow(normalizedCursorY, 2))
		radius = float32(cursorDistanceFromCenter) * 15

		deltaCursorX := cursorX - previousCursorX
		angle += float32(deltaCursorX * 0.01)
		previousCursorX = cursorX

		projection := glm.Perspective(glm.DegToRad(90), aspectRatio, 0.01, 1000.0)
		view := glm.LookAtV(glm.Vec3{10, 5, 10}, glm.Vec3{0, 0, 0}, glm.Vec3{0, 1, 0})
		model := glm.Translate3D(0, 0, 0).Mul4(glm.HomogRotate3DY(angle))

		mvp := projection.Mul4(view).Mul4(model)
		gl.UniformMatrix4fv(mvpLocation, 1, false, &mvp[0])

		gl.Uniform1i(burgerCountLocation, burgerCount)
		gl.Uniform1f(radiusLocation, radius)

		// gl.DrawElements(gl.TRIANGLES, int32(len(indices)), gl.UNSIGNED_INT, nil)
		// gl.DrawArrays(gl.TRIANGLES, 0, int32(len(vertices)))
		gl.DrawArraysInstanced(gl.TRIANGLES, 0, int32(len(vertices)), burgerCount)

		glfw.PollEvents()
		window.SwapBuffers()
	}
}

// dirty code
func parseObj(obj *[]byte) []float32 {
	source := string(*obj)
	lines := strings.Split(source, "\n")

	var positions []float32
	var textureCoordinates []float32
	var vertices []float32

	for _, line := range lines {
		components := strings.Split(line, " ")
		dataType := components[0]

		if dataType == "v" {
			// obj stores uv coordinates and other values as 16 bit floats (i think)
			x := util.ParseFloat(components[1], 16)
			y := util.ParseFloat(components[2], 16)
			z := util.ParseFloat(components[3], 16)

			positions = append(positions, float32(x), float32(y), float32(z))
		} else if dataType == "vt" {
			u := util.ParseFloat(components[1], 16)
			v := util.ParseFloat(components[2], 16)

			textureCoordinates = append(textureCoordinates, float32(u), float32(v))
		} else if dataType == "f" {
			for v := 1; v <= 3; v++ {
				indices := strings.Split(components[v], "/")
				
				positionIndex := util.ParseInt(indices[0], 10, 32)
				textureCoordinateIndex := util.ParseInt(indices[1], 10, 32)

				// Have to subtract one because the indices in .obj files are 1-based
				x := positions[(positionIndex - 1) * 3 + 0]
				y := positions[(positionIndex - 1) * 3 + 1]
				z := positions[(positionIndex - 1) * 3 + 2]

				u := textureCoordinates[(textureCoordinateIndex - 1) * 2 + 0]
				v := textureCoordinates[(textureCoordinateIndex - 1) * 2 + 1]

				vertices = append(vertices, x, y, z, u, v)
			}
		}
	}

	return vertices
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
