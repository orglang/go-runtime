# Orglang

This is a programming language to support organization development and management.

## Design

Any software, in essence, is a pile of abstractions.

### Principles

Project architecture reflects following design principles.

#### Explicit Abstraction Elaboration

- Domain-Driven Design

#### Core and Periphery Separation

- Functional Core and Imperative Shell
- Hexagonal/Onion Architecture

#### Data and Behavior Separation

- Functional Programming

#### Vertical then Horizontal Decomposition

- Vertical Slice Architecture
- Layered Architecture

### Kinds

Repository structure reflects abstraction kinds.

- `app`: Runnable program abstraction
  - `web`: Web application program
- `adt`: Reusable data abstraction
  - `identity`: Identification value type
  - `polarity`: Polarization value type
  - `pooldec`: Pool declaration aggregate type
  - `poolexec`: Pool execution aggregate type
  - `poolexp`: Pool expression value type
  - `procdec`: Process declaration aggregate type
  - `procdef`: Process definition entity type
  - `procexec`: Process execution aggregate type
  - `procexp`: Process expression value type
  - `procstep`: Process step value type
  - `revnum`: Revision number value type
  - `symbol`: Symbol value type
  - `syndec`: Synonym declaration value type
  - `procbind`: Term context value type
  - `typedef`: Type definition aggregate type
  - `typeexp`: Type expression value type
  - `uniqref`: Unique refererence value type
  - `uniqsym`: Unique symbol value type
- `lib`: Reusable behavior abstraction
  - `db`: Database drivers
  - `kv`: Key-value store drivers
  - `lf`: Logging framework harness
  - `te`: Template engine harness
  - `ws`: Web server harness
- `db`: Storage schema definition
  - `postgres`: PostgreSQL schema
- `orch`: Orchestration harness definition
  - `task`: Task (build tool) harness definition
- `proto`: Prototype endeavors
- `stack`: System level definition
- `test`: Test level harness
  - `e2e`: End-to-end tests

### Layers

Package structure reflects abstraction layers.

#### Toolkit agnostic

- `core.go`: Pure domain logic
    - Domain models (core models)
    - API interfaces (primary ports)
    - Service structs (core behaviors)
- `me.go`: Pure message exchange (ME) logic
    - Message related DTO's (edge models)
- `vp.go`: Pure view presentation (VP) logic
    - View related DTO's (edge models)
- `ds.go`: Pure data storage (DS) logic
    - Data related DTO's (edge models)
    - Repository interfaces (secondary ports)
- `iv.go`: Pure input validation (IV) logic
    - Message related validation
    - Config related validation
- `cs.go`: Pure config storage (CS) logic
    - Config related DTO's (edge models)
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
    - Specified by `API` interfaces
    - Implemented by `service` structs
- `entity`: Identity-aware abstraction
    - Consumed by `service` structs
    - Specified by `Repo` interfaces
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
