# faas

## Run

git clone https://github.com/Sten-H/faas.git

cd faas && docker-compose up (should publish on localhost:80)

## Services

### Ping

Accessed through /lambda/ping?host=http://someaddress.com

### Factorial

Accessed through /lambda/factorial?n=5

## About Features
 
* **Load balancing** by round robin method every time a service is requested
* **Service discovery** Will search for new containers within an interval and if they have faas labels they will be
entered into the routing table.
* **HTTP methods** GET, POST, PUT, DELETE are supported (in theory). The method type of a function is set with docker 
label faas.method, right now a function only accepts a single 
method.

## Issues

* A dead container path can be routed to if container has been taken down but not removed from the routing table yet. 
Might have to include some sort of check. Maybe you can ping it in some form.
* Containers can get entered into the routing table before they have started properly, checking  the container state for
anything other than "running" before entering them doesn't seem to help.