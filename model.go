package main

import (
	"time"
)

// Dataset model
type Dataset struct {
	UploadedAt  time.Time      `json:"uploaded_at"`
	Name        string         `json:"name"`
	Size        int            `json:"size"`
	Documents   []Document     `json:"documents"`
	GroundTruth []TruthElement `json:"ground_truth" bson:"ground_truth"`
}

//TruthElement model
type TruthElement struct {
	Id    string `json:"id" bson:"id"`
	Value string `json:"value"  bson:"value"`
}

// Document model
type Document struct {
	Number int    `json:"number"`
	Text   string `json:"text"`
	Id     string `json:"id"`
}

// Result model
type Result struct {
	Method      string                 `json:"method"`
	Status      string                 `json:"status"`
	StartedAt   time.Time              `json:"started_at"`
	DatasetName string                 `json:"dataset_name"`
	Params      map[string]string      `json:"params"`
	Topics      map[string]interface{} `json:"topics"`
	DocTopic    map[string]interface{} `json:"doc_topic"`
	Metrics     map[string]interface{} `json:"metrics"`
	Name        string                 `json:"name"`
}

// Run model
type Run struct {
	Method  string            `json:"method"`
	Dataset Dataset           `json:"dataset"`
	Params  map[string]string `json:"params"`
}

// ObservableTwitter model
type ObservableTwitter struct {
	AccountName string `json:"account_name"`
	Interval    string `json:"interval"`
	Lang        string `json:"lang"`
}

func (o ObservableTwitter) isIdentical(accountName, interval, lang string) bool {
	if o.AccountName == accountName && o.Interval == interval && o.Lang == lang {
		return true
	}
	return false
}

// ResponseMessage model
type ResponseMessage struct {
	Message string `json:"message"`
	Status  bool   `json:"status"`
}
