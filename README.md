# linux-process-runner
A small library that runs arbitrary Linux commands

## Prerequisites
```
- a recent Go version, this project has been tested on go1.17.2 darwin/amd64
- OpenSSL 3.0.0 7 sep 2021 (or LibreSSL 2.8.3)
- libprotoc 3.17.3
```

## Building

To build the project, run the following command:
```bash
> make build-all
```
This generates the protobufs, certs, logs folder (under `/var/log/linux-process-runner/`)
and the server + client binaries

## Running

Once the project has been built, run the server like so:
```bash
> sudo ./bin/server
2021/10/27 05:53:07 starting server...
```
the server writes logs to `/var/log/linux-process-runner` so you will need to run as sudo unless
the running user has the appropriate permissions to write files

## Testing

To run the tests for this project, run the go tests:
```bash
> make tests
```
