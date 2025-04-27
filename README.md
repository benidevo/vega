# ProspecTor

![Build Status](https://github.com/benidevo/prospector/workflows/Go%20Docker%20CI/badge.svg)

## Overview

ProspecTor is an application designed for job prospecting. It utilizes a monolithic architecture, runs within Docker containers managed by Docker Compose, uses the Gin web framework, and persists data using SQLite.

## Technology Stack

* Go
* Gin (Web Framework)
* SQLite (Database)
* Docker

## Getting Started

### Prerequisites

* Docker
* Docker Compose

### Running the Application

1. Clone the repository.
2. Navigate to the project directory.
3. Run the application using the Makefile:

    ```sh
    make run
    ```

    This command will build the Docker image (if necessary) and start the application container in detached mode.

## Available Commands (via Makefile)

* `make run`: Builds and starts the application containers.
* `make test`: Runs the test suite within the application container.
* `make stop`: Stops the application containers.
* `make logs`: Tails the logs from the application container.
* `make enter-app`: Opens a shell inside the running application container.
* `make format`: Formats the Go code within the application container using `go fmt` and `go vet`.

## Development

### Git Hooks

This project uses Git hooks to ensure code quality. To set up:

```bash
make setup-hooks
```

This will install the pre-commit hooks defined in the `.pre-commit-config.yaml` file. The hooks will run automatically on commit and push, ensuring that code quality checks are performed.
