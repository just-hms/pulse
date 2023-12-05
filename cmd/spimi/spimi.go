package main

import (
	"bufio"
	"flag"
	"log"
	"os"
	"runtime/pprof"

	"github.com/just-hms/pulse/pkg/readers"
	"github.com/just-hms/pulse/pkg/spimi"
)

var cpuprofile = flag.Bool("profile", false, "write cpu profile to \"cpu.prof\"")

func main() {
	flag.Parse()

	args := flag.Args()

	if *cpuprofile {
		f, err := os.Create("cpu.prof")
		if err != nil {
			log.Fatal(err)
		}
		pprof.StartCPUProfile(f)
		defer pprof.StopCPUProfile()
	}

	f := os.Stdin
	if len(args) > 1 {
		var err error
		f, err = os.Open(args[1])
		if err != nil {
			panic(err)
		}
	}

	r := readers.NewMsMarco(bufio.NewReader(f), 50_000)

	spimi.Load(r, 16)
}
