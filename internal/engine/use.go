package engine

import (
	"errors"
	"fmt"
	"os"
	"os/exec"
	"os/signal"

	"github.com/nchandur/ranga-tools/internal/registry"
)

// launches registered engine with terminal
func Use(name string) error {
	store, err := registry.Default()
	if err != nil {
		return err
	}
	entry, err := store.Get(name)
	if err != nil {
		if errors.Is(err, registry.ErrNotFound) {
			return fmt.Errorf("no engine named %q is registered", name)
		}
		return err
	}

	cmd := exec.Command(entry.Path)
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, os.Interrupt)
	defer signal.Stop(sigs)

	if err := cmd.Run(); err != nil {
		var exitErr *exec.ExitError
		if errors.As(err, &exitErr) {
			return fmt.Errorf("engine %q exited with code %d", entry.Name, exitErr.ExitCode())
		}
		return fmt.Errorf("run engine %q: %w", entry.Name, err)
	}
	return nil
}
