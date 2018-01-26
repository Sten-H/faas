# faas

## Run
git clone https://github.com/Sten-H/faas.git

cd faas && docker-compose up (should publish on localhost:80)

running something like docker-compose scale factorial=0 and then trying /lambda/factorial?n=5 will tell you that
function does not exist which I think should be expected behaviour. docker-compose scale factorial=3 and try GET 
again will give you result of function.

## Services

### Ping

Accessed through /lambda/ping?address=http://someaddress.com

### Factorial

Accessed through /lambda/factorial?n=5

## Missing Features

* HTML methods. Right now only accepting GET requests

## Issues

* Load balancing. I'm not sure if it is done at all right now. Printing out which container responds on request it seems
to always be the same one (lets say factorial_3 out of 1-3). This could be because the request are so staggered that 
factorial_3 is always available. Not sure how to try this out to be certain. I imagined this would be done under the hood
by docker-compose.
* routingTable is not thread safe right now. Should use lock.
* Flat project structure. Began with a flat structure to simplify docker-compose things as I learned it. Should modify it to reflect case specification.
* Right now http:// has to be included in ping address query. Don't think that should be required.
* Ping function only does one ping at the moment. Should use Pinger.RunLoop to ping for a given time maybe.

### Odd stuff

* json.Marshal seems to not work properly if struct field names do not have capitalized first
 letter? Look this up, seems odd. struct field names in general don't seem to have a capitalized
  first letter convention.