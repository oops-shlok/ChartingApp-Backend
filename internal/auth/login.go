package auth

import (
	"context"
	"encoding/json"
	"fmt"
	"time"
	"log"
	"net/http"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"github.com/dgrijalva/jwt-go"
	"golang.org/x/crypto/bcrypt"
	"ChartingApp-Backend/internal/database"
)

type AuthRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

type AuthResponse struct {
	Token string `json:"token"`
	User  User   `json:"user"`
}

var ErrUserNotFound = fmt.Errorf("user not found")

func NewAuthService(secretKey string) func(token string) (User, error) {
	return func(tokenString string) (User, error) {
		token, err := jwt.ParseWithClaims(tokenString, &jwt.StandardClaims{}, func(token *jwt.Token) (interface{}, error) {
			return []byte(secretKey), nil
		})

		if err != nil {
			return User{}, fmt.Errorf("failed to parse JWT token: %v", err)
		}

		if !token.Valid {
			return User{}, fmt.Errorf("invalid JWT token")
		}

		claims, ok := token.Claims.(*jwt.StandardClaims)
		if !ok {
			return User{}, fmt.Errorf("failed to extract claims from JWT token")
		}

		user, err := GetUserByUsername(claims.Subject)
		if err != nil {
			return User{}, fmt.Errorf("failed to fetch user details: %v", err)
		}

		return user, nil
	}
}

func AuthenticateHandler(w http.ResponseWriter, r *http.Request) {
	var authRequest AuthRequest
	if err := json.NewDecoder(r.Body).Decode(&authRequest); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	if authRequest.Username == "" || authRequest.Password == "" {
		http.Error(w, "Username and password are required", http.StatusBadRequest)
		return
	}

	user, err := AuthenticateUser(authRequest)
	if err != nil {
		http.Error(w, "Authentication failed", http.StatusUnauthorized)
		return
	}

	token, err := generateJWTToken(user)
	if err != nil {
		http.Error(w, "Failed to generate JWT token", http.StatusInternalServerError)
		return
	}

	response := AuthResponse{
		Token: token,
		User:  user,
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(response)
}

func AuthenticateUser(authRequest AuthRequest) (User, error) {
	user, err := GetUserByUsername(authRequest.Username)
	if err != nil {
		return User{}, err
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(authRequest.Password)); err != nil {
		return User{}, err
	}

	return user, nil
}

func GetUserByUsername(username string) (User, error) {
	collection := database.GetDB().Database("ChartKraft").Collection("Users")

	var user User
	err := collection.FindOne(context.TODO(), bson.M{"username": username}).Decode(&user)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return User{}, ErrUserNotFound
		}
		log.Printf("Error finding user by username '%s': %v\n", username, err)
		return User{}, err
	}

	return user, nil
}

func generateJWTToken(user User) (string, error) {
	secretKey := "YMinzYNPlPw8dnljfXhSMsVwdVgpnXfI"

	expirationTime := time.Now().Add(7 * 24 * time.Hour)
	claims := &jwt.StandardClaims{
		ExpiresAt: expirationTime.Unix(),
		IssuedAt:  time.Now().Unix(),
		Subject:   user.Username,
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString([]byte(secretKey))
	if err != nil {
		log.Println("Error signing JWT token:", err)
		return "", err
	}

	return tokenString, nil
}
