# gomelon [![Build Status](https://travis-ci.org/goburrow/gomelon.svg)](https://travis-ci.org/goburrow/gomelon)
Lightweight Go framework for building web services, inspired by Dropwizard.

## Example
See [example/example.go](https://github.com/goburrow/gomelon/blob/master/example/example.go)

```
INFO  [2015-02-04T22:46:18.062+10:00] gomelon.server: starting MyApp
    ______
   /\   __\______
  /..\  \  \     \
 /....\_____\  \  \
 \..../ / / /\_____\
  \../ / / /./   __/
   \/_____/./__   /
          \/_____/

INFO  [2015-02-04T22:46:18.063+10:00] gomelon.assets: registering AssetsBundle for path /static/
INFO  [2015-02-04T22:46:18.063+10:00] gomelon.admin: tasks =

    POST    /tasks/gc (*gomelon.DefaultTask)
    POST    /tasks/log (*gomelon.DefaultTask)
    POST    /tasks/task1 (*main.MyTask)

DEBUG [2015-02-04T22:46:18.063+10:00] gomelon.admin: health checks = [MyHealthCheck]
INFO  [2015-02-04T22:46:18.063+10:00] example: started MyComponent
```
