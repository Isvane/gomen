package main

import (
	"fmt"
	"html"
	"log"
	"net/http"
	"strconv"
	"sync"
)

type Database struct {
	mu       sync.RWMutex
	UserInfo map[string]int
}

func (d *Database) registerUserHandler(w http.ResponseWriter, r *http.Request) {
	name := r.PathValue("name")
	ageStr := r.PathValue("age")

	age, err := strconv.Atoi(ageStr)
	if err != nil {
		http.Error(w, "Age must be a valid number", http.StatusBadRequest)
		return
	}

	d.mu.Lock()
	d.UserInfo[name] = age
	defer d.mu.Unlock()

	log.Printf("Successfully registered %s with age %d", name, age)
}

func (d *Database) getUserHandler(w http.ResponseWriter, r *http.Request) {
	name := r.PathValue("name")

	d.mu.RLock()
	defer d.mu.RUnlock()

	value, ok := d.UserInfo[name]
	if ok {
		fmt.Fprintf(w, "Found %q! Age: %d", name, value)
	} else {
		http.Error(w, "User not found", http.StatusNotFound)
		return
	}
}

func (d *Database) deleteUserHandler(w http.ResponseWriter, r *http.Request) {
	name := r.PathValue("name")

	d.mu.Lock()
	defer d.mu.Unlock()

	_, ok := d.UserInfo[name]
	if ok {
		delete(d.UserInfo, name)
		fmt.Fprintf(w, "Deleted %q", name)
	} else {
		http.Error(w, "User not found", http.StatusNotFound)
		return
	}
}

func echoHandler(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, "Hello, %q", html.EscapeString(r.PathValue("arg")))
}

func main() {
	db := &Database{
		UserInfo: make(map[string]int),
	}
	mux := http.NewServeMux()

	mux.HandleFunc("GET /echo/{arg}", echoHandler)
	mux.HandleFunc("POST /register/{name}/{age}", db.registerUserHandler)
	mux.HandleFunc("GET /user/{name}", db.getUserHandler)
	mux.HandleFunc("DELETE /user/{name}", db.deleteUserHandler)

	log.Fatal(http.ListenAndServe(":8080", mux))
}
