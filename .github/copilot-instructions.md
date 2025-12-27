# Orglang

This a visual programming language to support organization development and management.

## Repository structure

- `app`: Process level definitions
    - `web`: Web application definitions
- `aat`: Reusable aggregate definitions
- `aet`: Reusable entity definitions
- `avt`: Reusable value definitions
- `db`: Storage schema definitions
    - `postgres`: PostgreSQL specific definitions
- `orch`: Orchestration harness definitions (local)
    - `task`: Task (build tool) harness definitions
- `proto`: Prototype endeavors
- `stack`: System level definitions
- `test`: End-to-end tests and harness

## Abstraction layers

### Definition abstractions

- `aat`: Abstract aggregate type
    - Consumed by controller adapter
    - Specified by `API` interface
    - Implemented by `Service` struct
- `aet`: Abstract entity type
    - Consumed by `Service` struct
    - Specified by `Repo` interface
    - Implemented by DAO adapter
- `avt`: Abstract value type
    - Specified by `ADT` type or interface

### Artifact abstractions

- `sources`: Human-readable source code artifacts
- `binaries`: Machine-readable binary artifacts (optional)
- `distros`: Distributable artifacts (images, archives, etc)
- `stacks`: Deployable artifacts (ansible playbooks, helm charts, etc)

### Execution abstractions

- `goroutine`: Framework level execution abstraction
- `process`: Application level execution abstraction
- `system`: Stack level execution abstraction

## Feature-slice structure

### Framework agnostic

- `core.go`: Pure domain logic
    - Domain models (core models)
    - API interfaces (primary ports)
    - Service structs (core behaviors)
- `me.go`: Pure message exchange (ME) logic
    - Exchange specific DTO's (borderline models)
    - Message validation harness
    - Message to domain mapping and vice versa
- `ds.go`: Pure data storage (DS) logic
    - Storage specific DTO's (borderline models)
    - Repository interfaces (secondary ports)
    - Domain to data mapping and vice versa
- `iv.go`: Pure input validation (IV) logic
- `tc.go`: Pure type conversion (TC) logic

### Framework specific

- `di_fx.go`: Fx (dependency injection system) specific component definitions
- `me_echo.go`: Echo (web framework) specific controllers (primary adapters)
- `ds_pgx.go`: pgx (PostgreSQL Driver and Toolkit) specific repository iplementations (secondary adapters for internal use)
- `sdk_resty.go`: Resty (HTTP client) specific API implementations (secondary adapters for external use)
- `iv_ozzo.go`: Ozzo (validation library) specific validation methods
- `tc_goverter.go`: Goverter (tool for creating type-safe converters) specific conversion methods

## Workflow

- `task sources` - before commit in task branch
- `task binaries` - before push in task branch
- `task packages` - before pull request
