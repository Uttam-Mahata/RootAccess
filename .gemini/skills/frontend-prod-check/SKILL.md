---
name: frontend-prod-check
description: Audit Angular frontend for production readiness. Checks for infinite loading states, empty data handling, state synchronization after CRUD operations, and performance best practices. Use when preparing for a release or investigating UI data refresh and loading issues.
---

# Frontend Production Readiness Check

This skill guides you through auditing an Angular application for production readiness, with a focus on reliability, performance, and user experience.

## Core Audit Areas

### 1. Reliable Loading States (Infinite Loading Prevention)
- **Error Handlers**: Ensure `isLoading` flags are reset to `false` in the `error` block of every `subscribe` or `tap`.
- **Completion**: Ensure `isLoading` is reset even if the stream completes without emitting (if applicable).
- **Timeouts**: For critical operations, implement a reasonable timeout to prevent permanent loading overlays.
- **Interceptors**: Check for global HTTP interceptors that might prevent a response from reaching the component.

### 2. Empty & Error States (Empty Data Handling)
- **Explicit Empty States**: Every list or data view must have an explicit `*ngIf="!isLoading && data.length === 0"` (or similar) section with a clear "No data" message.
- **Error feedback**: Use toast notifications or error banners instead of silent failure to let the user know why data isn't appearing.

### 3. State Synchronization (CRUD Operations)
- **Local Cache Update**: After a successful `POST`, `PUT`, or `DELETE`, ensure the local component state is updated immediately or a re-fetch is triggered.
- **Global Subjects**: Prefer using `BehaviorSubject` in services to provide data streams. Components should subscribe to these (`service.data$`) rather than storing local copies that can get out of sync.
- **Manual Refetch**: If a CRUD operation affects another part of the app, ensure a re-fetch is triggered for that part (e.g., updating user points after a challenge solve).

### 4. Memory & Performance
- **Subscriptions**: Use `takeUntilDestroyed()`, `take(1)`, or the `async` pipe to prevent memory leaks from dangling subscriptions.
- **Change Detection**: Favor `ChangeDetectionStrategy.OnPush` where possible for performance, but ensure state updates trigger a new check via `ChangeDetectorRef` or by updating Signals.
- **Signals**: If on Angular 16+, migrate simple UI state to Signals to simplify change detection and state tracking.

## Usage Guide

1. **Scan Components**: Identify components with `isLoading` or data-fetching logic.
2. **Audit Subscriptions**: Check `subscribe()` calls for missing error handling or state-reset logic.
3. **Check Templates**: Verify that `*ngIf="isLoading"` is correctly balanced with `*ngIf="!isLoading"` and "No data" sections.
4. **Test CRUD Flow**: Trace a CRUD operation from service to component to template to ensure the UI updates without a manual refresh.

For a detailed checklist, see [references/checklist.md](references/checklist.md).
