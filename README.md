this is my attempt at getting familiar with opengl

why did I choose golang? I don't know (my experience has been mixed at best)

the code is a mess because I chose not to abstract anything, I wanted to learn about the actual function calls you had to do to make stuff happen, and abstracting that away wouldn't really do much other than cleaning up the code

requirements:

- go-bindata (if it doesn't get installed in the build step)
- valid c compiler
- go
- opengl
- glfw

TODO:

- build tool
- make obj parsing work with index buffer
- batched or indexed rendering?
- triangulation? how to handle faces with 5 vertices
- proper shading
- anti aliasing
- ability to render multiple objects
- culling, stuff like that
- better graphics
- just general optimization stuff
