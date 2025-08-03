#version 460

layout(location = 0) in vec2 position;
layout(location = 1) in vec3 color;

out outblock {
    layout(location = 0) vec3 color;
} outputs;

void main() {
    outputs.color = color;
    gl_Position = vec4(position, 0.0, 1.0); // 2D, z=0, w=1
}