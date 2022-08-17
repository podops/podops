# podops

PodOps is a set of backend services and command line utilities to create podcast feeds from simple markdown files.

### TL;DR

... or how to build a podcast in 5 steps:

```shell
cd /the/podcast/location

# prepare the podcast repo
po new

# ... do some yaml editing

# create an episode
po template episode

# ... more yaml editing

# build the feed
po build

# register the podcast with the CDN at https://cdn.podops.dev
po init

# sync the podcast with the CDN
po sync
```

### Installation

The command line utility `po` can be installed from the source code location:

```shell
$ git clone https://github.com/podops/podops.git <some_location>

$ cd <some_location>

$ make cli
```

### Local development

Get the source code

```shell
$ git clone https://github.com/podops/podops.git <some_location>

$ cd <some_location>
$ go mod tidy
$ make test_build
```

#### Command Line

Run the CLI from local source code:

```shell
cd cmd/cli
go run cli.go <command>
```

In order to target a local API service, set the `PODOPS_API_ENDPOINT` environment variable:

```shell
cd cmd/cli
PODOPS_API_ENDPOINT=http://localhost:8080 go run cli.go <command>
```

#### API Service

Run the API service from local source code:

```shell
cd cmd/api
go run main.go
```

In order to set the location of the CDN folders, use environment variables `PODOPS_STORAGE_LOCATION` and `PODOPS_STATIC_LOCATION`, e.g.

```shell

PODOPS_STORAGE_LOCATION=/path/to/cdn PODOPS_STATIC_LOCATION=/path/to/public go run main.go

```
