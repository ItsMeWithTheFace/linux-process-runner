# Linux Process Runner Design Doc

The goal of the Linux Process Runner is to provide users
with the ability to run Linux processes in the form of a job.

## Library

First we will define the concept of a job for this library. A job will
consist of a unique ID, the command, supplied arguments, owning user
and state (i.e., running, stopped).

The library will contain the bulk of the logic to start, stop, and query
for job info as well as getting the output of a running job.

Users will be able to start a single job by providing a command and arguments
in a request. To start multiple jobs, the user will need to make multiple independent requests.
To keep things simple, we will run these commands on the host instance itself under a provisioned
user that is not `root`. Jobs started by any user will be run as this user internally. Ideally, we
would want to isolate ownership of the jobs like having each Linux process running as the 
authenticated user. We could go further by spawning a separate instance to run each user's job on, 
to prevent multiple users' jobs from interfering with each other.

When a job starts, its output will be appended to a file under `/var/log/<job_id>.log`. Users
looking to stream the output of a job will be met with the output of this log file
from the beginning of execution.

We can also stop a job by supplying the unique ID. This will terminate the job regardless of
what state it is in, but the log will persist on the disk for users looking to stream the output.
For simplicity, once a job is stopped, it cannot be started again. The user will need to start a new 
job with the same arguments if they would like to recreate their job. 

Finally, the job metadata can be queried for by supplying the unique ID. This will return the
job data itself such as the command, arguments, owner and state.

## Authentication

The API will serve as a simple layer to authenticate with the library and run jobs through the
client. We will use golang's `crypto/tls` library to setup an mTLS configuration with TLS 1.3
and the default cipher suites supported by the library which would be:

```
TLS_AES_128_GCM_SHA256
TLS_AES_256_GCM_SHA384
TLS_CHACHA20_POLY1305_SHA256
```

This setup will sacrifice compatibility with clients but ensures that more vulnerable ciphers
are not included.

## Authorization

Any user will be able to query for a job's metadata. Each user will only be able to start and stop
their own jobs.
