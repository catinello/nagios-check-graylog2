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

var (
	// command line arguments
	link *string
	user *string
	pass *string
	// using ssl to avoid name conflict with tls
	ssl  *bool
	// env debugging variable
	debug string
)

// handle args
func init() {
	link = flag.String("l", "http://localhost:12900", "Graylog2 API URL")
	user = flag.String("u", "", "API username")
	pass = flag.String("p", "", "API password")
	ssl = flag.Bool("insecure", false, "Accept insecure SSL/TLS certificates.")
	debug = os.Getenv(DEBUG)
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

	fmt.Printf("%s - %s|\n", ev, message)
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

	return l.Scheme + "://" + l.Host
}

func main() {
	flag.Parse()

	if len(*user) == 0 || len(*pass) == 0 {
		flag.PrintDefaults()
		os.Exit(3)
	}

	c := parse(link)

	start := time.Now()
	m := query(c+"/index.json", *user, *pass)
	fmt.Println(m["id"].(float64))
	fmt.Println(m["name"].(string))
	fmt.Println(m["tags"].([]interface{}))
	elapsed := time.Since(start)

	quit(OK, fmt.Sprintf(" took %v", elapsed), nil)
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

	res, err := client.Do(req)
	defer res.Body.Close()
	if err != nil {
		quit(CRITICAL, "Can not connect to Graylog2 API", err)
	}

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
