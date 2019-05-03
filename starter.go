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

	router := mux.NewRouter()
	router.HandleFunc("/hitec/orchestration/twitter/observe/tweet/account/{account_name}/interval/{interval}/lang/{lang}", postObserveTweets).Methods("POST")
	router.HandleFunc("/hitec/orchestration/twitter/process/tweet/account/{account_name}/lang/{lang}/{fast}", postProcessTweets).Methods("POST")

	// restart observation here? In case this MS needs to be restarted
	fmt.Println("Restart the Observation")
	go RestartObservation()
	go RetrieveAndProcessUnclassifiedTweets()
	fmt.Println("MS started")
	log.Fatal(http.ListenAndServe(":9703", handlers.CORS(allowedHeaders, allowedOrigins, allowedMethods)(router)))
}

/*
* This method calls for each step the responsible MS
*
* Steps:
*  1. store twitter account to to observe
*  2. notify the observer (crawler)
*  2.1 notify the processing layer to classify the newly addded tweets
 */
func postObserveTweets(w http.ResponseWriter, r *http.Request) {
	params := mux.Vars(r)
	accountName := params["account_name"]
	interval := params["interval"] // possible intervals: minutely, hourly, daily, monthly
	lang := params["lang"]

	fmt.Printf("1.0 postObserveTweets called with accountName: %s, interval: %s, lang: %s \n", accountName, interval, lang)

	// 1. store app to observe
	ok := RESTPostStoreObserveTwitterAccount(ObservableTwitter{AccountName: accountName, Interval: interval, Lang: lang})
	w.Header().Set("Content-Type", "application/json")
	if !ok {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(ResponseMessage{Status: false, Message: "storage layer unreachable"})
		return
	}

	fmt.Printf("1.1 restart observation \n")

	// 2. notify the observer (crawler)
	RestartObservation()

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(ResponseMessage{Status: true, Message: "observation successfully initiated"})
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
	if accountName == "" || lang == "" {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(ResponseMessage{Status: false, Message: "account name or language are empty"})
		return
	}

	processTweets(accountName, lang, fast)

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(ResponseMessage{Status: true, Message: "tweets successfully processed"})
}
