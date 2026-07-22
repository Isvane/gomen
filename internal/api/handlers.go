package api

import (
	"encoding/json"
	"fmt"
	"html"
	"net/http"

	models "github.com/Isvane/gohan/internal/model"
	"github.com/Isvane/gohan/internal/repository"
)

type Database struct {
	Repo *repository.UserRepo
}

func (d *Database) RegisterUserHandler(w http.ResponseWriter, r *http.Request) {
	var user models.User

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

	err = d.Repo.Create(user)
	if err == repository.ErrUserAlreadyExists {
		http.Error(w, "User already exists", http.StatusConflict)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(user)
}

func (d *Database) GetUserHandler(w http.ResponseWriter, r *http.Request) {
	name := r.PathValue("name")

	user, err := d.Repo.Get(name)
	if err == repository.ErrUserNotFound {
		http.Error(w, "User not found", http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(user)
}

func (d *Database) UpdateUserHandler(w http.ResponseWriter, r *http.Request) {
	name := r.PathValue("name")

	var updatedUser models.User
	err := json.NewDecoder(r.Body).Decode(&updatedUser)
	if err != nil {
		http.Error(w, "Invalid JSON payload: "+err.Error(), http.StatusBadRequest)
		return
	}

	if updatedUser.Age < 0 {
		http.Error(w, "Age must be a positive number", http.StatusBadRequest)
		return
	}

	res, err := d.Repo.Update(name, updatedUser)
	if err == repository.ErrUserNotFound {
		http.Error(w, "User not found", http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(res)
}

func (d *Database) DeleteUserHandler(w http.ResponseWriter, r *http.Request) {
	name := r.PathValue("name")

	err := d.Repo.Delete(name)
	if err == repository.ErrUserNotFound {
		http.Error(w, "User not found", http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(models.MessageResponse{
		Message: fmt.Sprintf("Successfully deleted user %q", name),
	})
}

func EchoHandler(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, "Hello, %q", html.EscapeString(r.PathValue("arg")))
}
