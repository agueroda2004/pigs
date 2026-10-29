export interface Weaging {
  id: string;
  farrowing_id: string;
  sow_id: string;
  sow_code: string;
  weaging_date: string;
  quantity: number;
  total_weight: number | null;
  destination: string | null;
  note: string | null;
}

export interface WeagingPage {
  items: Weaging[];
  total: number;
  page: number;
  page_size: number;
  total_pages: number;
}

export interface CreateWeagingRequest {
  sow_id: string;
  weaging_date: string;
  quantity: number;
  total_weight?: number;
  destination?: string;
  note?: string;
}

export interface UpdateWeagingRequest {
  total_weight?: number;
  clear_total_weight?: boolean;
  destination?: string;
  note?: string;
}

export interface WeagingFilters {
  sow_code?: string;
  from?: string;
  to?: string;
}
