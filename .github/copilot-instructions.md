# Orglang

This a visual programming language to support organization development and management.

## Repository structure

- `app`: Runnable application definitions
  - `web`: Web application definitions
- `adt`: Reusable abstract data types
  - `typedef`: Type definition aggregate
  - `typealias`: Type alias entity
  - `procdecl`: Process declaration aggregate
  - `procdef`: Process definition entity
  - `procexec`: Process execution aggregate
  - `pooldecl`: Pool declaration aggregate
  - `pooldef`: Pool definition entity
  - `poolexec`: Pool execution aggregate
  - `expctx`: Expression context value object
  - `identity`: Identification value object
  - `polarity`: Polarization value object
  - `qualsym`: Qualified symbol value object
  - `revnum`: Revision number value object
- `lib`: Reusable behavior definitions
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

## Abstraction layers

### Abstract data types

- `aggregates`: Abstract aggregate types (AAT)
    - Consumed by controller adapters
    - Specified by `API` interfaces
    - Implemented by `Service` structs
- `entities`: Abstract entity types (AET)
    - Consumed by `Service` structs of `aggregate` types
    - Specified by `Repo` interfaces
    - Implemented by DAO adapters
- `values`: Abstract value types (AVT)
    - Specified by `ADT` types and/or interfaces
    - Used in `entity` or `aggregate` types

### Artifact abstractions

- `sources`: Human-readable source code artifacts
- `binaries`: Machine-readable binary artifacts (optional)
- `distros`: Distributable artifacts (images, archives, etc.)
- `stacks`: Deployable artifacts (ansible playbooks, helm charts, etc.)

### Execution abstractions

- `goroutine`: Framework level execution abstraction
- `process`: Application level execution abstraction
- `system`: Stack level execution abstraction

## Feature-sliced structure

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
- `pc.go`: Pure properties configuration (PC) logic
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
