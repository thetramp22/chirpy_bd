package main

import (
	"encoding/json"
	"net/http"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/thetramp22/chirpy_bd/internal/auth"
	"github.com/thetramp22/chirpy_bd/internal/database"
)

type Chirp struct {
	ID        uuid.UUID `json:"id"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
	Body      string    `json:"body"`
	UserID    uuid.UUID `json:"user_id"`
}

func (cfg *apiConfig) handlerGetChirps(w http.ResponseWriter, req *http.Request) {
	dbChirps, err := cfg.dbQueries.GetAllChirps(req.Context())
	if err != nil {
		errMsg := "Error getting chirps:"
		respondWithError(w, http.StatusInternalServerError, errMsg, err)
		return
	}

	var chirps []Chirp

	for _, chirp := range dbChirps {
		chirps = append(chirps, Chirp{
			ID:        chirp.ID,
			CreatedAt: chirp.CreatedAt,
			UpdatedAt: chirp.UpdatedAt,
			Body:      chirp.Body,
			UserID:    chirp.UserID,
		})
	}

	respondWithJSON(w, http.StatusOK, chirps)
}

func (cfg *apiConfig) handlerGetChirpByID(w http.ResponseWriter, req *http.Request) {
	chirpID, err := uuid.Parse(req.PathValue("chirpID"))
	if err != nil {
		errMsg := "Error parsing uuid:"
		respondWithError(w, http.StatusInternalServerError, errMsg, err)
		return
	}
	dbChirp, err := cfg.dbQueries.GetChirpByID(req.Context(), chirpID)
	if err != nil {
		errMsg := "Error getting chirp:"
		respondWithError(w, http.StatusNotFound, errMsg, err)
		return
	}

	chirp := Chirp{
		ID:        dbChirp.ID,
		CreatedAt: dbChirp.CreatedAt,
		UpdatedAt: dbChirp.UpdatedAt,
		Body:      dbChirp.Body,
		UserID:    dbChirp.UserID,
	}

	respondWithJSON(w, http.StatusOK, chirp)

}

func (cfg *apiConfig) handlerCreateChirp(w http.ResponseWriter, req *http.Request) {
	type parameters struct {
		Body string `json:"body"`
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
	if len(params.Body) > 140 {
		errMsg := "Chirp is too long"
		respondWithError(w, http.StatusBadRequest, errMsg, err)
		return
	}

	cleanedBody := replaceBadWords(params.Body)

	dbChirp, err := cfg.dbQueries.CreateChirp(req.Context(), database.CreateChirpParams{
		Body:   cleanedBody,
		UserID: userID,
	})
	if err != nil {
		errMsg := "Error creating chirp:"
		respondWithError(w, http.StatusInternalServerError, errMsg, err)
		return
	}

	chirp := Chirp{
		ID:        dbChirp.ID,
		CreatedAt: dbChirp.CreatedAt,
		UpdatedAt: dbChirp.UpdatedAt,
		Body:      dbChirp.Body,
		UserID:    dbChirp.UserID,
	}

	respondWithJSON(w, http.StatusCreated, chirp)
}

func replaceBadWords(msg string) string {
	cleanMsgSlice := []string{}
	for _, word := range strings.Split(msg, " ") {
		if strings.ToLower(word) == "kerfuffle" || strings.ToLower(word) == "sharbert" || strings.ToLower(word) == "fornax" {
			word = "****"
		}
		cleanMsgSlice = append(cleanMsgSlice, word)
	}
	cleanMsg := strings.Join(cleanMsgSlice, " ")
	return cleanMsg
}
