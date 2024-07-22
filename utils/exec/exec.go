package exec

import (
	"context"
	"os/exec"

	"github.com/google/safetext/shsprintf"
)

type Cmd struct {
	*exec.Cmd
}

func Command(name string, args ...string) *Cmd {
	return CommandContext(context.Background(), name, args...)
}

func CommandContext(ctx context.Context, name string, args ...string) *Cmd {
	cmd := exec.CommandContext(ctx, name, args...)
	_, err := shsprintf.Sprintf("%s", name)
	if err != nil {
		cmd.Err = err
	}
	for _, arg := range args {
		_, err := shsprintf.Sprintf("%s", arg)
		if err != nil {
			cmd.Err = err
		}
	}
	return &Cmd{Cmd: cmd}
}
