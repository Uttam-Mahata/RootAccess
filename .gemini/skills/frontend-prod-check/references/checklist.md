# Frontend Production Checklist

Use this checklist during an audit to systematically identify common UI/UX and performance issues.

## 1. Reliability (Loading & Error Handling)

- [ ] Every `isLoading = true` has a corresponding `this.isLoading = false` in:
    - [ ] `next` handler
    - [ ] `error` handler
    - [ ] `complete` handler (if it can finish without emitting)
- [ ] Form submission buttons are disabled while `isLoading` is true.
- [ ] API error responses are shown to the user (e.g., via a toast or error property in the component).
- [ ] Critical background tasks (e.g., refreshing a token) do not block the UI unnecessarily.

## 2. Empty States & Visibility

- [ ] Every `*ngFor` has a fallback `*ngIf="data.length === 0"` with a descriptive message.
- [ ] Search/filter UI shows "No results matching '[query]'" when results are empty.
- [ ] "No results" messages are not visible *while* the data is still loading.
- [ ] Skeleton screens or spinners are appropriately sized and centered.

## 3. Data Flow & CRUD Consistency

- [ ] After `POST` (Create): New item is added to the local list or the entire list is re-fetched.
- [ ] After `PUT`/`PATCH` (Update): The updated object is updated in the local list or the entire list is re-fetched.
- [ ] After `DELETE`: The item is removed from the local list or the entire list is re-fetched.
- [ ] Multiple components showing the same data (e.g., user points in header vs. profile) use a shared service stream (e.g., `BehaviorSubject`) to stay synchronized.

## 4. Angular Best Practices (v18+)

- [ ] Subscriptions are properly managed via `takeUntilDestroyed()`, `async` pipe, or `take(1)`.
- [ ] Reactive Forms are used instead of Template-driven Forms for complex logic.
- [ ] `ChangeDetectionStrategy.OnPush` is considered for performance-critical lists.
- [ ] Modern control flow (`@if`, `@for`, `@switch`) is used instead of `*ngIf`, `*ngFor` where appropriate.
- [ ] Signals are used for local component state that affects rendering.

## 5. Performance

- [ ] Images have appropriate `alt` tags and are correctly sized.
- [ ] Large lists use `virtualScroll` or pagination if they exceed ~100 items.
- [ ] Unused dependencies (e.g., large libraries for small tasks) are avoided.
- [ ] Bundle size is optimized by lazy-loading routes.
