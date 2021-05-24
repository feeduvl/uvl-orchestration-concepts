package main

import (
	"encoding/json"
)

func saveDataset(dataset Dataset) error {

	sendData, err := json.Marshal(dataset)
	if err != nil {
		return err
	}

	err = RESTPostStoreDataset(sendData)
	if err != nil {
		return err
	}

	return nil
}
