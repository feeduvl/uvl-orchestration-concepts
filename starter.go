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

	initialize()

	router := makeRouter()
	log.Fatal(http.ListenAndServe(":9703", handlers.CORS(allowedHeaders, allowedOrigins, allowedMethods)(router)))
}

func initialize() {
	// restart observation here? In case this MS needs to be restarted
	fmt.Println("Init the Observation")
	InitObservation()
	ObserveUnclassifiedTweets()
}

func makeRouter() *mux.Router {
	router := mux.NewRouter()
	router.HandleFunc("/hitec/orchestration/twitter/observe/tweet/account/{account_name}/interval/{interval}/lang/{lang}", postObservableTwitterAccount).Methods("POST")
	router.HandleFunc("/hitec/orchestration/twitter/observe/account/{account_name}", postDeleteObservableTwitterAccount).Methods("DELETE")
	router.HandleFunc("/hitec/orchestration/twitter/process/tweet/account/{account_name}/lang/{lang}/{fast}", postProcessTweets).Methods("POST")
	router.HandleFunc("/hitec/orchestration/twitter/process/tweet/unclassified", postProcessUnclassifiedTweets).Methods("POST")
	return router
}

/*
* This method calls for each step the responsible MS
*
* Steps:
*  1. check if twitter account exists
*  2. store twitter account to to observe
*  3. add the observer (crawler)
 */
func postObservableTwitterAccount(w http.ResponseWriter, r *http.Request) {
	params := mux.Vars(r)
	accountName := params["account_name"]
	interval := params["interval"] // possible intervals: minutely, hourly, daily, monthly
	lang := params["lang"]

	fmt.Printf("1.0 postObserveTweets called with accountName: %s, interval: %s, lang: %s \n", accountName, interval, lang)

	w.Header().Set("Content-Type", "application/json")
	// 1. check if twitter account exists
	crawlerResponseMessage := RESTGetTwitterAccountNameExists(accountName)
	if !crawlerResponseMessage.AccountExists {
		fmt.Printf("1.1 account %s does not exist. The system will not be updated.\n", accountName)
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(crawlerResponseMessage)
		return
	}

	observable := ObservableTwitter{AccountName: accountName, Interval: interval, Lang: lang}

	// 2. store app to observe
	ok := RESTPostStoreObserveTwitterAccount(observable)
	if !ok {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(ResponseMessage{Status: false, Message: "storage layer unreachable"})
		return
	}

	fmt.Printf("1.1 restart observation \n")

	// 3. add the observer (crawler)
	AddObservable(observable)

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(ResponseMessage{Status: true, Message: "observation successfully initiated"})
}

func postDeleteObservableTwitterAccount(w http.ResponseWriter, r *http.Request) {
	params := mux.Vars(r)
	accountName := params["account_name"]

	fmt.Printf("1.0 postDeleteObservableTwitterAccount called for %v \n", accountName)

	// 3. add the observer (crawler)
	RemoveObservable(accountName)

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
}

/*
* This method calls for each step the responsible MS
*
* Steps:
*  1. crawl tweets
*  2. classify tweets
*  3. store tweets
 */
func postProcessTweets(w http.ResponseWriter, r *http.Request) {
	params := mux.Vars(r)
	accountName := params["account_name"]
	lang := params["lang"]
	fast := params["fast"]

	w.Header().Set("Content-Type", "application/json")

	crawlerResponseMessage := RESTGetTwitterAccountNameExists(accountName)
	if !crawlerResponseMessage.AccountExists {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(crawlerResponseMessage)
		return
	}

	processTweets(accountName, lang, fast)

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(ResponseMessage{Status: true, Message: "tweets successfully processed"})
}

func postProcessUnclassifiedTweets(w http.ResponseWriter, r *http.Request) {
	retrieveAndProcessUnclassifiedTweets()

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(ResponseMessage{Status: true, Message: "unclassified tweets successfully processed"})
}
