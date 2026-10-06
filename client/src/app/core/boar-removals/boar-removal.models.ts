import { BoarState } from '../boars/boar.models';

export type RemovalType = 'Muerte' | 'Desecho' | 'Sacrificio';

export type RemovalReason =
  | 'Edad_Paridad'
  | 'Fallo_Reproductivo'
  | 'Baja_Productividad'
  | 'Problema_Locomotor'
  | 'Enfermedad'
  | 'Muerte_Subita'
  | 'Otro';

export interface BoarRemoval {
  id: string;
  boar_id: string;
  removal_date: string;
  type: RemovalType;
  reason: RemovalReason;
  note: string | null;
  last_state: BoarState;
}

export interface CreateBoarRemovalRequest {
  boar_id: string;
  removal_date: string;
  type: RemovalType;
  reason: RemovalReason;
  note?: string;
}

export interface UpdateBoarRemovalRequest {
  removal_date?: string;
  type?: RemovalType;
  reason?: RemovalReason;
  note?: string;
}

export interface BoarRemovalFilters {
  boar_id?: string;
}
