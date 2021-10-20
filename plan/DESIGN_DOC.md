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
started as). To start multiple jobs, the user will need to make multiple independent requests. Job
metadata will be stored in-memory as a mapping of the Job ID to JobInfo and `Cmd` (i.e., the Command
struct from `os/exec`).

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

As mentioned in the [Library](#library) section, we will store an in-memory mapping of each job's 
unique ID to JobInfo and Command. The `owner` field in JobInfo will be the Serial Number of the 
authenticated x509 client certificate. For this prototype, we'll have a self-signed CA and generate 
client certificates from it, which should be unique within the CA. When a job is created we will add 
a new record to this store. When a stop/stream request is made, we will check the authenticated 
certificate's serial number (and CA) against the job's serial number and server CA. This will return
the expected output for stop/stream if successful and an unauthorized error if not.

## Client Usage

The client will have 4 commands, `start`, `stop`, `get`, and `stream`. Here are some examples of 
client usage that will interact with the API:

### Start Job
`usage: ./client start --cert CERT_PATH --cert-key KEY_PATH COMMAND [ARGUMENTS ...]`
```bash
./client start ls -l
----
Job ID: ed954e11-ef5a-4b61-b698-8537055d0adc
Client ID (owner): 123456

Command: ls
Arguments: [-l]

State: CREATED
----
```

### Stop Job
`usage: ./client stop --cert CERT_PATH --cert-key KEY_PATH UUID`
```bash
./client stop ed954e11-ef5a-4b61-b698-8537055d0adc
----
Job ID: ed954e11-ef5a-4b61-b698-8537055d0adc
Client ID (owner): 123456

Command: ls
Arguments: [-l]

State: STOPPED
----
```

### Get Job Info
`usage: ./client get --cert CERT_PATH --cert-key KEY_PATH UUID`
```bash
./client get ed954e11-ef5a-4b61-b698-8537055d0adc
----
Job ID: ed954e11-ef5a-4b61-b698-8537055d0adc
Client ID (owner): 123456

Command: ls
Arguments: [-l]

State: RUNNING
----
```

### Stream Job Output
`usage: ./client stream --cert CERT_PATH --cert-key KEY_PATH UUID`
```bash
./client stream ed954e11-ef5a-4b61-b698-8537055d0adc
total 8
-rw-r--r--  1 rakinuddin  staff   74 15 Oct 17:06 README.md
drwxr-xr-x  4 rakinuddin  staff  128 15 Oct 17:05 plan
```

## Improvements/Out of Scope Features

Ideally, we would want to avoid running the provided commands as `root` (or the OS user that the 
server runs as). One improvement would be mapping each authenticated user to an internal OS user with 
some permission type (i.e., standard, admin) which limits the types of commands it can run. We could 
go further by spawning a separate instance to run each user's job on, to prevent multiple users' jobs 
from interfering with each other.

In the [Authentication](#authentication) section we mentioned generating a self-signed CA and then 
generating client certificates from it to use in mTLS authentication. We could provide an API that
allows users to register and retrieve client certificates to authenticate with the main API.
This way we control unique Serial Number generation for each client as well as handle certificate
revocation and rotation.

I mentioned storing job data in-memory in a few places. On a production-ready system we would replace
these with persistent datastores (such as a database) to maintain data in the event of a reboot and
allow the system to scale properly.

The current implementation of having the job output piped into logs (i.e., in files under `/var/log/linux-process-runner/`)
can fill up the disk if we were to run a command with an extremely large output. We could mitigate
this on a production system by streaming the output directly to a scalable storage solution (for 
example, AWS S3) or compressing and shipping logs in batches to a persistent datastore via something 
like `logrotate`.
