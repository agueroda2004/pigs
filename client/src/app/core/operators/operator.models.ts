export interface Operator {
  id: string;
  name: string;
  active: boolean;
}

export interface OperatorDropdown {
  id: string;
  name: string;
  active: boolean;
}

export interface CreateOperatorRequest {
  name: string;
}

export interface UpdateOperatorRequest {
  name?: string;
  active?: boolean;
}
