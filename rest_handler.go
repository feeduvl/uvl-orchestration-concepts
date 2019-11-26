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
	"time"

	"log"
)

var baseURL = os.Getenv("BASE_URL")
var bearerToken = "Bearer " + os.Getenv("BEARER_TOKEN")

const (
	// analytics layer
	endpointPostClassificationTwitter = "/ri-analytics-classification-twitter/lang/"
	endpointPostExtractTweetTopics    = "/analytics-backend/tweetClassification"

	// collection layer
	endpointGetCrawlTweets              = "/ri-collection-explicit-feedback-twitter/mention/%s/lang/%s/fast"
	endpointGetCrawlAllAvailableTweets  = "/ri-collection-explicit-feedback-twitter/mention/%s/lang/%s"
	endpointGetTwitterAccountNameExists = "/ri-collection-explicit-feedback-twitter/%s/exists"

	// storage layer
	endpointPostObserveTwitterAccount        = "/ri-storage-twitter/store/observable/"
	endpointGetObservablesTwitterAccounts    = "/ri-storage-twitter/observables"
	endpointDeleteObservablesTwitterAccounts = "/ri-storage-twitter/observables"
	endpointGetUnclassifiedTweets            = "/ri-storage-twitter/account_name/%s/lang/%s/unclassified"
	endpointPostTweet                        = "/ri-storage-twitter/store/tweet/"
	endpointPostClassifiedTweet              = "/ri-storage-twitter/store/classified/tweet/"
	endpointPostTweetTopics                  = "/ri-storage-twitter/store/topics"

	GET           = "GET"
	POST          = "POST"
	DELETE        = "DELETE"
	AUTHORIZATION = "Authorization"
	ACCEPT        = "Accept"
	TYPE_JSON     = "application/json"

	errJsonMessageTemplate = "ERR - json formatting error: %v\n"
)

var client = getHTTPClient()

func getHTTPClient() *http.Client {
	pwd, _ := os.Getwd()
	caCert, err := ioutil.ReadFile(pwd + "/ca_chain.crt")
	if err != nil {
		log.Fatal(err)
	}
	caCertPool := x509.NewCertPool()
	caCertPool.AppendCertsFromPEM(caCert)
	timeout := time.Duration(4 * time.Minute)

	client := &http.Client{
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{
				RootCAs: caCertPool,
				// InsecureSkipVerify: true,
			},
		},
		Timeout: timeout,
		CheckRedirect: func(req *http.Request, _ []*http.Request) error {
			req.Header.Add(AUTHORIZATION, bearerToken)
			return nil
		},
	}

	return client
}

// RESTPostStoreObserveTwitterAccount returns ok
func RESTPostStoreObserveTwitterAccount(obserable ObservableTwitter) bool {
	requestBody := new(bytes.Buffer)
	err := json.NewEncoder(requestBody).Encode(obserable)
	if err != nil {
		log.Printf(errJsonMessageTemplate, err)
	}

	url := baseURL + endpointPostObserveTwitterAccount
	req, _ := http.NewRequest(POST, url, requestBody)
	req.Header.Set(AUTHORIZATION, bearerToken)
	req.Header.Add(ACCEPT, TYPE_JSON)
	res, err := client.Do(req)
	if err != nil {
		log.Printf("ERR post store observable %v\n", err)
	}
	defer res.Body.Close()

	return res.StatusCode == 200
}

// RESTGetObservablesTwitterAccounts retrieve all observables from the storage layer
func RESTGetObservablesTwitterAccounts() []ObservableTwitter {
	var obserables []ObservableTwitter

	url := baseURL + endpointGetObservablesTwitterAccounts

	req, _ := http.NewRequest(GET, url, bytes.NewBuffer(nil))
	req.Header.Set(AUTHORIZATION, bearerToken)
	req.Header.Add(ACCEPT, TYPE_JSON)
	res, err := client.Do(req)
	if err != nil {
		fmt.Println("ERR cannot send observable account get request", err)
		return obserables
	}
	defer res.Body.Close()

	err = json.NewDecoder(res.Body).Decode(&obserables)
	if err != nil {
		fmt.Println("ERR cannot decode twitter observable json", err)
		return obserables
	}

	return obserables
}

// RESTDeleteObservablesTwitterAccounts returns ok
func RESTDeleteObservablesTwitterAccounts(observable ObservableTwitter) bool {
	requestBody := new(bytes.Buffer)
	err := json.NewEncoder(requestBody).Encode(observable)
	if err != nil {
		log.Printf(errJsonMessageTemplate, err)
	}

	url := baseURL + endpointDeleteObservablesTwitterAccounts
	req, _ := http.NewRequest(DELETE, url, requestBody)
	req.Header.Set(AUTHORIZATION, bearerToken)
	req.Header.Add(ACCEPT, TYPE_JSON)
	res, err := client.Do(req)
	if err != nil {
		log.Printf("ERR cannot send request to delte observable %v\n", err)
	}
	defer res.Body.Close()

	return res.StatusCode == 200
}

// RESTGetTwitterAccountNameExists check if a twitter account exists
func RESTGetTwitterAccountNameExists(accountName string) CrawlerResponseMessage {
	var response CrawlerResponseMessage

	endpoint := fmt.Sprintf(endpointGetTwitterAccountNameExists, accountName)
	url := baseURL + endpoint
	req, _ := http.NewRequest(GET, url, bytes.NewBuffer(nil))
	req.Header.Set(AUTHORIZATION, bearerToken)
	req.Header.Add(ACCEPT, TYPE_JSON)
	res, err := client.Do(req)
	if err != nil {
		fmt.Println("ERR cannot send get request to check if Twitter account exists", err)
		return response
	}
	defer res.Body.Close()

	err = json.NewDecoder(res.Body).Decode(&response)
	if err != nil {
		fmt.Println("ERR cannot decode response from the twitter crawler json", err)
		return response
	}

	return response
}

