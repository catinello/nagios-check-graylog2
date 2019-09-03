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
	"strconv"
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
	author       = "Antonino Catinello"
	license      = "BSD"
	year         = "2016 - 2018"
	copyright    = "\u00A9"
	contributers = "kahluagenie, theherodied"
)

var (
	// command line arguments
	link    *string
	user    *string
	pass    *string
	version *bool
	// using ssl to avoid name conflict with tls
	ssl *bool
	// env debugging variable
	debug string
	// performence data
	pdata string
	// version value
	id        string
	indexwarn *string
	indexcrit *string

)

// handle performence data output
func perf(elapsed, total, inputs, tput, index float64) {
	pdata = fmt.Sprintf("time=%f;;;; total=%.f;;;; sources=%.f;;;; throughput=%.f;;;; index_failures=%.f;;;;", elapsed, total, inputs, tput, index)
}

// handle args
func init() {
	link = flag.String("l", "http://localhost:12900", "Graylog API URL")
	user = flag.String("u", "", "API username")
	pass = flag.String("p", "", "API password")
	ssl = flag.Bool("insecure", false, "Accept insecure SSL/TLS certificates. (optional)")
	version = flag.Bool("version", false, "Display version and license information. (info)")
	debug = os.Getenv(DEBUG)
	perf(0, 0, 0, 0, 0)
	indexwarn = flag.String("w", "", "Index error warning limit. (optional)")
	indexcrit = flag.String("c", "", "Index error critical limit. (optional)")
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
		quit(UNKNOWN, "Cannot parse given URL.", err)
	}

	if !strings.Contains(l.Host, ":") {
		quit(UNKNOWN, "Port number is missing. Please try "+l.Scheme+"://hostname:port", err)
	}

	if !strings.HasPrefix(l.Scheme, "HTTP") && !strings.HasPrefix(l.Scheme, "http") {
		quit(UNKNOWN, "Only HTTP is supported as protocol.", err)
	}

	return l.Scheme + "://" + l.Host + l.Path
}

func main() {
	flag.Parse()

	if *version {
		fmt.Printf("Version: %v License: %v %v %v %v\nContributers: %v\n", id, license, copyright, year, author, contributers)
		os.Exit(3)
	}

	if len(*user) == 0 || len(*pass) == 0 {
		fmt.Println("API Username/Password is mandatory.")
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

	// generate performance data output
	perf(elapsed.Seconds(), total["events"].(float64), inputs["total"].(float64), tput["throughput"].(float64), index["total"].(float64))

	// fix for backwards compatiblity if no index error threshold is set
	if len(*indexwarn) == 0 || len(*indexcrit) == 0 {
		quit(OK, fmt.Sprintf("Service is running!\n%.f total events processed\n%.f index failures\n%.f throughput\n%.f sources\nCheck took %v\n",
			total["events"].(float64), index["total"].(float64), tput["throughput"].(float64), inputs["total"].(float64), elapsed), nil)
	}

	if strings.HasSuffix((*indexcrit), "%") && strings.HasSuffix((*indexcrit), "%") {
		percentage := index["total"].(float64)/total["events"].(float64)
		// convert indexwarn and indexcrit strings to float64 variables for comparison below
		indexwarn2, err := strconv.ParseFloat((*indexwarn)[0:len(*indexwarn)-2], 64)
		if err != nil {
			quit(UNKNOWN, "Cannot parse given index warning error value.", err)
		}
		indexcrit2, err := strconv.ParseFloat((*indexcrit)[0:len(*indexcrit)-2], 64)
		if err != nil {
			quit(UNKNOWN, "Cannot parse given index critical error value.", err)
		}
		if percentage*100 > indexcrit2 {
			quit(CRITICAL, fmt.Sprintf("Index Failure above Critical Limit!\nService is running\n%.f total events processed\n%.f index failures\n%.f throughput\n%.f sources\nCheck took %v\n",
				total["events"].(float64), index["total"].(float64), tput["throughput"].(float64), inputs["total"].(float64), elapsed), nil)
		}
		if percentage*100 > indexwarn2 {
			quit(WARNING, fmt.Sprintf("Index Failure above Warning Limit!\nService is running\n%.f total events processed\n%.f index failures\n%.f throughput\n%.f sources\nCheck took %v\n",
				total["events"].(float64), index["total"].(float64), tput["throughput"].(float64), inputs["total"].(float64), elapsed), nil)
		}
		quit(OK, fmt.Sprintf("Service is running!\n%.f total events processed\n%.f index failures\n%.f throughput\n%.f sources\nCheck took %v\n",
			total["events"].(float64), index["total"].(float64), tput["throughput"].(float64), inputs["total"].(float64), elapsed), nil)
	}
	// convert indexwarn and indexcrit strings to float64 variables for comparison below
	indexwarn2, err := strconv.ParseFloat((*indexwarn), 64)
	if err != nil {
		quit(UNKNOWN, "Cannot parse given index warning error value.", err)
	}
	indexcrit2, err := strconv.ParseFloat((*indexcrit), 64)
	if err != nil {
		quit(UNKNOWN, "Cannot parse given index critical error value.", err)
	}

	// handle index thresholds
	if index["total"].(float64) < indexwarn2 && index["total"].(float64) < indexcrit2 {
		quit(OK, fmt.Sprintf("Service is running!\n%.f total events processed\n%.f index failures\n%.f throughput\n%.f sources\nCheck took %v\n",
			total["events"].(float64), index["total"].(float64), tput["throughput"].(float64), inputs["total"].(float64), elapsed), nil)
	}
	if index["total"].(float64) >= indexwarn2 && index["total"].(float64) < indexcrit2 {
		quit(WARNING, fmt.Sprintf("Index Failure above Warning Limit!\nService is running\n%.f total events processed\n%.f index failures\n%.f throughput\n%.f sources\nCheck took %v\n",
			total["events"].(float64), index["total"].(float64), tput["throughput"].(float64), inputs["total"].(float64), elapsed), nil)
	}
	if index["total"].(float64) >= indexcrit2 {
		quit(CRITICAL, fmt.Sprintf("Index Failure above Critical Limit!\nService is running\n%.f total events processed\n%.f index failures\n%.f throughput\n%.f sources\nCheck took %v\n",
			total["events"].(float64), index["total"].(float64), tput["throughput"].(float64), inputs["total"].(float64), elapsed), nil)
	}

}

// call Graylog HTTP API
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
		quit(CRITICAL, "Cannot connect to Graylog API", err)
	}
	defer res.Body.Close()

	body, err := ioutil.ReadAll(res.Body)
	if err != nil {
		quit(CRITICAL, "No response received from Graylog API", err)
	}

	if len(debug) != 0 {
		fmt.Println(string(body))
	}

	err = json.Unmarshal(body, &data)
	if err != nil {
		quit(UNKNOWN, "Cannot parse JSON from Graylog API", err)
	}

	if res.StatusCode != 200 {
		quit(CRITICAL, fmt.Sprintf("Graylog API replied with HTTP code %v", res.StatusCode), err)
	}

	return data
}
