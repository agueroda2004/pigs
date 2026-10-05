CREATE TYPE piglet_death_cause AS ENUM (
    'Aplastado',
    'Debilidad',
    'Diarrea',
    'Deformidad',
    'Otro',
    'Canibalismo',
    'Pata_abierta',
    'Bacteria',
    'Reaccion_medicamento'
);

CREATE TYPE turn AS ENUM (
    'Mañana',
    'Tarde',
    'Madrugada',
    'Madrugada_no_asistida'
);

CREATE TABLE piglet_deaths (
    id UUID PRIMARY KEY,
    farrowing_id UUID NOT NULL,
    operator_id UUID NOT NULL,
    death_date DATE NOT NULL,
    quantity INTEGER NOT NULL,
    cause piglet_death_cause NOT NULL,
    turn turn NOT NULL,
    note VARCHAR(500),
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    created_by UUID NOT NULL,
    updated_by UUID NOT NULL,
    CONSTRAINT piglet_deaths_farrowing_id_fkey
        FOREIGN KEY (farrowing_id) REFERENCES farrowings (id) ON DELETE CASCADE,
    CONSTRAINT piglet_deaths_operator_id_fkey
        FOREIGN KEY (operator_id) REFERENCES operators (id),
    CONSTRAINT piglet_deaths_created_by_fkey
        FOREIGN KEY (created_by) REFERENCES users (id),
    CONSTRAINT piglet_deaths_updated_by_fkey
        FOREIGN KEY (updated_by) REFERENCES users (id)
);

CREATE INDEX piglet_deaths_farrowing_id_idx ON piglet_deaths (farrowing_id);
CREATE INDEX piglet_deaths_death_date_idx ON piglet_deaths (death_date);
