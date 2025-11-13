# GLSL to SPIV-V Mass Compiler Tool

**Brief**
A simple CLI tool to compile glsl shaders quickly and efficiently.

Expects glslc to be installed on your system and accessable in the system PATH.

**Usage**
It will scan for shaders in the cwd (and if the recursively (`-r`) flag is set all subdirectories), producing .spv files alongside the shader source files.

**CLI Flags**
`-r` <- recursive
` -f` <- forces compilation (ignores cache)
`-b` <filepath> <- sets base filepath
`-e` <filepath(s)> <- excludes filepaths
`-i` <filepath(s)> <- includes filespaths
`-ei` <- exlcusive include. If specified, only the include filepaths will be searched
`-help` <- show help menu
`-v` <- show version