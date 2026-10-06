export interface Weaging {
  id: string;
  farrowing_id: string;
  sow_id: string;
  weaging_date: string;
  quantity: number;
  total_weight: number | null;
  destination: string | null;
  note: string | null;
  created_at: string;
  updated_at: string;
  created_by: string;
  updated_by: string;
}

export interface CreateWeagingRequest {
  sow_id: string;
  weaging_date: string;
  quantity: number;
  total_weight?: number;
  destination?: string;
  note?: string;
}

export interface WeagingFilters {
  sow_id?: string;
  from?: string;
  to?: string;
}
