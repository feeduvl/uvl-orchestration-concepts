package main

import (
	"math"
	"testing"
)

func TestUpdateAccount(t *testing.T) {
	// Should not crash
	updateAccount("WindItalia", "it")
}

func TestChunkTweets(t *testing.T) {
	for _, tweetCount := range []int{0, 1, 15, 30, 50} {
		tweets := make([]Tweet, 0, tweetCount)
		for i := 0; i < tweetCount; i += 1 {
			tweets = append(tweets, Tweet{})
		}

		chunks := chunkTweets(tweets)

		expectedChunkCount := int(math.Ceil(float64(tweetCount) / 25.0))
		if len(chunks) != expectedChunkCount {
			t.Errorf("Wrong number of chunks.\nExpected %d but got %d.\n", expectedChunkCount, len(chunks))
		}

		chunkedTweetsCount := 0
		for _, chunk := range chunks {
			chunkedTweetsCount += len(chunk)
		}

		if chunkedTweetsCount != tweetCount {
			t.Errorf("Wrong number of tweets in chunk.\nExpected %d but got %d.\n", tweetCount, chunkedTweetsCount)
		}
	}
}
