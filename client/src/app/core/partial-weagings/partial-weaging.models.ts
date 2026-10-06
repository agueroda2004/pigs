export type PartialWeagingType = 'Normal' | 'Nodriza' | 'Baja_Viabilidad';

export interface PartialWeaging {
  id: string;
  farrowing_id: string;
  sow_id: string;
  weaging_date: string;
  quantity: number;
  total_weight: number | null;
  type: PartialWeagingType;
  note: string | null;
  created_at: string;
  updated_at: string;
  created_by: string;
  updated_by: string;
}

export interface CreatePartialWeagingRequest {
  sow_id: string;
  weaging_date: string;
  quantity: number;
  total_weight?: number;
  type: PartialWeagingType;
  note?: string;
}

export interface PartialWeagingFilters {
  sow_id?: string;
  from?: string;
  to?: string;
}
