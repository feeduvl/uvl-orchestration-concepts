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
	endpointPostStartConceptDetection             = "/hitec/classify/concepts/"
	endpointPostStartAcceptanceCriteriaGeneration = "/hitec/generate/"

	// storage layer
	endpointPostStoreDataset         = "/hitec/repository/concepts/store/dataset/"
	endpointPostStoreGroundTruth     = "/hitec/repository/concepts/store/groundtruth/"
	endpointPostStoreDetectionResult = "/hitec/repository/concepts/store/detection/result/"
	endpointGetDataset               = "/hitec/repository/concepts/dataset/name/"
	// annotation
	endpointPostStoreAnnotation    = "/hitec/repository/concepts/store/annotation/"
	endpointPostAnnotationTokenize = "/hitec/annotation/tokenize/"

	// agreement
	endpointPostStoreAgreement            = "/hitec/repository/concepts/store/agreement/"
	endpointInfoFromAnnotations           = "/hitec/agreement/annotationinfo/"
	endpointCreateAnnotationFromAgreement = "/hitec/agreement/annotationexport/"
	endpointCalculateKappaFromAgreement   = "/hitec/agreement/calculateKappa/"

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
	timeout := 15 * time.Minute

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
	req, _ := http.NewRequest(method, url, payload)
	req.Header.Set(AUTHORIZATION, bearerToken)
	req.Header.Add(ACCEPT, TYPE_JSON)
	return req, nil
}

// RESTPostStoreAnnotation returns err
func RESTPostStoreAnnotation(annotation Annotation) error {
	requestBody := new(bytes.Buffer)
	_ = json.NewEncoder(requestBody).Encode(annotation)
	url := baseURL + endpointPostStoreAnnotation
	req, _ := createRequest(POST, url, requestBody)
	res, err := client.Do(req)
	if err != nil {
		log.Printf("ERR post store annotation %v\n", err)
		return err
	}
	defer func(Body io.ReadCloser) {
		_ = Body.Close()
	}(res.Body)

	return nil
}

// RESTPostStoreAgreement returns err
func RESTPostStoreAgreement(agreement Agreement) error {
	requestBody := new(bytes.Buffer)
	_ = json.NewEncoder(requestBody).Encode(agreement)
	url := baseURL + endpointPostStoreAgreement
	req, _ := createRequest(POST, url, requestBody)
	res, err := client.Do(req)
	if err != nil {
		log.Printf("ERR post store agreement %v\n", err)
		return err
	}
	defer func(Body io.ReadCloser) {
		_ = Body.Close()
	}(res.Body)

	return nil
}

// RESTPostStoreDataset returns err
func RESTPostStoreDataset(dataset Dataset) error {
	requestBody := new(bytes.Buffer)
	_ = json.NewEncoder(requestBody).Encode(dataset)
	url := baseURL + endpointPostStoreDataset
	req, _ := createRequest(POST, url, requestBody)
	res, err := client.Do(req)
	if err != nil {
		log.Printf("ERR post store dataset %v\n", err)
		return err
	}
	defer func(Body io.ReadCloser) {
		_ = Body.Close()
	}(res.Body)

	return nil
}

