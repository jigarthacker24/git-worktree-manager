package ide

import "os/exec"

func execLookPath(name string) (string, error) {
	return exec.LookPath(name)
}

func start(name string, args ...string) error {
	cmd := exec.Command(name, args...)
	return cmd.Start()
}
