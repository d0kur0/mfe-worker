package shell

import (
	"log"
	"os/exec"
	"strings"
)

type ExecShellCommandArgs struct {
	Cwd   string
	Debug bool
}

func ExecShellCommand(path string, args []string, eArgs ExecShellCommandArgs) (out string, err error) {
	cmd := exec.Command(path, args...)

	if len(eArgs.Cwd) > 0 {
		cmd.Dir = eArgs.Cwd
	}

	var b []byte
	b, err = cmd.CombinedOutput()
	out = string(b)

	if eArgs.Debug {
		log.Println(strings.Join(cmd.Args[:], " "))

		if err != nil {
			log.Printf("ExecShellCommand error: %s; %s", err, out)
		}
	}

	return
}
