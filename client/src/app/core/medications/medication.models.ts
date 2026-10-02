export interface Medication {
  id: string;
  name: string;
  active: boolean;
  created_at: string;
  updated_at: string;
  created_by: string;
  updated_by: string;
}

export interface MedicationOption {
  id: string;
  name: string;
}

export interface MedicationFilters {
  name?: string;
  active?: boolean;
}

export interface CreateMedicationRequest {
  name: string;
}

export interface UpdateMedicationRequest {
  name?: string;
  active?: boolean;
}
