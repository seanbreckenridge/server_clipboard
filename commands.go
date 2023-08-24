package server_clipboard

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

func Copy(password string, serverAddress string, clipboard string) (string, error) {

	// construct body
	postModel := CopyInput{Text: clipboard}
	postBody, err := json.Marshal(postModel)
	if err != nil {
		return "", err
	}
	buf := bytes.NewBuffer(postBody)

	// construct request
	req, err := http.NewRequest("POST", fmt.Sprintf("%s/copy", serverAddress), buf)
	if err != nil {
		return "", err
	}
	req.Header.Set("accept", "application/json")
	req.Header.Set("password", password)

	// send the HTTP request
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	// read response
	respBody, err := io.ReadAll(resp.Body)
	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("error: %s", string(respBody))
	}
	return string(respBody), nil
}

func Paste(password string, serverAddress string) (string, error) {
	// send request
	req, err := http.NewRequest("GET", fmt.Sprintf("%s/paste", serverAddress), nil)
	if err != nil {
		return "", err
	}
	req.Header.Set("password", password)

	// send the HTTP request
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	// read response
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("error: %s", string(body))
	}

	return string(body), nil
}
