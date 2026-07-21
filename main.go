package main

import (
	"encoding/json"
	"fmt"
	"html"
	"log"
	"net/http"
	"sync"
	"time"
)

type Database struct {
	mu       sync.RWMutex
	UserInfo map[string]int
}

type User struct {
	Name string `json:"name"`
	Age  int    `json:"age"`
}

type MessageResponse struct {
	Message string `json:"message"`
}

func (d *Database) registerUserHandler(w http.ResponseWriter, r *http.Request) {
	var user User

	err := json.NewDecoder(r.Body).Decode(&user)
	if err != nil {
		http.Error(w, "Invalid JSON payload: "+err.Error(), http.StatusBadRequest)
		return
	}

	if user.Name == "" {
		http.Error(w, "Name cannot be empty", http.StatusBadRequest)
		return
	}

	if user.Age < 0 {
		http.Error(w, "Age must be a positive number", http.StatusBadRequest)
		return
	}

	d.mu.Lock()
	defer d.mu.Unlock()

	if _, exists := d.UserInfo[user.Name]; exists {
		http.Error(w, "User already exists", http.StatusConflict)
		return
	}

	d.UserInfo[user.Name] = user.Age

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(user)
}

func (d *Database) getUserHandler(w http.ResponseWriter, r *http.Request) {
	name := r.PathValue("name")

	d.mu.RLock()
	value, ok := d.UserInfo[name]
	d.mu.RUnlock()

	if !ok {
		http.Error(w, "User not found", http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(User{Name: name, Age: value})
}

func (d *Database) updateUserHandler(w http.ResponseWriter, r *http.Request) {
	name := r.PathValue("name")

	var updatedUser User
	err := json.NewDecoder(r.Body).Decode(&updatedUser)
	if err != nil {
		http.Error(w, "Invalid JSON payload: "+err.Error(), http.StatusBadRequest)
		return
	}

	if updatedUser.Age < 0 {
		http.Error(w, "Age must be a positive number", http.StatusBadRequest)
		return
	}

	d.mu.Lock()
	defer d.mu.Unlock()

	_, ok := d.UserInfo[name]
	if !ok {
		http.Error(w, "User not found", http.StatusNotFound)
		return
	}

	if updatedUser.Name != "" && updatedUser.Name != name {
		delete(d.UserInfo, name)
		d.UserInfo[updatedUser.Name] = updatedUser.Age
	} else {
		updatedUser.Name = name
		d.UserInfo[name] = updatedUser.Age
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(updatedUser)
}

func (d *Database) deleteUserHandler(w http.ResponseWriter, r *http.Request) {
	name := r.PathValue("name")

	d.mu.Lock()
	defer d.mu.Unlock()

	_, ok := d.UserInfo[name]
	if !ok {
		http.Error(w, "User not found", http.StatusNotFound)
		return
	}

	delete(d.UserInfo, name)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(MessageResponse{
		Message: fmt.Sprintf("Successfully deleted user %q", name),
	})
}

func echoHandler(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, "Hello, %q", html.EscapeString(r.PathValue("arg")))
}

func logMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		log.Printf("%s %s Kyaa~", r.Method, r.URL.Path)
		next.ServeHTTP(w, r)
	})
}

func main() {
	db := &Database{
		UserInfo: make(map[string]int),
	}
	mux := http.NewServeMux()

	mux.HandleFunc("GET /echo/{arg}", echoHandler)
	mux.HandleFunc("POST /user", db.registerUserHandler)
	mux.HandleFunc("GET /user/{name}", db.getUserHandler)
	mux.HandleFunc("DELETE /user/{name}", db.deleteUserHandler)
	mux.HandleFunc("PUT /user/{name}", db.updateUserHandler)

	s := &http.Server{
		Addr:           ":8080",
		Handler:        logMiddleware(mux),
		ReadTimeout:    10 * time.Second,
		WriteTimeout:   10 * time.Second,
		MaxHeaderBytes: 1 << 20,
	}

	log.Printf("Server listening on http://localhost:8080")
	log.Fatal(s.ListenAndServe())
}
