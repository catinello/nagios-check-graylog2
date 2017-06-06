package main

import (
	"crypto/tls"
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"os"
	"strings"
	"time"
)

// nagios exit codes
const (
	OK = iota
	WARNING
	CRITICAL
	UNKNOWN
)

// export NCG2=debug
const DEBUG = "NCG2"

// license information
const (
        author = "Antonino Catinello"
        license = "BSD"
        year = "2016"
        copyright = "\u00A9"
)


var (
	// command line arguments
	link *string
	user *string
	pass *string
	version *bool
	// using ssl to avoid name conflict with tls
	ssl *bool
	// env debugging variable
	debug string
	// performence data
	pdata string
	// version value
	id string
)

// handle performence data output
func perf(elapsed, total, inputs, tput, index float64) {
	pdata = fmt.Sprintf("time=%f;;;; total=%.f;;;; sources=%.f;;;; throughput=%.f;;;; index_failures=%.f;;;;", elapsed, total, inputs, tput, index)
}

// handle args
func init() {
	link = flag.String("l", "http://localhost:12900", "Graylog2 API URL")
	user = flag.String("u", "", "API username")
	pass = flag.String("p", "", "API password")
	ssl = flag.Bool("insecure", false, "Accept insecure SSL/TLS certificates.")
	version = flag.Bool("version", false, "Display version and license information.")
	debug = os.Getenv(DEBUG)
	perf(0, 0, 0, 0, 0)
}

// return nagios codes on quit
func quit(status int, message string, err error) {
	var ev string

	switch status {
	case OK:
		ev = "OK"
	case WARNING:
		ev = "WARNING"
	case CRITICAL:
		ev = "CRITICAL"
	case UNKNOWN:
		ev = "UNKNOWN"
	}

	// if debugging is enabled
	// print errors
	if len(debug) != 0 {
		fmt.Println(err)
	}

	fmt.Printf("%s - %s|%s\n", ev, message, pdata)
	os.Exit(status)
}

// parse link
func parse(link *string) string {
	l, err := url.Parse(*link)
	if err != nil {
		quit(UNKNOWN, "Can not parse given URL.", err)
	}

	if !strings.Contains(l.Host, ":") {
		quit(UNKNOWN, "Port number is missing. Try "+l.Scheme+"://hostname:port", err)
	}

	if !strings.HasPrefix(l.Scheme, "HTTP") && !strings.HasPrefix(l.Scheme, "http") {
		quit(UNKNOWN, "Only HTTP is supported as protocol.", err)
	}

	return l.Scheme + "://" + l.Host + l.Path
}

func main() {
	flag.Parse()

	if *version {
		fmt.Printf("Version: %v License: %v %v %v %v\n", id, license, copyright, year, author)
		os.Exit(3)
	}

	if len(*user) == 0 || len(*pass) == 0 {
		flag.PrintDefaults()
		os.Exit(3)
	}

	c := parse(link)
	start := time.Now()

	system := query(c+"/system", *user, *pass)
	if system["is_processing"].(bool) != true {
		quit(CRITICAL, "Service is not processing!", nil)
	}
	if strings.Compare(system["lifecycle"].(string), "running") != 0 {
		quit(WARNING, fmt.Sprintf("lifecycle: %v", system["lifecycle"].(string)), nil)
	}
	if strings.Compare(system["lb_status"].(string), "alive") != 0 {
		quit(WARNING, fmt.Sprintf("lb_status: %v", system["lb_status"].(string)), nil)
	}

	index := query(c+"/system/indexer/failures", *user, *pass)
	tput := query(c+"/system/throughput", *user, *pass)
	inputs := query(c+"/system/inputs", *user, *pass)
	total := query(c+"/count/total", *user, *pass)

	elapsed := time.Since(start)

	perf(elapsed.Seconds(), total["events"].(float64), inputs["total"].(float64), tput["throughput"].(float64), index["total"].(float64))
	quit(OK, fmt.Sprintf("Service is running!\n%.f total events processed\n%.f index failures\n%.f throughput\n%.f sources\nCheck took %v",
		total["events"].(float64), index["total"].(float64), tput["throughput"].(float64), inputs["total"].(float64), elapsed), nil)
}

// call Graylog2 HTTP API
func query(target string, user string, pass string) map[string]interface{} {
	var client *http.Client
	var data map[string]interface{}

	if *ssl {
		tp := &http.Transport{
			// keep this necessary evil for internal servers with custom certs?
			TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
		}

		client = &http.Client{Transport: tp}
	} else {
		client = &http.Client{}
	}

	req, err := http.NewRequest("GET", target, nil)
	req.SetBasicAuth(user, pass)
	req.Header.Set("Accept", "application/json")

	res, err := client.Do(req)
	if err != nil {
		quit(CRITICAL, "Can not connect to Graylog2 API", err)
	}
	defer res.Body.Close()

	body, err := ioutil.ReadAll(res.Body)
	if err != nil {
		quit(CRITICAL, "No response received from Graylog2 API", err)
	}

	if len(debug) != 0 {
		fmt.Println(string(body))
	}

	err = json.Unmarshal(body, &data)
	if err != nil {
		quit(UNKNOWN, "Can not parse JSON from Graylog2 API", err)
	}

	if res.StatusCode != 200 {
		quit(CRITICAL, fmt.Sprintf("Graylog2 API replied with HTTP code %v", res.StatusCode), err)
	}

	return data
}
