# protobuf-workspace

Repository to centralize all APIs and protobuf specifications.

## Using the repository

The repository management and task execution are entirely performed through the
command line using the make command.

By running the following command, you can list all the currently supported
operations in the repository:
```!bash
make help
```

## Installing dependencies

To generate code and documentation from protobuf files, some tools need to be
installed in the development environment.

This can be done by running the command:
```!bash
make setup
```

## Generating source files and documentation

Protobuf files are used as the source of definitions within systems, aiming to
centralize the specification of the following functionalities:

* Definition of common values for services and systems, such as events and error
codes.
* Declaration of entities, their fields, and database-related properties, such
as indexes, primary keys, etc.
* Declaration of APIs for internal and external services.

These files use a specific syntax for their manipulation, which can be referenced
[here](https://protobuf.dev/programming-guides/proto3/).

In this repository, protobuf files serve not only as a source for backend services
but also as integration APIs for the frontend. Additionally, they are used to
generate documentation in the [OpenAPI](https://swagger.io/specification/v3/) format.

This entire process is executed by running the following command at the root of
the repository:
```!bash
make
```
