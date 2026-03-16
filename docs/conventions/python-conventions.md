# Python Conventions

> **When writing or reviewing any Python code, read this file first.**
>
> This consolidates Python-specific style, testing, architecture, and coding guidelines.
> These rules apply to all Python code in this repository.

---

## Table of Contents

- [General Coding Guidelines](#general-coding-guidelines)
- [Architecture & Layering Rules](#architecture--layering-rules)
- [Implementing a Feature (Standard Path)](#implementing-a-feature-standard-path)
- [Testing Expectations](#testing-expectations)
- [Code Quality Checklist](#code-quality-checklist)
- [Python Style Guidelines (PEP 8 & Beyond)](#python-style-guidelines-pep-8--beyond)

---

## General Coding Guidelines

- Write clean, readable, and well-structured code.
- Prefer explicit over implicit — avoid magic numbers, unclear abbreviations, and unexplained side effects.
- Follow the conventions already established in the codebase.
- Keep functions and modules small and focused on a single responsibility.
- Write meaningful commit messages using the [Conventional Commits](https://www.conventionalcommits.org/) format (e.g. `feat:`, `fix:`, `docs:`, `chore:`).

---

## Architecture & Layering Rules

### Allowed dependency direction

```
entry points  ➜  wiring/DI  ➜  domain modules
handlers      ➜  domain services (via abstract base classes / protocols)
integrations  ➜  external services
domain        ➜  shared ports/types + infrastructure abstractions
```

### Avoid / disallow

- Handlers calling data stores directly (unless explicitly part of a thin read-only endpoint).
- Domain A importing concrete packages from Domain B — use shared protocols/abstract classes instead.
- Shared/utility packages accumulating business logic (avoid "god modules").
- Introducing global state or module-level singletons without a documented reason.

---

## Implementing a Feature (Standard Path)

### Step 1 — Start from the user-facing surface

- Identify the target entry point: HTTP handler, background worker, CLI command, or a combination.
- If it changes an API contract, update the API spec (e.g. OpenAPI) first.
- Add or modify the handler/command: parse and validate input, call a domain service, return a standard response shape. Keep transport concerns out of the domain.

### Step 2 — Implement domain logic

- Add or modify code in the relevant domain module.
- Keep domain logic free of transport (HTTP, CLI) concerns.
- For cross-domain behaviour, define a `Protocol` or abstract base class and inject it via DI.

### Step 3 — Integrations

- Add or modify external client/adapters in the integrations layer.
- Domain code must depend on a `Protocol` or abstract interface, not the concrete integration.

### Step 4 — Wiring / DI

- Register new dependencies in the composition root.
- Keep wiring code outside domain packages.
- Prefer constructor injection; avoid module-level globals for service instances.

### Step 5 — Persistence & migrations

- If schema changes are required, add a migration with safe, forward-only changes.
- Ensure the migration can be applied in CI and locally before opening a PR.

---

## Testing Expectations

- **Domain logic**: unit tests alongside the domain module. Aim for high coverage of business rules.
- **Handler/command logic**: prefer parametrised tests for input validation and response mapping; use fakes or mocks for domain services via `unittest.mock` or `pytest-mock`.
- **Integration adapters**: unit-test adapters with mocked HTTP/SDK responses (`responses`, `httpretty`, or `unittest.mock`). Do not call real external services in automated tests.
- Avoid flaky tests; tests must be deterministic and independent of external state.
- **Parametrised tests**: use `@pytest.mark.parametrize` for repeated logic with multiple inputs:
  ```python
  @pytest.mark.parametrize("give,want,want_err", [
      ("valid@email.com", True, None),
      ("", False, ValueError),
  ])
  def test_validate_email(give, want, want_err):
      if want_err:
          with pytest.raises(want_err):
              validate_email(give)
      else:
          assert validate_email(give) == want
  ```
- **High coverage**: tests must cover all meaningful code paths. Focus on high-risk areas and edge cases (boundary values, error paths, `None`/empty inputs) rather than chasing 100 % line coverage. A well-chosen set of targeted tests is more valuable than exhaustive but shallow ones.
- **Readability and maintainability**: tests must be clean and structured. Use descriptive test names (`test_order_service__create_order__returns_not_found`), keep each test case focused on a single behaviour, and avoid complex conditional logic or branching inside the test body — split into separate functions instead.
- **Fixtures**: use `pytest` fixtures for shared setup/teardown. Keep fixtures scoped as narrowly as possible (`function` scope by default; widen only when performance justifies it).

---

## Code Quality Checklist

Before requesting a review, verify:

- [ ] No circular imports between modules.
- [ ] Shared/utility packages contain only protocols, types, and utilities — no domain behaviour.
- [ ] Handlers do not embed business rules.
- [ ] Integration clients are isolated behind `Protocol` or abstract base classes.
- [ ] Configuration is read from a single, explicit source (no scattered `os.environ` reads).
- [ ] All public functions, classes, and modules have docstrings.
- [ ] Type annotations are present on all function signatures.
- [ ] Tests are added or updated for any new or modified logic.
- [ ] A migration is included if the schema changed, and it is backwards-compatible if needed.

---

## Python Style Guidelines (PEP 8 & Beyond)

> Follow [PEP 8](https://peps.python.org/pep-0008/) as the baseline.
> The rules below refine or extend it for this codebase.

### Tooling & Linting

- Format all code with [Ruff](https://docs.astral.sh/ruff/) (formatter + linter). Run on save and in CI.
- Use [mypy](https://mypy.readthedocs.io/) for static type checking. All code must pass `mypy --strict` (or the project's configured strictness level).
- Use [pre-commit](https://pre-commit.com/) hooks to enforce formatting and linting before every commit.
- Required Ruff rules (at minimum): `E`, `F`, `I` (isort), `UP` (pyupgrade), `B` (bugbear), `C4`, `SIM`.

### Type Annotations

- Annotate all function parameters and return types:
  ```python
  def get_user(user_id: int) -> User:
      ...
  ```
- Use `from __future__ import annotations` at the top of every module to enable postponed evaluation and allow forward references.
- Prefer `X | Y` union syntax (Python 3.10+) over `Optional[X]` or `Union[X, Y]`.
- Use `typing.Protocol` to define structural interfaces instead of abstract base classes where duck typing is sufficient.
- Never use `Any` unless interfacing with untyped third-party code; document the reason with an inline comment.

### Naming Conventions

| Element | Convention | Example |
|---|---|---|
| Module | `snake_case` | `user_service.py` |
| Package | `snake_case`, short, no underscores preferred | `auth/` |
| Class | `PascalCase` | `UserService` |
| Function / method | `snake_case` | `get_user_by_id` |
| Variable | `snake_case` | `user_count` |
| Constant (module-level) | `UPPER_SNAKE_CASE` | `MAX_RETRIES = 3` |
| Private attribute / method | `_single_leading_underscore` | `_cache` |
| "Dunder" / magic method | `__double_underscore__` | `__init__`, `__str__` |
| Type variable | `PascalCase` with `T` suffix | `UserT = TypeVar("UserT")` |
| Protocol | `PascalCase` | `UserRepository` |

### Imports

- Group imports in this order, separated by a blank line (enforced by Ruff/isort):
  1. Standard library
  2. Third-party packages
  3. Local application imports
- Use absolute imports for application code; relative imports only within the same package and only one level deep.
- Never use wildcard imports (`from module import *`).
- Import only what you need; avoid importing an entire module just to use one attribute.

### Functions & Methods

- **Single responsibility**: each function or method does one thing.
- **Short functions**: aim for functions that fit in a single screen; extract helpers rather than nesting logic.
- **Return early**: guard with early returns to reduce nesting rather than deeply nested `if/else` blocks.
- **No mutable default arguments**: use `None` as the default and initialise inside the function:
  ```python
  # Bad
  def append_item(item: str, items: list[str] = []) -> list[str]:
      ...

  # Good
  def append_item(item: str, items: list[str] | None = None) -> list[str]:
      if items is None:
          items = []
      ...
  ```
- **Keyword-only arguments**: use `*` to force keyword-only arguments for functions with more than two or three optional parameters to improve call-site clarity:
  ```python
  def create_user(*, name: str, email: str, is_active: bool = True) -> User:
      ...
  ```

### Classes

- Prefer `dataclasses.dataclass` or `pydantic.BaseModel` for plain data containers; avoid writing boilerplate `__init__`, `__repr__`, and `__eq__` by hand.
- Use `__slots__` for performance-sensitive, frequently instantiated classes.
- Keep class hierarchies shallow; favour composition over inheritance.
- Do not expose internal helpers as public methods — prefix with `_`.

### Errors & Exceptions

- Define custom exception classes in a dedicated `exceptions.py` module per domain package. Inherit from a base domain exception:
  ```python
  class DomainError(Exception):
      """Base class for all domain errors."""

  class UserNotFoundError(DomainError):
      """Raised when a user cannot be found."""
  ```
- **Catch specific exceptions**: never use bare `except:` or `except Exception:` unless re-raising.
- **Context chaining**: use `raise NewError(...) from original_error` to preserve the traceback chain.
- **Avoid swallowing errors**: always log or re-raise; never silently pass.
- Do not use exceptions for normal control flow (e.g. to signal an empty result — return `None` or an empty collection instead).

### Concurrency & Async

- Use `async`/`await` (asyncio) for I/O-bound work. Do not mix sync blocking I/O inside an async function.
- Use `asyncio.gather` or task groups (`asyncio.TaskGroup`, Python 3.11+) to run independent coroutines concurrently.
- Use `anyio` or `asyncio.Lock` / `asyncio.Semaphore` to protect shared mutable state in async code.
- For CPU-bound work, use `concurrent.futures.ProcessPoolExecutor` and avoid blocking the event loop.
- Every background task spawned with `asyncio.create_task` must be awaited or cancelled during shutdown; store task references to prevent garbage collection.

### Resource Management

- Use context managers (`with` / `async with`) to manage all resources: files, database connections, HTTP sessions, locks.
- Define `__enter__`/`__exit__` (or `__aenter__`/`__aexit__`) for any class that manages a resource.
- Prefer `contextlib.contextmanager` or `contextlib.asynccontextmanager` for lightweight context managers.

### Constants & Configuration

- Module-level constants use `UPPER_SNAKE_CASE`.
- Application configuration is read once at startup via a dedicated settings module (e.g. using `pydantic-settings`). Never scatter `os.environ` reads across the codebase.
- Avoid hardcoding environment-specific values (hostnames, ports, credentials) anywhere outside the config module.

### Logging

- Use the standard `logging` module; never use `print` for diagnostic output in production code.
- Obtain a module-level logger: `logger = logging.getLogger(__name__)`.
- Use structured logging (e.g. `structlog`) when the project requires it — follow `logging-conventions.md` for details.
- Never log sensitive data (passwords, tokens, PII).

### Performance (Hot Path Only)

- Prefer list/dict/set comprehensions over `map`/`filter` with `lambda` for clarity.
- Use generators (`yield`) for large sequences to avoid loading everything into memory.
- Use `collections.defaultdict`, `collections.Counter`, and `functools.lru_cache` / `functools.cache` where appropriate.
- Profile before optimising; do not sacrifice readability for micro-optimisations.

### Style

- **Line length**: 99 characters maximum (configure in `pyproject.toml` / `ruff.toml`).
- **String quotes**: use double quotes `"` consistently (Ruff default).
- **Trailing commas**: use trailing commas in multi-line collections and function signatures to produce cleaner diffs.
- **f-strings**: prefer f-strings over `%`-formatting or `.format()` for all new code.
- **Avoid bare `pass`**: replace with an explanatory docstring or comment inside the body (e.g. in an intentionally empty `except` block).
- **Boolean comparisons**: never compare to `True`/`False` with `==`; use truthiness directly (`if flag:` not `if flag == True:`).
- **`None` checks**: always use `is None` / `is not None`, never `== None`.
- **`__all__`**: define `__all__` in every public module to explicitly declare its public API.
