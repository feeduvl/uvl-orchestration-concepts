package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/pkg/errors"
	"github.com/stretchr/testify/assert"
	"io"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
	"time"

	"github.com/gorilla/mux"
)

var router *mux.Router
var stopTestServer func()

var documents []Document
var mockDataset Dataset
var mockResult Result
var invalidPayloadString = "payload"
var invalidPayload []byte

func setupDataset() {
	documents = append(documents, Document{
		Id:   "0",
		Text: "Text 1",
	})
	documents = append(documents, Document{
		Id:   "1",
		Text: "Text 2",
	})
	documents = append(documents, Document{
		Id:   "2",
		Text: "Text 3",
	})

	mockDataset.Documents = documents
	mockDataset.UploadedAt = time.Now()
	mockDataset.Name = "test"
	mockDataset.Size = 3
}

func TestMain(m *testing.M) {
	fmt.Println("--- Start Tests")
	setup()

	// run the test cases defined in this file
	retCode := m.Run()

	tearDown()

	// call with result of m.Run()
	os.Exit(retCode)
}

func setup() {
	fmt.Println("--- --- setup")
	router = makeRouter()
	setupMockClient()
	setupDataset()
}

func setupMockClient() {
	fmt.Println("Mocking client")
	handler := makeMockHandler()
	s := httptest.NewServer(handler)
	stopTestServer = s.Close
	baseURL = s.URL
}

func makeMockHandler() http.Handler {
	r := mux.NewRouter()
	mockStorageConcepts(r)
	mockDetectionService(r)
	r.NotFoundHandler = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Println(errors.Errorf("Service method not mocked: %s", r.URL))
		w.WriteHeader(http.StatusNotFound)
	})
	return r
}

func mockStorageConcepts(r *mux.Router) {
	// endpointPostStoreDataset        = "/hitec/repository/concepts/store/dataset/"
	r.HandleFunc("/hitec/repository/concepts/store/dataset/", func(w http.ResponseWriter, request *http.Request) {
		respond(w, http.StatusOK, nil)
	})

	// endpointPostStoreGroundTruth        = "/hitec/repository/concepts/store/groundtruth/"
	r.HandleFunc("/hitec/repository/concepts/store/groundtruth/", func(w http.ResponseWriter, request *http.Request) {
		respond(w, http.StatusOK, nil)
	})

	// endpointPostStoreDetectionResult        = "/hitec/repository/concepts/store/detection/result/"
	r.HandleFunc("/hitec/repository/concepts/store/detection/result/", func(w http.ResponseWriter, request *http.Request) {
		respond(w, http.StatusOK, nil)
	})

	// endpointGetDataset     = "/hitec/repository/concepts/dataset/name/"
	r.HandleFunc("/hitec/repository/concepts/dataset/name/test", func(w http.ResponseWriter, request *http.Request) {
		respond(w, http.StatusOK, mockDataset)
	})
	r.HandleFunc("/hitec/repository/concepts/dataset/name/failed", func(w http.ResponseWriter, request *http.Request) {
		respond(w, http.StatusBadRequest, nil)
	})
	r.HandleFunc("/hitec/repository/concepts/dataset/name/failed2", func(w http.ResponseWriter, request *http.Request) {
		respond(w, http.StatusOK, invalidPayload)
	})
	r.HandleFunc("/hitec/repository/concepts/dataset/name/failed3", func(w http.ResponseWriter, request *http.Request) {
		respond(w, 0, nil)
	})
}

func mockDetectionService(r *mux.Router) {
	// endPointPostStartConceptDetection = "/hitec/classify/concepts/"
	r.HandleFunc("/hitec/classify/concepts/method/run", func(w http.ResponseWriter, request *http.Request) {
		respond(w, http.StatusOK, mockResult)
	})
	r.HandleFunc("/hitec/classify/concepts/fail/run", func(w http.ResponseWriter, request *http.Request) {
		respond(w, http.StatusNotFound, nil)
	})
	r.HandleFunc("/hitec/classify/concepts/fail2/run", func(w http.ResponseWriter, request *http.Request) {
		respond(w, 0, nil)
	})
}

