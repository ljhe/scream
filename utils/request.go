package utils

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
)

func Post(url string, data map[string]interface{}) (error, []byte) {
	jsonData, err := json.Marshal(data)
	if err != nil {
		return err, nil
	}
	resp, err := http.Post(url, "application/json", bytes.NewBuffer(jsonData))
	if err != nil {
		return err, nil
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return err, nil
	}

	return nil, body
}

func Get() {

}
