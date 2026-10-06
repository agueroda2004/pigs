DROP TABLE IF EXISTS partial_weagings;

ALTER TABLE farrowings
    DROP COLUMN IF EXISTS is_nurse,
    DROP COLUMN IF EXISTS nurse_start_date;

DROP TYPE IF EXISTS partial_weaging_type;
