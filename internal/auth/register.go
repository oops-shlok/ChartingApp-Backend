package auth

import (
	"context"
	"encoding/json"
	"log"
	"net/http"
	"golang.org/x/crypto/bcrypt"
	"ChartingApp-Backend/internal/database"
)

type User struct {
	Username string `json:"username"`
	Password string `json:"password"`
	Name     string `json:"name"`
	Email    string `json:"email"`
	Phone    string `json:"phone"`
}

type RegistrationResponse struct {
	Message string `json:"message"`
	User    User   `json:"user"`
}

func CreateUserHandler(w http.ResponseWriter, r *http.Request) {
	var user User
	if err := json.NewDecoder(r.Body).Decode(&user); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	if user.Username == "" || user.Password == "" || user.Name == "" || user.Email == "" || user.Phone == "" {
		http.Error(w, "All fields are required", http.StatusBadRequest)
		return
	}

	err := CreateUser(&user)
	if err != nil {
		http.Error(w, "Failed to create user", http.StatusInternalServerError)
		return
	}

	response := RegistrationResponse{
		Message: "User created successfully",
		User:    user,
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(response)
}

func CreateUser(user *User) error {
	collection := database.GetDB().Database("ChartKraft").Collection("Users")

	hashedPassword, err := hashPassword(user.Password)
	if err != nil {
		log.Println("Error hashing password:", err)
		return err
	}

	user.Password = hashedPassword

	_, err = collection.InsertOne(context.TODO(), user)
	if err != nil {
		log.Printf("Error creating user '%s': %v\n", user.Username, err)
		return err
	}

	return nil
}

func hashPassword(password string) (string, error) {
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return "", err
	}
	return string(hashedPassword), nil
}
