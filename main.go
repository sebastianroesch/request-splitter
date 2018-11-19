package main

import (
	"fmt"
	"io/ioutil"
	"os"
    "log"
	"net/http"
	"time"

	"github.com/gorilla/mux"
	"github.com/tkanos/gonfig"
)

var client *http.Client


func loadConfig() Configuration {
	env := os.Getenv("ENV")
	if len(env) == 0 {
		env = "development"
	}

	configuration := Configuration{}
	var err error
	if(env == "development") {
		err = gonfig.GetConf("config/config.development.json", &configuration)
	} else {
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
	configuration := loadConfig()

	router := mux.NewRouter().StrictSlash(true)
	client = &http.Client{}

	for _, element := range configuration.Endpoints {
		currentEndpoint := element
		fmt.Println("Endpoint configured:", currentEndpoint.Endpoint)

		router.HandleFunc(currentEndpoint.Endpoint, func(w http.ResponseWriter, r *http.Request) {
			fmt.Println("Endpoint called:", currentEndpoint.Endpoint)
			ch := make(chan bool)

			// Split the request.
			for _, element := range currentEndpoint.Redirects {
				// Run the upstream requests in parallel.
				go MakeRequest(element.URL, currentEndpoint.Endpoint, ch)
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
func MakeRequest(url string, forwardedEndpoint string, ch chan<-bool) {
	start := time.Now()
	
	req, _ := http.NewRequest("GET", url, nil)
	req.Header.Add("X-Forwarded-For", forwardedEndpoint)
	resp, err := client.Do(req)
  
	secs := time.Since(start).Seconds()

	if err != nil {
		// Handle errors..
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