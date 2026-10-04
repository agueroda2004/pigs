ALTER TABLE farrowing_medications
    ALTER COLUMN dose TYPE INTEGER USING dose::integer;
