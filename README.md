# faas

## Services

### Ping

Accessed through /lambda/ping?address=http://someaddress.com

### Factorial

Accessed through /lambda/factorial?n=5

## Missing Features

* Gateway is very simplified
    * Does not use docker api client
    * Gateway does not provide service discovery
    * Not scalable right now. Atleast not in any meaningful way.

## Issues

* Right now http:// has to be included in ping address query
* Ping function only does one ping at the moment. Should use Pinger.RunLoop to ping for a given time maybe.

### Odd stuff

* json.Marshal seems to not work properly if struct field names does not have capitalized first
 letter? Look this up, seems odd. struct field names in general don't seem to have a capitalized
  first letter convention.
