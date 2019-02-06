package main

import (
	"bytes"
	"crypto/tls"
	"crypto/x509"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"path/filepath"

	"log"
)

var baseURL = os.Getenv("BASE_URL")

const (
	// analytics layer
	endpointPostClassificationTwitter = "/ri-analytics-classification-twitter/lang/"

	// collection layer
	endpointGetCrawlTweets             = "/ri-collection-explicit-feedback-twitter/mention/%s/lang/%s/fast"
	endpointGetCrawlAllAvailableTweets = "/ri-collection-explicit-feedback-twitter/mention/%s/lang/%s"

	// storage layer
	endpointPostObserveTwitterAccount     = "/ri-storage-twitter/store/observable/"
	endpointGetObservablesTwitterAccounts = "/ri-storage-twitter/observables"
	endpointPostTweet                     = "/ri-storage-twitter/store/tweet/"
)

var client = getHTTPClient()

func getHTTPClient() *http.Client {
	ex, err := os.Executable()
	if err != nil {
		panic(err)
	}
	exPath := filepath.Dir(ex)
	caCert, err := ioutil.ReadFile(exPath + "ca_chain.crt")
	if err != nil {
		log.Fatal(err)
	}
	caCertPool := x509.NewCertPool()
	caCertPool.AppendCertsFromPEM(caCert)

	client := &http.Client{
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{
				RootCAs: caCertPool,
			},
		},
	}

	return client
}

// RESTPostStoreObserveTwitterAccount returns ok
func RESTPostStoreObserveTwitterAccount(obserable ObservableTwitter) bool {
	requestBody := new(bytes.Buffer)
	err := json.NewEncoder(requestBody).Encode(obserable)
	if err != nil {
		log.Printf("ERR - json formatting error: %v\n", err)
	}

	url := baseURL + endpointPostObserveTwitterAccount
	res, err := client.Post(url, "application/json; charset=utf-8", requestBody)
	if err != nil {
		log.Printf("ERR post store observable %v\n", err)
	}
	if res.StatusCode == 200 {
		return true
	}

	return false
}

// RESTGetObservablesTwitterAccounts retrieve all observables from the storage layer
func RESTGetObservablesTwitterAccounts() []ObservableTwitter {
	var obserables []ObservableTwitter

	url := baseURL + endpointGetObservablesTwitterAccounts
	res, err := client.Get(url)
	if err != nil {
		fmt.Println("ERR cannot send observable account get request", err)
		return obserables
	}

	err = json.NewDecoder(res.Body).Decode(&obserables)
	if err != nil {
		fmt.Println("ERR cannot decode twitter observable json", err)
		return obserables
	}

	return obserables
}

// RESTGetCrawlTweets retrieve all tweets from the collection layer that addresses the given account name
func RESTGetCrawlTweets(accountName string, lang string) []Tweet {
	var tweets []Tweet

	endpoint := fmt.Sprintf(endpointGetCrawlTweets, accountName, lang)
	url := baseURL + endpoint
	res, err := client.Get(url)
	if err != nil {
		fmt.Println("ERR cannot send request to tweet crawler", err)
		return tweets
	}

	err = json.NewDecoder(res.Body).Decode(&tweets)
	if err != nil {
		fmt.Println("ERR cannot decode crawled tweets", err)
		return tweets
	}

	return tweets
}

func RESTGetCrawlMaximumNumberOfTweets(accountName string, lang string) []Tweet {
	var tweets []Tweet

	endpoint := fmt.Sprintf(endpointGetCrawlAllAvailableTweets, accountName, lang)
	url := baseURL + endpoint
	res, err := client.Get(url)
	if err != nil {
		fmt.Println("ERR crawl max number of tweets", err)
		return tweets
	}

	err = json.NewDecoder(res.Body).Decode(&tweets)
	if err != nil {
		fmt.Println("ERR cannot decode tweets", err)
		return tweets
	}

	return tweets
}

// RESTPostClassifyTweets returns ok
func RESTPostClassifyTweets(tweets []Tweet, lang string) []Tweet {
	var classifiedTweets []Tweet

	requestBody := new(bytes.Buffer)
	err := json.NewEncoder(requestBody).Encode(tweets)
	if err != nil {
		log.Printf("ERR - json formatting error: %v\n", err)
	}

	url := baseURL + endpointPostClassificationTwitter + lang
	res, err := client.Post(url, "application/json; charset=utf-8", requestBody)
	if err != nil {
		log.Printf("ERR %v\n", err)
	}

	err = json.NewDecoder(res.Body).Decode(&classifiedTweets)
	if err != nil {
		log.Printf("ERR cannot decode classified tweets %v\n", err)
	}

	return classifiedTweets
}

// RESTPostStoreTweets returns ok
func RESTPostStoreTweets(tweets []Tweet) bool {
	requestBody := new(bytes.Buffer)
	err := json.NewEncoder(requestBody).Encode(tweets)
	if err != nil {
		log.Printf("ERR - json formatting error: %v\n", err)
	}

	url := baseURL + endpointPostTweet
	res, err := client.Post(url, "application/json; charset=utf-8", requestBody)
	if err != nil {
		log.Printf("ERR cannot send request to store tweets %v\n", err)
	}
	if res.StatusCode == 200 {
		return true
	}

	return false
}
