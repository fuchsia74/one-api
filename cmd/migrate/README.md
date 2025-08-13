# One API Database Migration Tool

A comprehensive command-line tool for migrating data between different database types supported by One API: SQLite, MySQL, and PostgreSQL.

## Features

- **Multi-database support**: Migrate between SQLite, MySQL, and PostgreSQL
- **Safe migration**: Comprehensive validation and dry-run capabilities
- **Batch processing**: Efficient handling of large datasets
- **Progress tracking**: Detailed logging and progress indicators
- **Data integrity**: Automatic validation of migration results
- **Backup recommendations**: Built-in safety recommendations

## Installation

Build the migration tool from the One API project root:

```bash
go build -o migrate ./cmd/migrate
```

## Usage

### Basic Migration

```bash
# Migrate from SQLite to MySQL
./migrate -source-type=sqlite -source-dsn="./one-api.db" \
          -target-type=mysql -target-dsn="user:pass@tcp(localhost:3306)/oneapi"

# Migrate from MySQL to PostgreSQL
./migrate -source-type=mysql -source-dsn="user:pass@tcp(localhost:3306)/oneapi" \
          -target-type=postgres -target-dsn="postgres://user:pass@localhost/oneapi?sslmode=disable"
```

### Safety Features

```bash
# Dry run - validate without making changes
./migrate -source-type=sqlite -source-dsn="./one-api.db" \
          -target-type=mysql -target-dsn="user:pass@tcp(localhost:3306)/oneapi" \
          -dry-run

# Validation only - check connections and compatibility
./migrate -source-type=sqlite -source-dsn="./one-api.db" \
          -target-type=mysql -target-dsn="user:pass@tcp(localhost:3306)/oneapi" \
          -validate-only

# Show migration plan
./migrate -source-type=sqlite -source-dsn="./one-api.db" \
          -target-type=mysql -target-dsn="user:pass@tcp(localhost:3306)/oneapi" \
          -show-plan
```

### Advanced Options

```bash
# Verbose logging
./migrate -source-type=sqlite -source-dsn="./one-api.db" \
          -target-type=mysql -target-dsn="user:pass@tcp(localhost:3306)/oneapi" \
          -verbose

# Skip validation (not recommended)
./migrate -source-type=sqlite -source-dsn="./one-api.db" \
          -target-type=mysql -target-dsn="user:pass@tcp(localhost:3306)/oneapi" \
          -skip-validation
```

## Command Line Options

| Option             | Description                                     |
| ------------------ | ----------------------------------------------- |
| `-source-type`     | Source database type (sqlite, mysql, postgres)  |
| `-source-dsn`      | Source database connection string               |
| `-target-type`     | Target database type (sqlite, mysql, postgres)  |
| `-target-dsn`      | Target database connection string               |
| `-dry-run`         | Perform validation without making changes       |
| `-validate-only`   | Only validate connections and compatibility     |
| `-show-plan`       | Show migration plan and exit                    |
| `-verbose`         | Enable verbose logging                          |
| `-skip-validation` | Skip pre-migration validation (not recommended) |
| `-help`            | Show help message                               |
| `-version`         | Show version information                        |

## Database Connection Strings

### SQLite

```
# File-based database
./path/to/database.db

# With options
./path/to/database.db?_busy_timeout=5000

# In-memory database (for testing)
:memory:
```

### MySQL

```
# Basic format
user:password@tcp(host:port)/database

# Examples
root:password@tcp(localhost:3306)/oneapi
user:pass@tcp(192.168.1.100:3306)/oneapi?charset=utf8mb4&parseTime=True&loc=Local
```

### PostgreSQL

```
# Basic format
postgres://user:password@host:port/database

# Examples
postgres://user:password@localhost:5432/oneapi
postgres://user:password@localhost:5432/oneapi?sslmode=disable
postgresql://user:password@localhost:5432/oneapi?sslmode=require
```

## Migration Process

The migration tool follows these steps:

