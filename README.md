# rediproxy
rediproxy is a proxy service for Redis. It provides an HTTP GET endpoint to read values from the backing Redis instance. It maintains an in-memory LRU cache to reduce the load on Redis and supports addition of data with an expiry. 


## Requirements
- docker
- docker-compose
- make

#### Run program
```sh
    make run
```
Endpoint to fetch data:

`http://localhost:8080/cache?key=<keyname>`

#### Run tests
```sh
    make tests
```
