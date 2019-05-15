package main

// Tweet model
type Tweet struct {
	CreatedAt           int      `json:"created_at"`
	CreatedAtFull       string   `json:"created_at_full"`
	FavoriteCount       int      `json:"favorite_count"`
	RetweetCount        int      `json:"retweet_count"`
	Text                string   `json:"text"`
	StatusID            string   `json:"status_id"`
	UserName            string   `json:"user_name"`
	InReplyToScreenName string   `json:"in_reply_to_screen_name"`
	Hashtags            []string `json:"hashtags"`
	Lang                string   `json:"lang"`
	Sentiment           string   `json:"sentiment" bson:"sentiment"`
	SentimentScore      int      `json:"sentiment_score" bson:"sentiment_score"`
	TweetClass          string   `json:"tweet_class"`
	ClassifierCertainty int      `json:"classifier_certainty"`
	Annotated           bool     `json:"is_annotated" bson:"is_annotated"`
}

// ObservableTwitter model
type ObservableTwitter struct {
	AccountName string `json:"account_name"`
	Interval    string `json:"interval"`
	Lang        string `json:"lang"`
}

// ResponseMessage model
type ResponseMessage struct {
	Message string `json:"message"`
	Status  bool   `json:"status"`
}

type CrawlerResponseMessage struct {
	AccountExists bool   `json:"account_exists"`
	Message       string `json:"message"`
}
