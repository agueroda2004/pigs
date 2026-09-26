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
  removal_date: string;
  type: RemovalType;
  reason: RemovalReason;
  note: string | null;
  last_state: SowState;
  created_at: string;
  updated_at: string;
  created_by: string;
  updated_by: string;
}

export interface CreateSowRemovalRequest {
  sow_id: string;
  removal_date: string;
  type: RemovalType;
  reason: RemovalReason;
  note?: string;
}

export interface SowRemovalFilters {
  sow_id?: string;
}
