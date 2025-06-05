package oscal

import (
  "tool/exec"
)

command: {
	// List available commands.
	help: exec.Run & {
		cmd: "cue help cmd"
	}
}