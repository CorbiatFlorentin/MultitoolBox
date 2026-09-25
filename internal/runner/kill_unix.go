//go:build !windows

package runner

import (
	"os/exec"
	"syscall"
)

// killTreeOnCancel lance la commande dans son propre groupe de processus et,
// à l'expiration du délai, tue tout le groupe : tuer seulement le processus
// lancé (ex. sh ou composer) laisserait tourner phpunit, pytest... en
// orphelins.
func killTreeOnCancel(cmd *exec.Cmd) {
	cmd.SysProcAttr = &syscall.SysProcAttr{Setpgid: true}
	cmd.Cancel = func() error {
		return syscall.Kill(-cmd.Process.Pid, syscall.SIGKILL)
	}
}
