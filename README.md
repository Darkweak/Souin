<p align="center"><a href="https://github.com/darkweak/souin"><img src="docs/img/logo.svg?sanitize=true" alt="Souin logo"></a></p>

# Souin Table of Contents
1. [Souin reverse-proxy cache](#project-description)
2. [Environment variables](#environment-variables)  
  2.1. [Required variables](#required-variables)  
  2.2. [Optional variables](#optional-variables)
3. [Cache system](#cache-system)
4. [Exemples](#exemples)  
  4.1. [Træfik container](#træfik-container)

[![Travis CI](https://travis-ci.com/Darkweak/Souin.svg?branch=master)](https://travis-ci.com/Darkweak/Souin)

# <img src="docs/img/logo.svg?sanitize=true" alt="Souin logo" width="30" height="30">ouin reverse-proxy cache

## Project description
Souin is a new cache system for every reverse-proxy. It will be placed on top of your reverse-proxy like Apache, NGinx or Traefik.  
As it's written in go, it can be deployed on any server and with docker integration, it will be easy to implement it on top of Swarm or kubernetes instance.

## Environment variables

### Required variables
|  Variable  |  Description  |  Value exemple  |
|:---:|:---:|:---:|
|`CACHE_PORT`|The HTTP port Souin will be running to|`80`|
|`CACHE_TLS_PORT`|The TLS port Souin will be running to|`443`|
|`REDIS_URL`|The redis instance URL|- `http://redis` (Container way)<br/>`http://localhost:6379` (Local way)|
|`TTL`|Duration to cache request (in seconds)|10|
|`REVERSE_PROXY`|The reverse-proxy instance URL like Apache, Nginx, Træfik, etc...|- `http://yourservice` (Container way)<br/>`http://localhost:81` (Local way)|

### Optional variables
|  Variable  |  Description  |  Value exemple  |
|:---:|:---:|:---:|
|`REGEX`|The regex to define URL to not store in cache|`http://domain.com/mypath`|

## Cache system
The cache is set into redis instance, because we can set, get, update and delete keys as easy as possible.  
To perform with that, redis should be on the same network than Souin instance if you are using docker-compose, then both should be on the same server if you use binaries  
Asynchronously, Souin will request redis instance and the reverse-proxy to get at least one valid response and return to the client the first response caught by Souin.

### Cache invalidation
The cache invalidation is made for CRUD requests, if you're doing a GET HTTP request, it will serve the cached response if exists then the reverse-proxy response will be served.  
If you're doing a POST, PUT, PATCH or DELETE HTTP request, the related cached get request will be dropped and the list endpoint will be dropped too  
It works very well with plain [API Platform](https://api-platform.com) integration (but not custom actions for now) and CRUD routes

## Exemples

### Træfik container
[Træfik](https://traefik.io) is a modern reverse-proxy and help you to manage full container architecure projects.

```yaml
# your-traefik-instance/docker-compose.yml
version: '3.4'

x-networks: &networks
  networks:
    - your_network

services:
  traefik:
    image: traefik:v2.0
    ports:
      - "81:80" # Note the 81 to 80 port declaration
      - "444:443" # Note the 444 to 443 port declaration
    command: --providers.docker
    volumes:
      - /var/run/docker.sock:/var/run/docker.sock
    <<: *networks

  # your other services here...

networks:
  your_network:
    external: true
```

```yaml
# your-souin-instance/docker-compose.yml
version: '3.4'

x-networks: &networks
  networks:
    - your_network

services:
  souin:
    build:
      context: .
    ports:
      - ${CACHE_PORT}:80
      - ${CACHE_TLS_PORT}:443
    depends_on:
      - redis
    environment:
      REDIS_URL: ${REDIS_URL}
      TTL: ${TTL}
      CACHE_PORT: ${CACHE_PORT}
      CACHE_TLS_PORT: ${CACHE_TLS_PORT}
      REVERSE_PROXY: ${REVERSE_PROXY}
      REGEX: ${REGEX}
      GOPATH: /app
    volumes:
      - ./cmd:/app/cmd
      - ./acme.json:/app/src/github.com/darkweak/souin/acme.json
    <<: *networks

  redis:
    image: redis:alpine
    <<: *networks

networks:
  your_network:
    external: true
```

## SSL

### Træfik
As Souin is compatible with Træfik, it can use (and it should use) acme.json provided on træfik. Souin will get new/updated certs from Træfik, then your SSL certs will be up to date as far as Træfik will be too
To provide, acme, use just have to map volume as above
```yaml
    volumes:
      - /anywhere/acme.json:/app/src/github.com/darkweak/souin/acme.json
```
At the moment you can't choose the path for the acme.json in the container, in the future you'll be able to do that just setting one env var
If none acme.json is provided to container, a default cert will be served.
