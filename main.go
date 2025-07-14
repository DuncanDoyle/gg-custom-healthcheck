package main

import (
	"errors"
	"io"
	"net/http"
	"os"
	"os/exec"

	log "github.com/sirupsen/logrus"
)

const LISTENER_IPADDRESS_VARNAME = "LISTENER_IPADDRESS"
const LISTENER_PORT_VARNAME = "LISTENER_PORT"
const HEALTHCHECK_ENDPOINT_VARNAME= "HEALTHCHECK_ENDPOINT"
const LOG_LEVEL_VARNAME = "LOG_LEVEL"

var listenerIpAddress string
var listenerPort string
var healtCheckEndpoint string
var logLevel string

func main() {
	log.Info("Starting healthcheck server.")

	parseConfig()

	http.HandleFunc("/", getRoot)
	http.HandleFunc("/healthz", getHealthz)

	httpServerAddress := listenerIpAddress + ":" + listenerPort
	
	log.Info("Starting HTTP Server on '" + httpServerAddress + "'")
	// Start the HTTP server
	err := http.ListenAndServe(httpServerAddress, nil)
	if errors.Is(err, http.ErrServerClosed) {
		log.Info("Server closed")
	} else if err != nil {
		log.Errorf("error starting server: %s\n", err)
		os.Exit(1)
	}
}

// HandlerFunc: https://pkg.go.dev/net/http#HandlerFunc
func getRoot(w http.ResponseWriter, r *http.Request) {
	log.Debug("got / request\n")
	io.WriteString(w, "Custom healthcheck pod. Please use the /healthz endpoint!\n")
}

func getHealthz(w http.ResponseWriter, r *http.Request) {
	log.Debug("got /healthz request\n")

	//Execute healthcheck cURL command
	curl := exec.Command("curl", "-k", "-s", "--haproxy-protocol", "-o", "/dev/null", "-w", "%{http_code}", healtCheckEndpoint)
	
	httpResponse, err := curl.Output()
	if err != nil {
		log.Error("error", err)
		//Set the response code to 503 - InternalServerError
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}
	//Check response code
	httpResponseString := string(httpResponse[:])
	if (httpResponseString == "200") {
		log.Debugf("Healthcheck response is OK: %s\n", httpResponseString)
		io.Writer.Write(w, httpResponse)
	} else {
		//We didn't get a 200, so we set the statuscode of our healthcheck to 503
		log.Warnf("Non 200 healthcheck response received. Received response code %s\n", httpResponseString)
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
	}
	log.Debug("finished /healthz request\n")
}

func parseConfig() {
	log.Info("Parsing configuration")

	logLevel, err := log.ParseLevel(os.Getenv(LOG_LEVEL_VARNAME))
	if err != nil {
		logLevel = log.InfoLevel
	}
	log.Info("Setting logLevel to: " + logLevel.String())
	log.SetLevel(logLevel)

	listenerIpAddress = os.Getenv(LISTENER_IPADDRESS_VARNAME)
	listenerPort = os.Getenv(LISTENER_PORT_VARNAME)
	healtCheckEndpoint = os.Getenv(HEALTHCHECK_ENDPOINT_VARNAME)

	//When listener ip-address has not been defined, we simply use empty string, as that indicates we want to listen on all ip-addresses.
	if listenerIpAddress == "" {
		log.Info("No Listener IP Address defined. Listening on all ip-addresses.")	
	}
	if listenerPort == "" {
		log.Fatal("Listener Port has not been configured.")
		os.Exit(1)
	}
	if healtCheckEndpoint == "" {
		log.Fatal("HealthCheck Endpoint has not been configured.")
		os.Exit(1)
	}

	log.Info("Initialized with the following values:" +
		"\n- Listener IP Address: " + listenerIpAddress +
		"\n- Listener Port: " + listenerPort +
		"\n- HealthCheck Endpoint: " + healtCheckEndpoint)
}