// RESTGetUnclassifiedTweets retrieve all tweets from a specified account that have not been classified yet
func RESTGetUnclassifiedTweets(accountName, lang string) []Tweet {
	var tweet []Tweet

	endpoint := fmt.Sprintf(endpointGetUnclassifiedTweets, accountName, lang)
	url := baseURL + endpoint
	req, _ := http.NewRequest(GET, url, bytes.NewBuffer(nil))
	req.Header.Set(AUTHORIZATION, bearerToken)
	req.Header.Add(ACCEPT, TYPE_JSON)
	res, err := client.Do(req)
	if err != nil {
		fmt.Println("ERR cannot send get request to get unclassified tweets", err)
		return tweet
	}
	defer res.Body.Close()

	err = json.NewDecoder(res.Body).Decode(&tweet)
	if err != nil {
		fmt.Println("ERR cannot decode unclassified tweets json", err)
		return tweet
	}

	return tweet
}

// RESTGetCrawlTweets retrieve all tweets from the collection layer that addresses the given account name
func RESTGetCrawlTweets(accountName string, lang string) []Tweet {
	var tweets []Tweet

	endpoint := fmt.Sprintf(endpointGetCrawlTweets, accountName, lang)
	url := baseURL + endpoint
	req, _ := http.NewRequest(GET, url, bytes.NewBuffer(nil))
	req.Header.Set(AUTHORIZATION, bearerToken)
	req.Header.Add(ACCEPT, TYPE_JSON)
	res, err := client.Do(req)
	if err != nil {
		fmt.Println("ERR cannot send request to tweet crawler", err)
		return tweets
	}
	defer res.Body.Close()

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
	req, _ := http.NewRequest(GET, url, bytes.NewBuffer(nil))
	req.Header.Set(AUTHORIZATION, bearerToken)
	req.Header.Add(ACCEPT, TYPE_JSON)
	res, err := client.Do(req)
	if err != nil {
		fmt.Println("ERR crawl max number of tweets", err)
		return tweets
	}
	defer res.Body.Close()

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
		log.Printf(errJsonMessageTemplate, err)
	}

	url := baseURL + endpointPostClassificationTwitter + lang
	req, _ := http.NewRequest(POST, url, requestBody)
	req.Header.Set(AUTHORIZATION, bearerToken)
	req.Header.Add(ACCEPT, TYPE_JSON)
	res, err := client.Do(req)
	if err != nil {
		log.Printf("ERR %v\n", err)
		return classifiedTweets
	}
	defer res.Body.Close()

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
		log.Printf(errJsonMessageTemplate, err)
	}

	url := baseURL + endpointPostTweet
	req, _ := http.NewRequest(POST, url, requestBody)
	req.Header.Set(AUTHORIZATION, bearerToken)
	req.Header.Add(ACCEPT, TYPE_JSON)
	res, err := client.Do(req)
	if err != nil {
		log.Printf("ERR cannot send request to store tweets %v\n", err)
	}
	defer res.Body.Close()

	return res.StatusCode == 200
}

// RESTPostStoreClassifiedTweets returns ok
func RESTPostStoreClassifiedTweets(tweets []Tweet) bool {
	requestBody := new(bytes.Buffer)
	err := json.NewEncoder(requestBody).Encode(tweets)
	if err != nil {
		log.Printf(errJsonMessageTemplate, err)
	}

	url := baseURL + endpointPostClassifiedTweet
	req, _ := http.NewRequest(POST, url, requestBody)
	req.Header.Set(AUTHORIZATION, bearerToken)
	req.Header.Add(ACCEPT, TYPE_JSON)
	res, err := client.Do(req)
	if err != nil {
		log.Printf("ERR cannot send request to store tweets %v\n", err)
	}
	defer res.Body.Close()

	return res.StatusCode == 200
}

// RESTPostExtractTweetTopics returns ok
func RESTPostExtractTweetTopics(tweet Tweet) TweetTopics {
	var tweetTopics TweetTopics

	requestBody := new(bytes.Buffer)
	err := json.NewEncoder(requestBody).Encode(TweetTopicExtractionPayload{Message: tweet.Text})
	if err != nil {
		log.Printf(errJsonMessageTemplate, err)
	}

	url := baseURL + endpointPostExtractTweetTopics
	req, _ := http.NewRequest(POST, url, requestBody)
	req.Header.Set(AUTHORIZATION, bearerToken)
	req.Header.Add(ACCEPT, TYPE_JSON)
	res, err := client.Do(req)
	if err != nil {
		log.Printf("ERR %v\n", err)
	}
	defer res.Body.Close()

	err = json.NewDecoder(res.Body).Decode(&tweetTopics)
	if err != nil {
		log.Printf("ERR cannot decode tweets with topics %v\n", err)
	}

	return tweetTopics
}

// RESTPostStoreTweetTopics returns ok
func RESTPostStoreTweetTopics(tweet Tweet) bool {
	requestBody := new(bytes.Buffer)
	err := json.NewEncoder(requestBody).Encode(tweet)
	if err != nil {
		log.Printf(errJsonMessageTemplate, err)
	}

	url := baseURL + endpointPostTweetTopics
	req, _ := http.NewRequest(POST, url, requestBody)
	req.Header.Set(AUTHORIZATION, bearerToken)
	req.Header.Add(ACCEPT, TYPE_JSON)
	res, err := client.Do(req)
	if err != nil {
		log.Printf("ERR cannot send request to store tweet topics %v\n", err)
	}
	defer res.Body.Close()

	return res.StatusCode == 200
}
