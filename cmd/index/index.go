package main

import (
	"bufio"
	"flag"
	"log"
	"os"
	"runtime/pprof"

	"github.com/just-hms/pulse/pkg/index"
	"github.com/just-hms/pulse/pkg/readers"
)

var cpuprofile = flag.Bool("cpuprofile", false, "write cpu profile to file")

func main() {
	flag.Parse()
	args := flag.Args()

	if *cpuprofile {
		f, err := os.Create("profile.out")
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

	r := readers.NewMsMarco(bufio.NewReader(f), 100_000)
	index.Load(r)

}
