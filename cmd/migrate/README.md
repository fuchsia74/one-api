# One API Database Migration Tool

A command-line tool for migrating One API data between SQLite, MySQL, and PostgreSQL.

## Features

- **Auto database type detection** via DSN scheme (`sqlite://`, `mysql://`, `postgres://`)
- **Validation modes**: plan, validate-only, and dry-run (no data changes)
- **Concurrent batch migration** (`-workers`, `-batch-size`) for large datasets
- **Idempotent**: safe to re-run; existing rows are skipped / merged by primary key logic
- **PostgreSQL sequence alignment** after data copy
- **Structured logging** (use `-verbose` for detailed per-table progress)

## Installation

From project root:

```bash
go build -o migrate ./cmd/migrate
```

## Quick Start

```bash
# Migrate from SQLite file to PostgreSQL
./migrate \
    -source-dsn="sqlite:///absolute/or/relative/path/one-api.db" \
    -target-dsn="postgres://user:pass@localhost:5432/oneapi?sslmode=disable"

# Migrate from MySQL to PostgreSQL with more workers & bigger batches
./migrate \
    -source-dsn="mysql://user:pass@tcp(localhost:3306)/oneapi" \
    -target-dsn="postgres://user:pass@localhost/oneapi?sslmode=disable" \
    -workers=8 -batch-size=2000
```

## Operation Modes

| Mode          | Flags            | Notes                                                                       |
| ------------- | ---------------- | --------------------------------------------------------------------------- |
| Show Plan     | `-show-plan`     | Calculates table list & record counts (target DSN optional but recommended) |
| Validate Only | `-validate-only` | Connects & runs compatibility checks; no schema or data changes             |
| Dry Run       | `-dry-run`       | Full pipeline minus writes (no INSERT/sequence updates)                     |
| Migration     | _(default)_      | Requires both DSNs; performs schema + data + post steps                     |

Mutual exclusivity rules (enforced):

- `-dry-run` cannot be combined with `-validate-only`
- `-show-plan` cannot be combined with either of the above

## Flags

| Flag               | Description                                                           |
| ------------------ | --------------------------------------------------------------------- |
| `-source-dsn`      | Source DB DSN (required)                                              |
| `-target-dsn`      | Target DB DSN (required except with `-show-plan` or `-validate-only`) |
| `-dry-run`         | Execute all logic without mutating target                             |
| `-validate-only`   | Connectivity & compatibility checks only                              |
| `-show-plan`       | Print migration plan (tables + counts) and exit                       |
| `-verbose`         | Extra logging (per-table batches etc.)                                |
| `-skip-validation` | Skip pre-migration validator (not recommended)                        |
| `-workers`         | Concurrent workers for data copy (default 4)                          |
| `-batch-size`      | Rows per fetch/insert batch (default 1000)                            |
| `-h`               | Help                                                                  |
| `-v`               | Version                                                               |

## DSN Formats

The tool infers database type from the DSN scheme prefix.

### SQLite

```text
sqlite:///absolute/or/relative/path/one-api.db
sqlite://./one-api.db
sqlite://:memory:          # (testing only)
```

### MySQL

```text
mysql://user:pass@tcp(host:3306)/oneapi
mysql://user:pass@tcp(host:3306)/oneapi?charset=utf8mb4&parseTime=True&loc=Local
```

### PostgreSQL

```text
postgres://user:pass@localhost:5432/oneapi
postgres://user:pass@localhost/oneapi?sslmode=disable
postgresql://user:pass@localhost/oneapi?sslmode=require   # alias scheme
```

## Migration Flow

1. Parse & validate flags
2. (Optional) Show plan / validate-only / dry-run gating
3. Pre-migration validation (unless `-skip-validation`): schema presence, counts, basic type checks
4. Target schema auto-migration (GORM) if performing real migration
5. Batched data copy using `-workers` \* `-batch-size`
6. PostgreSQL sequence alignment (if target is PostgreSQL and not dry-run)
7. Post-migration validation (record counts)

## Tables Processed (ordered)

1. `users`
2. `options`
3. `tokens`
4. `channels`
5. `redemptions`
6. `abilities`
7. `logs`
8. `user_request_costs`
9. `traces`

## PostgreSQL Sequence Alignment

After copying data into PostgreSQL the tool sets each sequence to `MAX(id)+1` so that new inserts do not collide. Logged per table (skipped in dry-run).

## Performance Tuning

- Increase `-workers` to leverage CPU / network parallelism
- Adjust `-batch-size` (larger batches reduce round trips but increase memory usage)
- Use `-verbose` during initial tuning, then omit for quieter runs

## Safety & Best Practices

Before:

1. Backup both source & (empty) target
2. Test with a copy and `-show-plan`, then `-validate-only`, then `-dry-run`
3. Schedule off-peak for large datasets

During:

1. Keep process uninterrupted
2. Monitor logs (errors halt the run)

After:

1. Review counts & app functionality
2. Retain backup until confidence established
3. Update application connection settings

## Troubleshooting

| Issue                      | Suggestions                                                                    |
| -------------------------- | ------------------------------------------------------------------------------ |
| Connection failures        | Check credentials / host / firewall / DSN scheme                               |
| Slow migration             | Increase `-workers`, tune `-batch-size`                                        |
| ID collisions (PostgreSQL) | Ensure run not done with `-dry-run`; sequence step only runs in real migration |
| Validation errors          | Fix underlying schema/data, re-run `-validate-only` first                      |

If migration stops: diagnose via logs, correct issue, and re-run (safe / idempotent).

## Example End-to-End Session

```bash
# 1. Plan
./migrate -source-dsn="sqlite:///./one-api.db" -target-dsn="postgres://user:pass@localhost/oneapi?sslmode=disable" -show-plan

# 2. Validate
./migrate -source-dsn="sqlite:///./one-api.db" -target-dsn="postgres://user:pass@localhost/oneapi?sslmode=disable" -validate-only

# 3. Dry run
./migrate -source-dsn="sqlite:///./one-api.db" -target-dsn="postgres://user:pass@localhost/oneapi?sslmode=disable" -dry-run -verbose

# 4. Real migration
./migrate -source-dsn="sqlite:///./one-api.db" -target-dsn="postgres://user:pass@localhost/oneapi?sslmode=disable" -workers=6 -batch-size=1500 -verbose
```

## Version

Current version: 1.0.0
