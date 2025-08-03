#version 460

in inblock {
	layout(location = 0) vec3 color;
} inputs;

layout(location = 0) out vec4 outColor; /* color attachment 0 */

void main() {
	outColor = vec4(inputs.color,1.0);
}