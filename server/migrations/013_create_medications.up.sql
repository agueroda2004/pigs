CREATE TABLE medications (
    id UUID PRIMARY KEY,
    name VARCHAR(50) NOT NULL UNIQUE,
    active BOOLEAN NOT NULL DEFAULT TRUE,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    created_by UUID NOT NULL,
    updated_by UUID NOT NULL,
    CONSTRAINT medications_created_by_fkey
        FOREIGN KEY (created_by) REFERENCES users (id),
    CONSTRAINT medications_updated_by_fkey
        FOREIGN KEY (updated_by) REFERENCES users (id)
);
