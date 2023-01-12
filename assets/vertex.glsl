// switch to higher version when needed, try to keep it low for compatibility
#version 330 core

layout(location = 0) in vec4 position;
layout(location = 1) in vec2 texCoord;

uniform mat4 u_MVP;

out vec2 v_TexCoord;

int radius = 10;

void main() {
  float angle = (2 * 3.14159265 / 25) * gl_InstanceID;
  vec4 offset = vec4(cos(angle) * radius, 0, sin(angle) * radius, 0);

  gl_Position = u_MVP * (position + offset);
  v_TexCoord = texCoord;
}