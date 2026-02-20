# Idiomatic Go & Production Patterns

## Error Handling
- **Wrapping**: Use `fmt.Errorf("context: %w", err)` to wrap errors.
- **Typed Errors**: Use `errors.Is(err, target)` or `errors.As(err, &target)` for checking specific error types.
- **Nil Checks**: Always check `err != nil` before proceeding.

## Concurrency
- **Context Propagation**: Pass `context.Context` through all layers (Handlers -> Services -> Repositories).
- **Graceful Shutdown**: Implement listeners for `SIGINT` and `SIGTERM` to close database connections and stop servers cleanly.
- **WaitGroups**: Use `sync.WaitGroup` to ensure background goroutines finish before the main process exits.

## Database (mongo-driver)
- **Projections**: Only fetch the fields you need (`bson.M{"field": 1}`).
- **Bulk Operations**: Use `BulkWrite` for high-volume inserts/updates.
- **Cursor Management**: Always `defer cursor.Close(ctx)` to prevent resource leaks.

## Project Structure
- **Internal vs. Cmd**: Keep core logic in `internal/` to prevent external packages from importing it. Use `cmd/` for entry points.
- **Interfaces**: Define interfaces at the consumer side to allow for easier mocking and testing.
