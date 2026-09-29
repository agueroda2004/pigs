CREATE TABLE boar_removals (
    id UUID PRIMARY KEY,
    boar_id UUID NOT NULL UNIQUE,
    removal_date DATE NOT NULL,
    type removal_type NOT NULL,
    reason removal_reason NOT NULL,
    note VARCHAR(500),
    last_state boar_state NOT NULL,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    created_by UUID NOT NULL,
    updated_by UUID NOT NULL,
    CONSTRAINT boar_removals_boar_id_fkey
        FOREIGN KEY (boar_id) REFERENCES boars (id),
    CONSTRAINT boar_removals_created_by_fkey
        FOREIGN KEY (created_by) REFERENCES users (id),
    CONSTRAINT boar_removals_updated_by_fkey
        FOREIGN KEY (updated_by) REFERENCES users (id)
);

CREATE INDEX boar_removals_boar_id_idx ON boar_removals (boar_id);
CREATE INDEX boar_removals_removal_date_idx ON boar_removals (removal_date);
