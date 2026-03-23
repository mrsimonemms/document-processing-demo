# Claude Instructions

This repository contains the **Document Processing Demo**, a canonical demo
showcasing how Temporal makes AI-powered document workflows more reliable,
resumable and observable.

The demo is designed to show that AI systems are useful but operationally
fragile without durable orchestration. The value of the project is not only in
document summarisation or question answering. It is in demonstrating retries,
failover, durable execution, progress tracking and recovery from failure.

Claude should treat this repository as a **production-shaped demo system**, not
as a toy app and not as a generic chat interface.

---

## Repository structure

This repository is intentionally split into top-level components.

Current implementation targets:

- `/golang` → Go backend and Temporal workflows
- `/ui` → SvelteKit frontend

All current Go backend code MUST live under `/golang`.
All current frontend application code MUST live under `/ui`.

Do not place application code at the repository root.

Expected current backend structure:

- golang/main.go
- golang/internal/api/handler.go
- golang/internal/activities/activities.go
- golang/internal/providers/
- golang/internal/workflows/document.go
- golang/internal/models/models.go
- golang/go.mod
- golang/go.sum

The `golang/` directory is a self-contained Go module rooted at
`github.com/mrsimonemms/document-processing-demo/golang`. All application
logic must live within it.

The repository root is for cross-repo files such as:

- CLAUDE.md
- README.md
- LICENSE
- CODE_OF_CONDUCT.md
- .pre-commit-config.yaml
- other repository-wide config files

The root also contains a minimal `go.mod` and an empty `main.go`. These are
infrastructure files and must not contain application logic. Do not add
imports, types or functions to the root `main.go`. Do not confuse the root
module with the backend module in `golang/`.

### Future language support

Today, the backend implementation is in Go under `/golang`.

Do not assume Go is the only language this repository will ever contain.
Design the current implementation cleanly within `/golang`, without making the
repository structure depend on Go being the only possible backend forever.

Do not scaffold additional language implementations unless explicitly asked.

---

## Project intent

The Document Processing Demo exists to demonstrate:

- Upload and ingestion of documents
- Long-running document processing workflows
- Document summarisation
- Question answering grounded in document content
- Recovery from network, provider and worker failures
- Explicit visibility into what Temporal is doing

The core story is:

- AI capabilities are valuable
- AI dependencies are unreliable
- Temporal makes the system dependable

This story should remain visible in code structure, naming, UI behaviour and
documentation.

---

## High-level architecture

The expected current architecture is:

- **Backend language:** Go
- **Workflow engine:** Temporal
- **Frontend:** SvelteKit
- **Document processing:** orchestrated through Temporal workflows
- **AI providers:** external services accessed through activities
- **Failure injection:** explicit demo controls, not hidden test-only behaviour

The backend should own orchestration, workflow lifecycle, failure handling and
provider abstraction.

The frontend should surface:

- document upload
- ingestion status
- summary view
- question and answer flow
- scenario selection and failure injection
- visible execution progress and recovery behaviour

Do not collapse the demo into a single opaque chatbot experience. The
orchestration story must remain visible.

---

## Temporal client setup

Use the existing Temporal connection helper pattern provided by the repository
author rather than introducing a new custom bootstrap layer.

The helper is available from:

<https://github.com/mrsimonemms/golang-helpers>

Guidelines:

- Prefer this helper for creating Temporal clients
- Do not reimplement client initialisation from scratch
- Do not introduce unnecessary abstraction layers around the Temporal client
- Keep setup simple and explicit
- Ensure it works for both local development and Temporal Cloud

The connection layer is considered infrastructure and should remain boring and
stable. Do not optimise or refactor it unless explicitly required.

---

## Workflow implementation approach

All workflows in this repository are currently implemented directly in Go using
the Temporal Go SDK.

Do not introduce:

- DSLs
- code generation layers
- declarative workflow specifications
- intermediate representations

Workflows should be:

- explicit Go functions
- easy to read top-to-bottom
- structured as a sequence of well-named steps
- composed using activities for side effects

The goal is for a developer reading the workflow code to immediately understand:

- what steps are executed
- where failures can occur
- how retries and failover are handled
- how progress is preserved

Prefer straightforward control flow over abstraction-heavy designs.

---

## Core product principles

Prioritise:

- Reliability over novelty
- Clarity over cleverness
- Determinism over convenience
- Explicit behaviour over hidden magic
- Demo value over feature sprawl

This project should feel intentionally scoped and easy to explain live.

Avoid:

- unnecessary agent frameworks
- abstract orchestration layers that hide Temporal concepts
- speculative features that weaken the main narrative
- over-engineering for hypothetical future requirements

---

## Demo philosophy

This is a demo, but it should reflect real production concerns.

The most important behaviours to demonstrate are:

- retries
- backoff
- failover between AI providers
- resumability after interruption
- partial progress preservation
- visible workflow state transitions
- durable execution across long-running tasks

The demo should be legible to an audience. A user should be able to understand
what failed, what Temporal did about it and why the system still completed.

Where there is a trade-off between "more AI capability" and "clearer Temporal
value", choose clearer Temporal value.

---

## Failure injection model

Failure injection is a first-class feature of this repository.

Failure scenarios are selected externally, for example via the UI, and passed
explicitly into workflows as part of their input.

Activities and workflow logic must read this scenario and simulate failures in a
deterministic and controlled way.

Supported demo scenarios may include:

- network timeout
- primary AI provider failure
- rate limiting
- malformed provider response
- temporary downstream service outage
- worker restart or interruption

Failure injection must be:

- explicit
- deterministic
- reproducible
- visible in execution logs or status

Prefer:

