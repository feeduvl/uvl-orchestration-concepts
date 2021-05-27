package main

func saveDataset(dataset Dataset) error {

	//sendData, err := json.Marshal(dataset)
	//if err != nil {
	//	return err
	//}

	err := RESTPostStoreDataset(dataset)
	if err != nil {
		return err
	}

	return nil
}
