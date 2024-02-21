# mikro\_s

## About

`mikro_s` is a Go framework for creating applications.

## Introduction

This framework is an API built to ease and standardize the creation of applications
that need to run for long periods, usually executing indefinitely, performing some
specific operation. But it also supports standalone applications that execute its
task and finishes right after.

Its main idea is to allow the user to create (or implement) an application, written
in Go, of the following categories:

* gRPC: an application with an API defined from a [protobuf](https://protobuf.dev) file.
* HTTP: an HTTP server-type application, with its API defined from a protobuf file.
* native: a general-purpose application, without a defined API, with the ability to execute any code for long periods.
* script: also a general-purpose application, without a defined API, but that only needs to execute a single function and stop.

### Service

Service, here, is considered an application that may or may not remain running
indefinitely, performing some type of task or waiting for commands to activate it.

The framework consists of an SDK that facilitates the creation of these applications
in a way that standardizes their code, so that they all perform tasks with the
same behavior and are written in a very similar manner. In addition to providing
flexibility, allowing these applications to also be customized when necessary.

Building a service using the framework's SDK must adhere to the following points:

* Have a struct where mandatory methods according to its category must be implemented;
* Initialize the SDK correctly;
* Have a configuration file, called `service.toml`, containing information about itself and its functionalities.

### Example of a service

The following example demonstrates how to create a service of a `script`
type. The `service` structure implements an [interface](apis/services/script/script.go)
that makes it being supported by this type of service inside the framework.

```golang
package main

import (
    "context"

    "github.com/somatech1/mikros"
    "github.com/somatech1/mikros/components/options"
)

// service is a structure that will hold all required data and information
// of the service itself.
//
// It must have declared, at least, a member of type *mikros.Service. This
// gives it the ability of being used and supported by the framework internals.
type service struct {
    *mikros.Service
}

func (s *service) Run(ctx context.Context) error {
    s.Logger().Info(ctx, "service Run method executed")
    return nil
}

func (s *service) Cleanup(ctx context.Context) error {
    s.Logger().Info(ctx, "cleaning up things")
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

When executed, it outputs the following (with a different time according the execution):

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

* Support for receiving custom 'service.toml' definition rules.
* Support for HTTP services without being declared in a protobuf file.
* Support for custom tags, key-value declared in the 'service.toml' file, to be added in each log line.
* Remove unnecessary Logger APIs.

## License

[Mozilla Public License 2.0](LICENSE)
