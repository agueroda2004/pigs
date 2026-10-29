-- Adds the "Sin Turno" option to the piglet death turn enum.
-- Applied migration 017 is left untouched; this is a forward-only enum extension.
ALTER TYPE turn ADD VALUE IF NOT EXISTS 'Sin Turno';
