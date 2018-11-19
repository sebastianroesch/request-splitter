package main

import (
	"fmt"
	"io/ioutil"
	"os"
    "log"
	"net/http"
	"time"
	"bytes"

	"github.com/gorilla/mux"
	"github.com/tkanos/gonfig"
)

var client *http.Client
var hostIp string

func LoadConfig() Configuration {
	env := os.Getenv("ENV")
	if len(env) == 0 {
		env = "development"
	}

	configuration := Configuration{}
	var err error

	// Load the configuration file based on the environment.
	if(env == "development") {
		err = gonfig.GetConf("config/config.development.json", &configuration)
	} else {
		// This is the path where a config file can be mounted.
		err = gonfig.GetConf("/config/config.json", &configuration)
	}

	if err != nil {
		fmt.Println(err)
		os.Exit(500)
	}

	return configuration
}

func main() {
	// Load the config from file and ENV variables.
	configuration := LoadConfig()
	hostIp = os.Getenv("HOST_IP")
	if len(hostIp) == 0 {
		hostIp = "0.0.0.0"
	}

	router := mux.NewRouter().StrictSlash(true)
	client = &http.Client{}

	for _, element := range configuration.Endpoints {
		currentEndpoint := element
		fmt.Println("Endpoint configured:", currentEndpoint.Endpoint)

		router.HandleFunc(currentEndpoint.Endpoint, func(w http.ResponseWriter, r *http.Request) {
			fmt.Println("Endpoint called:", currentEndpoint.Endpoint)
			ch := make(chan bool)

			// Read the content
			var bodyBytes []byte
			if r.Body != nil {
				bodyBytes, _ = ioutil.ReadAll(r.Body)
			}

			// Split the request.
			for _, element := range currentEndpoint.Redirects {

				// Check if the method has an override.
				method := r.Method
				if len(element.Method) != 0 {
					method = element.Method
				}

				// Run the upstream requests in parallel.
				go MakeRequest(element.URL, currentEndpoint.Endpoint, method, r, bodyBytes, ch)
			}

			// Wait for the upstream requests and check for success.
			allSuccess := true
			for range currentEndpoint.Redirects {
				success := <-ch
				fmt.Println(success)
				allSuccess = allSuccess && success
			}

			// Return a response based on the redirect success.
			if(allSuccess) {
				w.WriteHeader(http.StatusOK)
			} else {
				w.WriteHeader(http.StatusMultiStatus)
			}
		})
	}

	fmt.Println("Request Splitter running:", fmt.Sprintf(":%d", configuration.Port))
    log.Fatal(http.ListenAndServe(fmt.Sprintf(":%d", configuration.Port), router))
}

// MakeRequest makes the request to the upstream endpoint.
func MakeRequest(url string, forwardedEndpoint string, method string, r *http.Request, bodyBytes []byte, ch chan<-bool) {
	start := time.Now()

	fmt.Println("-->", method, url)

	// Create a new request. Forward the body, if set.
	reader := bytes.NewReader(bodyBytes)
	req, _ := http.NewRequest(method, url, reader)

	// Set the forwarded headers.
	req.Header.Add("X-Forwarded-For", hostIp)
	req.Header.Add("X-Forwarded-Url", forwardedEndpoint)

	// Add the original headers.
	for k, v := range r.Header {
		req.Header.Add(k, v[0])
	}
	// Make the request.
	resp, err := client.Do(req)
  
	secs := time.Since(start).Seconds()

	if err != nil {
		// Handle errors.
		fmt.Println("Redirect failure:", err)

		ch <- false
	}

	if resp != nil {
		// Handle success response.
		fmt.Println("Redirect success:", resp.Status)

		body, _ := ioutil.ReadAll(resp.Body)
		fmt.Println("%.2f elapsed with response length: %d %s", secs, len(body), url)

		ch <- true
	}
  }