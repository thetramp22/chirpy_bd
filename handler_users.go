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
	ID           uuid.UUID `json:"id"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
	Email        string    `json:"email"`
	Token        string    `json:"token"`
	RefreshToken string    `json:"refresh_token"`
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

	ExpiresIn := time.Duration(1) * time.Hour

	accessToken, err := auth.MakeJWT(dbUser.ID, cfg.jwtSecret, ExpiresIn)
	if err != nil {
		errMsg := "Error getting token:"
		respondWithError(w, http.StatusInternalServerError, errMsg, err)
		return
	}

	refreshToken, err := cfg.dbQueries.CreateRefreshToken(req.Context(), database.CreateRefreshTokenParams{
		Token:     auth.MakeRefreshToken(),
		UserID:    dbUser.ID,
		ExpiresAt: time.Now().UTC().AddDate(0, 0, 60),
	})
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Couldn't save refresh token", err)
		return
	}

	user := User{
		ID:           dbUser.ID,
		CreatedAt:    dbUser.CreatedAt,
		UpdatedAt:    dbUser.UpdatedAt,
		Email:        dbUser.Email,
		Token:        accessToken,
		RefreshToken: refreshToken.Token,
	}

	respondWithJSON(w, http.StatusOK, user)
}

func (cfg *apiConfig) handlerRefresh(w http.ResponseWriter, req *http.Request) {
	refreshToken, err := auth.GetBearerToken(req.Header)
	if err != nil {
		errMsg := "Error getting token:"
		respondWithError(w, http.StatusUnauthorized, errMsg, err)
		return
	}

	dbToken, err := cfg.dbQueries.GetRefreshToken(req.Context(), refreshToken)
	if err != nil ||
		dbToken.Token != refreshToken ||
		time.Now().UTC().After(dbToken.ExpiresAt) ||
		(dbToken.RevokedAt.Valid == true && time.Now().UTC().After(dbToken.RevokedAt.Time)) {
		errMsg := "Invalid token:"
		respondWithError(w, http.StatusUnauthorized, errMsg, err)
		return
	}

	user, err := cfg.dbQueries.GetUserFromRefreshToken(req.Context(), refreshToken)
	if err != nil {
		errMsg := "Error getting user:"
		respondWithError(w, http.StatusUnauthorized, errMsg, err)
		return
	}

	ExpiresIn := time.Duration(1) * time.Hour
	accessToken, err := auth.MakeJWT(user.ID, cfg.jwtSecret, ExpiresIn)
	if err != nil {
		errMsg := "Error getting token:"
		respondWithError(w, http.StatusInternalServerError, errMsg, err)
		return
	}

	payload := struct {
		Token string `json:"token"`
	}{
		Token: accessToken,
	}

	respondWithJSON(w, http.StatusOK, payload)
}

func (cfg *apiConfig) handlerRevoke(w http.ResponseWriter, req *http.Request) {
	refreshToken, err := auth.GetBearerToken(req.Header)
	if err != nil {
		errMsg := "Error getting token:"
		respondWithError(w, http.StatusUnauthorized, errMsg, err)
		return
	}

	dbToken, err := cfg.dbQueries.GetRefreshToken(req.Context(), refreshToken)
	if err != nil ||
		dbToken.Token != refreshToken ||
		time.Now().UTC().After(dbToken.ExpiresAt) ||
		(dbToken.RevokedAt.Valid == true && time.Now().UTC().After(dbToken.RevokedAt.Time)) {
		errMsg := "Invalid token:"
		respondWithError(w, http.StatusUnauthorized, errMsg, err)
		return
	}

	cfg.dbQueries.RevokeRefreshToken(req.Context(), refreshToken)

	respondWithJSON(w, http.StatusNoContent, nil)
}

func (cfg *apiConfig) handlerUpdateCredentials(w http.ResponseWriter, req *http.Request) {
	type parameters struct {
		Password string `json:"password"`
		Email    string `json:"email"`
	}

	token, err := auth.GetBearerToken(req.Header)
	if err != nil {
		errMsg := "Error getting token:"
		respondWithError(w, http.StatusUnauthorized, errMsg, err)
		return
	}
	userID, err := auth.ValidateJWT(token, cfg.jwtSecret)
	if err != nil {
		errMsg := "Error validating user token:"
		respondWithError(w, http.StatusUnauthorized, errMsg, err)
		return
	}

	decoder := json.NewDecoder(req.Body)
	params := parameters{}
	err = decoder.Decode(&params)
	if err != nil {
		errMsg := "Error decoding parameters:"
		respondWithError(w, http.StatusInternalServerError, errMsg, err)
		return
	}

	if params.Email == "" || params.Password == "" {
		errMsg := "Invalid email or password:"
		respondWithError(w, http.StatusUnauthorized, errMsg, err)
		return
	}

	hashedPassword, err := auth.HashPassword(params.Password)

	cfg.dbQueries.UpdateUser(req.Context(), database.UpdateUserParams{
		Email:          params.Email,
		HashedPassword: hashedPassword,
		ID:             userID,
	})

	respondWithJSON(w, http.StatusOK, nil)
}
