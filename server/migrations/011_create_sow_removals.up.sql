CREATE TYPE removal_type AS ENUM (
    'Muerte',
    'Desecho',
    'Sacrificio'
);

CREATE TYPE removal_reason AS ENUM (
    'Edad_Paridad',
    'Fallo_Reproductivo',
    'Baja_Productividad',
    'Problema_Locomotor',
    'Enfermedad',
    'Muerte_Subita',
    'Otro'
);

CREATE TABLE sow_removals (
    id UUID PRIMARY KEY,
    sow_id UUID NOT NULL UNIQUE,
    removal_date DATE NOT NULL,
    type removal_type NOT NULL,
    reason removal_reason NOT NULL,
    note VARCHAR(500),
    last_state sow_state NOT NULL,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    created_by UUID NOT NULL,
    updated_by UUID NOT NULL,
    CONSTRAINT sow_removals_sow_id_fkey
        FOREIGN KEY (sow_id) REFERENCES sows (id),
    CONSTRAINT sow_removals_created_by_fkey
        FOREIGN KEY (created_by) REFERENCES users (id),
    CONSTRAINT sow_removals_updated_by_fkey
        FOREIGN KEY (updated_by) REFERENCES users (id)
);

CREATE INDEX sow_removals_sow_id_idx ON sow_removals (sow_id);
CREATE INDEX sow_removals_removal_date_idx ON sow_removals (removal_date);
