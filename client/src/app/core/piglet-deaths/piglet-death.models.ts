export type PigletDeathCause =
  | 'Aplastado'
  | 'Debilidad'
  | 'Diarrea'
  | 'Deformidad'
  | 'Otro'
  | 'Canibalismo'
  | 'Pata_abierta'
  | 'Bacteria'
  | 'Reaccion_medicamento'
  | 'Desaparecido';

export type Turn = 'Mañana' | 'Tarde' | 'Madrugada' | 'Madrugada_no_asistida' | 'Sin Turno';

export interface PigletDeath {
  id: string;
  farrowing_id: string;
  sow_id: string;
  sow_code: string;
  operator_id: string;
  operator_name: string;
  death_date: string;
  quantity: number;
  weight: number | null;
  cause: PigletDeathCause;
  turn: Turn;
  note: string | null;
}

export interface PigletDeathPage {
  items: PigletDeath[];
  total: number;
  page: number;
  page_size: number;
  total_pages: number;
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

export interface UpdatePigletDeathRequest {
  operator_id: string;
  weight?: number;
  clear_weight?: boolean;
  cause: PigletDeathCause;
  turn: Turn;
  note?: string;
}

export interface PigletDeathFilters {
  sow_code?: string;
  from?: string;
  to?: string;
}
