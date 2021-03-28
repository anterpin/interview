package manager

import (
	"bytes"
	"os"
	"os/exec"
)

type Process struct {
	cmd    *exec.Cmd
	buffer *bytes.Buffer
	name   string
}

// try to create a process given args[0] as command
// and []args as second parameter
func Create(command string, args ...string) (Process, error) {
	process := Process{
		cmd:    exec.Command(command, args...),
		buffer: new(bytes.Buffer),
		name:   command,
	}
	// redirect stdin and stderr
	process.cmd.Stdout = process.buffer
	process.cmd.Stderr = process.buffer
	err := process.cmd.Start()
	if err == nil {
		go func() {
			_ = process.cmd.Wait()
		}()
	}
	return process, err
}

// kill the given process
// sending a sigkill signal
func (process Process) Kill() error {
	return process.cmd.Process.Kill()
}

// retrieve the state of the gven process
// nil if the process is still active
func (process Process) Status() *os.ProcessState {
	return process.cmd.ProcessState
}

// retrive the combined stdout and stderr of the given process
// TODO cast output buffer into a file to avoid increasing RAM usage
// TODO create a stream accepting the request context
func (process Process) Log() string {
	return process.buffer.String()
}
