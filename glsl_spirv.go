package main

import (
	"bufio"
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
	file, err := os.Open(absFilepath)
	if err != nil {
		fmt.Println("Error opening file:", err)
		return false
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if strings.HasPrefix(line, "#version") {
			//fmt.Println("Found version directive:", line)
			return true
		}
	}

	if err := scanner.Err(); err != nil {
		fmt.Println("Error reading file:", err)
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

func compileShaders(shadersToCompile []string) {
	compiledShaderCount := 0
	failedToCompileCount := 0

	for _, shaderPath := range shadersToCompile {
		var stdout, stderr bytes.Buffer

		ext := filepath.Ext(shaderPath)
		pathNoExt := strings.TrimSuffix(shaderPath, filepath.Ext(shaderPath))
		finalPath := pathNoExt + ext + ".spv"

		if fileExists(finalPath) {
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
			fmt.Printf("Compiled shader %s\n", shaderPath)
			compiledShaderCount++
		}
	}
	if compiledShaderCount > 0 {
		fmt.Printf("Successfully compiled %d shaders", compiledShaderCount)
	}
	if failedToCompileCount > 0 {
		fmt.Printf("Failed to compile compiled %d shaders", failedToCompileCount)
	}
	if compiledShaderCount == 0 && failedToCompileCount == 0 {
		fmt.Printf("All shaders are up to date.")
	}
}

func printHelpMenu() {

	fmt.Print(
		`
		GLSL to SPRIV help menu:\n
		----------------------------\n
		-r <- recursive
		-b <filepath> <- sets base filepath
		-e <filepath(s)> <- excludes filepaths
		-i <filepath(s)> <- includes filespaths
		-ei <- exlcusive include. If specified, only the include filepaths will be searched
		--help <- show help menu
		--version <- show version
		`)
}

func isArgOption(arg string) bool {
	if arg == "-r" {
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

func parseArguments(context *CompileContext, args []string) {
	basepathSet := false
	for i := 0; i < len(args); i++ {
		if args[i] == "--help" {
			printHelpMenu()
		}
		if args[i] == "--version" || args[i] == "-v" {
			fmt.Println("GLSL To SPIR-V Mass Compiler")
		}
		if args[i] == "-r" {
			context.recursiveSearch = true
		}
		if args[i] == "-b" {
			if !basepathSet {
				i++
				if len(args) <= i {
					return
				}
				context.baseFilepath = args[i]
			}
		}
		if args[i] == "-e" {
			i++
			if len(args) <= i {
				return
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
					return
				}
			}
		}
		if args[i] == "-i" {
			i++
			if len(args) <= i {
				return
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
					return
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
}

func main() {
	/*
		argsWithProg := os.Args
		argsWithoutProg := os.Args[1:]

		arg := os.Args[3]

		fmt.Println(argsWithProg)
		fmt.Println(argsWithoutProg)
		fmt.Println(arg)

		wd, err := os.Getwd()
		if err != nil {
			fmt.Println("Error getting current working directory:", err)
			os.Exit(1)
		}
	*/
	printHelpMenu()
	args := os.Args[1:]

	var context CompileContext
	context.recursiveSearch = true
	parseArguments(&context, args)

	context.baseFilepath = "C:\\Users\\TrippR\\OneDrive\\Documents\\REPOS\\neo-chalk\\ck\\lib\\ck\\render_backend\\pipelines"
	shadersToCompile := getShadersToCompile(&context)

	compileShaders(shadersToCompile)

	os.Exit(0)
}
