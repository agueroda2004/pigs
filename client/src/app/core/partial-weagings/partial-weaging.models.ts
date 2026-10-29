export type PartialWeagingType = 'Normal' | 'Nodriza' | 'Baja_Viabilidad';

export interface PartialWeaging {
  id: string;
  farrowing_id: string;
  sow_id: string;
  sow_code: string;
  weaging_date: string;
  quantity: number;
  total_weight: number | null;
  type: PartialWeagingType;
  note: string | null;
}

export interface PartialWeagingPage {
  items: PartialWeaging[];
  total: number;
  page: number;
  page_size: number;
  total_pages: number;
}

export interface CreatePartialWeagingRequest {
  sow_id: string;
  weaging_date: string;
  quantity: number;
  total_weight?: number;
  type: PartialWeagingType;
  note?: string;
}

export interface UpdatePartialWeagingRequest {
  total_weight?: number;
  clear_total_weight?: boolean;
  note?: string;
}

export interface PartialWeagingFilters {
  sow_code?: string;
  from?: string;
  to?: string;
}
