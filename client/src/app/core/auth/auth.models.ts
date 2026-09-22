export type UserRole = 'User' | 'Admin';

export interface AuthUser {
  id: string;
  username: string;
  role: UserRole;
}

export interface LoginCredentials {
  username: string;
  password: string;
}
