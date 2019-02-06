package main

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/robfig/cron"

	"github.com/gorilla/mux"
)

var router *mux.Router

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
	setupRouter()
}

func setupRouter() {
	router = mux.NewRouter()
	router.HandleFunc("/hitec/orchestration/twitter/observe/tweet/account/{account_name}/interval/{interval}/lang/{lang}", MockPostObserveTweets).Methods("POST")
	router.HandleFunc("/hitec/orchestration/twitter/process/tweet/account/{account_name}/lang/{lang}/{fast}", MockPostProcessTweets).Methods("POST")
}

func tearDown() {
	fmt.Println("--- --- tear down")
}

func buildRequest(method, endpoint string, payload io.Reader, t *testing.T) *http.Request {
	req, err := http.NewRequest(method, endpoint, payload)
	if err != nil {
		t.Errorf("An error occurred. %v", err)
	}

	return req
}

func executeRequest(req *http.Request) *httptest.ResponseRecorder {
	rr := httptest.NewRecorder()
	router.ServeHTTP(rr, req)

	return rr
}

func TestPostObserveTweets(t *testing.T) {
	fmt.Println("start TestPostObserveTweets")
	var method = "POST"
	var endpoint = "/hitec/orchestration/twitter/observe/tweet/account/%s/interval/%s/lang/%s"

	/*
	 * test for faillure
	 */
	endpointFail := fmt.Sprintf(endpoint, "should", "fail", "30h")
	req := buildRequest(method, endpointFail, nil, t)
	rr := executeRequest(req)

	//Confirm the response has the right status code
	if status := rr.Code; status != http.StatusInternalServerError {
		t.Errorf("Status code differs. Expected %d .\n Got %d instead", http.StatusInternalServerError, status)
	}

	/*
	 * test for success
	 */
	endpointSuccess := fmt.Sprintf(endpoint, "WindItalia", "2h", "it")
	req = buildRequest(method, endpointSuccess, nil, t)
	rr = executeRequest(req)

	//Confirm the response has the right status code
	if status := rr.Code; status != http.StatusOK {
		t.Errorf("Status code differs. Expected %d .\n Got %d instead", http.StatusOK, status)
	}
}

func MockPostObserveTweets(w http.ResponseWriter, r *http.Request) {
	params := mux.Vars(r)
	accountName := params["account_name"]
	interval := params["interval"] // possible intervals: minutely, hourly, daily, monthly
	lang := params["lang"]

	fmt.Printf("1.0 postObserveTweets called with accountName: %s, interval: %s, lang: %s \n", accountName, interval, lang)

	// 1. store app to observe

	var ok bool
	allowedLanguages := map[string]bool{
		"en": true,
		"it": true,
	}
	allowedSpcialIntervals := map[string]bool{
		"minutely": true,
		"hourly":   true,
		"daily":    true,
		"weekly":   true,
		"monthly":  true,
		"6h":       true,
		"2h":       true,
	}
	_, err := cron.Parse(interval)
	if accountName == "" || !allowedLanguages[lang] || (err != nil && !allowedSpcialIntervals[interval]) {
		ok = false
	} else {
		ok = true
	}
	w.Header().Set("Content-Type", "application/json")
	if !ok {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(ResponseMessage{Status: false, Message: "storage layer unreachable"})
		return
	}

	fmt.Printf("1.1 restart observation \n")

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(ResponseMessage{Status: true, Message: "observation successfully initiated"})
}

func TestPostProcessTweets(t *testing.T) {
	fmt.Println("start TestPostProcessTweets")
	var method = "POST"
	var endpoint = "/hitec/orchestration/twitter/process/tweet/account/%s/lang/%s/%s"

	/*
	 * test for faillure
	 */
	endpointFail := fmt.Sprintf(endpoint, "", "fail", "error")
	req := buildRequest(method, endpointFail, nil, t)
	rr := executeRequest(req)

	//Confirm the response has the right status code
	if status := rr.Code; status != http.StatusMovedPermanently {
		t.Errorf("Status code differs. Expected %d .\n Got %d instead", http.StatusMovedPermanently, status)
	}

	/*
	 * test for success
	 */
	endpointSuccess := fmt.Sprintf(endpoint, "WindItalia", "it", "fast")
	req = buildRequest(method, endpointSuccess, nil, t)
	rr = executeRequest(req)

	//Confirm the response has the right status code
	if status := rr.Code; status != http.StatusOK {
		t.Errorf("Status code differs. Expected %d .\n Got %d instead", http.StatusOK, status)
	}
}

func MockPostProcessTweets(w http.ResponseWriter, r *http.Request) {
	params := mux.Vars(r)
	accountName := params["account_name"]
	lang := params["lang"]
	fast := params["fast"]

	var ok bool
	allowedLanguages := map[string]bool{
		"en": true,
		"it": true,
	}
	if accountName == "" || !allowedLanguages[lang] || fast == "" {
		ok = false
	} else {
		ok = true
	}

	w.Header().Set("Content-Type", "application/json")
	if !ok {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(ResponseMessage{Status: false, Message: "account name or language are empty"})
		return
	}

	fmt.Printf("1.1 restart observation \n")

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(ResponseMessage{Status: true, Message: "tweets successfully processed"})
}
