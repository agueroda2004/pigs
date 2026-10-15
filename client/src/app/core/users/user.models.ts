import { UserRole } from '../auth/auth.models';

export interface User {
  id: string;
  name: string;
  username: string;
  role: UserRole;
  active: boolean;
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
