package client

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"os"
	"path/filepath"

	"github.com/txsvc/httpservice/pkg/api"

	"github.com/podops/podops"
	"github.com/podops/podops/config"
)

// Get is used to request data from the API. No payload, only queries!
func Get(url, cmd, token string, response interface{}) (int, error) {
	uri := url + cmd

	req, err := http.NewRequest("GET", uri, nil)
	if err != nil {
		return http.StatusBadRequest, err
	}

	return invoke(token, req, response)
}

// Post is used to invoke an API method using http POST
func Post(url, cmd, token string, request, response interface{}) (int, error) {
	uri := url + cmd

	payload, err := json.Marshal(&request)
	if err != nil {
		return http.StatusInternalServerError, err
	}

	req, err := http.NewRequest("POST", uri, bytes.NewBuffer(payload))
	if err != nil {
		return http.StatusBadRequest, err
	}

	return invoke(token, req, response)
}

// Put is used to invoke an API method using http PUT
func Put(url, cmd, token string, request, response interface{}) (int, error) {
	var req *http.Request
	uri := url + cmd

	if request != nil {
		p, err := json.Marshal(&request)
		if err != nil {
			return http.StatusInternalServerError, err
		}

		req, err = http.NewRequest("PUT", uri, bytes.NewBuffer(p))
		if err != nil {
			return http.StatusBadRequest, err
		}
	} else {
		var err error
		req, err = http.NewRequest("PUT", uri, nil)
		if err != nil {
			return http.StatusBadRequest, err
		}
	}

	return invoke(token, req, response)
}

// DELETE is used to request the deletion of a resource. Maybe apayload, no response!
func Del(url, cmd, token string, request interface{}) (int, error) {
	var req *http.Request
	uri := url + cmd

	if request != nil {
		p, err := json.Marshal(&request)
		if err != nil {
			return http.StatusInternalServerError, err
		}

		req, err = http.NewRequest("DELETE", uri, bytes.NewBuffer(p))
		if err != nil {
			return http.StatusBadRequest, err
		}
	} else {
		var err error
		req, err = http.NewRequest("DELETE", uri, nil)
		if err != nil {
			return http.StatusBadRequest, err
		}
	}

	return invoke(token, req, nil)
}

// Creates a new file upload http request with optional extra params
func UploadRequest(url, token, form, path string) (*http.Request, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	body := bytes.Buffer{}
	writer := multipart.NewWriter(&body)
	part, err := writer.CreateFormFile(form, filepath.Base(path))
	if err != nil {
		return nil, err
	}

	_, err = io.Copy(part, file)
	if err != nil {
		return nil, err
	}

	err = writer.Close()
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequest("POST", url, &body)
	req.Header.Set("Content-Type", writer.FormDataContentType())
	req.Header.Set("User-Agent", config.UserAgentString)
	if token != "" {
		req.Header.Set("Authorization", "Bearer "+token)
	}

	return req, err
}

func invoke(token string, req *http.Request, response interface{}) (int, error) {

	req.Header.Set("Content-Type", "application/json; charset=utf-8")
	req.Header.Set("User-Agent", config.UserAgentString)
	if token != "" {
		req.Header.Set("Authorization", "Bearer "+token)
	}

	// perform the request
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		if resp == nil {
			return http.StatusInternalServerError, err
		}
		return resp.StatusCode, err
	}

	defer resp.Body.Close()

	// anything other than OK, Created, Accepted, NoContent is treated as an error
	if resp.StatusCode > http.StatusNoContent {
		if response != nil {
			// as we expect a response, there might be a StatusObject
			status := api.StatusObject{}
			err = json.NewDecoder(resp.Body).Decode(&status)
			if err != nil {
				return resp.StatusCode, fmt.Errorf(podops.MsgStatus, resp.StatusCode)
			}
			return status.Status, fmt.Errorf(status.Message)
		}
		return resp.StatusCode, podops.ErrApiError
	}

	// unmarshal the response if one is expected
	if response != nil {
		err = json.NewDecoder(resp.Body).Decode(response)
		if err != nil {
			return http.StatusInternalServerError, err
		}
	}

	return resp.StatusCode, nil
}
