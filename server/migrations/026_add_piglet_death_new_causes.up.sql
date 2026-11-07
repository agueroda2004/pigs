-- Adds the "Muerte Súbita" and "Baja Viabilidad" options to the piglet death cause enum.
-- Applied migration 025 is left untouched; this is a forward-only enum extension.
ALTER TYPE piglet_death_cause ADD VALUE IF NOT EXISTS 'Muerte_subita';
ALTER TYPE piglet_death_cause ADD VALUE IF NOT EXISTS 'Baja_viabilidad';