import { SowState } from '../sows/sow.models';

export type RemovalType = 'Muerte' | 'Desecho' | 'Sacrificio';

export type RemovalReason =
  | 'Edad_Paridad'
  | 'Fallo_Reproductivo'
  | 'Baja_Productividad'
  | 'Problema_Locomotor'
  | 'Enfermedad'
  | 'Muerte_Subita'
  | 'Otro';

export interface SowRemoval {
  id: string;
  sow_id: string;
  sow_code: string;
  removal_date: string;
  type: RemovalType;
  reason: RemovalReason;
  note: string | null;
  last_state: SowState;
}

export interface CreateSowRemovalRequest {
  sow_id: string;
  removal_date: string;
  type: RemovalType;
  reason: RemovalReason;
  note?: string;
}

export interface UpdateSowRemovalRequest {
  removal_date?: string;
  type?: RemovalType;
  reason?: RemovalReason;
  note?: string;
}

export interface SowRemovalFilters {
  sow_code?: string;
}
