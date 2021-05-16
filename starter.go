package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"

	"os"

	"github.com/gorilla/handlers"
	"github.com/gorilla/mux"
)

func main() {
	log.SetOutput(os.Stdout)
	allowedHeaders := handlers.AllowedHeaders([]string{"X-Requested-With"})
	allowedOrigins := handlers.AllowedOrigins([]string{"*"})
	allowedMethods := handlers.AllowedMethods([]string{"GET", "HEAD", "POST", "PUT", "DELETE", "OPTIONS"})

	router := makeRouter()
	log.Fatal(http.ListenAndServe(":9703", handlers.CORS(allowedHeaders, allowedOrigins, allowedMethods)(router)))
}

func makeRouter() *mux.Router {
	router := mux.NewRouter()
	router.HandleFunc("/hitec/repository/concepts/store/dataset/", postNewDataset).Methods("POST")
	router.HandleFunc("/hitec/classification/concepts/", postStartNewDetection).Methods("POST")
	return router
}

func postNewDataset() {
	// Receive new dataset
	// Process it
	// Store dataset in database
}

func postStartNewDetection() {
	// Get Dataset from Database
	// Call detection MS
	// Store results in database
}
