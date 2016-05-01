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
>10/Oct/2000:13:55:46 -0700 - 10/Oct/2000:13:55:56 -0700
>	Section Hits: map[http://my.site.com/pages:1]
```

## Features
TODO expand on each of these points

[/] Consume an actively written-to w3c-formatted HTTP access log (https://en.wikipedia.org/wiki/Common_Log_Format)

[/] Every 10s, display in the console the sections of the web site with the most hits (a section is defined as being what's before the second '/' in a URL. i.e. the section for "http://my.site.com/pages/create' is "http://my.site.com/pages"), as well as interesting summary statistics on the traffic as a whole.

TODO more stats (method and protocol summary)

[/] Make sure a user can keep the console app running and monitor traffic on their machine

[/] Whenever total traffic for the past 2 minutes exceeds a certain number on average, add a message saying that “High traffic generated an alert - hits = {value}, triggered at {time}”

[/] Whenever the total traffic drops again below that value on average for the past 2 minutes, add another message detailing when the alert recovered

[/] Make sure all messages showing when alerting thresholds are crossed remain visible on the page for historical reasons.

[/] Write a test for the alerting logic

## Future Improvements

TODO Explain how you’d improve on this application design