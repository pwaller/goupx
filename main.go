package main

import (
	"flag"
	"log"
	"os"
	"os/exec"

	"github.com/pwaller/goupx/hemfix"
)

const usageText = "usage: goupx [args...] path\n"

// usage prints some nice output instead of panic stacktrace when an user calls goupx without arguments
func usage() {
	os.Stderr.WriteString(usageText)
	flag.PrintDefaults()
}

func main() {

	run_strip := flag.Bool("s", false, "run strip")
	run_upx := flag.Bool("u", true, "run upx")

	flag.Parse()

	if flag.NArg() != 1 {
		usage()
		return
	}

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
		cmd := exec.Command("strip", "-s", input_file)
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		err = cmd.Run()
		if err != nil {
			log.Panic("strip failed: ", err)
		}
	}

	if *run_upx {
		cmd := exec.Command("upx", input_file)
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		err = cmd.Run()
		if err != nil {
			log.Panic("upx failed: ", err)
		}
	}

}
