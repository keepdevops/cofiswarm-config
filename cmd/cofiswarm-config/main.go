// Command cofiswarm-config is the swarm-config build tool (Go port of
// scripts/build_swarm_config.py + scripts/migrate_swarm_config.py).
//
//	cofiswarm-config build [--root DIR]
//	    Assemble swarm-config.json from config/coordinator.json + config/agents/*.json.
//	cofiswarm-config split [--source FILE] [--out-dir DIR] [--dry-run]
//	    Split a monolithic swarm-config.json into per-agent files.
package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"path/filepath"

	"github.com/keepdevops/cofiswarm-config/internal/build"
)

func main() {
	log.SetFlags(0)
	log.SetPrefix("[cofiswarm-config] ")
	if len(os.Args) < 2 {
		usage()
	}
	switch os.Args[1] {
	case "build":
		fs := flag.NewFlagSet("build", flag.ExitOnError)
		root := fs.String("root", defaultRoot(), "repo root")
		_ = fs.Parse(os.Args[2:])
		n, err := build.Build(*root)
		if err != nil {
			log.Fatal(err)
		}
		log.Printf("wrote swarm-config.json (%d agents)", n)
	case "split":
		fs := flag.NewFlagSet("split", flag.ExitOnError)
		root := defaultRoot()
		source := fs.String("source", filepath.Join(root, "swarm-config.json"), "monolithic swarm-config.json")
		outDir := fs.String("out-dir", filepath.Join(root, "config", "agents"), "per-agent output dir")
		dryRun := fs.Bool("dry-run", false, "report without writing")
		_ = fs.Parse(os.Args[2:])
		n, err := build.Split(*source, *outDir, *dryRun)
		if err != nil {
			log.Fatal(err)
		}
		log.Printf("split %d agents from %s", n, *source)
	default:
		usage()
	}
}

func usage() {
	fmt.Fprintln(os.Stderr, "usage: cofiswarm-config <build|split> [flags]")
	os.Exit(2)
}

// defaultRoot is the parent of the executable's dir's parent when run from the repo,
// falling back to the current working directory.
func defaultRoot() string {
	if wd, err := os.Getwd(); err == nil {
		return wd
	}
	return "."
}
