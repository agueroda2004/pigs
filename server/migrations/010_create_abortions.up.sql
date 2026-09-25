CREATE TYPE abortion_cause AS ENUM (
    'Desconocido',
    'Infeccioso',
    'Traumatismo',
    'Manejo',
    'Nutricional',
    'Otro'
);

CREATE TABLE abortions (
    id UUID PRIMARY KEY,
    sow_id UUID NOT NULL,
    service_id UUID NOT NULL,
    abortion_date DATE NOT NULL,
    cause abortion_cause NOT NULL,
    note VARCHAR(500),
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    created_by UUID NOT NULL,
    updated_by UUID NOT NULL,
    CONSTRAINT abortions_sow_id_fkey
        FOREIGN KEY (sow_id) REFERENCES sows (id),
    CONSTRAINT abortions_service_id_fkey
        FOREIGN KEY (service_id) REFERENCES services (id),
    CONSTRAINT abortions_created_by_fkey
        FOREIGN KEY (created_by) REFERENCES users (id),
    CONSTRAINT abortions_updated_by_fkey
        FOREIGN KEY (updated_by) REFERENCES users (id)
);

CREATE INDEX abortions_sow_id_idx ON abortions (sow_id);
CREATE INDEX abortions_service_id_idx ON abortions (service_id);
