import { UserRole } from '../auth/auth.models';

export interface User {
  id: string;
  name: string;
  username: string;
  role: UserRole;
  active: boolean;
  created_at: string;
  updated_at: string;
  created_by?: string;
  updated_by?: string;
}

export interface CreateUserRequest {
  name: string;
  username: string;
  password: string;
  role: UserRole;
}

export interface UpdateUserRequest {
  name?: string;
  username?: string;
  password?: string;
  role?: UserRole;
  active?: boolean;
}