// RESTPostStoreGroundTruth returns err
func RESTPostStoreGroundTruth(dataset Dataset) error {
	requestBody := new(bytes.Buffer)
	_ = json.NewEncoder(requestBody).Encode(dataset)

	url := baseURL + endpointPostStoreGroundTruth
	req, _ := createRequest(POST, url, requestBody)
	res, err := client.Do(req)
	if err != nil {
		log.Printf("ERR post store groundtruth %v\n", err)
		return err
	}
	defer func(Body io.ReadCloser) {
		_ = Body.Close()
	}(res.Body)

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

// RESTPostStartNewDetection returns Result ,err
func RESTPostStartNewDetection(result Result, run Run) (Result, error) {
	requestBody := new(bytes.Buffer)

	_ = json.NewEncoder(requestBody).Encode(run)

	url := baseURL + endpointPostStartConceptDetection + run.Method + "/run"
	if run.Method == "acceptance-criteria" {
		url = baseURL + endpointPostStartAcceptanceCriteriaGeneration + run.Method + "/run"
	}
	log.Printf("PostStartNewDetection url: %s\n", url)
	req, _ := createRequest(POST, url, requestBody)
	res, err := client.Do(req)
	if err != nil {
		log.Printf("ERR post start new detection %v\n", err)
		log.Printf("Note: If the request timed out, the method microservice may take too long to process the" +
			" request. Consider increasing timeout in rest_handler->getHTTPClient.")

		return result, err
	}
	defer func(Body io.ReadCloser) {
		_ = Body.Close()
	}(res.Body)

	_res := new(Result)
	err = json.NewDecoder(res.Body).Decode(&_res)
	if err != nil {
		log.Printf("ERR parsing response %v\n", err)
		return result, err
	}

	result.Topics = _res.Topics
	result.DocTopic = _res.DocTopic
	result.Metrics = _res.Metrics

	return result, nil
}

// RESTPostStoreResult returns ,err
func RESTPostStoreResult(result Result) error {
	requestBody := new(bytes.Buffer)
	_ = json.NewEncoder(requestBody).Encode(result)

	url := baseURL + endpointPostStoreDetectionResult
	req, _ := createRequest(POST, url, requestBody)
	res, err := client.Do(req)
	if err != nil {
		log.Printf("ERR post store result %v\n", err)
		return err
	}
	defer func(Body io.ReadCloser) {
		_ = Body.Close()
	}(res.Body)

	return nil
}

// RESTGetInfoFromAnnotations returns info from annotation,
// including docs, tokens, toreAlternatives, wordCodeAlternatives, relationshipAlternatives, err
func RESTGetInfoFromAnnotations(
	annotationNames []string,
	completeConcurrences bool,
) (
	RelevantAgreementFields,
	error,
) {
	var relevantAgreementFields RelevantAgreementFields

	// make request
	requestBody := new(bytes.Buffer)
	var data = map[string]interface{}{
		"annotationNames":      annotationNames,
		"completeConcurrences": completeConcurrences,
	}
	_ = json.NewEncoder(requestBody).Encode(data)

	url := baseURL + endpointInfoFromAnnotations
	req, _ := createRequest(POST, url, requestBody)
	res, err := client.Do(req)
	if err != nil {
		log.Printf("ERR get annotation %v\n", err)
		return relevantAgreementFields, err
	}
	// parse result
	err = json.NewDecoder(res.Body).Decode(&relevantAgreementFields)
	if err != nil {
		log.Printf("ERR parsing dataset %v\n", err)
		return relevantAgreementFields, err
	}
	return relevantAgreementFields, err
}

// RESTCreateAnnotationFromAgreement returns error,
func RESTCreateAnnotationFromAgreement(
	agreementName string,
	newAnnotationName string,
) error {
	// make request
	requestBody := new(bytes.Buffer)
	var data = map[string]interface{}{
		"agreementName":     agreementName,
		"newAnnotationName": newAnnotationName,
	}
	_ = json.NewEncoder(requestBody).Encode(data)

	url := baseURL + endpointCreateAnnotationFromAgreement
	req, _ := createRequest(POST, url, requestBody)
	res, err := client.Do(req)
	if err != nil {
		log.Printf("ERR creating annotation from agreement %v\n", err)
		return err
	}
	defer func(Body io.ReadCloser) {
		_ = Body.Close()
	}(res.Body)

	return nil
}

// RESTCalculateKappaFromAgreement returns kappas, error
func RESTCalculateKappaFromAgreement(
	agreement Agreement,
) (map[string]float64, error) {
	requestBody := new(bytes.Buffer)
	_ = json.NewEncoder(requestBody).Encode(agreement)

	var data map[string]float64

	url := baseURL + endpointCalculateKappaFromAgreement
	req, _ := createRequest(POST, url, requestBody)
	res, err := client.Do(req)
	if err != nil {
		log.Printf("ERR get annotation %v\n", err)
		return data, err
	}
	// parse result
	err = json.NewDecoder(res.Body).Decode(&data)
	if err != nil {
		log.Printf("ERR parsing dataset %v\n", err)
		return data, err
	}
	return data, err
}
