package main

import (
	"encoding/json"
	"net/http"
	"strings"
)

func handlerValidateChirp(w http.ResponseWriter, req *http.Request) {
	type parameters struct {
		Body string `json:"body"`
	}
	type returnVals struct {
		CleanedBody string `json:"cleaned_body"`
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
	cleanedBody := replaceBadWords(params.Body)

	respBody := returnVals{
		CleanedBody: cleanedBody,
	}

	respondWithJSON(w, http.StatusOK, respBody)
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
