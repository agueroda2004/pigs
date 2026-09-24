export type SowState =
  | 'Viva'
  | 'Muerta'
  | 'Desecho'
  | 'Sacrificada'
  | 'Abortada'
  | 'Gestando'
  | 'Lactando'
  | 'Destetada';

export type SowOrigin = 'Propio' | 'Externo';

export interface Sow {
  id: string;
  code: string;
  location: string | null;
  active: boolean;
  entry_date: string;
  birth_date: string | null;
  note: string | null;
  state: SowState;
  origin: SowOrigin;
  parity: number;
  breed_id: string;
  created_at: string;
  updated_at: string;
  created_by: string;
  updated_by: string;
}

export interface CreateSowRequest {
  code: string;
  location?: string;
  entry_date: string;
  birth_date?: string;
  note?: string;
  origin: SowOrigin;
  parity: number;
  breed_id: string;
}

export interface UpdateSowRequest {
  code?: string;
  location?: string;
  active?: boolean;
  entry_date?: string;
  birth_date?: string;
  note?: string;
  origin?: SowOrigin;
  breed_id?: string;
}

export interface SowFilters {
  code?: string;
  breed_id?: string;
  origin?: SowOrigin;
  active?: boolean;
  state?: SowState;
}
