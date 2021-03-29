package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	"github.com/anterpin/interview/server/apiobj"
)

// schedule a process owned by the calling client
// return the process id
func start(rw http.ResponseWriter, r *http.Request) {
	userid, err := getUserId(r)
	if err != nil {
		rw.WriteHeader(http.StatusForbidden)
		_ = json.NewEncoder(rw).Encode(apiobj.Error{Err: "forbidden"})
		return
	}
	commandObj := apiobj.Command{}
	decoder := json.NewDecoder(r.Body)
	err = decoder.Decode(&commandObj)

	if err != nil {
		rw.WriteHeader(http.StatusBadRequest)
		_ = json.NewEncoder(rw).Encode(apiobj.Error{Err: err.Error()})
		return
	}

	command := strings.TrimSpace(commandObj.Command)
	id, err := _manager.Start(command, userid)
	if err != nil {
		rw.WriteHeader(http.StatusBadRequest)
		_ = json.NewEncoder(rw).Encode(apiobj.Error{Err: err.Error()})
		return
	}

	_ = json.NewEncoder(rw).Encode(apiobj.UUID{UUID: id})
}

// stop a process given a id owned by the calling client
func stop(rw http.ResponseWriter, r *http.Request) {
	userid, err := getUserId(r)
	if err != nil {
		rw.WriteHeader(http.StatusForbidden)
		_ = json.NewEncoder(rw).Encode(apiobj.Error{Err: "forbidden"})
		fmt.Fprint(rw, err.Error())
		return
	}

	idObj := apiobj.UUID{}
	decoder := json.NewDecoder(r.Body)
	err = decoder.Decode(&idObj)

	if err != nil {
		rw.WriteHeader(http.StatusBadRequest)
		_ = json.NewEncoder(rw).Encode(apiobj.Error{Err: err.Error()})
		return
	}

	id := strings.TrimSpace(idObj.UUID)
	err = _manager.Stop(id, userid)
	if err != nil {
		rw.WriteHeader(http.StatusBadRequest)
		_ = json.NewEncoder(rw).Encode(apiobj.Error{Err: err.Error()})
		return
	}

	_ = json.NewEncoder(rw).Encode(apiobj.Status{Status: "ok"})
}

// list all the processes owned by the calling client
func list(rw http.ResponseWriter, r *http.Request) {
	userid, err := getUserId(r)
	if err != nil {
		rw.WriteHeader(http.StatusForbidden)
		_ = json.NewEncoder(rw).Encode(apiobj.Error{Err: "forbidden"})
		return
	}

	strArr := _manager.List(userid)
	_ = json.NewEncoder(rw).Encode(apiobj.List{List: strArr})
}

// return the process state object of the process having that id and owned by the client
func status(rw http.ResponseWriter, r *http.Request) {
	userid, err := getUserId(r)
	if err != nil {
		rw.WriteHeader(http.StatusForbidden)
		_ = json.NewEncoder(rw).Encode(apiobj.Error{Err: "forbidden"})
		return
	}

	ids, ok := r.URL.Query()["id"]
	if !ok || len(ids) != 1 || len(ids[0]) < 1 {
		rw.WriteHeader(http.StatusBadRequest)
		_ = json.NewEncoder(rw).Encode(apiobj.Error{Err: "missing get parameter id"})
		return
	}

	id := strings.TrimSpace(ids[0])
	status, err := _manager.Status(id, userid)

	if err != nil {
		rw.WriteHeader(http.StatusBadRequest)
		_ = json.NewEncoder(rw).Encode(apiobj.Error{Err: err.Error()})
		return
	}
	_ = json.NewEncoder(rw).Encode(apiobj.State{State: status})
}

// return the output of the process given the id and owned by the client
func _log(rw http.ResponseWriter, r *http.Request) {
	userid, err := getUserId(r)
	if err != nil {
		rw.WriteHeader(http.StatusForbidden)
		_ = json.NewEncoder(rw).Encode(apiobj.Error{Err: "forbidden"})
		return
	}

	ids, ok := r.URL.Query()["id"]
	if !ok || len(ids) != 1 || len(ids[0]) < 1 {
		rw.WriteHeader(http.StatusBadRequest)
		_ = json.NewEncoder(rw).Encode(apiobj.Error{Err: "missing get parameter id"})
		return
	}

	id := strings.TrimSpace(ids[0])
	str, err := _manager.Log(id, userid)
	if err != nil {
		rw.WriteHeader(http.StatusBadRequest)
		_ = json.NewEncoder(rw).Encode(apiobj.Error{Err: err.Error()})
		return
	}
	_ = json.NewEncoder(rw).Encode(apiobj.Log{Log: str})
}
