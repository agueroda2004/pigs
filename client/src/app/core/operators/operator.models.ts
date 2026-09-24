export interface Operator {
  id: string;
  name: string;
  active: boolean;
  created_at: string;
  updated_at: string;
  created_by: string;
  updated_by: string;
}

export interface CreateOperatorRequest {
  name: string;
}

export interface UpdateOperatorRequest {
  name?: string;
  active?: boolean;
}
