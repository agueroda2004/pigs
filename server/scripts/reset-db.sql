-- Resets operational farm data for local development.
--
-- Preserved (never truncated by this script):
--   users, refresh_tokens, operators, medications, breeds
--
-- Cleared:
--   boars, sows, services, mounts, abortions, sow_removals, boar_removals,
--   farrowings, farrowing_operators, farrowing_medications, piglet_deaths,
--   piglet_fosterings, partial_weagings, weagings
--
-- Every table that references the tables below is listed explicitly, so no
-- CASCADE is needed. CASCADE is intentionally omitted: if a future migration
-- adds a table that references operational data, this reset fails loudly
-- instead of silently truncating a table that may need to be preserved.
--
-- TRUNCATE is transactional in PostgreSQL. Wrapping it in BEGIN/COMMIT means
-- any error leaves the database untouched.
--
-- Run this only against a local development database, after taking a backup.
-- See migrations/README.md for the full procedure and verification queries.

BEGIN;

TRUNCATE TABLE
    boars,
    sows,
    services,
    mounts,
    abortions,
    sow_removals,
    boar_removals,
    farrowings,
    farrowing_operators,
    farrowing_medications,
    piglet_deaths,
    piglet_fosterings,
    partial_weagings,
    weagings
RESTART IDENTITY;

COMMIT;
