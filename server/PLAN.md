# Service Farrowing Schedule Plan

## Goal
Prepare a read-only service schedule that identifies confirmed services due to farrow today through the next 7 days, plus confirmed services that are overdue. Include each sow's average gestation duration from completed historical farrowings.

This phase covers only the service **domain** and **infrastructure** packages. The application use case, HTTP endpoint, authorization, and client work are deferred.

## Current state
- `service/domain` models a service and its state. A new confirmed service's expected farrowing date is derived as the last mount date plus 114 days.
- `service/infrastructure` contains the PostgreSQL repository and knows the `services`, `mounts`, `farrowings`, and `sows` tables.
- `services.expected_farrowing_date` is a `DATE`; the service table currently has an index on `sow_id`, but not on state/date.
- `farrowings` stores `service_id`, `sow_id`, and the actual `farrow_date`. The current schema has indexes on `sow_id` and `farrow_date`, but not `service_id`.
- The existing service list is paginated and ordered by creation time, so the schedule should be a separate query/read model rather than changing its behavior.

## Domain work
1. Add domain read-model types for a schedule result with separate upcoming and overdue entries. Each entry should expose the service and sow identifiers/code, expected farrowing date, and nullable average gestation duration in days.
2. Keep this as a read-only projection; do not change service creation, state transitions, or the 114-day expected-date rule.
3. Define and document the date classification: upcoming includes expected dates from today through today + 7 days, inclusive; overdue includes dates before today. A service due today belongs only to upcoming.
4. Represent an unavailable historical average as `nil` when the sow has no eligible completed farrowing history.

## Infrastructure work
1. Add a repository query method for the schedule and implement it in `PostgresServiceRepository`.
2. Use one SQL query to filter `Confirmado` services, join the sow code, classify upcoming and overdue entries, and calculate historical averages in the database. Do not load complete service/farrowing datasets into Go for filtering or averaging.
3. Calculate each historical gestation duration as actual `farrow_date` minus the last mount date for that farrowing's service. Average eligible completed cycles per sow; exclude the current confirmed service and cycles without a recorded farrowing.
4. Order each group by expected farrowing date ascending, then sow code ascending. Return empty slices rather than `nil` slices when a group has no entries.
5. Pass the reference date into the query from the caller instead of relying on the database session's `CURRENT_DATE`, so the configured application clock/time zone can define “today” consistently.
6. Review the query plan with representative data. Add database indexes only if `EXPLAIN ANALYZE` shows they are needed; candidates include `(state, expected_farrowing_date)` on `services` and `service_id` on `farrowings`.

## Deferred beyond the domain and infrastructure phase
- Client pages and rendering.
- Any materialized view, background aggregation, or stored average.

## Acceptance criteria for this phase
- Domain read-model types clearly represent upcoming and overdue schedule entries, including a nullable average.
- The PostgreSQL repository can return both groups using one database query and accepts an explicit reference date.
- Only confirmed services with an expected farrowing date are considered.
- Upcoming is inclusive of today and the following seven days; overdue is strictly before today; groups do not overlap.
- Historical averages use completed recorded farrowings and the last mount date for each service.
- Ordering is deterministic within both groups.
- Query and repository tests cover the date boundaries, empty groups, and missing historical averages.

# Application and Server Wiring Plan

## Goal
Expose the farrowing schedule through the server by adding the application use case, wiring it through the service module and application container, and registering an authenticated HTTP endpoint. This phase builds on the domain and infrastructure read model/repository query described above.

## Application work
1. Add a `GetFarrowingScheduleService` use case in `service/application` that depends on the service repository and the injected server clock.
2. In `Execute`, derive the reference calendar date from the injected clock and pass it to the repository. Normalize the date consistently with the server's configured time zone; do not use the database server's local date.
3. Return the domain schedule read model, keeping date grouping and historical gestation calculations in the domain/infrastructure responsibilities already specified.
4. Keep this use case read-only and do not add authorization decisions to the application layer.

## Service module and container wiring
1. Expose the new use case on `service/infrastructure.Module` alongside the existing service use cases.
2. Construct it with the same service repository and clock already provided to `serviceinfrastructure.NewModule`.
3. Pass the use case into `ServiceHandler` and register the schedule route with the existing service routes.
4. Confirm the root server container already constructs the service module with the shared PostgreSQL repository and clock; update container wiring only if the new dependency requires it.

## HTTP endpoint
1. Add an authenticated read endpoint, `GET /api/v1/services/farrowing-schedule`, without admin-only middleware so regular authenticated users can use the report.
2. Return a stable JSON object with `upcoming` and `overdue` arrays. Each entry should include service ID, sow ID and code, expected farrowing date, and nullable `average_gestation_days`.
3. Serialize dates as `YYYY-MM-DD` and return empty arrays rather than `null` when either group has no entries.
4. Map repository failures through the existing service error handling; do not expose database errors in the response.

