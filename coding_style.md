# Go Development Standards

## 4.1. Formatting

All Go source code in this project MUST be formatted using the standard `gofmt` tool. This is a non-negotiable standard in the Go community that ensures consistency and readability across the entire codebase. Code should be formatted before every commit. Most Go-compatible editors can be configured to run `gofmt` on save.

## 4.2. Linting

To maintain code quality and catch common programming errors, this project SHOULD use a linter. `golangci-lint` is the recommended tool. It aggregates many different linters and provides a single, configurable interface. A default configuration should be used to check for issues such as:

- Unused variables or imports (go vet)
- Code complexity
- Improper error handling
- Inefficient code constructs

## 4.3. Naming Conventions

All identifiers MUST follow standard Go naming conventions:

- **Packages**: Package names should be short, concise, and all lowercase (e.g., handlers, tools).
- **Variables, Functions, and Methods**: Identifiers intended for internal use within a package should use camelCase. Identifiers intended to be exported for use by other packages must use PascalCase.
- **Interfaces**: Interfaces should be named with the -er suffix if they contain a single method (e.g., Reader, Writer). For interfaces with multiple methods, choose a name that represents its purpose (e.g., MCPService).

## 4.4. Error Handling

Error handling is a critical aspect of robust Go applications.

- Errors MUST NOT be ignored or discarded. The blank identifier (`_`) should not be used to discard an error unless there is an explicit and documented reason.
- Errors returned from function calls MUST be checked.
- When returning an error from a function, it should be the last return value.
- When an error is propagated up the call stack, it SHOULD be wrapped with additional context using `fmt.Errorf` with the `%w` verb. This preserves the original error while adding contextual information, which is invaluable for debugging.

Example:

```go
// Bad:
// return err

// Good:
return fmt.Errorf("failed to process records for user %s: %w", userID, err)
```

## 4.5. Concurrency

The current implementation of this connector is synchronous and does not require concurrency primitives. However, if the project is extended with features that require concurrent operations, the following rules MUST be applied:

- Race conditions must be prevented. Access to shared memory MUST be synchronized using either channels (the preferred Go-idiomatic approach) or mutexes from the sync package.
- The `go vet -race` command should be used during testing to detect potential race conditions.

## 4.6. Package Structure

The project is organized into packages based on their functional responsibility. This separation of concerns makes the code easier to navigate, test, and maintain.

- **main**: The main package, containing the application entrypoint.
- **handlers**: Contains Gin handlers that bridge HTTP requests to the application logic.
- **middleware**: Contains Gin middleware for concerns like authentication and logging.
- **tools**: Contains the core business logic for the tools exposed by the MCP server.

New functionality should be placed in the appropriate existing package or in a new package if it represents a distinct domain of responsibility.