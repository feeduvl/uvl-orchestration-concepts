package main

import (
	"encoding/csv"
	"encoding/json"
	"fmt"
	"time"

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

	fmt.Println("uvl-orchestration-concepts MS running")
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
		panic(err)
	}

	file, header, err := r.FormFile("file")
	if err != nil {
		json.NewEncoder(w).Encode(ResponseMessage{Status: true, Message: "File error"})
		w.WriteHeader(http.StatusInternalServerError)
		panic(err)
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
	reader := csv.NewReader(file)
	reader.Comma = '|'
	reader.LazyQuotes = true
	lines, err := reader.ReadAll()
	if err != nil {
		json.NewEncoder(w).Encode(ResponseMessage{Status: true, Message: "Processing error"})
		w.WriteHeader(http.StatusInternalServerError)
		panic(err)
	}
	var a []Document
	for i, line := range lines {
		var d = Document{i, line[0]}
		a = append(a, d)
	}
	d := Dataset{Name: header.Filename, Size: len(a), Documents: a, UploadedAt: time.Now()}

	// Store dataset in database
	err = RESTPostStoreDataset(d)
	if err != nil {
		json.NewEncoder(w).Encode(ResponseMessage{Status: true, Message: "Error saving dataset"})
		w.WriteHeader(http.StatusInternalServerError)
		panic(err)
	}

	json.NewEncoder(w).Encode(ResponseMessage{Status: true, Message: "Dataset successfully uploaded"})
	w.WriteHeader(http.StatusOK)
	return

}

func postStartNewDetection(w http.ResponseWriter, r *http.Request) {

	var body map[string]interface{}
	err := json.NewDecoder(r.Body).Decode(&body)
	fmt.Printf("postStartNewDetection called. Request Body: %v\n", r.Body)
	fmt.Printf("postStartNewDetection called. Parsed Body: %v\n", body)
	fmt.Printf("postStartNewDetection called. Error decoding body: %s\n", err)

	datasetName := body["dataset"].(string)
	method := body["method"].(string)
	fmt.Printf("postStartNewDetection called. Method: %v, Dataset: %v\n", method, datasetName)

	// Get Dataset from Database
	dataset, err := RESTGetDataset(datasetName)
	if err != nil {
		fmt.Printf("ERROR retrieving dataset (postStartNewDetection) %s\n", err)
		w.WriteHeader(http.StatusBadRequest)
		panic(err)
	}

	var params = make(map[string]string)
	for key, value := range body {
		s := fmt.Sprintf("%v", value)
		params[key] = s
	}

	delete(params, "method")
	delete(params, "dataset")

	fmt.Printf("postStartNewDetection Params: %v\n", params)

	result := new(Result)
	result.Method = method
	result.DatasetName = dataset.Name
	result.Status = "scheduled"
	result.StartedAt = time.Now()
	result.Params = params

	run := new(Run)
	run.Method = method
	run.Params = params
	run.Dataset = dataset

	fmt.Printf("postStartNewDetection, calling MS and waiting for response\n")
	// Call detection MS
	endResult, err := RESTPostStartNewDetection(*result, *run)
	if err != nil {
		fmt.Printf("ERROR starting new detection %s\n", err)
		w.WriteHeader(http.StatusBadRequest)
		panic(err)
	}
	// Store results in database
	fmt.Printf("Response received, Topcis: %s\n", endResult.Topics)
	err = RESTPostStoreResult(endResult)
	if err != nil {
		json.NewEncoder(w).Encode(ResponseMessage{Status: true, Message: "Error saving result"})
		w.WriteHeader(http.StatusInternalServerError)
		panic(err)
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(ResponseMessage{Status: true, Message: "Detection started"})
	return
}