1. **Connection Validation**: Verify connections to both source and target databases
2. **Pre-migration Validation**: Check compatibility, data integrity, and potential issues
3. **Schema Migration**: Create tables in target database using GORM auto-migration
4. **Data Migration**: Transfer data in batches with progress tracking
5. **PostgreSQL Sequence Fix**: Update PostgreSQL sequences to match maximum ID values (PostgreSQL targets only)
6. **Post-migration Validation**: Verify data integrity and record counts

## Tables Migrated

The tool migrates all One API tables in the correct order to respect foreign key constraints:

1. `users` - User accounts and settings
2. `options` - System configuration options
3. `tokens` - API tokens and access keys
4. `channels` - API provider channels
5. `redemptions` - Redemption codes
6. `abilities` - Channel abilities and permissions
7. `logs` - Usage and system logs
8. `user_request_costs` - Request cost tracking
9. `traces` - Request tracing and performance monitoring

## PostgreSQL Sequence Management

When migrating **to PostgreSQL** from other databases, the tool automatically handles sequence management:

- **Problem**: PostgreSQL uses sequences for auto-increment columns, but migrated data retains original ID values while sequences start from 1
- **Solution**: The tool automatically updates all PostgreSQL sequences to start from the maximum migrated ID value + 1
- **Tables Handled**: All tables with auto-increment ID columns (users, tokens, channels, options, redemptions, abilities, logs, user_request_costs)
- **Logging**: Detailed logs show which sequences were updated and their new starting values

This ensures that new records created after migration will have correct, non-conflicting ID values.

## Best Practices

### Before Migration

1. **Always backup your databases** before starting migration
2. **Test on a copy** of your production data first
3. **Run validation** to check for potential issues
4. **Use dry-run mode** to verify the migration plan
5. **Schedule during off-peak hours** for large datasets

### During Migration

1. **Monitor the process** and watch for errors
2. **Ensure stable network connection** for remote databases
3. **Have sufficient disk space** for both source and target
4. **Don't interrupt the process** once started

### After Migration

1. **Verify data integrity** using the built-in validation
2. **Test application functionality** with the new database
3. **Keep backups** until you're confident in the migration
4. **Update connection strings** in your application

## Troubleshooting

### Common Issues

**Connection Errors**

- Verify database credentials and network connectivity
- Check firewall settings for remote databases
- Ensure target database exists and user has proper permissions

**Large Dataset Migration**

- Use verbose mode to monitor progress
- Ensure sufficient disk space and memory
- Consider running during off-peak hours

**Data Type Compatibility**

- Review warnings from pre-migration validation
- Some data types may be automatically converted
- Test thoroughly after migration

### Error Recovery

If migration fails:

1. Check the error logs for specific issues
2. Restore from backup if necessary
3. Fix the underlying issue
4. Re-run with dry-run mode to verify
5. Retry the migration

## Examples

### Complete Migration Workflow

```bash
# 1. Show migration plan
./migrate -source-type=sqlite -source-dsn="./one-api.db" \
          -target-type=mysql -target-dsn="user:pass@tcp(localhost:3306)/oneapi" \
          -show-plan

# 2. Run validation
./migrate -source-type=sqlite -source-dsn="./one-api.db" \
          -target-type=mysql -target-dsn="user:pass@tcp(localhost:3306)/oneapi" \
          -validate-only

# 3. Dry run
./migrate -source-type=sqlite -source-dsn="./one-api.db" \
          -target-type=mysql -target-dsn="user:pass@tcp(localhost:3306)/oneapi" \
          -dry-run -verbose

# 4. Actual migration
./migrate -source-type=sqlite -source-dsn="./one-api.db" \
          -target-type=mysql -target-dsn="user:pass@tcp(localhost:3306)/oneapi" \
          -verbose
```

## Support

For issues and questions:

- Check the troubleshooting section above
- Review the verbose logs for detailed error information
- Ensure you're using compatible database versions
- Test with a small dataset first

## Version

Current version: 1.0.0
