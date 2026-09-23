export type BoarState = 'Vivo' | 'Muerto' | 'Desecho' | 'Sacrificado';

export type BoarOrigin = 'Propio' | 'Externo';

export interface Boar {
  id: string;
  code: string;
  location: string | null;
  active: boolean;
  entry_date: string;
  birth_date: string | null;
  note: string | null;
  state: BoarState;
  origin: BoarOrigin;
  breed_id: string;
  created_at: string;
  updated_at: string;
  created_by: string;
  updated_by: string;
}

export interface CreateBoarRequest {
  code: string;
  location?: string;
  entry_date: string;
  birth_date?: string;
  note?: string;
  origin: BoarOrigin;
  breed_id: string;
}

export interface UpdateBoarRequest {
  code?: string;
  location?: string;
  active?: boolean;
  entry_date?: string;
  birth_date?: string;
  note?: string;
  origin?: BoarOrigin;
  breed_id?: string;
}

export interface BoarFilters {
  code?: string;
  breed_id?: string;
  origin?: BoarOrigin;
  active?: boolean;
}
