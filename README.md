# IronFunctions

[![CircleCI](https://circleci.com/gh/iron-io/functions.svg?style=svg)](https://circleci.com/gh/iron-io/functions)
[![GoDoc](https://godoc.org/github.com/iron-io/functions?status.svg)](https://godoc.org/github.com/iron-io/functions)

Welcome to IronFunctions! The open source serverless platform.

## What is IronFunctions?

IronFunctions is an open source serverless platform, or as we like to refer to it, Functions as a
Service (FaaS) platform that you can run anywhere.

* [Run anywhere](docs/faq.md#where-can-i-run-ironfunctions)
  * Public cloud, hybrid, on-premise
  * [Import Lambda functions](docs/lambda/import.md) from AWS and run them wherever you want
* [Any language](docs/faq.md#which-languages-are-supported)
  * [AWS Lambda support](docs/lambda/README.md)
* Easy to use
* Easy to scale

## What is Serverless/FaaS?

Serverless is a new paradigm in computing that enables simplicity, efficiency and scalability for both developers
and operators. It's important to distinguish the two, because the benefits differ:

### Benefits for developers

The main benefits that most people refer to are on the developer side and they include:

* No servers to manage (serverless) -- you just upload your code and the platform deals with the infrastructure
* Super simple coding -- no more monoliths! Just simple little bits of code
* Pay by the milliseconds your code is executing -- unlike a typical application that runs 24/7, and you're paying
  24/7, functions only run when needed

Since you'll be running IronFunctions yourself, the paying part may not apply, but it does apply to
cost savings on your infrastructure bills as you'll read below.

### Benefits for operators

If you will be operating IronFunctions (the person who has to manage the servers behind the serverless),
then the benefits are different, but related.

* Extremely efficient use of resources
  * Unlike an app/API/microservice that consumes resources 24/7 whether they
    are in use or not, functions are time sliced across your infrastructure and only consume resources while they are
    actually doing something
* Easy to manage and scale
  * Single system for code written in any language or any technology
  * Single system to monitor
  * Scaling is the same for all functions, you don't scale each app independently
  * Scaling is simply adding more IronFunctions nodes

There is a lot more reading you can do on the topic, just search for ["what is serverless"](https://www.google.com/webhp?sourceid=chrome-instant&ion=1&espv=2&ie=UTF-8#q=what%20is%20serverless)
and you'll find plenty of information. We have pretty thorough post on the Iron.io blog called [What is Serverless Computing and Why is it Important].

## Join Our Community

First off, join the community!

[![Slack Status](https://open-iron.herokuapp.com/badge.svg)](http://get.iron.io/open-slack)

## Quickstart

This guide will get you up and running in a few minutes.

### Run IronFunctions Container

To get started quickly with IronFunctions, you can just fire up an `iron/functions` container:

```sh
docker run --rm -it --name functions --privileged -v $PWD/data:/app/data -p 8080:8080 iron/functions
```

**Note**: A list of configurations via env variables can be found [here](docs/options.md).

### CLI tool

The IronFunctions CLI tool is optional, but it makes things easier. Install it with:

```sh
curl -sSL http://get.iron.io/fnctl | sh
```

### Create an Application

An application is essentially a grouping of functions, that put together, form an API. Here's how to create an app.

```sh
fnctl apps create myapp
```

Or using a cURL:

```sh
curl -H "Content-Type: application/json" -X POST -d '{
    "app": { "name":"myapp" }
}' http://localhost:8080/v1/apps
```

[More on apps](docs/apps.md).

Now that we have an app, we can map routes to functions.

### Add a Route

A route is a way to define a path in your application that maps to a function. In this example, we'll map
`/hello` to a simple `Hello World!` function called `iron/hello` which is a function we already made that you
can use -- yes, you can share functions! The source code for this function is in the [examples directory](examples/hello-go).
You can read more about [writing your own functions here](docs/writing.md).

```sh
fnctl routes create myapp /hello iron/hello
```

Or using cURL:

```sh
curl -H "Content-Type: application/json" -X POST -d '{
    "route": {
        "path":"/hello",
        "image":"iron/hello"
    }
}' http://localhost:8080/v1/apps/myapp/routes
```

[More on routes](docs/routes.md).

### Calling your Function

Calling your function is as simple as requesting a URL. Each app has it's own namespace and each route mapped to the app.
The app `myapp` that we created above along with the `/hello` route we added would be called via the following
URL: http://localhost:8080/r/myapp/hello

Either surf to it in your browser or use `fnctl`:

```sh
fnctl routes run myapp /hello
```

Or using a cURL:

```sh
curl http://localhost:8080/r/myapp/hello
```

### Passing data into a function

Your function will get the body of the HTTP request via STDIN, and the headers of the request will be passed in
as env vars. You can test a function with the CLI tool:

```sh
echo '{"name":"Johnny"}' | fnctl routes run myapp /hello
```

Or using cURL:

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

![async log](docs/assets/async-log.png)

Read more on [logging](docs/logging.md).

## Writing Functions

See [Writing Functions](docs/writing.md).

## More Documentation

See [docs/](docs/README.md) for full documentation.

## Want to contribute to IronFunctions?

See [contributing](CONTRIBUTING.md).

## Support

You can get community support via:

* [Stack Overflow](http://stackoverflow.com/questions/tagged/ironfunctions)
* [Slack](https://get.iron.io/open-slack)

You can get commercial support by contacting [Iron.io](https://iron.io)
