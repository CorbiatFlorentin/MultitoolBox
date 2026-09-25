package runner

import (
	"os/exec"
	"strconv"
)

// killTreeOnCancel fait tuer, à l'expiration du délai, le processus et tous
// ses descendants : tuer seulement le processus lancé (ex. cmd.exe ou
// composer) laisserait tourner phpunit, pytest... en orphelins.
func killTreeOnCancel(cmd *exec.Cmd) {
	cmd.Cancel = func() error {
		return exec.Command("taskkill", "/T", "/F", "/PID", strconv.Itoa(cmd.Process.Pid)).Run()
	}
}
