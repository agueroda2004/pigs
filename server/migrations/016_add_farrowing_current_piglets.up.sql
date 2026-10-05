ALTER TABLE farrowings
    ADD COLUMN current_piglets INTEGER NOT NULL DEFAULT 0;

UPDATE farrowings
SET current_piglets = live_born;
