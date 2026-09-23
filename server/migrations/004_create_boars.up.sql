CREATE TYPE boar_state AS ENUM ('Vivo', 'Muerto', 'Desecho', 'Sacrificado');

CREATE TABLE boars (
    id UUID PRIMARY KEY,
    code VARCHAR(50) NOT NULL UNIQUE,
    location VARCHAR(100),
    active BOOLEAN NOT NULL DEFAULT TRUE,
    entry_date DATE NOT NULL,
    birth_date DATE,
    note VARCHAR(500),
    state boar_state NOT NULL DEFAULT 'Vivo',
    breed_id UUID NOT NULL,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    created_by UUID NOT NULL,
    updated_by UUID NOT NULL,
    CONSTRAINT boars_breed_id_fkey
        FOREIGN KEY (breed_id) REFERENCES breeds (id),
    CONSTRAINT boars_created_by_fkey
        FOREIGN KEY (created_by) REFERENCES users (id),
    CONSTRAINT boars_updated_by_fkey
        FOREIGN KEY (updated_by) REFERENCES users (id)
);

CREATE INDEX boars_breed_id_idx ON boars (breed_id);
