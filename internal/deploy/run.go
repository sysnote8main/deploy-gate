package deploy

import (
	"context"
	"fmt"
	"os/exec"
	"time"
)

const timeout = 5 * time.Minute

func Run(script string) (string, error) {
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	cmd := exec.CommandContext(ctx, script)

	output, err := cmd.CombinedOutput()
	out := string(output)
	if ctx.Err() == context.DeadlineExceeded {
		return out, fmt.Errorf("deploy timed out")
	}
	if err != nil {
		return out, fmt.Errorf("deploy failed: %w", err)
	}

	return out, nil
}