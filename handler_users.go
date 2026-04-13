package main

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/google/uuid"
	"github.com/thetramp22/chirpy_bd/internal/auth"
	"github.com/thetramp22/chirpy_bd/internal/database"
)

type User struct {
	ID        uuid.UUID `json:"id"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
	Email     string    `json:"email"`
}

func (cfg *apiConfig) handlerCreateUser(w http.ResponseWriter, req *http.Request) {
	type parameters struct {
		Password string `json:"password"`
		Email    string `json:"email"`
	}

	decoder := json.NewDecoder(req.Body)
	params := parameters{}
	err := decoder.Decode(&params)
	if err != nil {
		errMsg := "Error decoding parameters:"
		respondWithError(w, http.StatusInternalServerError, errMsg, err)
		return
	}

	hashedPassord, err := auth.HashPassword(params.Password)
	if err != nil {
		errMsg := "Error hashing password:"
		respondWithError(w, http.StatusInternalServerError, errMsg, err)
		return
	}

	dbUser, err := cfg.dbQueries.CreateUser(req.Context(), database.CreateUserParams{
		Email:          params.Email,
		HashedPassword: hashedPassord,
	})
	if err != nil {
		errMsg := "Error creating user:"
		respondWithError(w, http.StatusInternalServerError, errMsg, err)
		return
	}

	user := User{
		ID:        dbUser.ID,
		CreatedAt: dbUser.CreatedAt,
		UpdatedAt: dbUser.UpdatedAt,
		Email:     dbUser.Email,
	}

	respondWithJSON(w, http.StatusCreated, user)
}

func (cfg *apiConfig) handlerLogin(w http.ResponseWriter, req *http.Request) {
	type parameters struct {
		Password string `json:"password"`
		Email    string `json:"email"`
	}

	decoder := json.NewDecoder(req.Body)
	params := parameters{}
	err := decoder.Decode(&params)
	if err != nil {
		errMsg := "Error decoding parameters:"
		respondWithError(w, http.StatusInternalServerError, errMsg, err)
		return
	}

	dbUser, err := cfg.dbQueries.GetUserByEmail(req.Context(), params.Email)
	if err != nil {
		errMsg := "Error getting user:"
		respondWithError(w, http.StatusUnauthorized, errMsg, err)
		return
	}

	passwordMatch, err := auth.CheckPasswordHash(params.Password, dbUser.HashedPassword)
	if err != nil || passwordMatch == false {
		errMsg := "Password does not match:"
		respondWithError(w, http.StatusUnauthorized, errMsg, err)
		return
	}

	user := User{
		ID:        dbUser.ID,
		CreatedAt: dbUser.CreatedAt,
		UpdatedAt: dbUser.UpdatedAt,
		Email:     dbUser.Email,
	}

	respondWithJSON(w, http.StatusOK, user)
}
