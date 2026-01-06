# Orglang

This a programming language to support organization development and management.

## Repository structure

- `app`: Runnable application definitions
  - `web`: Web application definitions
- `adt`: Reusable data types
  - `typedef`: Type definition aggregate
  - `typeexp`: Type expression value
  - `procdec`: Process declaration aggregate
  - `procdef`: Process definition entity
  - `procexp`: Process expression value
  - `procexec`: Process execution aggregate
  - `pooldec`: Pool declaration aggregate
  - `pooldef`: Pool definition entity
  - `poolexec`: Pool execution aggregate
  - `syndec`: Synonym declaration value
  - `termctx`: Term context value
  - `identity`: Identification value
  - `polarity`: Polarization value
  - `qualsym`: Qualified symbol value
  - `revnum`: Revision number value
- `lib`: Reusable behavior types
  - `ck`: Configuration keeper harness
  - `lf`: Logging framework harness
  - `sd`: Storage driver harness
  - `te`: Template engine harness
  - `ws`: Web server harness
- `db`: Storage schema definitions
  - `postgres`: PostgreSQL specific definitions
- `orch`: Orchestration harness definitions (local)
  - `task`: Task (build tool) harness definitions
- `proto`: Prototype endeavors
- `stack`: System level definitions
- `test`: End-to-end tests and harness

## Abstraction aspects

### Abstraction scale

- `aggregate`: Concurrency-aware abstraction
    - Consumed by controller adapters
    - Specified by `API` interfaces
    - Implemented by `Service` structs
- `entity`: Identity-aware abstraction
    - Consumed by `Service` structs
    - Specified by `Repo` interfaces
    - Implemented by DAO adapters
- `value`: Classical data abstraction
    - Consumed by `entity` or `aggregate` structs
    - Specified by `ADT` types and/or interfaces
    - Implemented by concrete types and/or structs

### Abstraction lifecycle

- `dec`: Declaration phase
- `def`: Definition phase
- `exp`: Expression phase
- `exec`: Execution phase

### Abstraction slice

- `ref`: Machine-readable pointer to an abstraction
- `spec`: Specification for abstraction creation
- `rec`: Record for abstraction retrieval (excluding sub abstractions)
- `mod`: Modification for abstraction update (including sub abstractions)
- `snap`: Snapshot for abstraction retrieval (including sub abstractions)

### Abstraction artifact 

- `sources`: Human-readable source code artifacts
- `binaries`: Machine-readable binary artifacts (optional)
- `distros`: Distributable artifacts (images, archives, etc.)
- `stacks`: Deployable artifacts (ansible playbooks, helm charts, etc.)

## Abstraction structure

### Toolkit agnostic

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
- `pc.go`: Pure property configuration (PC) logic
    - Configuration specific DTO's (borderline models)
- `tc.go`: Pure type conversion (TC) logic
    - Domain to domain conversions
    - Domain to message conversions and vice versa
    - Domain to data conversions and vice versa

### Toolkit specific

- `di_fx.go`: Fx (dependency injection library) specific component definitions
- `me_echo.go`: Echo (web framework) specific controller definitions (primary adapters)
- `vp_echo.go`: Echo (web framework) specific presenter definitions (primary adapters)
- `me_resty.go`: Resty (HTTP library) specific client definitions (secondary adapters for external use)
- `ds_pgx.go`: pgx (PostgreSQL driver and toolkit) specific DAO definitions (secondary adapters for internal use)
- `iv_ozzo.go`: Ozzo (validation library) specific validation definitions
- `tc_goverter.go`: Goverter (type conversion tool) specific conversion definitions
- `vp/bs5/*.html`: Go's built-in `html/template` and Bootstrap 5 (frontend toolkit) specific presentation definitions

## Workflow

- `task sources` - before commit to task or feature branch
- `task binaries` - before push to task or feature branch
- `task distros` - before pull request to feature branch
- `task stacks` - before pull request to main branch
