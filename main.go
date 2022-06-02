package main

import (
	"log"
	"main/build"
	"main/serve"
	"os"
)

func main() {
	switch os.Args[1] {
	case "build":
		if err := build.All(); err != nil {
			panic(err)
		}
	case "serve":
		if err := serve.Start(); err != nil {
			panic(err)
		}
	default:
		log.Fatalf("unrecognized command %s", os.Args[1])
	}
}
