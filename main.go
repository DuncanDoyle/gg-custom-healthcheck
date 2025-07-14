package main

import (
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"

	log "github.com/sirupsen/logrus"
)

const LISTENER_IPADDRESS_VARNAME = "LISTENER_IPADDRESS"
const LISTENER_PORT_VARNAME = "LISTENER_PORT"
const LOG_LEVEL_VARNAME = "LOG_LEVEL"

var listenerIpAddress string
var listenerPort string
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
		fmt.Printf("server closed\n")
		log.Info("Server closed")
	} else if err != nil {
		fmt.Printf("error starting server: %s\n", err)
		// log.Errorf("error starting server: %w\n", err)
		os.Exit(1)
	}
}

// HandlerFunc: https://pkg.go.dev/net/http#HandlerFunc
func getRoot(w http.ResponseWriter, r *http.Request) {
	fmt.Printf("got / request\n")
	io.WriteString(w, "Custom healthcheck pod. Please use the /healthz endpoint!\n")
}

func getHealthz(w http.ResponseWriter, r *http.Request) {
	fmt.Printf("got /healthz request\n")

	//ddoyle: Execute a cURL command
	// We need to execute:  curl -k -s --haproxy-protocol -o /dev/null -w %{http_code} https://127.0.0.1:8443/envoy-hc
	curl := exec.Command("curl", "-k", "-s", "--haproxy-protocol", "-o", "/dev/null", "-w", "%{http_code}", "https://127.0.0.1:8443/envoy-hc")
	// curl := exec.Command("curl", "-s", "-o", "/dev/null", "-w", "%{http_code}", "http://httpbin.org/status/200")
	
	httpResponse, err := curl.Output()
	if err != nil {
		fmt.Println("erorr" , err)
		//Set the response code to 503 - InternalServerError
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}
	//Check response code
	httpResponseString := string(httpResponse[:])
	if (httpResponseString == "200") {
		fmt.Printf("Healthcheck response is OK: %s\n", httpResponseString)
		io.Writer.Write(w, httpResponse)
	} else {
		//We didn't get a 200, so we set the statuscode of our healthcheck to 503
		fmt.Printf("We didn't get a 200! We got a %s\n", httpResponseString)
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
	}

	fmt.Printf("finished /healthz request\n")
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

	//ddoyle: When listener ip-address has not been defined, we simply use empty string, as that indicates we want to listen on all ip-addresses.
	if listenerIpAddress == "" {
		log.Info("No Listener IP Address defined. Listening on all ip-addresses.")
		
	}
	if listenerPort == "" {
		log.Fatal("Listener Port has not been configured.")
		os.Exit(1)
	}
	log.Info("Initialized with the following values:" +
		"\n- Listener IP Address: " + listenerIpAddress +
		"\n- Listener Port: " + listenerPort)
}