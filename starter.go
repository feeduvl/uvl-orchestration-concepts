package main

import (
	"bytes"
	"encoding/csv"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"mime/multipart"
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

	// Init
	router.HandleFunc("/hitec/orchestration/concepts/annotationinit/", makeNewAnnotation).Methods("POST")
	router.HandleFunc("/hitec/orchestration/concepts/store/dataset/", postNewDataset).Methods("POST")
	router.HandleFunc("/hitec/orchestration/concepts/store/groundtruth/", postAddGroundTruth).Methods("POST")
	router.HandleFunc("/hitec/orchestration/concepts/detection/", postStartNewDetection).Methods("POST")
	return router
}

func handleErrorWithResponse(w http.ResponseWriter, err error, message string) {
	if err != nil {
		_ = json.NewEncoder(w).Encode(ResponseMessage{Status: true, Message: message})
		w.WriteHeader(http.StatusInternalServerError)
		panic(err)
	}
}

func postNewDataset(w http.ResponseWriter, r *http.Request) {

	w.Header().Set("Content-Type", "application/json")

	// Receive new dataset
	err := r.ParseMultipartForm(32 << 20)
	if err != nil {
		_ = json.NewEncoder(w).Encode(ResponseMessage{Status: true, Message: "Form data could not be retrieved"})
		w.WriteHeader(http.StatusInternalServerError)
		fmt.Printf("ERROR parsing form data: %s for request body %v\n", err, r.Body)
		return
	}

	file, header, err := r.FormFile("file")
	handleErrorWithResponse(w, err, "File error")
	defer func(file multipart.File) {
		_ = file.Close()
	}(file)

	name := strings.Split(header.Filename, ".")
	fmt.Printf("postNewDataset called. File name: %s\n", name[0])

	if name[1] != "csv" && name[1] != "txt" && name[1] != "xlsx" {
		_ = json.NewEncoder(w).Encode(ResponseMessage{Status: true, Message: "Filetype not supported"})
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	// Process it
	var d = Dataset{}
	if name[1] == "xlsx" {
		f, err := excelize.OpenReader(file)
		handleErrorWithResponse(w, err, "Error reading xlsx file")
		sheetName := f.GetSheetList()[0]
		cols, err := f.GetCols(sheetName)
		handleErrorWithResponse(w, err, "Error reading xlsx columns")

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
		handleErrorWithResponse(w, err, "Csv processing error")
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
	handleErrorWithResponse(w, err, "Error saving dataset")

	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(ResponseMessage{Status: true, Message: "Dataset successfully uploaded"})
	return

}

func postAddGroundTruth(w http.ResponseWriter, r *http.Request) {

	w.Header().Set("Content-Type", "application/json")

	// Receive groundtruth
	err := r.ParseMultipartForm(32 << 20)
	if err != nil {
		_ = json.NewEncoder(w).Encode(ResponseMessage{Status: true, Message: "Form data could not be retrieved"})
		w.WriteHeader(http.StatusInternalServerError)
		fmt.Printf("ERROR decoding json: %s for request body %v\n", err, r.Body)
		return
	}

	file, header, err := r.FormFile("file")
	handleErrorWithResponse(w, err, "File error")
	defer func(file multipart.File) {
		_ = file.Close()
	}(file)

	datasetName := r.FormValue("dataset")
	name := strings.Split(header.Filename, ".")
	fmt.Printf("postAddGroundTruth called. File name: %s, Dataset: %s.\n", name[0], datasetName)

	// Process file content
	var d = Dataset{}
	if name[1] == "xlsx" {
		f, err := excelize.OpenReader(file)
		handleErrorWithResponse(w, err, "Error reading xlsx file")
		sheetName := f.GetSheetList()[0]
		cols, err := f.GetCols(sheetName)
		handleErrorWithResponse(w, err, "Error reading xlsx columns")

		var a []TruthElement
		var ids = false
		if len(cols) > 1 {
			ids = true
		}
		for i, rowCell := range cols[0] {
			var s string
			if ids {
				s = cols[1][i]
			} else {
				s = ""
			}
			if rowCell != "" {
				var t = TruthElement{s, rowCell}
				a = append(a, t)
			} else {
				break
			}
		}
		d = Dataset{Name: datasetName, GroundTruth: a}

	} else {
		reader := csv.NewReader(file)
		reader.Comma = '|'
		reader.LazyQuotes = true
		lines, err := reader.ReadAll()
		handleErrorWithResponse(w, err, "Csv processing error")
		var a []TruthElement
		for _, line := range lines {
			var s string
			if len(line) == 1 {
				s = ""
			} else {
				s = line[1]
			}
			var t = TruthElement{s, line[0]}
			a = append(a, t)
		}
		d = Dataset{Name: datasetName, GroundTruth: a}
	}

	// Store groundtruth in database
	err = RESTPostStoreGroundTruth(d)
	handleErrorWithResponse(w, err, "Error saving groundtruth")
	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(ResponseMessage{Status: true, Message: "GroundTruth successfully uploaded"})
	return
}

func postStartNewDetection(w http.ResponseWriter, r *http.Request) {

	var body map[string]interface{}
	err := json.NewDecoder(r.Body).Decode(&body)
	if err != nil {
		fmt.Printf("ERROR decoding body: %s, body: %v\n", err, r.Body)
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	datasetName := body["dataset"].(string)
	if datasetName == "" {
		_ = json.NewEncoder(w).Encode(ResponseMessage{Status: true, Message: "Cannot start detection with no dataset."})
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	method := body["method"].(string)
	fmt.Printf("postStartNewDetection called. Method: %v, Dataset: %v\n", method, datasetName)

	name := body["name"].(string)

	// Get Dataset from Database
	dataset, err := RESTGetDataset(datasetName)
	handleErrorWithResponse(w, err, "ERROR retrieving dataset")

	// Get parameters
	var params = make(map[string]string)
	for key, value := range body {
		s := fmt.Sprintf("%v", value)
		params[key] = s
	}

	delete(params, "method")
	delete(params, "dataset")
	delete(params, "name")

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
	handleErrorWithResponse(w, err, "Error saving to database")

	go _startNewDetection(result, run)

	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(ResponseMessage{Status: true, Message: "Detection started"})
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
	fmt.Printf("Response received, Topics: %s\n", endResult.Topics)
	err = RESTPostStoreResult(endResult)
	if err != nil {
		fmt.Printf("ERROR storing final result %s\n", err)
		panic(err)
	}

	// What to do when storing the result fails?
}

func createKeyValuePairs(m map[string]interface{}) string {
	b := new(bytes.Buffer)
	for key, value := range m {
		fmt.Fprintf(b, "%s=\"%#v\"\n", key, value)
	}
	return b.String()
}

// makeNewAnnotation make and return a new document annotation
func makeNewAnnotation(w http.ResponseWriter, r *http.Request) {
	var body map[string]interface{}
	err := json.NewDecoder(r.Body).Decode(&body)
	fmt.Printf("postAnnotationTokenize called: %s", createKeyValuePairs(body))
	if err != nil {
		fmt.Printf("ERROR decoding body: %s, body: %v\n", err, r.Body)
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	annotationName := body["name"].(string)
	datasetName := body["dataset"].(string)
	if datasetName == "" {
		_ = json.NewEncoder(w).Encode(ResponseMessage{Status: true, Message: "Cannot start detection with no dataset."})
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	tokenizationJsonBytes, err := getNewAnnotation(w, datasetName)
	if err != nil {
		fmt.Printf("Error getting tokenization, returning")
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	var annotation Annotation
	if err := json.Unmarshal(tokenizationJsonBytes, &annotation); err != nil {
		fmt.Printf("Failed to parse annotation bytes")
		return
	}

	// initialize basic fields
	annotation.UploadedAt = time.Now()
	annotation.Name = annotationName
	annotation.Dataset = datasetName

	err = RESTPostStoreAnnotation(annotation)
	if err != nil {
		fmt.Printf("Failed to POST new annotation")
		return
	}

	finalAnnotation, err := json.Marshal(annotation)
	if err != nil {
		fmt.Printf("Failed to marshal annotation")
	}
	w.Write(finalAnnotation)
}

// postAnnotationTokenize Tokenize a document and return the result
func getNewAnnotation(w http.ResponseWriter, datasetName string) ([]byte, error) {
	dataset, err := RESTGetDataset(datasetName)
	handleErrorWithResponse(w, err, "ERROR retrieving dataset")
	if err != nil {
		return *new([]byte), err
	}

	log.Printf("Tokenizing: " + datasetName)

	requestBody := new(bytes.Buffer)

	url := baseURL + endpointPostAnnotationTokenize
	_ = json.NewEncoder(requestBody).Encode(dataset)
	req, _ := createRequest(POST, url, requestBody)

	res, err := client.Do(req)

	defer res.Body.Close()

	if err != nil {
		log.Printf("ERR getting tokens for annotation %v\n", err)
		log.Printf("Note: If the request timed out, the method microservice may take too long to process the" +
			" request. Consider increasing timeout in rest_handler->getHTTPClient.")
		return *new([]byte), err
	}

	w.WriteHeader(res.StatusCode)

	b, err := ioutil.ReadAll(res.Body)

	if err != nil {
		log.Fatalln(err)
		return *new([]byte), err
	}

	log.Printf("Got response: " + string(b))
	return b, nil
}
