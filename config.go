package main

type Configuration struct {
	Endpoints []EndpointConfig
	Port      int
}

type EndpointConfig struct {
	Endpoint  string
	Redirects []RedirectConfig
}

type RedirectConfig struct {
	URL    string
	Method string
}