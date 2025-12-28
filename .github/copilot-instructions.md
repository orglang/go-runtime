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

## Feature-sliced structure

### Framework agnostic

- `core.go`: Pure domain logic
    - Domain models (core models)
    - API interfaces (primary ports)
    - Service structs (core behaviors)
- `me.go`: Pure message exchange (ME) logic
    - Exchange specific DTO's (borderline models)
- `vp.go`: Pure view presentation (VP) logic
    - Presentation specific DTO's (borderline models)
- `ds.go`: Pure data storage (DS) logic
    - Storage specific DTO's (borderline models)
    - Repository interfaces (secondary ports)
- `iv.go`: Pure input validation (IV) logic
    - Message validation harness
    - Props validation harness
- `oc.go`: Pure option configuration (OC) logic
    - Configuration specific DTO's (borderline models)
- `tc.go`: Pure type conversion (TC) logic
    - Domain to domain conversions
    - Domain to message conversions and vice versa
    - Domain to data conversions and vice versa

### Framework specific

- `di_fx.go`: Fx (dependency injection library) specific component definitions
- `me_echo.go`: Echo (web framework) specific controllers (primary adapters)
- `vp_echo.go`: Echo (web framework) specific presenters (primary adapters)
- `me_resty.go`: Resty (HTTP client library) specific API implementations (secondary adapters for external use)
- `ds_pgx.go`: pgx (PostgreSQL Driver and Toolkit) specific repository iplementations (secondary adapters for internal use)
- `iv_ozzo.go`: Ozzo (validation library) specific validation harness
- `oc_viper.go`: Viper (configuration library) specific configuration harness
- `tc_goverter.go`: Goverter (type conversion tool) specific conversion harness
- `vp_template.html`: HTML template engine (standart library) specific views

## Workflow

- `task sources` - before commit to task or feature branch
- `task binaries` - before push to task or feature branch
- `task distros` - before pull request to feature branch
- `task stacks` - before pull request to main branch
