package apiobj

import (
	"os"
)

// wrap an error string
// used in all endpoints
type Error struct {
	Err string `json:"error"`
}

// wrap the command to execute
// used in the /start endpoint
type Command struct {
	Command string `json:"command"`
}

// wrap the uuid of the process
// used in /start /stop /log /status endpoints
type UUID struct {
	UUID string `json:"uuid"`
}

// wrap the string ok
// used in the /stop endpoint
type Status struct {
	Status string `json:"status"`
}

// wrap a list of strings
// used in the /list endpoint
type List struct {
	List []string `json:"list"`
}

// wrap the process output
// used in the /log endpoint
type Log struct {
	Log string `json:"log"`
}

// wrap the os.ProcessState object
// used in the /status endpoint
type State struct {
	State *os.ProcessState `json:"status"`
}
