# Orglang

This is a programming language to support organization development and management.

## Design Principles

### Explicit Abstraction Elaboration

- Domain-Driven Design

### Core and Periphery Separation

- Functional Core and Imperative Shell
- Hexagonal/Onion/Clean Architecture

### Data and Behavior Separation

- Functional Programming

### Vertical and Horizontal Slicing

- Vertical Slice Architecture
- Layered Architecture

## Conceptualization

Any software, in essence, is a pile of abstractions.

### Kinds

Repository structure reflects abstraction kinds.

- `app`: Runnable program abstractions
  - `web`: Web application program
- `adt`: Reusable data abstractions
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
- `lib`: Reusable behavior abstractions
  - `ck`: Configuration keeper harness
  - `lf`: Logging framework harness
  - `sd`: Storage driver harness
  - `te`: Template engine harness
  - `ws`: Web server harness
- `db`: Storage schema definitions
  - `postgres`: PostgreSQL specific definitions
- `orch`: Orchestration harness definitions
  - `task`: Task (build tool) harness definitions
- `proto`: Prototype endeavors
- `stack`: System level definitions
- `test`: End-to-end tests and harness

### Layers

Package structure reflects abstraction layers.

#### Toolkit agnostic

- `core.go`: Pure domain logic
    - Domain models (core models)
    - API interfaces (primary ports)
    - Service structs (core behaviors)
- `me.go`: Pure message exchange (ME) logic
    - Exchange specific DTO's (edge models)
- `vp.go`: Pure view presentation (VP) logic
    - Presentation specific DTO's (edge models)
- `ds.go`: Pure data storage (DS) logic
    - Storage specific DTO's (edge models)
    - Repository interfaces (secondary ports)
- `iv.go`: Pure input validation (IV) logic
    - Message validation harness
    - Props validation harness
- `pc.go`: Pure property configuration (PC) logic
    - Configuration specific DTO's (edge models)
- `tc.go`: Pure type conversion (TC) logic
    - Domain to domain conversions
    - Domain to message conversions and vice versa
    - Domain to data conversions and vice versa

#### Toolkit specific

- `di_fx.go`: Fx (dependency injection library) specific component definitions
- `me_echo.go`: Echo (web framework) specific controller definitions (primary adapters)
- `vp_echo.go`: Echo (web framework) specific presenter definitions (primary adapters)
- `me_resty.go`: Resty (HTTP library) specific client definitions (secondary adapters for external use)
- `ds_pgx.go`: pgx (PostgreSQL driver and toolkit) specific DAO definitions (secondary adapters for internal use)
- `iv_ozzo.go`: Ozzo (validation library) specific validation definitions
- `tc_goverter.go`: Goverter (type conversion tool) specific conversion definitions
- `vp/bs5/*.html`: Go's built-in `html/template` and Bootstrap 5 (frontend toolkit) specific presentation definitions

### Aspects

Code structure reflects abstraction aspects. 

#### Scale

- `aggregate`: Concurrency-aware abstraction
    - Consumed by controller adapters
    - Specified by `api` interfaces
    - Implemented by `service` structs
- `entity`: Identity-aware abstraction
    - Consumed by `service` structs
    - Specified by `repo` interfaces
    - Implemented by DAO adapters
- `value`: Classical data abstraction
    - Consumed by `entity` or `aggregate` abstractions
    - Specified by `ADT` types and/or interfaces
    - Implemented by concrete types and/or structs

#### Lifecycle

- `dec`: Abstraction declaration phase
- `def`: Abstraction definition phase
- `exp`: Abstraction expression phase
- `exec`: Abstraction execution phase

#### Slice

- `ref`: Machine-readable pointer to an abstraction
- `spec`: Specification to create an abstraction
- `rec`: Record for abstraction retrieval (excluding sub abstractions)
- `mod`: Modification to change an abstraction (including sub abstractions)
- `snap`: Snapshot for abstraction retrieval (including sub abstractions)

#### Artifact

- `sources`: Human-readable code of abstraction
- `binaries`: Machine-readable code of abstraction
- `distros`: Distribution-friendly binaries (images, archives, etc.)
- `stacks`: Deployment-friendly definitions (ansible playbooks, helm charts, etc.)

## Development

### Workflow

- `task sources` - before commit to task branch
- `task binaries` - before push to task branch
- `task distros` - before push or merge to feature branch
- `task stacks` - before merge to main branch
