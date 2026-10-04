CREATE TABLE farrowings (
    id UUID PRIMARY KEY,
    service_id UUID NOT NULL UNIQUE,
    sow_id UUID NOT NULL,
    farrow_date DATE NOT NULL,
    start_time VARCHAR(20),
    end_time VARCHAR(20),
    location VARCHAR(100),
    live_born INTEGER NOT NULL DEFAULT 0,
    stillborn INTEGER NOT NULL DEFAULT 0,
    mummified INTEGER NOT NULL DEFAULT 0,
    litter_weight NUMERIC,
    stillborn_weight NUMERIC,
    is_manipulated BOOLEAN NOT NULL DEFAULT FALSE,
    note VARCHAR(500),
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    created_by UUID NOT NULL,
    updated_by UUID NOT NULL,
    CONSTRAINT farrowings_service_id_fkey
        FOREIGN KEY (service_id) REFERENCES services (id),
    CONSTRAINT farrowings_sow_id_fkey
        FOREIGN KEY (sow_id) REFERENCES sows (id),
    CONSTRAINT farrowings_created_by_fkey
        FOREIGN KEY (created_by) REFERENCES users (id),
    CONSTRAINT farrowings_updated_by_fkey
        FOREIGN KEY (updated_by) REFERENCES users (id)
);

CREATE INDEX farrowings_sow_id_idx ON farrowings (sow_id);
CREATE INDEX farrowings_farrow_date_idx ON farrowings (farrow_date);

CREATE TABLE farrowing_operators (
    id UUID PRIMARY KEY,
    farrowing_id UUID NOT NULL,
    operator_id UUID NOT NULL,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    CONSTRAINT farrowing_operators_farrowing_id_fkey
        FOREIGN KEY (farrowing_id) REFERENCES farrowings (id) ON DELETE CASCADE,
    CONSTRAINT farrowing_operators_operator_id_fkey
        FOREIGN KEY (operator_id) REFERENCES operators (id),
    CONSTRAINT farrowing_operators_unique UNIQUE (farrowing_id, operator_id)
);

CREATE INDEX farrowing_operators_farrowing_id_idx ON farrowing_operators (farrowing_id);
CREATE INDEX farrowing_operators_operator_id_idx ON farrowing_operators (operator_id);

CREATE TABLE farrowing_medications (
    id UUID PRIMARY KEY,
    farrowing_id UUID NOT NULL,
    medication_id UUID NOT NULL,
    dose INTEGER NOT NULL,
    applied_by UUID NOT NULL,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    CONSTRAINT farrowing_medications_farrowing_id_fkey
        FOREIGN KEY (farrowing_id) REFERENCES farrowings (id) ON DELETE CASCADE,
    CONSTRAINT farrowing_medications_medication_id_fkey
        FOREIGN KEY (medication_id) REFERENCES medications (id),
    CONSTRAINT farrowing_medications_applied_by_fkey
        FOREIGN KEY (applied_by) REFERENCES operators (id),
    CONSTRAINT farrowing_medications_unique UNIQUE (farrowing_id, medication_id)
);

CREATE INDEX farrowing_medications_farrowing_id_idx ON farrowing_medications (farrowing_id);
CREATE INDEX farrowing_medications_medication_id_idx ON farrowing_medications (medication_id);
CREATE INDEX farrowing_medications_applied_by_idx ON farrowing_medications (applied_by);
