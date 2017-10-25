package main

/*
**   Copyright 2017 Telenor Digital AS
**
**  Licensed under the Apache License, Version 2.0 (the "License");
**  you may not use this file except in compliance with the License.
**  You may obtain a copy of the License at
**
**      http://www.apache.org/licenses/LICENSE-2.0
**
**  Unless required by applicable law or agreed to in writing, software
**  distributed under the License is distributed on an "AS IS" BASIS,
**  WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
**  See the License for the specific language governing permissions and
**  limitations under the License.
 */

// This is a simple daemon that will forward expvar variables to AWS CloudWatch
// -- ie a poor man's monitoring. The functionality is as basic as possible:
// Only whole numbers (aka floats) are supported. The defaults will work for
// a Congress server running with the expvar values at http://localhost:8081/debug/vars
//
// It will log to syslog if the -syslog flag is set.
//
import (
	"flag"
	"io/ioutil"
	"log"
	"log/syslog"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"time"
)

const metadataInstanceURL = "http://169.254.169.254/latest/meta-data/instance-id"

var expvarURI = "http://localhost:8081/debug/vars"
var interval = 60
var filterList = "\\.total$"
var metricName = "my-service"
var useSyslog = false

func init() {
	flag.StringVar(&expvarURI, "expvar-uri", expvarURI, "The expvar URI to read from")
	flag.IntVar(&interval, "interval", interval, "Interval (in seconds) for the polling process")
	flag.StringVar(&filterList, "filters", filterList, "Regexp filters for list. Separate filters with semicolons")
	flag.StringVar(&metricName, "metricname", metricName, "Name of metric (aka group of metrics in AWS)")
	flag.BoolVar(&useSyslog, "syslog", false, "Use syslog when logging")
	flag.Parse()
}

// Pull the instance ID from the instance metadata. If there's no endpoint
// (or there's an error) just assume this is running locally and name it
// "local-test"
func awsInstanceID() string {
	const testInstanceID = "local-test"
	resp, err := http.Get(metadataInstanceURL)
	if err != nil || resp.StatusCode != http.StatusOK {
		return testInstanceID
	}

	buf, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return testInstanceID
	}
	return strings.TrimSpace(string(buf))
}

func main() {
	if useSyslog {
		syslogLogger, err := syslog.New(syslog.LOG_ERR|syslog.LOG_DAEMON, "cloudwatchexporter")
		if err != nil {
			log.Printf("Unable to set up error syslog: %v", err)
			return
		}
		log.SetOutput(syslogLogger)
		log.SetFlags(log.Lshortfile)
	}
	log.Printf("Starting expvars -> cloudwatch exporter")

	terminate := make(chan bool)

	sigch := make(chan os.Signal, 2)
	signal.Notify(sigch, os.Interrupt, os.Kill)
	go func() {
		sig := <-sigch
		log.Printf("Caught signal '%v' - terminating", sig)
		terminate <- true
	}()

	filters, err := NewFilter(strings.Split(filterList, ";")...)
	if err != nil {
		log.Printf("Invalid filter statement: %v - terminating", err)
		return
	}

	forwarder, err := NewForwarder(expvarURI, filters, metricName, awsInstanceID())
	if err != nil {
		log.Printf("Unable to create forwarder: %v - terminating", err)
		return
	}

	if err := forwarder.ReadAndForward(); err != nil {
		log.Printf("Unable to forward: %v", err)
	}

	for {
		select {
		case <-terminate:
			return
		case <-time.After(time.Duration(interval) * time.Second):
			if err := forwarder.ReadAndForward(); err != nil {
				log.Printf("Unable to forward: %v", err)
			}
		}
	}
}
