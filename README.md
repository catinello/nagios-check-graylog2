nagios-check-graylog2
===

Nagios Graylog2 checks via REST API the availability of the service. 

Is the service processing data? How long does the check take? We monitor performance through the number of data sources, total of processed messages, index failures and the actual throughput.

This plugin is written in standard Go which means there are no third party libraries used and it is plattform independant. It can compile on all available Go architecutres and operating systems (Linux, *BSD, Mac OS X, Windows).

##Installation:##

Binary packages are available for linux (x86_64) and windows (x86_64) in the [release section][RELEASES].

Just download the archive and extract the binary to your drive.

    # instruction on linux
    $ curl -O https://github.com/catinello/nagios-check-graylog2/releases/download/16.vkXMDT/check_graylog2-16.vkXMDT-linux-amd64.tar.xz
    $ tar -xvf check_graylog2-16.vkXMDT-linux-amd64.tar.xz -C /#path#to#/#your#/nagios/libexec/ check_graylog2

##Usage:##

    check_graylog2
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

##Examples:##

    $ ./check_graylog2 -l http://localhost:12900 -u USERNAME -p PASSWORD
    OK - Service is running!
    768764376 total events processed
    0 index failures
    297 throughput
    1 sources
    Check took 94ms|time=0.0094;;;; total=768764376;;;; sources=1;;;; throughput=297;;;; index_failures=0;;;;

    $ ./check_graylog2 -l http://localhost:12900 -u USERNAME -p PASSWORD
    CRITICAL - Can not connect to Graylog2 API|time=0.000000;;;; total=0;;;; sources=0;;;; throughput=0;;;; index_failures=0;;;;

    $ ./check_graylog2 -l https://localhost -insecure -u USERNAME -p PASSWORD
    UNKNOWN - Port number is missing. Try https://hostname:port|time=0.000000;;;; total=0;;;; sources=0;;;; throughput=0;;;; index_failures=0;;;;

##Return Values:##

Nagios return codes are used.

    0 = OK
    1 = WARNING
    2 = CRITICAL
    3 = UNKNOWN

##License:##

&copy; [Antonino Catinello][HOME] - [BSD-License][BSD]

[BSD]:https://github.com/catinello/nagios-check-graylog2/blob/master/LICENSE
[HOME]:http://antonino.catinello.eu
[RELEASES]:https://github.com/catinello/nagios-check-graylog2/releases
