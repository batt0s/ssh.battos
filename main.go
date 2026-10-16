package main

import (
	"flag"
	"log"

	"github.com/batt0s/ssh.battos/tui"
)

func main() {
	local := flag.Bool("local", false, "run the TUI locally instead of the SSH server")
	flag.Parse()

	if *local {
		if err := tui.StartTea(); err != nil {
			log.Fatal(err)
		}
		return
	}

	tui.SSSHMain()
}
