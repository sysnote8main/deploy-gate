package deploy

import (
	"fmt"
	"os/exec"
)

func Run(script string) error {
	cmd := exec.Command(script)

	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("%w: %s", err, output)
	}

	return nil
}