export interface PigletFostering {
  id: string;
  donor_farrowing_id: string;
  receiver_farrowing_id: string;
  donor_sow_id: string;
  receiver_sow_id: string;
  movement_date: string;
  quantity: number;
  note: string | null;
  created_at: string;
  updated_at: string;
  created_by: string;
  updated_by: string;
}

export interface CreatePigletFosteringRequest {
  donor_sow_id: string;
  receiver_sow_id: string;
  movement_date: string;
  quantity: number;
  note?: string;
}

export interface PigletFosteringFilters {
  donor_sow_id?: string;
  receiver_sow_id?: string;
  from?: string;
  to?: string;
}
