package main

import (
	"crypto/tls"
	"crypto/x509"
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
	endpointConceptDetection = "/analytics-backend/concepts/detection/"

	// storage layer
	endpointStoreDataset         = "/hitec/repository/concepts/store/dataset/"
	endpointStoreDetectionResult = "/hitec/repository/concepts/store/detection/result/"

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
