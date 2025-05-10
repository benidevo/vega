# ProspecTor Scheduler

The ProspecTor Scheduler is a command-line tool for managing administrative tasks and scheduled operations in the ProspecTor system.

## Overview

The scheduler provides a flexible command registry system for executing administrative and maintenance tasks that don't belong in the main application flow. It is particularly useful for operations that:

- Need to be run on a schedule (via cron)
- Require privileged access
- Perform system maintenance
- Configure system resources

## Usage

The scheduler is available as a separate executable at `cmd/scheduler/main.go`. To use the scheduler:

```bash
# Run directly (during development)
go run cmd/scheduler/main.go <command> [arguments]

# Or use the compiled binary (in production)
./scheduler <command> [arguments]
```

If run without arguments, the scheduler will display a list of available commands.

## Available Commands

### sync-users

Synchronizes user accounts from a configuration file with the database.

```bash
./scheduler sync-users /path/to/config/users.yaml
```

#### Configuration File Format

The user configuration file can be in YAML or JSON format and should follow this structure:

```yaml
users:
  - username: admin
    password: secure_password
    email: admin@example.com
    reset_on_next_run: false

  - username: user2
    password: password2
    email: user2@example.com
    reset_on_next_run: true
```

**Fields:**

- `username`: The login username (required)
- `password`: Plain text password (will be hashed when stored in DB)
- `email`: User's email address
- `reset_on_next_run`: When true, the password will be updated during sync

#### Behavior

- If a user doesn't exist, a new user will be created
- If a user exists and `reset_on_next_run` is true, their password will be updated
- If a user exists and `reset_on_next_run` is false, no changes will be made

## Security Considerations

### Password Security

- Passwords in config files are stored in plain text for simplicity
- All passwords are hashed before being stored in the database
- In production, consider encrypting the configuration files and using restricted file permissions
- Config files should not be checked into version control

### Access Control

- The scheduler CLI should have restricted access in production environments
- Consider using a designated service account with appropriate permissions

## Adding New Commands

To add a new command to the scheduler:

1. Create a new file in the `internal/scheduler/commands/` directory
2. Implement the `Command` interface:

   ```go
   type Command interface {
       Name() string
       Description() string
       Execute(args []string) error
   }
   ```

3. Register your command in `internal/scheduler/app.go`

   ```go
   func (a *App) registerCommands() {
       // Existing commands
       a.RegisterCommand(commands.NewUserSyncCommand(a.db, a.config))

       // Add your new command
       a.RegisterCommand(commands.NewYourCommand(a.db, a.config))
   }
   ```

## Running as a Scheduled Task

To run the scheduler as a cron job:

```bash
# Example crontab entry (runs user sync daily at 2 AM)
0 2 * * * /path/to/prospector/scheduler sync-users /path/to/config/users.yaml >> /var/log/prospector/scheduler.log 2>&1
```

For Docker environments, you can use the provided cron configuration in the Docker Compose setup.