## Verification
1. Add application tests for forwarding the clock-derived reference date, propagating repository errors, and returning the schedule result.
2. Add handler tests for the authenticated route, response JSON/date formatting, empty arrays, and server error mapping.
3. Add module/container wiring coverage where existing patterns require it, and confirm the route is reachable through the assembled server.
4. Run the focused service application and infrastructure tests, then the server test suite.

## Acceptance criteria
- An authenticated `User` or `Admin` can request `GET /api/v1/services/farrowing-schedule`.
- The use case supplies a reference date derived from the shared server clock.
- The response contains `upcoming` and `overdue` arrays with the agreed fields and date format.
- Existing service endpoints and their authorization behavior remain unchanged.
- The endpoint, application service, module, and container are wired through the normal server construction path.
- Focused tests and the server test suite pass.



# Client User Home Schedule Plan

## Goal
Update the User Home schedule to show both farrowing estimates returned by the server and add a sow-code filter over the schedule already loaded on the page.

## Current state
- `client/src/app/features/user-home` already requests the farrowing schedule once when the page loads.
- The page already renders upcoming and overdue lists and has loading, error, retry, and empty states.
- `ServicesService.getFarrowingSchedule()` and the `FarrowingSchedule` client model already exist.
- The client schedule entry currently models and displays `expected_farrowing_date` and `average_gestation_days`; it does not yet model `average_expected_farrowing_date`.
- The existing sow and service lists trim sow-code input and apply search through their current list/filter flows. Home's new filter should apply to the schedule response already loaded, without additional requests.

## Step-by-step client plan
1. **Update the response model.** Add nullable `average_expected_farrowing_date` to `FarrowingScheduleEntry`, matching the server's JSON field and `YYYY-MM-DD` format.
2. **Display both dates.** Update each upcoming and overdue card to show the existing 114-day `expected_farrowing_date` and the new average-based expected date as separately labeled values. Keep the average gestation duration visible as context.
3. **Handle missing history.** When `average_expected_farrowing_date` or `average_gestation_days` is `null`, show a clear unavailable/history fallback. Continue showing the standard 114-day expected date.
4. **Add a sow-code filter to User Home.** Add a labeled text input and clear action. Trim surrounding whitespace and match sow codes case-insensitively by substring, consistent with the existing code search interaction.
5. **Filter the current schedule locally.** Derive filtered upcoming and overdue lists from the already fetched `schedule()` data. Do not make another HTTP request when the filter changes and do not change the server endpoint's date range or grouping.
6. **Keep list feedback accurate.** Base section counts on filtered results. Distinguish an actually empty server group from a group with no matches for the current code filter, and provide a clear way to clear the filter.
7. **Preserve request states.** Keep existing loading, request error, retry, and server-empty behavior while adding filter-specific no-match feedback.
8. **Update focused client tests.** Cover both dates on cards, null history values, case-insensitive/trimmed substring matching across both sections, clear filter, no-match states, and that filtering does not trigger another schedule request.

## Acceptance criteria
- User Home displays both the existing 114-day date and the new average-based date with distinct labels.
- The sow-code filter searches both upcoming and overdue entries from the current response, ignoring case and surrounding whitespace.
- Filtering does not refetch the schedule or alter server grouping/order.
- Clearing the filter restores all entries from the loaded schedule.
- Missing historical averages do not hide the standard expected date.
- Loading, request failure, truly empty groups, and filter no-match states remain distinguishable.
# Add the Sow Average-Based Farrowing Date

## Goal
Extend the existing farrowing schedule response so each service exposes both:

- `expected_farrowing_date`: the existing standard estimate, based on 114 days after the current service's last mount.
- `average_expected_farrowing_date`: a second estimate based on this sow's average gestation duration across completed historical cycles.

Keep the existing `average_gestation_days` value in the response so the user can understand the basis for the second date.

## Findings from the current server code
- `service/domain/service.go` computes and persists `ExpectedFarrowingDate` as the latest mount date plus 114 days when a service is created or its mounts are updated.
- The existing `GET /api/v1/services/farrowing-schedule` endpoint and application use case are already implemented.
- `service/domain/farrowing_schedule.go` already defines `FarrowingScheduleEntry` with `ExpectedFarrowingDate` and nullable `AverageGestationDays`.
- `service/infrastructure/postgres_repository.go` already calculates each sow's historical average from completed (`Terminado`) services with a recorded farrowing, using actual farrowing date minus that service's last mount date.
- The schedule query does not yet select the current service's last mount date or calculate a second, average-based expected date.
- The HTTP response currently exposes `expected_farrowing_date` and `average_gestation_days` only.

## Calculation rule
For the current confirmed service, calculate:

`average_expected_farrowing_date = current_service_last_mount_date + rounded(average_gestation_days)`

