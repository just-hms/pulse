package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"runtime/pprof"

	"github.com/just-hms/pulse/pkg/engine"
)

var cpuprofile = flag.Bool("profile", false, "write cpu profile to \"data/cpu.prof\"")

func main() {
	flag.Parse()

	args := flag.Args()

	if *cpuprofile {
		f, err := os.Create("data/cpu.prof")
		if err != nil {
			log.Fatal(err)
		}

		pprof.StartCPUProfile(f)
		defer pprof.StopCPUProfile()
	}

	if len(args) != 1 {
		panic("give me the query")
	}

	q := args[0]

	res, err := engine.Search(q, "data/dump", 10)
	if err != nil {
		panic(err)
	}

	fmt.Println(res)
}
