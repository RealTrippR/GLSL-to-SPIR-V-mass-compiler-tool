/*
Copyright (C) 2025 Tripp R.

Permission is hereby granted, free of charge, to any person obtaining
a copy of this software and associated documentation files (the “Software”),
to deal in the Software without restriction, including without limitation
the rights to use, copy, modify, merge, publish, distribute, sublicense,
and/or sell copies of the Software, and to permit persons to whom the Software
is furnished to do so, subject to the following conditions:

The above copyright notice and this permission notice shall be included
in all copies or substantial portions of the Software.

THE SOFTWARE IS PROVIDED “AS IS”, WITHOUT WARRANTY OF ANY KIND,
EXPRESS OR IMPLIED, INCLUDING BUT NOT LIMITED TO THE
WARRANTIES OF MERCHANTABILITY, FITNESS FOR A PARTICULAR PURPOSE
AND NONINFRINGEMENT. IN NO EVENT SHALL THE AUTHORS OR COPYRIGHT
HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER LIABILITY,
WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER
DEALINGS IN THE SOFTWARE.
*/

package main

import (
	"bytes"
	"fmt"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

type ShaderData struct {
	filepath string
}

type SearchPath struct {
	path      string
	recursive bool
}
type CompileContext struct {
	recursiveSearch   bool
	exclusiveIncude   bool
	ignoreCache       bool
	excludeFilepaths  []SearchPath
	includeFilespaths []SearchPath
	baseFilepath      string //if empty, defaults to working directory
}

func getFiledateSinceEpoch(abs_filename string) uint64 {
	fileInfo, err := os.Stat(abs_filename)
	if err != nil {
		log.Println("getFiledateSinceEpoch:", err)
		return 0x0
	}

	modTime := fileInfo.ModTime()
	epoch := modTime.Unix() // seconds since Jan 1, 1970 UTC
	return uint64(epoch)
}

func fileExists(filename string) bool {
	_, err := os.Stat(filename)
	if err == nil {
		return true // file exists
	}
	if os.IsNotExist(err) {
		return false // file does not exist
	}
	return false
}

func contains(slice []SearchPath, item string) bool {
	for _, v := range slice {
		if v.path == item {
			return true
		}
	}
	return false
}

func isFileshader(absFilepath string) bool {
	data, err := os.ReadFile(absFilepath)
	if err != nil {
		fmt.Println("Error opening file:", err)
		return false
	}

	lines := strings.Split(string(data), "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if strings.HasPrefix(line, "#version") {
			return true
		}
	}

	return false
}

func getShadersToCompileStageII(context *CompileContext, WorkingDir string, recursive bool) []string {

	entries, err := os.ReadDir(WorkingDir)
	if err != nil {
		log.Fatalf("Error reading directory %s: %v", WorkingDir, err)
	}

	shadersToCompile := []string{}

	for _, entry := range entries {
		fileDir := filepath.Join(WorkingDir, entry.Name())
		if !entry.IsDir() {
			if !contains(context.excludeFilepaths, fileDir) {
				if isFileshader(fileDir) {
					shadersToCompile = append(shadersToCompile, fileDir)
				}
			}
		} else {
			if !contains(context.excludeFilepaths, fileDir) {
				if recursive {
					shadersToCompile = append(shadersToCompile, getShadersToCompileStageII(context, fileDir, recursive)...)
				}
			}
		}
	}

	return shadersToCompile
}

func getShadersToCompile(context *CompileContext) []string {
	shadersToCompile := []string{}

	if !context.exclusiveIncude {
		shadersToCompile = append(shadersToCompile, getShadersToCompileStageII(context, context.baseFilepath, context.recursiveSearch)...)
	}
	for _, path := range context.includeFilespaths {
		shadersToCompile = append(shadersToCompile, getShadersToCompileStageII(context, context.baseFilepath, path.recursive)...)
	}

	return shadersToCompile
}

