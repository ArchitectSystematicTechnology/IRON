# IronFunctions

Welcome to IronFunctions!  The open source Functions as a Service platform.

## Join Our Community

First off, join the community!

[![Slack Status](https://open-iron.herokuapp.com/badge.svg)](https://open-iron.herokuapp.com)

## Quickstart

This guide will get you up and running in a few minutes.

### Run IronFunctions Container

To get started quickly with IronFunctions, you can just fire up an `iron/functions` container:

```sh
docker run --rm -it --name functions --privileged -v $PWD/data:/app/data -p 8080:8080 iron/functions
```

**Note**: A list of configurations via env variables can be found [here](docs/options.md).*

### CLI tool

You can easily operate IronFunctions with its CLI tool. Install it with:

```sh
curl -sSL https://fn.iron.io/install | sh
```

If you're concerned about the [potential insecurity](http://curlpipesh.tumblr.com/)
of using `curl | sh`, feel free to use a two-step version of our installation and examine our
installation script:

```bash
curl -f -sSL https://fn.iron.io/install -O
sh install
```

### Create an Application

An application is essentially a grouping of functions, that put together, form an API. Here's how to create an app.

```sh
fnctl apps create myapp
```

Or using a cURL call:
```sh
curl -H "Content-Type: application/json" -X POST -d '{
    "app": { "name":"myapp" }
}' http://localhost:8080/v1/apps
```

Now that we have an app, we can map routes to functions.

### Add a Route

A route is a way to define a path in your application that maps to a function. In this example, we'll map
`/path` to a simple `Hello World!` image called `iron/hello`.

```sh
fnctl routes create myapp /hello iron/hello
```

Or using a cURL call:
```sh
curl -H "Content-Type: application/json" -X POST -d '{
    "route": {
        "path":"/hello",
        "image":"iron/hello"
    }
}' http://localhost:8080/v1/apps/myapp/routes
```

### Calling your Function

Calling your function is as simple as requesting a URL. Each app has it's own namespace and each route mapped to the app.
The app `myapp` that we created above along with the `/hello` route we added would be called via the following URL.

```sh
fnctl routes run myapp /hello
```

Or using a cURL call:
```sh
curl http://localhost:8080/r/myapp/hello
```

You also may just surf to it: http://localhost:8080/r/myapp/hello

### Passing data into a function

Your function will get the body of the HTTP request via STDIN, and the headers of the request will be passed in as env vars. Try this:

```sh
echo '{"name":"Johnny"}' | fnctl routes run myapp /hello
```

Or using a cURL call:
```sh
curl -H "Content-Type: application/json" -X POST -d '{
    "name":"Johnny"
}' http://localhost:8080/r/myapp/hello
```

You should see it say `Hello Johnny!` now instead of `Hello World!`.

### Add an asynchronous function

IronFunctions supports synchronous function calls like we just tried above, and asynchronous for background processing.

Asynchronous function calls are great for tasks that are CPU heavy or take more than a few seconds to complete.
For instance, image processing, video processing, data processing, ETL, etc.
Architecturally, the main difference between synchronous and asynchronous is that requests
to asynchronous functions are put in a queue and executed on upon resource availability so that they do not interfere with the fast synchronous responses required for an API.
Also, since it uses a message queue, you can queue up millions of function calls without worrying about capacity as requests will
just be queued up and run at some point in the future.

To add an asynchronous function, create another route with the `"type":"async"`, for example:

```sh
curl -H "Content-Type: application/json" -X POST -d '{
    "route": {
        "type": "async",
        "path":"/hello-async",
        "image":"iron/hello"
    }
}' http://localhost:8080/v1/apps/myapp/routes
```

Now if you request this route:

```sh
curl -H "Content-Type: application/json" -X POST -d '{
    "name":"Johnny"
}' http://localhost:8080/r/myapp/hello-async
```

You will get a `call_id` in the response:

```json
{"call_id":"572415fd-e26e-542b-846f-f1f5870034f2"}
```

If you watch the logs, you will see the function actually runs in the background:

![async log](docs/async-log.png)

Read more on [logging](docs/logging.md).

## Client Libraries

- [Go](https://github.com/iron-io/functions_go)
- [Ruby](https://github.com/iron-io/functions_ruby)
- [Javascript](https://github.com/iron-io/functions_js)

## Writing Functions

See [Writing Functions)(docs/writing.md).

## More Documentation

See [docs/](docs/README.md) for full documentation.