- scenario flags or structured scenario types
- named failure modes
- deterministic failure triggers such as "fail first attempt"

Avoid:

- hidden randomness
- non-deterministic failure behaviour
- global mutable state controlling failures
- behaviour that changes between runs without an explicit scenario

The selected scenario must be visible in workflow inputs and explainable during
a demo.

---

## Temporal-specific constraints

Temporal workflow code must remain deterministic.

Do not:

- use wall-clock time directly inside workflows
- introduce randomness into workflow logic
- perform network calls from workflows
- access mutable global state from workflows
- hide side effects inside workflow code

Do:

- keep side effects in activities
- make activity inputs and outputs serializable
- use explicit retry policies
- model failover decisions clearly
- preserve workflow history readability

If a workflow becomes long-running or large, prefer explicit workflow design
choices such as chunking, signals, queries or Continue-As-New where appropriate.

Temporal should remain visible as the orchestration layer, not buried behind
generic helper abstractions.

---

## AI integration rules

AI providers are dependencies, not trusted control planes.

Claude should assume:

- provider calls can fail
- provider responses may be slow
- provider responses may be malformed
- providers may need fallback behaviour
- summarisation and Q&A quality should be "good enough" for the demo
- operational resilience matters more than model sophistication

Do not tightly couple the system to a single provider's SDK or response shape.

Prefer:

- provider abstractions with explicit request and response contracts
- validation of provider outputs
- fallback from a primary provider to a secondary provider
- clear logging and surfaced status for provider failures

Do not build the architecture around speculative autonomous multi-agent systems.
If "agentic" behaviour is implemented, it must still map cleanly to explicit,
observable workflow steps.

---

## Frontend rules

The frontend is part of the demo story.

It should make the following obvious:

- what document is being processed
- what stage processing is in
- whether the system is healthy or recovering
- which scenario is active
- what happened when something failed
- how the final result was still produced

Prefer a UI with:

- a clear document upload flow
- a visible progress and status area
- scenario controls for failure injection
- a summary panel
- a question and answer panel
- a timeline or event log showing recovery steps

Do not reduce the UI to just a chat box.

Do not hide important system state behind developer tools or logs only.

---

## API and backend design guidance

Prefer small, explicit services and handlers.

Suggested responsibilities:

- document upload and registration
- workflow start and tracking
- summary retrieval
- question submission
- answer retrieval
- scenario selection and failure injection configuration
- status and event reporting

Keep request and response shapes stable and easy to inspect.

Avoid premature microservice decomposition. This is a demo and should remain easy
to run locally and easy to explain.

---

## Code style and structure

Prefer:

- explicit types and structs
- focused packages
- readable naming
- small composable functions
- straightforward control flow
- standard library solutions where reasonable

Avoid:

- clever abstractions
- deep inheritance-style layering
- unnecessary interfaces
- broad utility packages with unclear ownership

Go-specific guidance:

- be explicit with error handling
- keep workflow and activity boundaries clear
- separate Temporal concerns from transport concerns
- do not hide core orchestration logic in generic helpers

Frontend guidance:

- keep state management simple
- prefer readable Svelte components over abstraction-heavy patterns
- build UI components around demo flows, not generic component systems

---

## Testing and validation

The repo should be easy to validate locally.

Prioritise tests that protect the demo story:

- workflow retry behaviour
- provider failover behaviour
- failure injection behaviour
- API contract stability
- frontend display of workflow state where practical

At a minimum, changes should preserve:

- deterministic workflow behaviour
- clear scenario handling
- stable demo flows
- readable logs and surfaced status

Do not add brittle tests that depend on real external AI providers.

Prefer mocks, fakes or controlled provider adapters for test coverage.

---

## Required local validation

Before considering any change complete, run the appropriate validation commands
for the parts of the repository you changed.

At a minimum:

- run `pre-commit run --all-files`
- ensure all checks pass
- fix any issues rather than reporting them without action

If you change Go code under `/golang`, also ensure the Go project builds and
tests cleanly from within that directory.

If you change frontend code under `/ui`, also ensure the frontend checks, build
and tests, if present, pass from within that directory.

Do not claim work is complete if `pre-commit` or the relevant project checks are
failing.

If a required tool is missing, state exactly what is missing and what command
failed.

---

## Documentation expectations

Documentation is part of the demo.

Docs should explain:

- what the demo does
- what scenario it is proving
- how to run it locally
- how to trigger failure scenarios
- what Temporal concepts are being demonstrated
- what is intentionally simplified

Examples and docs should reflect actual supported behaviour.

Do not write aspirational documentation for features that do not exist.

If behaviour changes intentionally, update docs. If behaviour changes
unintentionally, fix the implementation rather than changing docs to match a
regression.

---

## Writing style rules

- Use British English spelling and punctuation.
- Do not use em dashes.
- Prefer clear, direct technical writing.
- Avoid marketing fluff.
- Avoid exaggerated claims about AI capabilities.
- Keep the tone grounded, practical and inspectable.

---

## Creative scope and delivery expectations

Unless explicitly instructed otherwise, assume:

- the shortest correct implementation is preferred
- the main story is reliability under failure
- architecture should remain easy to explain in a demo
- backwards compatibility inside the repo matters when modifying existing code
- visible behaviour is more important than internal cleverness

When implementing a feature:

1. start with the smallest complete version
2. preserve the demo narrative
3. keep Temporal concepts visible
4. avoid expanding scope without a good reason

Do not invent major new product features without discussion.

---

## When unsure

If something is unclear, prefer:

- preserving determinism
- preserving demo clarity
- making failure behaviour explicit
- asking before expanding scope significantly

Correctness beats convenience.
Reliability beats novelty.
Clarity beats magic.
