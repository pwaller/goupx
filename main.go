package main

import (
	"flag"
	"fmt"
	"github.com/pwaller/goupx/hemfix"
	"log"
	"os"
	"os/exec"
)

const usageText = "usage: goupx [args...] path\n"

var (
	run_strip               = flag.Bool("s", false, "run strip")
	run_upx                 = flag.Bool("u", true, "run upx")
	manual_compression      = flag.Uint("c", 0, "Manual Compression Level: (1-9)")
	best_compression        = flag.Bool("best", false, "Best Compression")
	brute_compression       = flag.Bool("brute", false, "Brute Compression")
	ultra_brute_compression = flag.Bool("ultra-brute", false, "Ultra Brute Compression")
)

// usage prints some nice output instead of panic stacktrace when an user calls
// goupx without arguments
func usage() {
	os.Stderr.WriteString(usageText)
	flag.PrintDefaults()
}

// checkCompressionLevel sets the compression level to 9 if it is higher than
// the maximum limit
func checkCompressionLevel() {
	if *manual_compression > uint(9) {
		*manual_compression = uint(9)
	}
}

// stripBinary attempts to strip the binary.
func stripBinary(input_file string) {
	cmd := exec.Command("strip", "-s", input_file)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		log.Panic("strip failed: ", err)
	}
}

// compressBinary attempts to compress the binary with UPX.
func compressBinary(input_file string) {
	var cmd *exec.Cmd
	if *best_compression {
		cmd = exec.Command("upx", "--best", input_file)
	} else if *brute_compression {
		cmd = exec.Command("upx", "--brute", input_file)
	} else if *ultra_brute_compression {
		cmd = exec.Command("upx", "--ultra-brute", input_file)
	} else if *manual_compression != 0 {
		cmd = exec.Command("upx",
			fmt.Sprintf("%s%d", "-", *manual_compression), input_file)
	} else {
		cmd = exec.Command("upx", input_file)
	}
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		log.Panic("upx failed: ", err)
	}
}

func main() {
	flag.Parse()

	if flag.NArg() != 1 {
		usage()
		return
	}

	checkCompressionLevel()

	defer func() {
		if err := recover(); err != nil {
			log.Print("Panicked. Giving up.")
			panic(err)
			return
		}
	}()

	input_file := flag.Arg(0)
	err := hemfix.FixFile(input_file)
	if err != nil {
		log.Panicf("Failed to fix '%s': %v", input_file, err)
	}
	log.Print("File fixed!")

	if *run_strip {
		stripBinary(input_file)
	}

	if *run_upx {
		compressBinary(input_file)
	}
}
