export type AbortionCause =
  'Desconocido' | 'Infeccioso' | 'Traumatismo' | 'Manejo' | 'Nutricional' | 'Otro';

export interface Abortion {
  id: string;
  sow_id: string;
  service_id: string;
  abortion_date: string;
  cause: AbortionCause;
  note: string | null;
  created_at: string;
  updated_at: string;
  created_by: string;
  updated_by: string;
}

export interface CreateAbortionRequest {
  sow_id: string;
  abortion_date: string;
  cause: AbortionCause;
  note?: string;
}

export interface AbortionFilters {
  sow_id?: string;
}
