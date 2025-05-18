# ProspecTor

![Build Status](https://github.com/benidevo/prospector/workflows/CI/badge.svg)

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

### Database Migration Commands

* `make migrate-create`: Create a new migration file (will prompt for migration name).
* `make migrate-up`: Apply all pending migrations.
* `make migrate-down`: Rollback the most recent migration.
* `make migrate-reset`: Rollback all migrations.
* `make migrate-force`: Set the migration version (will prompt for version).

## Development

### Database Migrations

ProspecTor uses [golang-migrate](https://github.com/golang-migrate/migrate) for database schema management. Migrations are automatically applied when the application starts.

Migration files are stored in the `migrations/sqlite` directory and follow the naming convention:

```
{version}_{description}.{up|down}.sql
```

For example:
* `000001_create_users_table.up.sql`: Creates the users table
* `000001_create_users_table.down.sql`: Drops the users table

#### Working with Migrations

1. **Creating a new migration**:

   ```bash
   make migrate-create
   # Enter migration name when prompted, e.g., "add_jobs_table"
   ```

2. **Edit the migration files**:
   After creation, edit the generated SQL files:
   * `{version}_add_jobs_table.up.sql`: Add SQL to create new tables/columns
   * `{version}_add_jobs_table.down.sql`: Add SQL to revert the changes

3. **Apply migrations**:

   ```bash
   make migrate-up
   ```

4. **Rollback migrations**:

   ```bash
   make migrate-down  # Rollback one migration
   make migrate-reset # Rollback all migrations
   ```

5. **Fix migration state**:
   If migrations get into a bad state, you can force a specific version:

   ```bash
   make migrate-force
   # Enter version number when prompted
   ```

### Git Hooks

This project uses Git hooks to ensure code quality. To set up:

```bash
make setup-hooks
```

This will install the pre-commit hooks defined in the `.githooks/pre-commit` file. The hooks will run automatically on commit and push, ensuring that code quality checks are performed.
