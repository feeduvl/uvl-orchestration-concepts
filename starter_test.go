package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/pkg/errors"
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
	router = makeRouter()
	setupMockClient()
	observer.Stop() // don't execute scheduled updates
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
	r.NotFoundHandler = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Println(errors.Errorf("Service method not mocked: %s", r.URL))
		w.WriteHeader(http.StatusNotFound)
	})
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

func assertJsonDecodes(t *testing.T, rr *httptest.ResponseRecorder, v interface{}) {
	err := json.Unmarshal(rr.Body.Bytes(), v)
	if err != nil {
		t.Error(errors.Wrap(err, "Expected valid json array"))
	}
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

func TestPostObservableTwitterAccount(t *testing.T) {
	ep := endpoint{method: "POST", url: "/hitec/orchestration/twitter/observe/tweet/account/%s/interval/%s/lang/%s"}
	assertFailure(t, ep.withVars("should", "fail", "en").mustExecuteRequest(nil))
	assertSuccess(t, ep.withVars("WindItalia", "2h", "it").mustExecuteRequest(nil))
}
