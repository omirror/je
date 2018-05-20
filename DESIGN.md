# Design

## Server interactions

### Creating a job

```#!bash
$ curl -v -o - -X POST http://localhost:8000/job/hello.sh
*   Trying ::1...
* TCP_NODELAY set
* Connected to localhost (::1) port 8000 (#0)
> POST /job/hello.sh HTTP/1.1
> Host: localhost:8000
> User-Agent: curl/7.54.0
> Accept: */*
>
< HTTP/1.1 302 Found
< Location: /job/1
< Date: Sun, 20 May 2018 07:38:36 GMT
< Content-Length: 0
<
* Connection #0 to host localhost left intact
```

### Viewing all jobs

```#!bash
$ curl -s -q -o - http://localhost:8000/jobs | jq '.'
[
  {
    "ID": 1,
    "Name": "hello.sh",
    "Status": 0,
    "Response": "Hello\n",
    "CreatedAt": "2018-05-20T00:38:36.616995092-07:00",
    "StartedAt": "2018-05-20T00:38:36.619777692-07:00",
    "EndedAt": "2018-05-20T00:38:36.627481175-07:00"
  }
]
```

### Viewing a specific job

```#!bash
$ curl -s -q -o - http://localhost:8000/job/1 | jq '.'
{
  "ID": 1,
  "Name": "hello.sh",
  "Status": 0,
  "Response": "Hello\n",
  "CreatedAt": "2018-05-20T00:38:36.616995092-07:00",
  "StartedAt": "2018-05-20T00:38:36.619777692-07:00",
  "EndedAt": "2018-05-20T00:38:36.627481175-07:00"
}
```
