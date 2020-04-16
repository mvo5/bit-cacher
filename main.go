// -*- Mode: Go; indent-tabs-mode: t -*-

/*
 * Copyright (C) 2020 Canonical Ltd
 *
 * This program is free software: you can redistribute it and/or modify
 * it under the terms of the GNU General Public License version 3 as
 * published by the Free Software Foundation.
 *
 * This program is distributed in the hope that it will be useful,
 * but WITHOUT ANY WARRANTY; without even the implied warranty of
 * MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
 * GNU General Public License for more details.
 *
 * You should have received a copy of the GNU General Public License
 * along with this program.  If not, see <http://www.gnu.org/licenses/>.
 *
 */
package main

import (
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net"
	"net/http"

	"github.com/gorilla/mux"
)

// XXX: add time.Now() to each cache key so that we can expire it
type Cache map[string]bool

type StateChange struct {
	Action string `json:",omitempty"`
	Name   string `json:",omitempty"`
}

var currentCache Cache = make(Cache)

/* show via:
 $ curl -f http://localhost:8080/api/1/get/$name
{} or 404
*/
func getName(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	_, ok := currentCache[vars["name"]]
	if !ok {
		w.WriteHeader(http.StatusNotFound)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	fmt.Println(w, "{}")
}

func readBody(w http.ResponseWriter, r *http.Request) ([]byte, error) {
	// limit to 1024 byte names
	body, err := ioutil.ReadAll(io.LimitReader(r.Body, 1024))
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return nil, err
	}
	if err := r.Body.Close(); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return nil, err
	}
	return body, nil
}

/*
 createName via:
 $ curl -i -H "Content-Type: application/json" -X POST -d '{"action": "create", "name": "some-name"}' http://localhost:8080/api/1/change
*/
func stateChange(w http.ResponseWriter, r *http.Request) {
	var stateChange StateChange
	body, err := readBody(w, r)
	if err != nil {
		log.Printf("readBody failed: %v", err)
		return
	}
	w.Header().Set("Content-Type", "application/json; charset=UTF-8")
	if err := json.Unmarshal(body, &stateChange); err != nil {
		log.Printf("body %q resulted in %v", body, err)
		w.WriteHeader(422) // unprocessable entity
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	if stateChange.Action != "create" {
		log.Printf("unknown action %q", stateChange.Action)
		w.WriteHeader(422) // unprocessable entity
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
	currentCache[stateChange.Name] = true
	w.WriteHeader(http.StatusCreated)
}

func makeRouter() *mux.Router {
	r := mux.NewRouter()
	r.HandleFunc("/api/1/get/{name}", getName).Methods("GET")
	r.HandleFunc("/api/1/change", stateChange).Methods("POST")

	return r
}

func main() {
	listen := ":8080"

	r := makeRouter()
	http.Handle("/", r)
	listener, err := net.Listen("tcp", listen)
	if err != nil {
		log.Fatal(err)
	}
	log.Fatal(http.Serve(listener, nil))
}
