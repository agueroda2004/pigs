export type PigletDeathCause =
  | 'Aplastado'
  | 'Debilidad'
  | 'Diarrea'
  | 'Deformidad'
  | 'Otro'
  | 'Canibalismo'
  | 'Pata_abierta'
  | 'Bacteria'
  | 'Reaccion_medicamento';

export type Turn = 'Mañana' | 'Tarde' | 'Madrugada' | 'Madrugada_no_asistida';

export interface PigletDeath {
  id: string;
  farrowing_id: string;
  sow_id: string;
  operator_id: string;
  operator_name: string;
  death_date: string;
  quantity: number;
  weight: number | null;
  cause: PigletDeathCause;
  turn: Turn;
  note: string | null;
  created_at: string;
  updated_at: string;
  created_by: string;
  updated_by: string;
}

export interface CreatePigletDeathRequest {
  sow_id: string;
  operator_id: string;
  death_date: string;
  quantity: number;
  weight?: number;
  cause: PigletDeathCause;
  turn: Turn;
  note?: string;
}

export interface PigletDeathFilters {
  sow_id?: string;
  from?: string;
  to?: string;
}