func respond(writer http.ResponseWriter, statusCode int, body interface{}) {
	var bodyData []byte
	var err error
	if body == nil {
		bodyData = make([]byte, 0)
	} else {
		switch body.(type) {
		case string:
			bodyData = []byte(body.(string))
		case []byte:
			bodyData = body.([]byte)
		default:
			bodyData, err = json.Marshal(body)
			if err != nil {
				panic(err)
			}
		}
	}

	writer.WriteHeader(statusCode)

	if _, err = writer.Write(bodyData); err != nil {
		panic(err)
	}
}

func tearDown() {
	fmt.Println("--- --- tear down")
	stopTestServer()
}

type endpoint struct {
	method string
	url    string
}

func (e endpoint) withVars(vs ...interface{}) endpoint {
	e.url = fmt.Sprintf(e.url, vs...)
	return e
}

func (e endpoint) executeRequest(payload interface{}) (error, *httptest.ResponseRecorder) {
	body := new(bytes.Buffer)
	err := json.NewEncoder(body).Encode(payload)
	if err != nil {
		return err, nil
	}

	req, err := http.NewRequest(e.method, e.url, body)
	if err != nil {
		return err, nil
	}

	rr := httptest.NewRecorder()
	router.ServeHTTP(rr, req)

	return nil, rr
}

func (e endpoint) mustExecuteRequest(payload interface{}) *httptest.ResponseRecorder {
	err, rr := e.executeRequest(payload)
	if err != nil {
		panic(errors.Wrap(err, `Could not execute request`))
	}

	return rr
}

func (e endpoint) executeRequestForm(payload *bytes.Buffer, writer *multipart.Writer) (error, *httptest.ResponseRecorder) {

	writer.Close()
	req, err := http.NewRequest(e.method, e.url, bytes.NewReader(payload.Bytes()))
	if err != nil {
		return err, nil
	}
	req.Header.Set("Content-Type", writer.FormDataContentType())

	rr := httptest.NewRecorder()
	router.ServeHTTP(rr, req)

	return nil, rr
}

func (e endpoint) mustExecuteRequestForm(payload *bytes.Buffer, writer *multipart.Writer) *httptest.ResponseRecorder {
	err, rr := e.executeRequestForm(payload, writer)
	if err != nil {
		panic(errors.Wrap(err, `Could not execute request`))
	}

	return rr
}

func isSuccess(code int) bool {
	return code >= 200 && code < 300
}

func assertSuccess(t *testing.T, rr *httptest.ResponseRecorder) {
	if !isSuccess(rr.Code) {
		t.Errorf("Status code differs. Expected success.\n Got status %d (%s) instead", rr.Code, http.StatusText(rr.Code))
	}
}
func assertFailure(t *testing.T, rr *httptest.ResponseRecorder) {
	if isSuccess(rr.Code) {
		t.Errorf("Status code differs. Expected failure.\n Got status %d (%s) instead", rr.Code, http.StatusText(rr.Code))
	}
}

func assertMessage(t *testing.T, rr *httptest.ResponseRecorder, message string) {
	if isSuccess(rr.Code) {
		var response map[string]interface{}
		_ = json.NewDecoder(rr.Body).Decode(&response)
		if !(response["message"] == message) {
			t.Errorf("Message differs! Got %s", response["message"])
		}
	}
}

/*
 * Test methods
 */
func TestPostNewDataset(t *testing.T) {
	ep := endpoint{method: "POST", url: "/hitec/orchestration/concepts/store/dataset/"}
	assertSuccess(t, ep.mustExecuteRequest(nil))

	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)
	defer writer.Close()
	fw, _ := writer.CreateFormFile("file", "test.csv")
	file, _ := os.Open("test/test.csv")
	_, _ = io.Copy(fw, file)

	assertSuccess(t, ep.mustExecuteRequestForm(body, writer))

	body = &bytes.Buffer{}
	writer = multipart.NewWriter(body)
	defer writer.Close()
	fw, _ = writer.CreateFormFile("file", "test2.csv")
	file, _ = os.Open("test/test2.csv")
	_, _ = io.Copy(fw, file)

	assertSuccess(t, ep.mustExecuteRequestForm(body, writer))

	body = &bytes.Buffer{}
	writer = multipart.NewWriter(body)
	defer writer.Close()
	fw, _ = writer.CreateFormFile("file", "test.xlsx")
	file, _ = os.Open("test/test.xlsx")
	_, _ = io.Copy(fw, file)

	assertSuccess(t, ep.mustExecuteRequestForm(body, writer))

	body = &bytes.Buffer{}
	writer = multipart.NewWriter(body)
	defer writer.Close()
	fw, _ = writer.CreateFormFile("file", "test2.xlsx")
	file, _ = os.Open("test/test2.xlsx")
	_, _ = io.Copy(fw, file)

	assertSuccess(t, ep.mustExecuteRequestForm(body, writer))

	body = &bytes.Buffer{}
	writer = multipart.NewWriter(body)
	defer writer.Close()
	fw, _ = writer.CreateFormFile("file", "test.dat")
	file, _ = os.Open("test/test.dat")
	_, _ = io.Copy(fw, file)

	assertSuccess(t, ep.mustExecuteRequestForm(body, writer))
}

