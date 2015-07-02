package main

import (
	"log"
	"os"
	"os/exec"

	"github.com/pwaller/goupx/hemfix"
)

const usageText = `usage: goupx [args...] files...

    --no-upx: Disables UPX from running.
    --strip-binary: Strips binaries before compressing them.

See UPX's documentation (man upx) for information on UPX's flags.
`

var run_strip = false
var run_upx = true
var upxPath string

// usage prints some nice output instead of panic stacktrace when an user calls
// goupx without arguments
func usage() {
	os.Stderr.WriteString(usageText)
}

// findUpxBinary searches for the upx binary in PATH.
func findUpxBinary() {
	var err error
	upxPath, err = exec.LookPath("upx")
	if err != nil {
		log.Fatal("Couldn't find upx binary in PATH")
	}
}

// parseArguments parses arguments from os.Args and separates the goupx flags
// from the UPX flags, as well as separating the files from the arguments.
func parseArguments() (args []string, files []string) {
	if len(os.Args) == 1 {
		usage()
	}
	args = append(args, upxPath)
	for _, arg := range os.Args[1:] {
		switch {
		case arg == "-h" || arg == "--help":
			usage()
		case arg == "--no-upx":
			run_upx = false
		case arg == "--strip-binary":
			run_strip = true
		case arg[0] != '-':
			files = append(files, arg)
		default:
			args = append(args, arg)
		}
	}
	return
}

// compressBinary attempts to compress the binary with UPX.
func compressBinary(input_file string, arguments []string) {
	if run_upx {
		cmd := &exec.Cmd{
			Path: upxPath,
			Args: append(arguments, input_file),
		}
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		if err := cmd.Run(); err != nil {
			log.Panic("upx failed: ", err)
		}
	}
}

// stripBinary attempts to strip the binary.
func stripBinary(input_file string) {
	if run_strip {
		cmd := exec.Command("strip", "-s", input_file)
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		if err := cmd.Run(); err != nil {
			log.Panic("strip failed: ", err)
		}
	}
}

// runHemfix will attempt to fix the current input file.
func runHemfix(input_file string) {
	if err := hemfix.FixFile(input_file); err != nil {
		log.Panicf("Failed to fix '%s': %v", input_file, err)
	}
	log.Print("File fixed!")
}

func main() {
	findUpxBinary()
	arguments, files := parseArguments()
	for _, file := range files {
		runHemfix(file)
		stripBinary(file)
		compressBinary(file, arguments)
	}
	if err := recover(); err != nil {
		log.Print("Panicked. Giving up.")
		panic(err)
		return
	}
}
