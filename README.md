nagios-check-graylog2
===

Nagios Graylog2 checks via REST API the availability of the service. 

- Is the service processing data?
- How long does the check take?
- Monitoring performance
  - through the number of data sources,
  - total processed messages, 
  - index failures
  - and the actual throughput.

This plugin is written in standard Go which means there are no third party libraries used and it is plattform independant. It can compile on all available Go architectures and operating systems (Linux, *BSD, Mac OS X, Windows, ...).

## Installation: 

Just download the source and build it yourself using the go-tools.

    $ go get github.com/catinello/nagios-check-graylog2
    $ mv $GOPATH/bin/nagios-check-graylog2 check_graylog2

## Usage:

    check_graylog2
      -c string
          	Index Critical Limit
      -l string
            Graylog2 API URL (default "http://localhost:12900")
      -p string
            API password
      -u string
            API username
      -insecure
            Accept insecure SSL/TLS certificates.
      -version
            Display version and license information.
      -w string
            Index Error Limit
 
## Debugging:

Please try your command with the environment variable set as `NCG2=debug` or prefixing your command for example on linux like this.

    NCG2=debug /usr/local/nagios/libexec/check_graylog2 -l http://localhost:9000/api/ -u USERNAME -p PASSWORD -w 10 -c 20

## Examples:

    $ ./check_graylog2 -l http://localhost:12900 -u USERNAME -p PASSWORD -w 10 -c 20
    OK - Service is running!
    768764376 total events processed
    0 index failures
    297 throughput
    1 sources
    Check took 94ms
    |time=0.0094;;;; total=768764376;;;; sources=1;;;; throughput=297;;;; index_failures=0;;;;

    $ ./check_graylog2 -l http://localhost:12900 -u USERNAME -p PASSWORD -w 10 -c 20
    CRITICAL - Can not connect to Graylog2 API|time=0.000000;;;; total=0;;;; sources=0;;;; throughput=0;;;; index_failures=0;;;;

    $ ./check_graylog2 -l https://localhost -insecure -u USERNAME -p PASSWORD -w 10 -c 20
    UNKNOWN - Port number is missing. Try https://hostname:port|time=0.000000;;;; total=0;;;; sources=0;;;; throughput=0;;;; index_failures=0;;;;
    
     $ ./check_graylog2 -l http://localhost:12900 -u USERNAME -p PASSWORD -w 10 -c 20
    CRITICAL - Indexer Failure Critical!
    Service is running
    533732628 total events processed
    21 index failures
    297 throughput
    1 sources
    Check took 94ms
    |time=0.0094;;;; total=533732628;;;; sources=1;;;; throughput=297;;;; index_failures=21;;;;


## Return Values:

Nagios return codes are used.

    0 = OK
    1 = WARNING
    2 = CRITICAL
    3 = UNKNOWN

## License:

&copy; [Antonino Catinello][HOME] - [BSD-License][BSD]

[BSD]:https://github.com/catinello/nagios-check-graylog2/blob/master/LICENSE
[HOME]:https://antonino.catinello.eu

