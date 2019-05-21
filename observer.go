package main

import (
	"fmt"

	"github.com/robfig/cron"
)

var observableManager map[string]ObservableTwitterManager
var observer *cron.Cron
var observerUnclassifiedTweets *cron.Cron
var observables = NewSet()

func InitObservation() {
	fmt.Println("2.1 initiate observation")
	observableManager = make(map[string]ObservableTwitterManager)

	loadObservables()
	observer = cron.New()

	for accountName := range observables.m {
		if _, ok := observableManager[accountName]; ok {
			continue //
		}
		AddObservable(observables.m[accountName])
	}
	observer.Start()
}

func AddObservable(observable ObservableTwitter) {
	fmt.Printf("[%s] 2.2: add observer\n", observable.AccountName)
	accountName := observable.AccountName
	lang := observable.Lang
	interval := observable.Interval

	if o, observableAlreadyExists := observableManager[accountName]; observableAlreadyExists {
		if o.Observable.isIdentical(accountName, interval, lang) {
			fmt.Printf("[%s] 2.2.1: identical observervable already exists\n", accountName)
			return // cronjob for identical observable already started
		} else {
			fmt.Printf("[%s] 2.2.2: observervable needs to be updated\n", accountName)
			// configuration for a cronjob changed and needs to be updated
			o.CronJob.Stop()
		}
	}

	observableManager[accountName] = ObservableTwitterManager{
		Observable: observable,
		CronJob:    cron.New(),
	}

	fmt.Printf("[%s] 2.2.3: add cron job\n", accountName)
	err := observableManager[accountName].CronJob.AddFunc(getObserverInterval(interval), func() {
		fmt.Printf("[%s] 2.3: crawl tweets\n", accountName)
		crawledTweets := crawlObservableTweets(accountName, lang)
		storeCrawledTweets(crawledTweets)
		fmt.Printf("[%s] 2.4: crawled and stored tweets: %v\n", accountName, len(crawledTweets))
		if len(crawledTweets) == 0 {
			return
		}

		fmt.Printf("[%s] 2.5: classify and store tweets \n", accountName)
		for _, chunkOfTweets := range chunkTweets(crawledTweets) {
			classifiedTweets := classifyTweets(chunkOfTweets, lang)
			storeClassifiedTweets(classifiedTweets)
		}
		fmt.Printf("[%s] 2.6: tweets classified and stored\n", accountName)
	})
	if err != nil {
		fmt.Printf("ERR - could not add %s as observer\nGot error: %v\n---\n", accountName, err)
	}
	observableManager[accountName].CronJob.Start()
}

func RemoveObservable(accountName string) bool {
	fmt.Printf("[%s] 2.1: removeObserver\n", accountName)
	if _, observableExists := observableManager[accountName]; observableExists {
		observableManager[accountName].CronJob.Stop()
		delete(observableManager, accountName)
		return RESTDeleteObservablesTwitterAccounts(observableManager[accountName].Observable)
	}

	fmt.Printf("[%s] 2.2: observer removed\n", accountName)
	return false
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
	storeCrawledTweets(crawledTweets)
	fmt.Printf("1.2. crawled and stored %v tweets: \n\n", len(crawledTweets))

	fmt.Printf("2.1. classify and store tweets: \n")
	for _, chunkOfTweets := range chunkTweets(crawledTweets) {
		classifiedTweets := classifyTweets(chunkOfTweets, lang)
		storeClassifiedTweets(classifiedTweets)
	}
	fmt.Printf("3.2 tweets classified and stored \n\n")
}

func loadObservables() {
	observables = NewSet()
	for _, observable := range RESTGetObservablesTwitterAccounts() {
		observables.Add(observable.AccountName, observable)
	}
	fmt.Printf("2.2 loadObservables lead to these accounts: %v \n", observables)
}

func getObserverInterval(interval string) string {
	switch interval {
	case "minutely":
		return "0 * * * * *"
	case "hourly":
		return "@hourly"
	case "daily":
		return "@daily"
	case "midnight":
		return "@midnight"
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

func storeClassifiedTweets(classifiedTweets []Tweet) {
	RESTPostStoreClassifiedTweets(classifiedTweets)
}

func ObserveUnclassifiedTweets() {
	observerUnclassifiedTweets = cron.New()
	err := observerUnclassifiedTweets.AddFunc(getObserverInterval("midnight"), func() {
		retrieveAndProcessUnclassifiedTweets()
	})
	if err != nil {
		fmt.Println("ERR - could not add the observer for unclassified tweets", err)
	}
	observerUnclassifiedTweets.Start()
}

func retrieveAndProcessUnclassifiedTweets() {
	fmt.Printf("1.0. retrieveAndProcessUnclassifiedTweets \n")
	for _, observable := range RESTGetObservablesTwitterAccounts() {
		fmt.Printf("1.1. get unclassified tweets for %v in lang %v \n", observable.AccountName, observable.Lang)
		tweets := RESTGetUnclassifiedTweets(observable.AccountName, observable.Lang)
		if len(tweets) == 0 {
			fmt.Printf("1.1.1 no unclassfied tweetsfound\n")
			continue
		}
		fmt.Printf("1.2. classify and store %v tweets: \n", len(tweets))
		for _, chunkOfTweets := range chunkTweets(tweets) {
			classifiedTweets := classifyTweets(chunkOfTweets, observable.Lang)
			storeClassifiedTweets(classifiedTweets)
		}
		fmt.Printf("1.3 tweets classified and stored \n\n")
	}
	fmt.Printf("1.4. done retrieveAndProcessUnclassifiedTweets \n")
}
