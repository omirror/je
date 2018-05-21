# API

je provides a HTTP API that the `client` library and `job` command-line client
uses to submit, view and manage jobs. This API can also be used with a regular
HTTP client such as `curl`.

## POST /:name?args=...

Submits a new job of type `name` with the *optional* arguments `?args=...`.

* **Returns**: 302 Found

> Redirects to `GET /search/:id` to view information on the newly submitted job. Arguments can be provided via the *optional* `?args=...` query string. Input to the job can be *optionally* provided by posting a request body that is passed as "standard input" to the job upon execution.

## GET /search/:id

Retrives informatoin about a job, its status, created and finished times and response.

* **Returns:** 200 OK

> Returns a JSON array (*only containing 1 element*) of [`Job`](#job) objects.

## GET /search

Returns all known jobs and their information.

* **Returns:** 200 OK

> Returns a JOSN array of all jobs as [`Job`](#job) objects.

# Appendix

## Job

```#!json
{
  "ID": 1,
  "Name": "hello.sh",
  "Status": -1,
  "Response": "Hello\n",
  "CreatedAt": "2018-05-20T22:09:13.272775819-07:00",
  "StartedAt": "2018-05-20T22:09:13.276717365-07:00",
  "EndedAt": "2018-05-20T22:09:13.284525808-07:00"
}
```