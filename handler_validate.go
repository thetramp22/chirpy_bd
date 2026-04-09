package main

import (
	"encoding/json"
	"net/http"
)

func handlerValidateChirp(w http.ResponseWriter, req *http.Request) {
	type parameters struct {
		Body string `json:"body"`
	}
	type returnVals struct {
		Valid bool `json:"valid"`
	}

	decoder := json.NewDecoder(req.Body)
	params := parameters{}
	err := decoder.Decode(&params)
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

	respBody := returnVals{
		Valid: true,
	}

	respondWithJSON(w, 200, respBody)
}
