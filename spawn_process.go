package main

import (
	"github.com/kardianos/osext"
	"os/exec"
)

func SpawnProcess(args ...string) error {
	a := make([]string, 0, len(args)+3)
	if Option.Debug {
		a = append(a, "--debug")
	}
	if Option.SettingsFile != "" {
		a = append(a, "--settings", Option.SettingsFile)
	}
	a = append(a, args...)

	file, err := osext.Executable()
	if err != nil {
		return err
	}
	return exec.Command(file, a...).Start()
}
