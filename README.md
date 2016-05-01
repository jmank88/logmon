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

logmon -f lib/testdata/input/high-traffic.txt
>10/Oct/2000:13:55:36 -0700 - 10/Oct/2000:13:55:46 -0700
>	Section Hits: map[http://my.site.com/pages:11]
>High traffic generated an alert - hits = 11, triggered at 10/Oct/2000:13:55:46 -0700
>10/Oct/2000:13:55:46 -0700 - 10/Oct/2000:13:55:56 -0700
>	Section Hits: map[http://my.site.com/pages:1]
>Recovered from high traffic at 10/Oct/2000:13:55:56 -0700
```

## Features
TODO expand on each of these points

[x] Consume an actively written-to w3c-formatted HTTP access log (https://en.wikipedia.org/wiki/Common_Log_Format)

Logs can be read from stdin, or from a file via the -f flag.

[x] Every 10s, display in the console the sections of the web site with the most hits (a section is defined as being what's before the second '/' in a URL. i.e. the section for "http://my.site.com/pages/create' is "http://my.site.com/pages"), as well as interesting summary statistics on the traffic as a whole.

Every 10s, a summary of hit counts per section is written.
TODO more stats (method and protocol summary)

[x] Make sure a user can keep the console app running and monitor traffic on their machine

The application remains running until it reaches EOF or an error is encountered.

[x] Whenever total traffic for the past 2 minutes exceeds a certain number on average, add a message saying that “High traffic generated an alert - hits = {value}, triggered at {time}”

A high traffic alert will be triggered when the average traffic per interval over the last httDuration exceeds the value specified by the -h flag.

[x] Whenever the total traffic drops again below that value on average for the past 2 minutes, add another message detailing when the alert recovered

[x] Make sure all messages showing when alerting thresholds are crossed remain visible on the page for historical reasons.

All alerts and summaries are written chronologically and can be easily searched/grepped for.

[x] Write a test for the alerting logic

## Future Improvements

TODO Explain how you’d improve on this application design