package oscal

import (
	"encoding/json"
	"tool/exec"
	"tool/file"
	"strings"
	oscalv1 "github.com/orvis98/daggerverse/oscal/v1alpha1"
)

_helpers: {
	getDate: exec.Run & {
		cmd:    #"date -u +"%Y-%m-%dT%H:%M:%SZ""#
		stdout: string
	}
	metadata: exec.Run & {
		stdin: json.Marshal({
			metadata: {
				version:         "v0.1.0"
				"last-modified": strings.Trim(getDate.stdout, "\n\"")
				"oscal-version": oscalv1.#OSCAL.version
			}
		})
		cmd:    "cue eval -p oscal -"
		stdout: string
	}
}

command: {
	// Generate the metadata.cue file.
	metadata: file.Create & {
		filename: "metadata.cue"
		contents: _helpers.metadata.stdout
	}
	// Export the component definition to stdout.
	export: exec.Run & {
		$after: [metadata]
		cmd: "cue export ."
	}
}
