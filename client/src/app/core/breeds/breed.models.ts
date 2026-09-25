export interface Breed {
  id: string;
  name: string;
  active: boolean;
  created_at: string;
  updated_at: string;
  created_by: string;
  updated_by: string;
}

export interface BreedOption {
  id: string;
  name: string;
}

export interface CreateBreedRequest {
  name: string;
}

export interface UpdateBreedRequest {
  name?: string;
  active?: boolean;
}
