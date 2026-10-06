CREATE TYPE partial_weaging_type AS ENUM (
    'Normal',
    'Nodriza',
    'Baja_Viabilidad'
);

ALTER TABLE farrowings
    ADD COLUMN is_nurse BOOLEAN NOT NULL DEFAULT FALSE,
    ADD COLUMN nurse_start_date DATE;

CREATE TABLE partial_weagings (
    id UUID PRIMARY KEY,
    farrowing_id UUID NOT NULL,
    weaging_date DATE NOT NULL,
    quantity INTEGER NOT NULL,
    total_weight NUMERIC,
    type partial_weaging_type NOT NULL,
    note VARCHAR(500),
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    created_by UUID NOT NULL,
    updated_by UUID NOT NULL,
    CONSTRAINT partial_weagings_farrowing_id_fkey
        FOREIGN KEY (farrowing_id) REFERENCES farrowings (id) ON DELETE CASCADE,
    CONSTRAINT partial_weagings_created_by_fkey
        FOREIGN KEY (created_by) REFERENCES users (id),
    CONSTRAINT partial_weagings_updated_by_fkey
        FOREIGN KEY (updated_by) REFERENCES users (id)
);

CREATE INDEX partial_weagings_farrowing_id_idx ON partial_weagings (farrowing_id);
CREATE INDEX partial_weagings_weaging_date_idx ON partial_weagings (weaging_date);
