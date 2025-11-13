# GLSL to SPIV-V Mass Compiler Tool

<ins> **Brief** </ins>

A simple CLI tool to compile glsl shaders quickly and efficiently.

Expects glslc to be installed on your system and accessable in the system PATH.

<ins> **Usage** </ins>

It will scan for shaders in the cwd (and if the recursively (`-r`) flag is set all subdirectories), producing .spv files alongside the shader source files.

<ins> **CLI Flags** </ins>

`-r` <- recursive <br>
`-f` <- forces compilation (ignores cache) <br>
`-b` <filepath> <- sets base filepath <br> 
`-e` <filepath(s)> <- excludes filepaths <br>
`-i` <filepath(s)> <- includes filespaths <br>
`-ei` <- exlcusive include. If specified, only the include filepaths will be searched <br>
`-help` <- show help menu <br>
`-v` <- show version <br>
