export interface Breed {
  id: string;
  name: string;
  active: boolean;
}

export interface BreedFilters {
  name?: string;
  active?: boolean;
}

export interface CreateBreedRequest {
  name: string;
}

export interface UpdateBreedRequest {
  name?: string;
  active?: boolean;
}
