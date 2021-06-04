package main

import (
	"bytes"
	"crypto/tls"
	"crypto/x509"
	"encoding/json"
	"io"
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
	endpointPostStartConceptDetection = "/analytics-backend/concepts/detection/"

	// storage layer
	endpointPostStoreDataset         = "/hitec/repository/concepts/store/dataset/"
	endpointPostStoreDetectionResult = "/hitec/repository/concepts/store/detection/result/"
	endpointGetDataset               = "/hitec/repository/concepts/dataset/name/"

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

func createRequest(method string, url string, payload io.Reader) (*http.Request, error) {
	req, err := http.NewRequest(method, url, payload)
	req.Header.Set(AUTHORIZATION, bearerToken)
	req.Header.Add(ACCEPT, TYPE_JSON)
	return req, err
}

// RESTPostStoreDataset returns err
func RESTPostStoreDataset(dataset Dataset) error {
	requestBody := new(bytes.Buffer)
	err := json.NewEncoder(requestBody).Encode(dataset)
	if err != nil {
		log.Printf(errJsonMessageTemplate, err)
		return err
	}
	url := baseURL + endpointPostStoreDataset
	req, _ := createRequest(POST, url, requestBody)
	res, err := client.Do(req)
	if err != nil {
		log.Printf("ERR post store dataset %v\n", err)
		return err
	}
	defer res.Body.Close()

	return nil
}

// RESTGetDataset returns dataset, err
func RESTGetDataset(datasetName string) (Dataset, error) {
	requestBody := new(bytes.Buffer)
	var dataset Dataset

	// make request
	url := baseURL + endpointGetDataset + datasetName
	req, _ := createRequest(GET, url, requestBody)
	res, err := client.Do(req)
	if err != nil {
		log.Printf("ERR get dataset %v\n", err)
		return dataset, err
	}
	// parse result
	err = json.NewDecoder(res.Body).Decode(&dataset)
	if err != nil {
		log.Printf("ERR parsing dataset %v\n", err)
		return dataset, err
	}
	return dataset, err
}

// RESTPostStartNewDetection returns err
func RESTPostStartNewDetection(result Result) (Result, error) {
	requestBody := new(bytes.Buffer)

	err := json.NewEncoder(requestBody).Encode(result)
	if err != nil {
		log.Printf(errJsonMessageTemplate, err)
		return result, err
	}
	url := baseURL + endpointPostStartConceptDetection + result.Method
	req, _ := createRequest(POST, url, requestBody)
	res, err := client.Do(req)
	if err != nil {
		log.Printf("ERR post start new detection %v\n", err)
		return result, err
	}
	defer res.Body.Close()

	// read response and add to result

	return result, nil
}
