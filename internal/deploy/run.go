package deploy

import (
	"context"
	"fmt"
	"os/exec"
	"time"
)

const timeout = 5 * time.Minute

func Run(script string) error {
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	cmd := exec.CommandContext(ctx, script)

	output, err := cmd.CombinedOutput()
	if ctx.Err() == context.DeadlineExceeded {
		return fmt.Errorf("deploy timed out: %s", output)
	}
	if err != nil {
		return fmt.Errorf("deploy failed: %w: %s", err, output)
	}

	return nil
}