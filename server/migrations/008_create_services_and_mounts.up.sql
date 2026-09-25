CREATE TYPE service_state AS ENUM (
    'Confirmado','Fallido','Aborto','Terminado'
);

CREATE TYPE mount_type AS ENUM (
    'Natural','Artificial'
);

CREATE TABLE services (
    id UUID PRIMARY KEY,
    sow_id UUID NOT NULL,
    expected_farrowing_date DATE,
    note VARCHAR(500),
    state service_state NOT NULL DEFAULT 'Confirmado',
    location VARCHAR(100),
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    created_by UUID NOT NULL,
    updated_by UUID NOT NULL,
    CONSTRAINT services_sow_id_fkey
        FOREIGN KEY (sow_id) REFERENCES sows (id),
    CONSTRAINT services_created_by_fkey
        FOREIGN KEY (created_by) REFERENCES users (id),
    CONSTRAINT services_updated_by_fkey
        FOREIGN KEY (updated_by) REFERENCES users (id)
);

CREATE INDEX services_sow_id_idx ON services (sow_id);

CREATE TABLE mounts (
    id UUID PRIMARY KEY,
    service_id UUID NOT NULL,
    boar_id UUID NOT NULL,
    operator_id UUID NOT NULL,
    mount_number INTEGER NOT NULL,
    mount_date TIMESTAMP NOT NULL,
    type mount_type NOT NULL DEFAULT 'Artificial',
    note VARCHAR(500),
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    created_by UUID NOT NULL,
    updated_by UUID NOT NULL,
    CONSTRAINT mounts_service_id_fkey
        FOREIGN KEY (service_id) REFERENCES services (id),
    CONSTRAINT mounts_boar_id_fkey
        FOREIGN KEY (boar_id) REFERENCES boars (id),
    CONSTRAINT mounts_operator_id_fkey
        FOREIGN KEY (operator_id) REFERENCES operators (id),
    CONSTRAINT mounts_created_by_fkey
        FOREIGN KEY (created_by) REFERENCES users (id),
    CONSTRAINT mounts_updated_by_fkey
        FOREIGN KEY (updated_by) REFERENCES users (id),
    CONSTRAINT mounts_service_number_unique UNIQUE (service_id, mount_number)
);

CREATE INDEX mounts_service_id_idx ON mounts (service_id);
CREATE INDEX mounts_boar_id_idx ON mounts (boar_id);
CREATE INDEX mounts_operator_id_idx ON mounts (operator_id);
