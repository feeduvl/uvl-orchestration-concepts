package main

import (
	"fmt"

	"github.com/robfig/cron"
)

var observableTwitterAccounts = NewSet()
var observer *cron.Cron

func startObsevation() {
	fmt.Println("2.1 initiate observation")
	loadObservableTwitterAccounts()

	// TODO: consider a real interval in the future. ATM I have to take a fixed interval and start the process
	// for alll accounts and tweets in sequential order, because the classification MS cannot handle simultaneuos
	// requests ATM
	observationInterval := getObserverInterval("2h")
	observer = cron.New()
	err := observer.AddFunc(observationInterval, func() {
		fmt.Printf("cron job triggered")
		for accountName := range observableTwitterAccounts.m {
			observable := observableTwitterAccounts.m[accountName]

			fmt.Printf("2.3.0 crawl tweets %s \n", observationInterval)
			crawledTweets := crawlObservableTweets(observable.AccountName, observable.Lang)
			fmt.Printf("2.3.1 crawled tweets: %v for %v \n", len(crawledTweets), observable.AccountName)
			if len(crawledTweets) == 0 {
				continue
			}

			fmt.Printf("2.3.2 classify and store tweets \n")
			for _, chunkOfTweets := range chunkTweets(crawledTweets) {
				classifiedTweets := classifyTweets(chunkOfTweets, observable.Lang)
				storeCrawledTweets(classifiedTweets)
			}
			fmt.Printf("2.3.3 tweets classified and stored \n")
		}
	})
	if err != nil {
		fmt.Println("ERR - could not add the observer", err)
	}
	observer.Start()
}

func processTweets(accountName, lang, fast string) {
	fmt.Printf("0.0. postProcessTweets called with accountName: %s, lang: %s \n", accountName, lang)
	fmt.Printf("1.1. crawl tweets for %s %s \n", accountName, fast)
	var crawledTweets []Tweet
	if fast == "fast" {
		crawledTweets = crawlObservableTweets(accountName, lang)
	} else {
		crawledTweets = RESTGetCrawlMaximumNumberOfTweets(accountName, lang)
	}
	fmt.Printf("1.2. crawled %v tweets: \n\n", len(crawledTweets))

	fmt.Printf("2.1. classify and store tweets: \n")
	for _, chunkOfTweets := range chunkTweets(crawledTweets) {
		classifiedTweets := classifyTweets(chunkOfTweets, lang)
		storeCrawledTweets(classifiedTweets)
	}
	fmt.Printf("3.2 tweets classified and stored \n\n")
}

func stopObservation() {
	fmt.Printf("2.0 stop observation \n")
	observer.Stop()
}

func loadObservableTwitterAccounts() {
	observableTwitterAccounts = NewSet()
	for _, observable := range RESTGetObservablesTwitterAccounts() {
		observableTwitterAccounts.Add(observable.AccountName, observable)
	}
	fmt.Printf("2.2 loadObservableTwitterAccounts lead to these accounts: %v \n", observableTwitterAccounts)
}

func getObserverInterval(interval string) string {
	switch interval {
	case "minutely":
		return "0 * * * * *"
	case "hourly":
		return "@hourly"
	case "daily":
		return "@daily"
	case "weekly":
		return "@weekly"
	case "monthly":
		return "@monthly"
	case "6h":
		return "@every 6h0m0s"
	case "2h":
		return "@every 2h0m0s"
	default:
		return interval // allows custom intervals to the cron job specification (https://godoc.org/github.com/robfig/cron) might thorw an error if the custom interval is wrong
	}
}

func crawlObservableTweets(accountName string, lang string) []Tweet {
	return RESTGetCrawlTweets(accountName, lang)
}

func chunkTweets(tweets []Tweet) [][]Tweet {
	var chunks [][]Tweet
	var chunk []Tweet
	for len(tweets) > 0 {
		a := tweets[len(tweets)-1]
		tweets = tweets[:len(tweets)-1]
		chunk = append(chunk, a)

		if len(chunk) == 25 {
			c := make([]Tweet, len(chunk))
			copy(c, chunk)
			chunks = append(chunks, c)
			chunk = chunk[:0]
		}
	}

	return chunks
}

func classifyTweets(tweets []Tweet, lang string) []Tweet {
	return RESTPostClassifyTweets(tweets, lang)
}

func storeCrawledTweets(crawledTweets []Tweet) {
	RESTPostStoreTweets(crawledTweets)
}

// RestartObservation stops the observation and starts it again
func RestartObservation() {
	if observer != nil {
		stopObservation()
	}
	startObsevation()
}
