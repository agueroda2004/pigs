export interface Medication {
  id: string;
  name: string;
  active: boolean;
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
