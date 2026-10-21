export interface FarrowingOperatorLink {
  id: string;
  operator_id: string;
}

export interface FarrowingMedicationLink {
  id: string;
  medication_id: string;
  dose: number;
  applied_by: string;
}

export interface Farrowing {
  id: string;
  sow_id: string;
  sow_code: string;
  service_id: string;
  farrow_date: string;
  start_time: string | null;
  end_time: string | null;
  location: string | null;
  live_born: number;
  stillborn: number;
  mummified: number;
  current_piglets: number;
  litter_weight: number | null;
  stillborn_weight: number | null;
  is_manipulated: boolean;
  is_nurse: boolean;
  nurse_start_date: string | null;
  note: string | null;
  operators: FarrowingOperatorLink[];
  medications: FarrowingMedicationLink[];
}

export interface FarrowingPage {
  items: Farrowing[];
  total: number;
  page: number;
  page_size: number;
  total_pages: number;
}

export interface CreateFarrowingOperatorRequest {
  operator_id: string;
}

export interface CreateFarrowingMedicationRequest {
  medication_id: string;
  dose: number;
  applied_by: string;
}

export interface CreateFarrowingRequest {
  sow_id: string;
  farrow_date: string;
  start_time?: string;
  end_time?: string;
  location?: string;
  live_born: number;
  stillborn: number;
  mummified: number;
  litter_weight: number;
  stillborn_weight: number;
  is_manipulated: boolean;
  note?: string;
  operators: CreateFarrowingOperatorRequest[];
  medications: CreateFarrowingMedicationRequest[];
}

export interface UpdateFarrowingRequest {
  start_time?: string;
  end_time?: string;
  location?: string;
  litter_weight: number;
  stillborn_weight: number;
  is_manipulated?: boolean;
  note?: string;
  operators?: CreateFarrowingOperatorRequest[];
  medications?: CreateFarrowingMedicationRequest[];
}

export interface FarrowingFilters {
  sow_code?: string;
  from?: string;
  to?: string;
}