func TestPostAddGroundTruth(t *testing.T) {
	ep := endpoint{method: "POST", url: "/hitec/orchestration/concepts/store/groundtruth/"}
	assertSuccess(t, ep.mustExecuteRequest(nil))

	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)
	defer writer.Close()
	fw, _ := writer.CreateFormFile("file", "test.csv")
	file, _ := os.Open("test/test.csv")
	_, _ = io.Copy(fw, file)

	body = &bytes.Buffer{}
	writer = multipart.NewWriter(body)
	defer writer.Close()
	fw, _ = writer.CreateFormFile("file", "test2.csv")
	file, _ = os.Open("test/test2.csv")
	_, _ = io.Copy(fw, file)

	assertSuccess(t, ep.mustExecuteRequestForm(body, writer))

	body = &bytes.Buffer{}
	writer = multipart.NewWriter(body)
	defer writer.Close()
	fw, _ = writer.CreateFormFile("file", "test.xlsx")
	file, _ = os.Open("test/test.xlsx")
	_, _ = io.Copy(fw, file)

	assertSuccess(t, ep.mustExecuteRequestForm(body, writer))

	body = &bytes.Buffer{}
	writer = multipart.NewWriter(body)
	defer writer.Close()
	fw, _ = writer.CreateFormFile("file", "test2.xlsx")
	file, _ = os.Open("test/test2.xlsx")
	_, _ = io.Copy(fw, file)

	assertSuccess(t, ep.mustExecuteRequestForm(body, writer))

	body = &bytes.Buffer{}
	writer = multipart.NewWriter(body)
	defer writer.Close()
	fw, _ = writer.CreateFormFile("file", "test.dat")
	file, _ = os.Open("test/test.dat")
	_, _ = io.Copy(fw, file)

	assertSuccess(t, ep.mustExecuteRequestForm(body, writer))
}

func TestPostStartNewDetection(t *testing.T) {
	ep := endpoint{method: "POST", url: "/hitec/orchestration/concepts/detection/"}

	assertFailure(t, ep.mustExecuteRequest(invalidPayloadString))

	var requestBodyFail = make(map[string]interface{})
	requestBodyFail["dataset"] = ""
	assertMessage(t, ep.mustExecuteRequest(requestBodyFail), "Cannot start detection with no dataset.")

	var requestBody = make(map[string]interface{})
	var params = make(map[string]string)
	params["alpha"] = "0.2"

	requestBody["dataset"] = "test"
	requestBody["method"] = "method"
	requestBody["name"] = "test_"
	requestBody["params"] = params
	assertSuccess(t, ep.mustExecuteRequest(requestBody))

	var result = new(Result)
	result.Method = "method"
	result.DatasetName = "test"
	result.Params = params
	var run = new(Run)
	run.Method = "method"
	run.Dataset = mockDataset
	run.Params = params

	assert.NotPanics(t, func() {
		_startNewDetection(result, run)
	})

	run.Method = "fail"

	assert.NotPanics(t, func() {
		_startNewDetection(result, run)
	})

	run.Method = "fail2"

	assert.NotPanics(t, func() {
		_startNewDetection(result, run)
	})
}

func TestRESTGetDataset(t *testing.T) {
	_, err := RESTGetDataset("failed")
	assert.Error(t, err)
	_, err = RESTGetDataset("failed2")
	assert.Error(t, err)
	_, err = RESTGetDataset("failed3")
	assert.Error(t, err)
}
