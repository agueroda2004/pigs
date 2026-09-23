CREATE TABLE breeds (
    id UUID PRIMARY KEY,
    name VARCHAR(100) NOT NULL UNIQUE,
    active BOOLEAN NOT NULL DEFAULT TRUE,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    created_by UUID NOT NULL,
    updated_by UUID NOT NULL,
    CONSTRAINT breeds_created_by_fkey
        FOREIGN KEY (created_by) REFERENCES users (id),
    CONSTRAINT breeds_updated_by_fkey
        FOREIGN KEY (updated_by) REFERENCES users (id)
);
