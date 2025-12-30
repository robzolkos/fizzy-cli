package main

import (
	"github.com/robzolkos/fizzy-cli/internal/commands"
	"runtime/debug"
)

var version = "dev"

func main() {
	if version == "dev" {
		if info, ok := debug.ReadBuildInfo(); ok && info.Main.Version != "" && info.Main.Version != "(devel)" {
			version = info.Main.Version
		}
	}

	commands.SetVersion(version)
	commands.Execute()
}
