CREATE TABLE piglet_fosterings (
    id UUID PRIMARY KEY,
    donor_farrowing_id UUID NOT NULL,
    receiver_farrowing_id UUID NOT NULL,
    movement_date DATE NOT NULL,
    quantity INTEGER NOT NULL,
    note VARCHAR(500),
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    created_by UUID NOT NULL,
    updated_by UUID NOT NULL,
    CONSTRAINT piglet_fosterings_donor_farrowing_id_fkey
        FOREIGN KEY (donor_farrowing_id) REFERENCES farrowings (id) ON DELETE CASCADE,
    CONSTRAINT piglet_fosterings_receiver_farrowing_id_fkey
        FOREIGN KEY (receiver_farrowing_id) REFERENCES farrowings (id) ON DELETE CASCADE,
    CONSTRAINT piglet_fosterings_created_by_fkey
        FOREIGN KEY (created_by) REFERENCES users (id),
    CONSTRAINT piglet_fosterings_updated_by_fkey
        FOREIGN KEY (updated_by) REFERENCES users (id)
);

CREATE INDEX piglet_fosterings_donor_farrowing_id_idx ON piglet_fosterings (donor_farrowing_id);
CREATE INDEX piglet_fosterings_receiver_farrowing_id_idx ON piglet_fosterings (receiver_farrowing_id);
CREATE INDEX piglet_fosterings_movement_date_idx ON piglet_fosterings (movement_date);
