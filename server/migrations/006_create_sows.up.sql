CREATE TYPE sow_state AS ENUM (
    'Viva',
    'Muerta',
    'Desecho',
    'Sacrificada',
    'Abortada',
    'Gestando',
    'Lactando',
    'Destetada'
);

CREATE TABLE sows (
    id UUID PRIMARY KEY,
    code VARCHAR(50) NOT NULL UNIQUE,
    location VARCHAR(100),
    active BOOLEAN NOT NULL DEFAULT TRUE,
    entry_date DATE NOT NULL,
    birth_date DATE,
    note VARCHAR(500),
    state sow_state NOT NULL DEFAULT 'Viva',
    origin boar_origin NOT NULL DEFAULT 'Propio',
    parity INTEGER NOT NULL DEFAULT 0,
    breed_id UUID NOT NULL,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    created_by UUID NOT NULL,
    updated_by UUID NOT NULL,
    CONSTRAINT sows_breed_id_fkey
        FOREIGN KEY (breed_id) REFERENCES breeds (id),
    CONSTRAINT sows_created_by_fkey
        FOREIGN KEY (created_by) REFERENCES users (id),
    CONSTRAINT sows_updated_by_fkey
        FOREIGN KEY (updated_by) REFERENCES users (id)
);

CREATE INDEX sows_breed_id_idx ON sows (breed_id);
