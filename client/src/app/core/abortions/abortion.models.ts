export type AbortionCause =
  'Desconocido' | 'Infeccioso' | 'Traumatismo' | 'Manejo' | 'Nutricional' | 'Otro';

export interface Abortion {
  id: string;
  sow_id: string;
  sow_code: string;
  service_id: string;
  abortion_date: string;
  cause: AbortionCause;
  note: string | null;
}

export interface AbortionPage {
  items: Abortion[];
  total: number;
  page: number;
  page_size: number;
  total_pages: number;
}

export interface CreateAbortionRequest {
  sow_id: string;
  abortion_date: string;
  cause: AbortionCause;
  note?: string;
}

export interface UpdateAbortionRequest {
  abortion_date?: string;
  cause?: AbortionCause;
  note?: string;
}

export interface AbortionFilters {
  sow_code?: string;
}
