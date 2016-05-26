nagios-check-graylog2
===

Nagios Graylog2 checks via REST API the availability of the service. 

Is the service processing data? How long does the check take? We monitor performance through the number of data sources, total of recorded messages, index failures and the throughput.

This plugin is written in standard go which means there are no third party libraries and it is plattform independant. It should run on all available go architecutres and operating systems (Linux, *BSD, Mac OS X, Windows).

##Installation:##

Static binaries are available for linux (32/64 bit) and windows (64bit).

    # curl -O https://...
    # tar -xvf FILENAME.tar.xz -C /#to#/#your#/nagios/libexec/

##Usage:##

    Usage: check_graylog2
      -l string
            Graylog2 API URL (default "http://localhost:12900")
      -p string
            API password
      -u string
            API username
      -insecure
            Accept insecure SSL/TLS certificates.

##Examples:##

    $ ./check_graylog2 -l http://localhost:12900 -u USERNAME -p PASSWORD
    HTTP OK: 174ms - STATUS|time=174ms;0;

##Return Values:##

Nagios return codes are used.

    0 = OK
    1 = WARNING
    2 = CRITICAL
    3 = UNKNOWN

##License:##

[&copy; Antonino Catinello][HOME] - [BSD-License][BSD]

[BSD]:https://github.com/catinello/nagios-check-graylog2/blob/master/LICENSE
[HOME]:http://antonino.catinello.eu
