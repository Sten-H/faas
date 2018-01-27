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

* Load balancing (by default it seems to only access one container of an image)
* HTML methods. Right now only accepting GET requests

## Issues

* Flat project structure. Began with a flat structure to simplify docker-compose things as I learned it. Should modify it to reflect case specification.
* Right now http:// has to be included in ping address query. Don't think that should be required.