Use the historical average from completed services of the same sow, excluding the current service. The current service's last mount date is the baseline because both the existing 114-day estimate and the historical gestation durations are measured from the last mount. Round the average to the nearest whole calendar day because the result is a `DATE`. If no eligible historical duration exists, return `null` for both the average and the average-based date; keep the standard 114-day date populated.

## Step-by-step server plan
1. **Preserve the existing estimate.** Keep `ExpectedFarrowingDate` and its 114-day domain calculation unchanged. Do not replace the stored date with an average-based date.
2. **Extend the schedule read model.** Add a nullable `AverageExpectedFarrowingDate` field to `FarrowingScheduleEntry`. Keep `AverageGestationDays` nullable and document that both new values depend on completed historical cycles.
3. **Extend the repository query.** In `farrowingScheduleQuery`, obtain the last mount date of each current confirmed service and use the already-computed per-sow average to calculate the second date. Select the field alongside the existing values in a stable order.
4. **Handle missing data explicitly.** If there is no current-service mount or no eligible completed-history average, leave the average-based date `NULL`. The existing domain validation should normally guarantee a mount for a valid service; keep the query safe for inconsistent legacy records.
5. **Scan the new query field.** Update the repository row scan and populate the domain read model without recalculating SQL dates in Go.
6. **Extend the HTTP response.** Add `average_expected_farrowing_date` to the schedule entry response, serialize it as `YYYY-MM-DD`, and emit JSON `null` when unavailable. Preserve the current `expected_farrowing_date` and `average_gestation_days` fields and their meanings.
7. **Leave application and container wiring unchanged.** The existing schedule use case already calls the repository and returns the domain read model; it should carry the added field through without new dependencies or endpoints.
8. **Update focused tests.** Cover the unchanged 114-day date, a sow with a fractional historical average and the agreed rounding, a different average-based date from the 114-day date, no historical data, and an overdue service. Verify the HTTP JSON includes both date fields and keeps nullable values as `null`.
9. **Review query cost.** Keep this in the existing schedule query so the API remains one request and one database query. Check the query plan with representative data before deciding whether new indexes are needed.

## Acceptance criteria
- Every schedule entry continues to return the standard 114-day `expected_farrowing_date`.
- Entries with completed sow history also return `average_gestation_days` and `average_expected_farrowing_date` based on the current service's last mount.
- The date calculation rounds the average to the nearest whole calendar day.
- Entries without completed history return `null` for the average and average-based date while retaining the standard expected date.
- Existing upcoming/overdue grouping, service state filtering, ordering, endpoint path, and authorization remain unchanged.
- No database migration or persisted average is introduced; the average-based date is calculated from current historical records at query time.


# Commit and Push Plan

## Current Git state
- The current branch is `main`; the configured remote is `origin` (`https://github.com/agueroda2004/pigs.git`).
- The worktree contains uncommitted server changes for the farrowing schedule endpoint and client changes for authentication/navigation, user-home schedule, and placeholder routes/pages.
- `server/PLAN.md` is currently untracked and contains the feature plans.
- No implementation changes have been committed or pushed as part of this plan.

## Step-by-step workflow
1. **Review the worktree.** Inspect `git diff` and all untracked files. Confirm each change belongs to the planned user navigation and farrowing schedule work; leave unrelated edits untouched.
2. **Create a feature branch.** Start a branch from the current `main` state while preserving the worktree, for example `feat/user-farrowing-schedule`.
3. **Run focused checks.** Run the relevant server service application/infrastructure tests and the client auth/navigation, services API, and user-home tests. Then run the broader server and client checks required by the project. Resolve failures before committing.
4. **Review the final diff.** Run `git diff --check`, inspect staged changes, and verify no generated files, secrets, or unrelated changes are included.
5. **Commit server work.** Stage only the service schedule domain/application/infrastructure code and relevant tests. Commit with a message such as `feat(service): add farrowing schedule endpoint`.
6. **Commit client work.** Stage only the client API model/service, user-home schedule UI, role-aware navigation, placeholder routes/pages, and relevant tests. Commit with a message such as `feat(client): show user farrowing schedule`.
7. **Commit the plan document.** Stage `server/PLAN.md` separately and commit with a message such as `docs: plan user farrowing schedule`.
8. **Verify commit contents.** Review `git status`, `git log`, and each commit's file list to ensure the commits contain only the intended changes and the worktree is clean.
9. **Push the feature branch.** Push the new branch to `origin` and set its upstream tracking branch. Do not push directly to `main` as part of this workflow.
10. **Report the result.** Provide the branch name, commit hashes/messages, checks run, and push status.

## Acceptance criteria
- All intended changes are committed on a feature branch based on current `main`.
- Server, client, and planning document changes are separated into reviewable commits.
- Relevant tests/checks pass and `git diff --check` is clean before committing.
- The branch is pushed to `origin` with upstream tracking configured.
- No unrelated work is staged, committed, or pushed.
