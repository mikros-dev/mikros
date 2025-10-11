# mikros

<p align="center">
  <a href="https://mikros.dev">
    <img src="https://raw.githubusercontent.com/mikros-dev/mikros/main/.assets/images/go-logo.png" alt="mikros logo" width="650"/>
  </a>
</p>

<h3 align="center">mikros is a Go framework for creating applications.</h3>

<p align="center">
  <a href="https://pkg.go.dev/github.com/mikros-dev/mikros"><img src="https://pkg.go.dev/badge/github.com/mikros-dev/mikros.svg" alt="Go Reference"></a>
  <a href="https://mikros.dev"><img src="https://img.shields.io/badge/site-mikros.dev-blue" alt="Docs"></a>
  <a href="https://goreportcard.com/report/github.com/mikros-dev/mikros"><img src="https://goreportcard.com/badge/github.com/mikros-dev/mikros" alt="Go Report Card"></a>
  <a href="https://github.com/mikros-dev/mikros/releases"><img src="https://img.shields.io/github/v/release/mikros-dev/mikros?sort=semver" alt="GitHub release"></a>
  <a href="https://github.com/mikros-dev/mikros/blob/main/LICENSE"><img src="https://img.shields.io/github/license/mikros-dev/mikros" alt="License"></a>
</p>

## Introduction

This framework is an API built to ease and standardize the creation of applications
that need to run for long periods, usually executing indefinitely, performing some
specific operation. But it also supports standalone applications that execute its
task and finish right after.

Its main idea is to allow the user to create (or implement) an application, written
in Go, of the following categories:

* gRPC: an application with an API defined from a [protobuf](https://protobuf.dev) file.
* HTTP (spec): an HTTP application with its API defined from a [protobuf](https://protobuf.dev) file.
* HTTP (std): an HTTP application built directly with Go’s standard `net/http` library (routers, handlers, middlewares).
* worker: a general-purpose application, without a defined API, with the ability to execute any code for long periods.
* script: also a general-purpose application, without a defined API, but that only needs to execute a single function and stop.

### Service

Service here is considered an application that may or may not remain running
indefinitely, performing some type of task or waiting for commands to activate it.

The framework consists of an SDK that facilitates the creation of these applications
in a way that standardizes their code so that they all perform tasks with the
same behavior and are written in a very similar manner. In addition to providing
flexibility, allowing these applications to also be customized when necessary.

Building a service using the framework's SDK must adhere to the following points:

* Having a structure where mandatory methods according to its category must be implemented;
* Initialize the SDK correctly;
* Have a configuration file, called `service.toml`, containing information about itself and its functionalities.

### Example of a service

The following example demonstrates how to create a service of a `script`
type. The `service` structure implements an [interface](apis/services/script/api.go)
that makes it being supported by this type of service inside the framework.

```golang
package main

import (
    "context"

    "github.com/mikros-dev/mikros"
    "github.com/mikros-dev/mikros/components/options"
    logger_api "github.com/mikros-dev/mikros/apis/features/logger"
)

// service is a structure that will hold all required data and information
// of the service itself. It is also the place to define all features that
// the service will use.
type service struct {
    Logger logger_api.LoggerAPI `mikros:"feature"`
}

func (s *service) Run(ctx context.Context) error {
    s.Logger.Info(ctx, "service Run method executed")
    return nil
}

func (s *service) Cleanup(ctx context.Context) error {
    s.Logger.Info(ctx, "cleaning up things")
    return nil
}

func main() {
    // Creates a new service using the framework API.
    svc := mikros.NewService(&options.NewServiceOptions{
        Service: map[string]options.ServiceOptions{
            "script": &options.ScriptServiceOptions{},
        },
    })

    // Puts it to execute.
    svc.Start(&service{})
}
```

It must have a `service.toml` file with the following content:

```toml
name = "script-example"
types = ["script"]
version = "v1.0.0"
language = "go"
product = "Matrix"
```

When executed, it outputs the following (with a different time according to the execution):

```bash
{"time":"2024-02-09T07:54:57.159265-03:00","level":"INFO","msg":"starting service","service.name":"script-example","service.type":"script","service.version":"v1.0.0","service.env":"local","service.product":"Matrix"}
{"time":"2024-02-09T07:54:57.159405-03:00","level":"INFO","msg":"starting dependent services","service.name":"script-example","service.type":"script","service.version":"v1.0.0","service.env":"local","service.product":"Matrix"}
{"time":"2024-02-09T07:54:57.159443-03:00","level":"INFO","msg":"service resources","service.name":"script-example","service.type":"script","service.version":"v1.0.0","service.env":"local","service.product":"Matrix","svc.http.auth":"false"}
{"time":"2024-02-09T07:54:57.159449-03:00","level":"INFO","msg":"service is running","service.name":"script-example","service.type":"script","service.version":"v1.0.0","service.env":"local","service.product":"Matrix","service.mode":"script"}
{"time":"2024-02-09T07:54:57.159458-03:00","level":"INFO","msg":"service Run method executed","service.name":"script-example","service.type":"script","service.version":"v1.0.0","service.env":"local","service.product":"Matrix"}
{"time":"2024-02-09T07:54:57.159464-03:00","level":"INFO","msg":"stopping service","service.name":"script-example","service.type":"script","service.version":"v1.0.0","service.env":"local","service.product":"Matrix"}
{"time":"2024-02-09T07:54:57.159467-03:00","level":"INFO","msg":"stopping dependent services","service.name":"script-example","service.type":"script","service.version":"v1.0.0","service.env":"local","service.product":"Matrix"}
{"time":"2024-02-09T07:54:57.159804-03:00","level":"INFO","msg":"cleaning up things","service.name":"script-example","service.type":"script","service.version":"v1.0.0","service.env":"local","service.product":"Matrix"}
{"time":"2024-02-09T07:54:57.159815-03:00","level":"INFO","msg":"service stopped","service.name":"script-example","service.type":"script","service.version":"v1.0.0","service.env":"local","service.product":"Matrix"}
```

## Roadmap

* ~~Support for custom tags, key-value, declared in the 'service.toml' file, to be added in each log line.~~
* ~~Support for receiving custom 'service.toml' definition rules.~~
* ~~Support for HTTP services without being declared in a protobuf file.~~
* ~~Remove unnecessary Logger APIs.~~

## License

[Mozilla Public License 2.0](LICENSE)
