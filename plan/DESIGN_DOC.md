# Linux Process Runner Design Doc

The goal of the Linux Process Runner is to provide users
with the ability to run Linux processes in the form of a job.

## Library

First we will define the concept of a job for this library. A job will
consist of a unique ID (which will be a generated UUID), the command, supplied 
arguments, owning user and state (i.e., running, stopped).

The library will contain the bulk of the logic to start, stop, and query
for job info as well as getting the output of a running job.

Users will be able to start a single job by providing a command and arguments
in a request. The command will be run as a new process as `root` (or whichever user the server is 
started as). To start multiple jobs, the user will need to make multiple independent requests.

When a job starts, its output will be appended to a file under `/var/log/linux-process-runner/<job_id>.log`. 
Users looking to stream the output of a job will be met with the output of this log file
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

Any user will be able to query for a job's metadata. Each user will only be able to start, stop and stream their own jobs.

We will store an in-memory mapping of each job's unique ID to the Serial Number of the authenticated 
x509 client certificate. For this prototype, we'll have a self-signed CA and generate client 
certificates from it, which should be unique within the CA. When a job is created we will add 
a new record to this store. When a stop/stream request is made, we will check for this mapping or 
return an unauthorized error if a mapping is not found.

## Improvements/Out of Scope Features

Ideally, we would want to avoid running the provided commands as `root` (or the OS user that the 
server runs as). One improvement would be mapping each authenticated user to an internal OS user with 
some permission type (i.e., standard, admin) which limits the type of commands it can run. We could 
go further by spawning a separate instance to run each user's job on, to prevent multiple users' jobs 
from interfering with each other.

In the [Authentication](#authentication) section we mentioned generating a self-signed CA and then 
generating client certificates from it to use in mTLS authentication. We could provide an API that
allows users to register and retrieve client certificates to authenticate with the main API.
This way we control unique Serial Number generation for each client as well as handle certificate
revocation and rotation.
