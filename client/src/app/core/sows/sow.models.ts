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

export const SERVICEABLE_SOW_STATES: SowState[] = ['Viva', 'Destetada', 'Abortada', 'Gestando'];
export const REMOVABLE_SOW_STATES: SowState[] = ['Viva', 'Destetada', 'Abortada', 'Gestando'];

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
}

export interface SowDropdown {
  id: string;
  code: string;
}

export interface SowPage {
  items: Sow[];
  total: number;
  page: number;
  page_size: number;
  total_pages: number;
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
  entry_date?: string;
  birth_date?: string;
  note?: string;
  origin?: SowOrigin;
  breed_id?: string;
  parity?: number;
}

export interface SowFilters {
  code?: string;
  breed_id?: string;
  origin?: SowOrigin;
  active?: boolean;
  state?: SowState;
}
