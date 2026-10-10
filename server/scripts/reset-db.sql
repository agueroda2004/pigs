-- Resets all business data for local development.
-- Keeps the users table and their active sessions (refresh_tokens) intact.
TRUNCATE TABLE
    breeds,
    boars,
    sows,
    operators,
    services,
    mounts,
    abortions,
    sow_removals,
    boar_removals,
    medications,
    farrowings,
    farrowing_operators,
    farrowing_medications,
    piglet_deaths,
    piglet_fosterings,
    partial_weagings,
    weagings
RESTART IDENTITY CASCADE;
