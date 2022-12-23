// switch to higher version when needed, try to keep it low for compatibility
#version 330 core

layout(location = 0) in vec4 position;
layout(location = 1) in vec2 texCoord;

out vec2 v_TexCoord;

void main() {
  gl_Position = position;
  v_TexCoord = texCoord;
}