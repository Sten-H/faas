# faas

## Run

git clone https://github.com/Sten-H/faas.git

cd faas && docker-compose up (should publish on localhost:80)

## Services

### Ping

Accessed through /lambda/ping?address=http://someaddress.com

### Factorial

Accessed through /lambda/factorial?n=5

## Missing Features

* HTML methods. Right now only accepting GET requests

## About Features

Right now a lot of the functionality is implemented in a very naive way, to be able to sketch out the form of it. 
* Load balancing is extremely naive, it treats list of containers of services as a FIFO queue. So every container of
"function_1" will be cycled through until a container is used again.
* Service discovery right now is done, I think, in a rather heavy handed way where it will rebuild the entire RouteTable
in an interval. Best would be if it could trigger add/remove  when new containers appear or existing containers 
disappear. In the implementation used right now since it rebuild the RouteTable it also resets the """load balancing"""

## Issues

* When a container gets entered into routingTable (and general docker environment) it doesn't seem to start until
getting its first request (which it will fail because it won't be ready)
* Right now http:// has to be included in ping address query