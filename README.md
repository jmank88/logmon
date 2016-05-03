# HTTP log monitoring console program

A simple console program that monitors HTTP traffic on your machine.

## Example

```
go build
./logmon --help
>Usage of ./logmon:
>  -f string
>    	Optional input file, to be used in place of stdin.
>  -h int
>    	High traffic threshold. A high traffic alert will be triggered when the average traffic per 10s over the last 2m0s exceeds this value. (default 10)

./logmon -f lib/testdata/input/high-traffic.txt
>10/Oct/2000:13:55:36 -0700 - 10/Oct/2000:13:55:46 -0700
>	Total Hits: 11
>	Top Sections: [{http://my.site.com/pages 11}]
>	Top Methods: [{GET 11}]
>	Top Protocols: [{HTTP/1.0 11}]
>	Top Status Codes: [{200 11}]
>High traffic generated an alert - hits = 11, triggered at 10/Oct/2000:13:55:46 -0700
>10/Oct/2000:13:57:36 -0700 - 10/Oct/2000:13:57:46 -0700
>	Total Hits: 1
>	Top Sections: [{http://my.site.com/pages 1}]
>	Top Methods: [{GET 1}]
>	Top Protocols: [{HTTP/1.0 1}]
>	Top Status Codes: [{200 1}]
>Recovered from high traffic at 10/Oct/2000:13:57:46 -0700
```

## Features

[x] Consume an actively written-to w3c-formatted HTTP access log (https://en.wikipedia.org/wiki/Common_Log_Format)

Logs can be read from stdin, or from a file via the -f flag.

[x] Every 10s, display in the console the sections of the web site with the most hits (a section is defined as being what's before the second '/' in a URL. i.e. the section for "http://my.site.com/pages/create' is "http://my.site.com/pages"), as well as interesting summary statistics on the traffic as a whole.

Every 10s, the total hits and a summary of hit counts per section, method, protocol, and status code is written.

[x] Make sure a user can keep the console app running and monitor traffic on their machine

The application remains running until it reaches EOF or an error is encountered.

[x] Whenever total traffic for the past 2 minutes exceeds a certain number on average, add a message saying that “High traffic generated an alert - hits = {value}, triggered at {time}”

A high traffic alert will be triggered when the average traffic per interval over the last httDuration exceeds the value specified by the -h flag.

[x] Whenever the total traffic drops again below that value on average for the past 2 minutes, add another message detailing when the alert recovered

A recovery message will be written when the average traffic per interval over the last httDuration drops back below the value specified by the -h flag.

[x] Make sure all messages showing when alerting thresholds are crossed remain visible on the page for historical reasons.

All alerts and summaries are written chronologically and can be easily searched/grepped for.

[x] Write a test for the alerting logic

The logmon_test.go file contains various tests, including one for high traffic alerts.

## Future Improvements

- Expose the interval duration and high traffic duration parameters as command line flags

- Make summary stats configurable and expose the parameters through the API and as command line flags

- Better handling of bad data and/or out of order logs. Bad lines and (most) lines from the 'past' are ignored. Instead, they could be an additional type of summary stat.