ALTER TABLE piglet_deaths
    DROP CONSTRAINT piglet_deaths_farrowing_id_fkey;
ALTER TABLE piglet_deaths
    ADD CONSTRAINT piglet_deaths_farrowing_id_fkey
        FOREIGN KEY (farrowing_id) REFERENCES farrowings (id) ON DELETE CASCADE;

ALTER TABLE piglet_fosterings
    DROP CONSTRAINT piglet_fosterings_donor_farrowing_id_fkey;
ALTER TABLE piglet_fosterings
    ADD CONSTRAINT piglet_fosterings_donor_farrowing_id_fkey
        FOREIGN KEY (donor_farrowing_id) REFERENCES farrowings (id) ON DELETE CASCADE;

ALTER TABLE piglet_fosterings
    DROP CONSTRAINT piglet_fosterings_receiver_farrowing_id_fkey;
ALTER TABLE piglet_fosterings
    ADD CONSTRAINT piglet_fosterings_receiver_farrowing_id_fkey
        FOREIGN KEY (receiver_farrowing_id) REFERENCES farrowings (id) ON DELETE CASCADE;

ALTER TABLE partial_weagings
    DROP CONSTRAINT partial_weagings_farrowing_id_fkey;
ALTER TABLE partial_weagings
    ADD CONSTRAINT partial_weagings_farrowing_id_fkey
        FOREIGN KEY (farrowing_id) REFERENCES farrowings (id) ON DELETE CASCADE;

ALTER TABLE weagings
    DROP CONSTRAINT weagings_farrowing_id_fkey;
ALTER TABLE weagings
    ADD CONSTRAINT weagings_farrowing_id_fkey
        FOREIGN KEY (farrowing_id) REFERENCES farrowings (id) ON DELETE CASCADE;
