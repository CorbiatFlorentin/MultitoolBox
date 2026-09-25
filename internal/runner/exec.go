package runner

import (
	"context"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"time"
)

// timeout borne la durée de chaque commande lancée par run (0 = illimité).
var timeout time.Duration

// SetTimeout fixe la durée maximale de chaque commande de test ; au-delà, le
// processus est tué et run renvoie une erreur. 0 désactive la limite.
func SetTimeout(d time.Duration) { timeout = d }

// run executes name+args in dir, streaming output live.
func run(dir, name string, args ...string) error {
	ctx := context.Background()
	if timeout > 0 {
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(ctx, timeout)
		defer cancel()
	}

	cmd := exec.CommandContext(ctx, name, args...)
	cmd.Dir = dir
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Stdin = os.Stdin
	if timeout > 0 {
		killTreeOnCancel(cmd)
		cmd.WaitDelay = time.Second
	}

	err := cmd.Run()
	if errors.Is(ctx.Err(), context.DeadlineExceeded) {
		return fmt.Errorf("%s: délai dépassé (%s)", name, timeout)
	}
	return err
}

func fileExists(dir, name string) bool {
	_, err := os.Stat(dir + string(os.PathSeparator) + name)
	return err == nil
}
