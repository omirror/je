# API

je provides a HTTP API that the `client` library and `job` command-line client
uses to submit, view and manage jobs. This API can also be used with a regular
HTTP client such as `curl`.

## POST /:name?args=...

Submits a new job of type `name` with the *optional* arguments `?args=...`.

* **Returns**: 302 Found

> Redirects to `GET /search/:id` to view information on the newly submitted job. Arguments can be provided via the *optional* `?args=...` query string. Input to the job can be *optionally* provided by posting a request body that is passed as "standard input" to the job upon execution.
