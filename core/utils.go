package core

import "os/exec"

func ExecCommand(args ...string) ([]byte, error) {
	cmd := exec.Command(args[0], args[1:]...)
	stdout, err := cmd.Output()
	if err != nil {
		return stdout, err
	}

	return stdout, nil
}
