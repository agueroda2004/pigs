export interface PigletFostering {
  id: string;
  donor_farrowing_id: string;
  receiver_farrowing_id: string;
  donor_sow_id: string;
  donor_sow_code: string;
  receiver_sow_id: string;
  receiver_sow_code: string;
  movement_date: string;
  quantity: number;
  note: string | null;
}

export interface PigletFosteringPage {
  items: PigletFostering[];
  total: number;
  page: number;
  page_size: number;
  total_pages: number;
}

export interface CreatePigletFosteringRequest {
  donor_sow_id: string;
  receiver_sow_id: string;
  movement_date: string;
  quantity: number;
  note?: string;
}

export interface UpdatePigletFosteringRequest {
  movement_date: string;
  quantity: number;
  note?: string;
}

export interface PigletFosteringFilters {
  sow_code?: string;
  from?: string;
  to?: string;
}
