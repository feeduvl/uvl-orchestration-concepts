package main

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/gorilla/mux"
)

var router *mux.Router
var stopTestServer func()

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
	setupMockClient()
	observer.Stop() // don't execute scheduled updates
}

func setupRouter() {
	router = mux.NewRouter()
	router.HandleFunc("/hitec/orchestration/twitter/observe/tweet/account/{account_name}/interval/{interval}/lang/{lang}", postObservableTwitterAccount).Methods("POST")
	router.HandleFunc("/hitec/orchestration/twitter/process/tweet/account/{account_name}/lang/{lang}/{fast}", postProcessTweets).Methods("POST")
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
	mockAnalyticsClassificationTwitter(r)
	mockAnalyticsBackend(r)
	mockCollectionExplicitFeedbackTwitter(r)
	mockStorageTwitter(r)
	return r
}

func mockAnalyticsClassificationTwitter(r *mux.Router) {
	// endpointPostClassificationTwitter = "/ri-analytics-classification-twitter/lang/"
}

func mockAnalyticsBackend(r *mux.Router) {
	// endpointPostExtractTweetTopics    = "/analytics-backend/tweetClassification"
}

func mockCollectionExplicitFeedbackTwitter(r *mux.Router) {
	// endpointGetCrawlTweets              = "/ri-collection-explicit-feedback-twitter/mention/%s/lang/%s/fast"
	// endpointGetCrawlAllAvailableTweets  = "/ri-collection-explicit-feedback-twitter/mention/%s/lang/%s"

	// endpointGetTwitterAccountNameExists = "/ri-collection-explicit-feedback-twitter/%s/exists"
	r.HandleFunc("/ri-collection-explicit-feedback-twitter/{account}/exists", func(w http.ResponseWriter, request *http.Request) {
		vars := mux.Vars(request)
		account := vars["account"]
		fmt.Printf("Check account %s exists\n", account)

		if account == "WindItalia" {
			respond(w, http.StatusOK, map[string]interface{}{
				"account_exists": true,
				"message":        fmt.Sprintf("Account %s exists on Twitter", account),
			})
		} else {
			respond(w, http.StatusOK, map[string]interface{}{
				"account_exists": false,
				"message":        fmt.Sprintf("Account %s does not exist", account),
			})
		}
	})
}

func mockStorageTwitter(r *mux.Router) {
	// endpointPostObserveTwitterAccount        = "/ri-storage-twitter/store/observable/"
	r.HandleFunc("/ri-storage-twitter/store/observable/", func(w http.ResponseWriter, request *http.Request) {
		fmt.Printf("Post observable\n")
		respond(w, http.StatusOK, nil)
	})

	// endpointGetObservablesTwitterAccounts    = "/ri-storage-twitter/observables"
	// endpointDeleteObservablesTwitterAccounts = "/ri-storage-twitter/observables"
	// endpointGetUnclassifiedTweets            = "/ri-storage-twitter/account_name/%s/lang/%s/unclassified"
	// endpointPostTweet                        = "/ri-storage-twitter/store/tweet/"
	// endpointPostClassifiedTweet              = "/ri-storage-twitter/store/classified/tweet/"
	// endpointPostTweetTopics                  = "/ri-storage-twitter/store/topics"
}

func respond(writer http.ResponseWriter, statusCode int, response interface{}) {
	var body []byte
	var err error
	if response == nil {
		body = make([]byte, 0)
	} else {
		switch response.(type) {
		case string:
			body = []byte(response.(string))
		case []byte:
			body = response.([]byte)
		default:
			body, err = json.Marshal(response)
			if err != nil {
				panic(err)
			}
		}
	}

	writer.WriteHeader(statusCode)

	if _, err = writer.Write(body); err != nil {
		panic(err)
	}
}

func tearDown() {
	fmt.Println("--- --- tear down")
	stopTestServer()
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

/*
 * Test methods
 */

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
	if status := rr.Code; status != http.StatusBadRequest {
		t.Errorf("Status code differs. Expected %d .\n Got %d instead", http.StatusBadRequest, status)
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
