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

	log.Fatal(http.ListenAndServe(":8080", mux))
}
