CREATE TYPE boar_origin AS ENUM ('Propio', 'Externo');

ALTER TABLE boars
    ADD COLUMN origin boar_origin NOT NULL DEFAULT 'Propio';
