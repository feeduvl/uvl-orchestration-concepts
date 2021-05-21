package main

import (
	"encoding/csv"
	"encoding/json"
	"fmt"
	//"io"
	"log"
	"net/http"
	"strings"

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
	log.Fatal(http.ListenAndServe(":9709", handlers.CORS(allowedHeaders, allowedOrigins, allowedMethods)(router)))
}

func makeRouter() *mux.Router {
	router := mux.NewRouter()
	router.HandleFunc("/hitec/orchestration/concepts/store/dataset/", postNewDataset).Methods("POST")
	router.HandleFunc("/hitec/orchestration/concepts/detection/", postStartNewDetection).Methods("POST")
	return router
}

func postNewDataset(w http.ResponseWriter, r *http.Request) {

	w.Header().Set("Content-Type", "application/json")

	// Receive new dataset
	err := r.ParseMultipartForm(32 << 20)
	if err != nil {
		json.NewEncoder(w).Encode(ResponseMessage{Status: true, Message: "Form data could not be retrieved"})
		w.WriteHeader(http.StatusInternalServerError)
		return
		//panic(err)
	}

	file, header, err := r.FormFile("file")
	if err != nil {
		json.NewEncoder(w).Encode(ResponseMessage{Status: true, Message: "File error"})
		w.WriteHeader(http.StatusInternalServerError)
		return
		//panic(err)
	}
	defer file.Close()

	name := strings.Split(header.Filename, ".")
	fmt.Printf("postNewDataset called. File name: %s\n", name[0])

	if name[1] != "csv" {
		json.NewEncoder(w).Encode(ResponseMessage{Status: true, Message: "Filetype not supported"})
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	// Process it
	// TODO: Remove? io.Copy(&buf, file)
	lines, err := csv.NewReader(file).ReadAll()
	if err != nil {
		json.NewEncoder(w).Encode(ResponseMessage{Status: true, Message: "Processing error"})
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	var a []Document
	for i, line := range lines {
		var d = Document{i, line[0]}
		a = append(a, d)
	}
	b, err := json.Marshal(a)
	if err != nil {
		json.NewEncoder(w).Encode(ResponseMessage{Status: true, Message: "Processing error"})
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	// Store dataset in database
	saveDataset(b)
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(ResponseMessage{Status: true, Message: "Dataset successfully uploaded"})
	return

}

func postStartNewDetection(w http.ResponseWriter, r *http.Request) {

	//params := mux.Vars(r)
	// Get Dataset from Database
	// Call detection MS
	// Store results in database

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(ResponseMessage{Status: true, Message: "Detection started"})
	return
}
