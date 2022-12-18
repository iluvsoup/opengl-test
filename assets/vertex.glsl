// switch to higher version when needed, try to keep it low for compatibility
#version 330 core

layout(location=0)in vec4 vp;

void main(){
  gl_Position=vp;
}