func compileShaders(shadersToCompile []string, context *CompileContext) {
	compiledShaderCount := 0
	failedToCompileCount := 0

	for _, shaderPath := range shadersToCompile {
		var stdout, stderr bytes.Buffer

		ext := filepath.Ext(shaderPath)
		pathNoExt := strings.TrimSuffix(shaderPath, filepath.Ext(shaderPath))
		finalPath := pathNoExt + ext + ".spv"

		/* if source is newer than compiled shader, recompile the shader */
		if !context.ignoreCache && fileExists(finalPath) {
			var compiledDate uint64 = getFiledateSinceEpoch(finalPath)
			var sourceDate uint64 = getFiledateSinceEpoch(shaderPath)

			if sourceDate < compiledDate {
				continue
			}
		}

		cmdHdl := exec.Command("glslc", shaderPath, "-o", finalPath)

		// Redirect output to console
		cmdHdl.Stdout = &stdout
		cmdHdl.Stderr = &stderr
		err := cmdHdl.Run()

		if err != nil {
			if exitError, ok := err.(*exec.ExitError); ok {
				exitCode := exitError.ExitCode()
				if exitCode != 0 {
					log.Println("Failed to compile shader: ", stderr.String())
					failedToCompileCount++
				}
			} else {
			}
		} else {
			fmt.Printf("Compiled shader: %s\n------------\n", shaderPath)
			compiledShaderCount++
		}
	}
	if compiledShaderCount > 0 {
		fmt.Printf("Successfully compiled %d shaders\n", compiledShaderCount)
	}
	if failedToCompileCount > 0 {
		fmt.Printf("Failed to compile compiled %d shaders\n", failedToCompileCount)
	}
	if compiledShaderCount == 0 && failedToCompileCount == 0 {
		fmt.Printf("All shaders are up to date. A total of %d shaders were found.", len(shadersToCompile))
	}
}

func printHelpMenu() {

	fmt.Print(
		`
		GLSL to SPRIV help menu:
		----------------------------
		-r <- recursive
		-f <- forces compilation (ignores cache)
		-b <filepath> <- sets base filepath
		-e <filepath(s)> <- excludes filepaths
		-i <filepath(s)> <- includes filespaths
		-ei <- exlcusive include. If specified, only the include filepaths will be searched
		-help <- show help menu
		-v <- show version
		`)
}

func isArgOption(arg string) bool {
	if arg == "-r" {
		return true
	}
	if arg == "-f" {
		return true
	}
	if arg == "-b" {
		return true
	}
	if arg == "-e" {
		return true
	}
	if arg == "-i" {
		return true
	}
	if arg == "-ei" {
		return true
	}
	return false
}

func parseArguments(context *CompileContext, args []string, errOut *string) bool {
	*errOut = ""
	basepathSet := false
	for i := 0; i < len(args); i++ {
		if args[i] == "-help" {
			printHelpMenu()
		}
		if args[i] == "-version" || args[i] == "-v" {
			fmt.Println("GLSL To SPIR-V Mass Compiler: Version 1.0, 13.11.2025")
		}
		if args[i] == "-r" {
			context.recursiveSearch = true
		}
		if args[i] == "-f" {
			context.ignoreCache = true
		}
		if args[i] == "-b" {
			if !basepathSet {
				i++
				if len(args) <= i {
					*errOut = "'-b': expected filepath to follow command."
					return false
				}
				context.baseFilepath = args[i]
				basepathSet = true
			}
		}
		if args[i] == "-e" {
			i++
			if len(args) <= i {
				return false
			}
			for isArgOption(args[i]) {
				recursive := false
				if args[i] == "-r" {
					recursive = true
				}
				tmp := SearchPath{
					path:      args[i+1],
					recursive: recursive,
				}
				context.excludeFilepaths = append(context.excludeFilepaths, tmp)
				i++
				if len(args) <= i {
					*errOut = "'-e': expected filepath to follow command."
					return false
				}
			}
		}
		if args[i] == "-i" {
			i++
			if len(args) <= i {
				*errOut = "'-i': expected filepath to follow command."
				return false
			}
			for isArgOption(args[i]) {
				recursive := false
				if args[i] == "-r" {
					recursive = true
				}
				tmp := SearchPath{
					path:      args[i+1],
					recursive: recursive,
				}
				context.includeFilespaths = append(context.excludeFilepaths, tmp)
				i++
				if len(args) <= i {
					return false
				}
			}
		}
		if args[i] == "-ei" {
			context.exclusiveIncude = true
		}
	}

	if !basepathSet {
		context.baseFilepath, _ = os.Getwd()
	}

	return true
}

func main() {

	args := os.Args[1:]

	var context CompileContext
	context.recursiveSearch = true

	var err string
	if parseArguments(&context, args, &err) {
		shadersToCompile := getShadersToCompile(&context)
		//fmt.Println("base:", context.baseFilepath)
		compileShaders(shadersToCompile, &context)
	} else {
		fmt.Println("Failed to parse arguments: ", err)
		printHelpMenu()
	}

	os.Exit(0)
}
