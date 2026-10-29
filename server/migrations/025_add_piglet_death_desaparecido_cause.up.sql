-- Adds the "Desaparecido" option to the piglet death cause enum.
-- Applied migration 017 is left untouched; this is a forward-only enum extension.
ALTER TYPE piglet_death_cause ADD VALUE IF NOT EXISTS 'Desaparecido';
