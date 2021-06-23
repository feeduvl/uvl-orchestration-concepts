package main

import (
	"encoding/csv"
	"encoding/json"
	"fmt"
	"strconv"
	"time"

	//"io"
	"log"
	"net/http"
	"strings"

	"os"

	"github.com/gorilla/handlers"
	"github.com/gorilla/mux"

	"github.com/360EntSecGroup-Skylar/excelize"
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

	if name[1] != "csv" && name[1] != "txt" && name[1] != "xlsx" {
		json.NewEncoder(w).Encode(ResponseMessage{Status: true, Message: "Filetype not supported"})
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	// Process it
	var d = Dataset{}
	if name[1] == "xlsx" {
		f, err := excelize.OpenReader(file)
		if err != nil {
			json.NewEncoder(w).Encode(ResponseMessage{Status: true, Message: "Error reading xlsx file"})
			w.WriteHeader(http.StatusInternalServerError)
			panic(err)
		}
		sheetName := f.GetSheetList()[0]
		cols, err := f.GetCols(sheetName)
		if err != nil {
			json.NewEncoder(w).Encode(ResponseMessage{Status: true, Message: "Error reading xlsx columns"})
			w.WriteHeader(http.StatusInternalServerError)
			panic(err)
		}

		var a []Document
		var ids = false
		if len(cols) > 1 {
			ids = true
		}
		for i, rowCell := range cols[0] {
			var s string
			if ids {
				s = cols[1][i]
			} else {
				s = strconv.Itoa(i)
			}
			if rowCell != "" {
				var d = Document{i, rowCell, s}
				a = append(a, d)
			} else {
				break
			}
		}
		d = Dataset{Name: name[0], Size: len(a), Documents: a, UploadedAt: time.Now()}

	} else {
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
			var s string
			if len(line) == 1 {
				s = strconv.Itoa(i)
			} else {
				s = line[1]
			}
			var d = Document{i, line[0], s}
			a = append(a, d)
		}
		d = Dataset{Name: name[0], Size: len(a), Documents: a, UploadedAt: time.Now()}
	}

	// Store dataset in database
	err = RESTPostStoreDataset(d)
	if err != nil {
		json.NewEncoder(w).Encode(ResponseMessage{Status: true, Message: "Error saving dataset"})
		w.WriteHeader(http.StatusInternalServerError)
		panic(err)
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(ResponseMessage{Status: true, Message: "Dataset successfully uploaded"})
	return

}

func postStartNewDetection(w http.ResponseWriter, r *http.Request) {

	var body map[string]interface{}
	err := json.NewDecoder(r.Body).Decode(&body)
	fmt.Printf("postStartNewDetection called. Parsed Body: %v\n", body)
	fmt.Printf("postStartNewDetection called. Error decoding body: %s\n", err)

	datasetName := body["dataset"].(string)
	if datasetName == "" {
		json.NewEncoder(w).Encode(ResponseMessage{Status: true, Message: "Cannot start detection with no dataset."})
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	method := body["method"].(string)
	fmt.Printf("postStartNewDetection called. Method: %v, Dataset: %v\n", method, datasetName)

	name := body["name"].(string)

	// Get Dataset from Database
	dataset, err := RESTGetDataset(datasetName)
	if err != nil {
		fmt.Printf("ERROR retrieving dataset (postStartNewDetection) %s\n", err)
		w.WriteHeader(http.StatusBadRequest)
		panic(err)
	}

	// Get parameters
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
	result.Name = name

	run := new(Run)
	run.Method = method
	run.Params = params
	run.Dataset = dataset

	// Store result object in database (prior to getting results)
	err = RESTPostStoreResult(*result)
	if err != nil {
		json.NewEncoder(w).Encode(ResponseMessage{Status: true, Message: "Error saving to database"})
		w.WriteHeader(http.StatusInternalServerError)
		panic(err)
	}

	go _startNewDetection(result, run)

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(ResponseMessage{Status: true, Message: "Detection started"})
	return
}

func _startNewDetection(result *Result, run *Run) {

	// Change status and save it to database
	result.Status = "started"
	_ = RESTPostStoreResult(*result)

	// Call detection MS
	fmt.Printf("_startNewDetection, calling MS and waiting for response\n")
	endResult, err := RESTPostStartNewDetection(*result, *run)
	if err != nil {
		fmt.Printf("ERROR with detection %s\n", err)
		endResult.Status = "failed"
		_ = RESTPostStoreResult(endResult)
		return
	}

	endResult.Status = "finished"

	// Store results in database
	fmt.Printf("Response received, Topcis: %s\n", endResult.Topics)
	_ = RESTPostStoreResult(endResult)
	if err != nil {
		fmt.Printf("ERROR storing final result %s\n", err)
		panic(err)
	}

	// What to do when storing the result fails?
}
