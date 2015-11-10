# Melon [![Build Status](https://travis-ci.org/goburrow/melon.svg)](https://travis-ci.org/goburrow/melon) [![GoDoc](https://godoc.org/github.com/goburrow/melon?status.svg)](https://godoc.org/github.com/goburrow/melon)
Lightweight Go framework for building web services, inspired by Dropwizard.

## Overview
Melon is a partial port of [Dropwizard](http://dropwizard.io/) in Go.
Besides of builtin Go packages, it utilizes a number of [libraries](https://github.com/goburrow/melon/blob/master/THIRDPARTY.md)
in order to build a server stack quickly, including:

* [goji](https://github.com/zenazn/goji): a robust web framework.
* [gol](https://github.com/goburrow/gol): a simple hierarchical logging API.
* [metrics](https://github.com/codahale/metrics): a minimalist instrumentation library.
* [validator](https://github.com/go-validator/validator): extensible value validations.

Features supported:

- Commands: for controlling your application from command line.
- Bundles: for modularizing your application.
- Managed Objects: for starting and stopping your components.
- HealthChecks: for checking health of your application in production.
- Metrics: for monitoring and statistics.
- Tasks: for administration.
- Resources: for RESTful endpoints.
- Filters: for injecting middlewares.
- Logging: for understanding behaviors of your application.
- Configuration: for application parameters.
- Banner: for fun. :)
- and more...

## Examples
See [example](https://github.com/goburrow/melon/tree/master/example)

```
INFO  [2015-02-04T12:00:01.289+10:00] melon/server: starting MyApp
    ______
   /\   __\______
  /..\  \  \     \
 /....\_____\  \  \
 \..../ / / /\_____\
  \../ / / /./   __/
   \/_____/./__   /
          \/_____/

INFO  [2015-02-04T12:00:01.289+10:00] melon/assets: registering AssetsBundle for path /static/
DEBUG [2015-02-04T12:00:01.289+10:00] melon/server: resources = [*rest.XMLProvider,*main.usersResource,*main.userResource]
INFO  [2015-02-04T12:00:01.289+10:00] melon/server: endpoints =

    GET     /users (*main.usersResource)
    POST    /users (*main.usersResource)
    GET     /user/:name (*main.userResource)
    POST    /user/:name (*main.userResource)
    DELETE  /user/:name (*main.userResource)

INFO  [2015-02-04T12:00:01.290+10:00] melon/admin: tasks =

    POST    /tasks/gc (*core.gcTask)
    POST    /tasks/log (*logging.logTask)
    POST    /tasks/rmusers (*main.usersTask)

DEBUG [2015-02-04T12:00:01.290+10:00] melon/admin: health checks = [UsersHealthCheck]
INFO  [2015-02-04T12:00:01.290+10:00] melon/server: listening :8080
INFO  [2015-02-04T12:00:01.290+10:00] melon/server: listening :8081
```

## Contributing
The project still lacks of coffee and swears. Comments, issues and pull requests are welcome.
