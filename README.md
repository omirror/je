# je - Job Engine

A distributed job execution engine for the execution of batch jobs, workflows,
remediations and more. You *could* also use `je` as a simple FaaS
(*Function as a Service*) or "Serverless Computing" aka "Lambda".

**Status:** Early alpha stages, in heavy development.

## Features

* Simple HTTP API
* Simple command-line client
* UNIX friendly

## Install

```#!bash
$ go install git.mills.io/prologic/je/...
```

## Usage

Run the je daemon/server:

```#!bash
$ je -d
INFO[0000] je 0.0.1-dev (HEAD) listening on 0.0.0.0:8000
```

Run a simple job:


```#!bash
$ job run -r echo -- 'hello world'
hello world
```

You should see something like this on the server side:

```
$ je -d
INFO[0000] je 0.0.1-dev (HEAD) listening on 0.0.0.0:8000
[je] 2018/05/20 20:33:40 ([::1]:50853) "POST /echo?args=hello+world HTTP/1.1" 302 0 10.342742ms
[je] 2018/05/20 20:33:40 ([::1]:50853) "GET /search/47 HTTP/1.1" 200 212 198.135µs
```

## License

je is licensed under the [MIT License](https://git.mills.io/prologic/je/blob/master/LICENSE)
