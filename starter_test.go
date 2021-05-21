package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/pkg/errors"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/gorilla/mux"
)

var router *mux.Router
var stopTestServer func()

var mockTweet = map[string]interface{}{
	"status_id":               "933476766408200200",
	"in_reply_to_screen_name": "musteraccount",
	"tweet_class":             "problem_report",
	"user_name":               "maxmustermann",
	"created_at":              20181201,
	"favorite_count":          1,
	"text":                    "@maxmustermann Thanks for your message!",
	"lang":                    "en",
	"retweet_count":           1,
}
var mockTweetList = []interface{}{mockTweet}

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
	r.HandleFunc("/ri-analytics-classification-twitter/lang/{lang}", func(w http.ResponseWriter, request *http.Request) {
		respond(w, http.StatusOK, mockTweetList)
	})
}

func mockAnalyticsBackend(r *mux.Router) {
	// endpointPostExtractTweetTopics    = "/analytics-backend/tweetClassification"
	r.HandleFunc("/analytics-backend/tweetClassification", func(w http.ResponseWriter, request *http.Request) {
		respond(w, http.StatusOK, `{"first_class": {"label": "label1", "score": 0.7}, "second_class":  {"label":  "label2", "score": 0.5}}`)
	})
}

func mockCollectionExplicitFeedbackTwitter(r *mux.Router) {
	// endpointGetCrawlTweets              = "/ri-collection-explicit-feedback-twitter/mention/%s/lang/%s/fast"
	// endpointGetCrawlAllAvailableTweets  = "/ri-collection-explicit-feedback-twitter/mention/%s/lang/%s"
	r.HandleFunc("/ri-collection-explicit-feedback-twitter/mention/{account}/lang/{lang}/{fast}", func(w http.ResponseWriter, request *http.Request) {
		respond(w, http.StatusOK, mockTweetList)
	})

	// endpointGetTwitterAccountNameExists = "/ri-collection-explicit-feedback-twitter/%s/exists"
	r.HandleFunc("/ri-collection-explicit-feedback-twitter/{account}/exists", func(w http.ResponseWriter, request *http.Request) {
		vars := mux.Vars(request)
		account := vars["account"]
		fmt.Printf("Check account %s exists\n", account)

		if account == "WindItalia" || account == "VodafoneUK" {
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
		respond(w, http.StatusOK, nil)
	})

	// endpointGetObservablesTwitterAccounts    = "/ri-storage-twitter/observables"
	// endpointDeleteObservablesTwitterAccounts = "/ri-storage-twitter/observables"
	r.HandleFunc("/ri-storage-twitter/observables", func(w http.ResponseWriter, request *http.Request) {
		if request.Method == "GET" {
			respond(w, http.StatusOK, `[{"account_name": "WindItalia", "interval": "midnight", "lang": "it"}]`)
		} else if request.Method == "DELETE" {
			respond(w, http.StatusOK, nil)
		} else {
			respond(w, http.StatusMethodNotAllowed, nil)
		}
	})

	// endpointGetUnclassifiedTweets            = "/ri-storage-twitter/account_name/%s/lang/%s/unclassified"
	r.HandleFunc("/ri-storage-twitter/account_name/{account}/lang/{lang}/unclassified", func(w http.ResponseWriter, request *http.Request) {
		respond(w, http.StatusOK, mockTweetList)
	})

	// endpointPostTweet                        = "/ri-storage-twitter/store/tweet/"
	r.HandleFunc("/ri-storage-twitter/store/tweet/", func(w http.ResponseWriter, request *http.Request) {
		respond(w, http.StatusOK, nil)
	})

	// endpointPostClassifiedTweet              = "/ri-storage-twitter/store/classified/tweet/"
	r.HandleFunc("/ri-storage-twitter/store/classified/tweet/", func(w http.ResponseWriter, request *http.Request) {
		respond(w, http.StatusOK, nil)
	})

	// endpointPostTweetTopics                  = "/ri-storage-twitter/store/topics"
	r.HandleFunc("/ri-storage-twitter/store/topics", func(w http.ResponseWriter, request *http.Request) {
		respond(w, http.StatusOK, nil)
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

/*
 * Test methods
 */

func TestPostObservableTwitterAccount(t *testing.T) {
	println("Running test TestPostObservableTwitterAccount")
	ep := endpoint{method: "POST", url: "/hitec/orchestration/twitter/observe/tweet/account/%s/interval/%s/lang/%s"}
	assertFailure(t, ep.withVars("should", "fail", "en").mustExecuteRequest(nil))
	assertSuccess(t, ep.withVars("VodafoneUK", "2h", "en").mustExecuteRequest(nil))
	assertSuccess(t, ep.withVars("VodafoneUK", "2h", "en").mustExecuteRequest(nil))                 // noop re-adding existing observable
	assertSuccess(t, ep.withVars("VodafoneUK", "30 3-6,20-23 * * *", "en").mustExecuteRequest(nil)) // update existing observable
}

func TestPostDeleteObservableTwitterAccount(t *testing.T) {
	println("Running test TestPostDeleteObservableTwitterAccount")
	ep := endpoint{method: "DELETE", url: "/hitec/orchestration/twitter/observe/account/%s"}
	assertSuccess(t, ep.withVars("VodafoneUK").mustExecuteRequest(nil))
}

func TestPostProcessTweets(t *testing.T) {
	println("Running test TestPostProcessTweets")
	ep := endpoint{method: "POST", url: "/hitec/orchestration/twitter/process/tweet/account/%s/lang/%s/%s"}
	assertFailure(t, ep.withVars("shouldfail", "en", "slow").mustExecuteRequest(nil))
	assertFailure(t, ep.withVars("shouldfail", "en", "fast").mustExecuteRequest(nil))
	assertSuccess(t, ep.withVars("WindItalia", "it", "slow").mustExecuteRequest(nil))
	assertSuccess(t, ep.withVars("WindItalia", "it", "fast").mustExecuteRequest(nil))
}

func TestPostProcessUnclassifiedTweets(t *testing.T) {
	println("Running test TestPostProcessUnclassifiedTweets")
	ep := endpoint{method: "POST", url: "/hitec/orchestration/twitter/process/tweet/unclassified"}
	assertSuccess(t, ep.mustExecuteRequest(nil))
}
