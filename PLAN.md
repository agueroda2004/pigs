## Plan: Server Farrowing History

### Objective

Extend the server-side sow history response so farrowing events expose complete farrowing details, including operator and medication information.

### Scope

1. **Review the existing server flow**
   - Inspect:
     - `server/internal/modules/service/application/services_history_by_sow.go`
     - `server/internal/modules/sow/application/sow_history_by_sow.go`
   - Trace the sow history application service, repository, HTTP handler, and domain read models.
   - Review the farrowing, operator, medication, user, and sow-history database relationships.

2. **Define the farrowing history read model**
   - Ensure the server response contains:
     - `farrow_date`
     - `start_time`
     - `end_time`
     - `duration`
     - `location`
     - `live_born`
     - `mummified`
     - `stillborn`
     - `litter_weight`
     - `stillborn_weight`
     - `is_manipulated`
     - `note`
     - `is_nurse`
     - operator names
     - medication names, doses, and operator names
   - Preserve nullable fields and existing JSON naming conventions.
   - Decide whether duration is persisted or derived from the start and end times, keeping the rule in the server read model.

3. **Implement repository queries**
   - Extend the sow history repository query or related history queries to load farrowing records for the requested sow.
   - Join farrowing operators with the operator/user source to return display names.
   - Join farrowing medications with the medication source to return medication names.
   - Return medication dose and the administering operator name.
   - Ensure one-to-many joins do not duplicate farrowing history events.
   - Return empty collections instead of null collections where required by the API contract.

4. **Integrate the farrowing event**
   - Add the enriched farrowing data to the server’s sow history event model.
   - Preserve existing event ordering and behavior for services, abortions, farrowings, and other history events.
   - Keep inactive-sow and not-found error behavior unchanged.

5. **Document server methods**
   - Add or update approximately two-line English documentation for every new or modified server method/function, following `server/AGENTS.MD`.

6. **Add focused server tests**
   - Add application tests for the enriched sow-history behavior.
   - Add infrastructure tests for:
     - Farrowing field mapping.
     - Operator and medication joins.
     - Multiple operators or medications without duplicate history events.
     - Empty operator/medication collections.
     - Nullable fields and duration calculation.
     - Not-found and inactive-sow behavior.
   - Place tests in the established module `tests` folders:
     - `application/tests`
     - `infrastructure/tests`
     - `domain/tests` when domain logic changes.

7. **Validate**
   - Run focused tests for the affected server modules.
   - Run formatting and static checks required by the Go project.
   - Verify the final HTTP sow-history response contains all requested farrowing fields and nested names.