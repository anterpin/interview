package manager

import (
	"errors"
	"fmt"
	"log"
	"os"
	"strings"
	"sync"

	uuid "github.com/satori/go.uuid"
)

type Manager struct {
	// map[userid] user process hashmap
	userProcesses map[int]*UserProcesses
	// used only to resize the userProcesse hashmap
	mutex sync.Mutex
}

func NewManager() Manager {
	return Manager{
		userProcesses: make(map[int]*UserProcesses),
	}
}

func (this *Manager) AddUser(userid int) {
	_, exists := this.userProcesses[userid]
	if exists {
		log.Printf("User id %d already exist", userid)
		return
	}

	this.mutex.Lock()
	defer this.mutex.Unlock()

	userProcessesPtr := new(UserProcesses)
	userProcessesPtr.processes = make(map[uuid.UUID]Process)
	this.userProcesses[userid] = userProcessesPtr
}

type UserProcesses struct {
	processes map[uuid.UUID]Process
	mutex     sync.Mutex
}

func (this *Manager) getUserProcesses(userid int) (*UserProcesses, bool) {

	userProcesses, exists := this.userProcesses[userid]
	if !exists {
		// if should not exist
		// it prints to warn the server
		// but in any case it will return the create a new user process hashmap
		log.Printf("Unknown user id %d in process hashmap", userid)
		this.AddUser(userid)
		userProcesses = this.userProcesses[userid]
	}
	return userProcesses, exists
}

func (this *Manager) getUserProcess(processId string, userid int, callback func(Process) (interface{}, error)) (interface{}, error) {
	id, err := uuid.FromString(processId)
	if err != nil {
		// not a valid v4 id, in this case it accepts every type of id
		return nil, err
	}
	userProcesses, _ := this.getUserProcesses(userid)

	userProcesses.mutex.Lock()
	defer userProcesses.mutex.Unlock()

	process, exists := userProcesses.processes[id]
	// program with this id does not exist
	if !exists {
		return nil, fmt.Errorf("Do not exist process id %s", processId)
	}
	return callback(process)
}

func (this *Manager) Start(command string, userid int) (string, error) {
	args := strings.Fields(command)
	// empty command
	if len(args) == 0 {
		return "", errors.New("Empty Command")
	}

	// generate the uuid
	processid := uuid.NewV4()
	process, err := Create(args[0], args[1:]...)
	if err != nil {
		return "", err
	}

	userProcesses, _ := this.getUserProcesses(userid)

	userProcesses.mutex.Lock()
	defer userProcesses.mutex.Unlock()

	userProcesses.processes[processid] = process

	return processid.String(), nil
}

func (this *Manager) Status(processId string, userid int) (*os.ProcessState, error) {
	result, err := this.getUserProcess(processId, userid, func(process Process) (interface{}, error) {
		return process.Status(), nil
	})
	if err != nil {
		return nil, err
	}

	if result == nil {
		return nil, nil
	}
	return result.(*os.ProcessState), nil
}

func (this *Manager) Stop(processId string, userid int) error {
	_, err := this.getUserProcess(processId, userid, func(process Process) (interface{}, error) {
		err := process.Kill()
		if err != nil {
			return nil, err
		}
		return nil, nil
	})

	return err
}

func (this *Manager) Log(processId string, userid int) (string, error) {
	result, err := this.getUserProcess(processId, userid, func(process Process) (interface{}, error) {
		return process.Log(), nil
	})

	if err != nil {
		return "", err
	}
	// cast to the type return value of process.Log
	return result.(string), nil
}

func (this *Manager) List(userid int) []string {
	userProcesses, _ := this.getUserProcesses(userid)

	userProcesses.mutex.Lock()
	defer userProcesses.mutex.Unlock()

	arr := make([]string, len(userProcesses.processes))
	i := 0
	for id, process := range userProcesses.processes {

		state := process.Status()
		// list fast status preview
		str := "ACTIVE"
		if state != nil {
			str = "TERMINATED"
		}

		arr[i] = fmt.Sprintf("%s %s %s", id.String(), process.name, str)
		i++
	}
	return arr
}
