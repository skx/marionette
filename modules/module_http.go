package modules

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"net/http"
	"strconv"
	"strings"

	"github.com/skx/marionette/config"
	"github.com/skx/marionette/environment"
)

// HTTPModule stores our state.
type HTTPModule struct {

	// cfg contains our configuration object.
	cfg *config.Config

	// Save the status-code, after our request was completed.
	statusCode int

	// Save the status-line, after our request was completed.
	statusLine string

	// Save the body, after our request was completed
	body string
}

// Check is part of the module-api, and checks arguments.
func (f *HTTPModule) Check(args map[string]interface{}) error {

	// Ensure we have a url to request.
	_, ok := args["url"]
	if !ok {
		return fmt.Errorf("missing 'url' parameter")
	}

	// Ensure the url is a string.
	url := StringParam(args, "url")
	if url == "" {
		return fmt.Errorf("failed to convert 'url' to string")
	}

	return nil
}

// Execute is part of the module-api, and is invoked to run a rule.
func (f *HTTPModule) Execute(env *environment.Environment, args map[string]interface{}) (bool, error) {

	// Get the url.
	url := StringParam(args, "url")
	if url == "" {
		return false, fmt.Errorf("missing 'url' parameter")
	}

	// Default to a GET request if method not supplied.
	method := StringParam(args, "method")
	if method == "" {
		method = "GET"
	}
	method = strings.ToUpper(method)

	body := StringParam(args, "body")

	request, err := http.NewRequest(method, url, bytes.NewBuffer([]byte(body)))
	if err != nil {
		return false, err
	}

	// Add any headers.
	headers := ArrayParam(args, "headers")
	if len(headers) > 0 {
		for _, header := range headers {
			parts := strings.SplitN(header, ":", 2)
			request.Header.Add(parts[0], parts[1])
		}
	}

	// Perform the request.
	client := http.Client{}
	response, err := client.Do(request)
	if err != nil {
		return false, err
	}

	// Make sure we close the body.
	defer response.Body.Close()

	// Read the response.
	var content []byte
	content, err = ioutil.ReadAll(response.Body)
	if err != nil {
		return false, err
	}

	// Check the response against the expected status code if supplied.
	expectedStatus := StringParam(args, "expect")
	if expectedStatus != "" {
		expectedInt, err := strconv.Atoi(expectedStatus)
		if err != nil {
			return false, err
		}

		if response.StatusCode != expectedInt {
			return false, fmt.Errorf("request returned unexpected status: %d, expected %d", response.StatusCode, expectedInt)
		}
	} else {
		// Otherwise, return error if not a 2xx status code.
		if response.StatusCode < 200 || response.StatusCode >= 300 {
			return false, fmt.Errorf("request returned unsuccessful status: %d", response.StatusCode)
		}
	}

	// Save the result values
	f.statusCode = response.StatusCode
	f.statusLine = response.Status
	f.body = string(content)

	return true, nil

}

// GetOutputs is an optional interface method which allows the
// module to return values to the caller - prefixed by the rule-name.
func (f *HTTPModule) GetOutputs() map[string]string {

	// Prepare a map of key->values to return
	m := make(map[string]string)

	// Populate with information from our execution.
	m["code"] = fmt.Sprintf("%d", f.statusCode)
	m["status"] = f.statusLine
	m["body"] = f.body

	return m
}

// init is used to dynamically register our module.
func init() {
	Register("http", func(cfg *config.Config) ModuleAPI {
		return &HTTPModule{cfg: cfg}
	})
}
