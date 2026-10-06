CREATE TABLE weagings (
    id UUID PRIMARY KEY,
    farrowing_id UUID NOT NULL UNIQUE,
    weaging_date DATE NOT NULL,
    quantity INTEGER NOT NULL,
    total_weight NUMERIC,
    destination VARCHAR(100),
    note VARCHAR(500),
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    created_by UUID NOT NULL,
    updated_by UUID NOT NULL,
    CONSTRAINT weagings_farrowing_id_fkey
        FOREIGN KEY (farrowing_id) REFERENCES farrowings (id) ON DELETE CASCADE,
    CONSTRAINT weagings_created_by_fkey
        FOREIGN KEY (created_by) REFERENCES users (id),
    CONSTRAINT weagings_updated_by_fkey
        FOREIGN KEY (updated_by) REFERENCES users (id)
);

CREATE INDEX weagings_weaging_date_idx ON weagings (weaging_date);
