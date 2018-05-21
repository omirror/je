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

> Returns a JSON array (*only containing 1 elemtn*) of `Job` structs that have the following attributes:

```#!go
type Job struct {
	ID        int    `storm:"id,increment"`
	Name      string `storm:"index"`
	Status    int    `storm:"index"`
	Response  string
	CreatedAt time.Time `storm:"index"`
	StartedAt time.Time `storm:"index"`
	EndedAt   time.Time `storm:"index"`
}
```

## GET /search

Returns all known jobs and their information.

* **Returns:** 200 OK

> Returns a JOSN array of all jobs as `Job` structs with the following attributes:

```#!go
type Job struct {
	ID        int    `storm:"id,increment"`
	Name      string `storm:"index"`
	Status    int    `storm:"index"`
	Response  string
	CreatedAt time.Time `storm:"index"`
	StartedAt time.Time `storm:"index"`
	EndedAt   time.Time `storm:"index"`
}
```