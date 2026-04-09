package main

import (
	"encoding/json"
	"fmt"
	"net/http"
)

func handlerValidateChirp(w http.ResponseWriter, req *http.Request) {
	type parameters struct {
		Body string `json:"body"`
	}

	decoder := json.NewDecoder(req.Body)
	params := parameters{}
	err := decoder.Decode(&params)
	if err != nil {
		respondWithError(w, 500, fmt.Sprint(err))
		return
	}
	if len(params.Body) > 140 {
		errMsg := "Chirp is too long"
		respondWithError(w, 400, errMsg)
		return
	}

	type returnVals struct {
		Valid bool `json:"valid"`
	}
	respBody := returnVals{
		Valid: true,
	}

	respondWithJSON(w, 200, respBody)
}
