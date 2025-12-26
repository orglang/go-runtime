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
    - Implemented by dao adapter
- `avt`: Abstract value type
    - Specified by `ADT` type or interface

### Execution abstractions

- `goroutine`: Framework level execution abstraction
- `process`: Application level execution abstraction
- `system`: Stack level execution abstraction

## Feature-slice structure

### Framework agnostic

- `me.go`: Pure message exchange (ME) logic
    - Exchange specific DTO's (borderline models)
    - Message validation harness
    - Message to domain mapping and vice versa
- `core.go`: Pure domain logic
    - Domain models (core models)
    - API interfaces (primary ports)
    - Service structs (core behaviors)
- `ds.go`: Pure data storage (DS) logic
    - Storage specific DTO's (borderline models)
    - Repository interfaces (secondary ports)
    - Domain to data mapping and vice versa

### Framework specific

- `me_echo.go`: Echo (web framework) specific controllers (primary adapters)
- `ds_pgx.go`: pgx (PostgreSQL Driver and Toolkit) specific repository iplementations (secondary adapters for internal use)
- `sdk_resty.go`: Resty (HTTP client) specific API implementations (secondary adapters for external use)
- `di_fx.go`: Fx (dependency injection system) specific component definitions
