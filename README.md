# Go expvars to AWS CloudWatch daemon

This is a simple daemon that reads the content of an expvar endpoint in 
go and forwards the output to CloudWatch. 

## Command line parameters
* `-expvar-uri` -- the expvar endpoint to read from. The default is 
  `http://localhost:8081/debug/vars`
* `-interval` -- interval (in seconds) for the polling. The default is 60 seconds
* `-filters` -- list of regexpes to include. Separate each filter with a semicolon. 
  The default is to pull all metrics ending with `.total` into CloudWatch.
* `-metricname` -- the top level name of the metrics in CloudWatch. The default 
  is `my-service`. You'll probably want to change this.
* `-syslog` -- log to syslog. The default is to log to stdout
