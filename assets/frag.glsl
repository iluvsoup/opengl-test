// switch to higher version when needed, try to keep it low for compatibility
#version 330 core

layout(location = 0) out vec4 frag_colour;

uniform vec4 u_Color;

void main() {
  frag_colour = u_Color;
}