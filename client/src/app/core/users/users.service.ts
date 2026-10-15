import { HttpClient } from '@angular/common/http';
import { Injectable, inject } from '@angular/core';
import { Observable } from 'rxjs';

import { environment } from '../../../environments/environment';
import { CreateUserRequest, UpdateUserRequest, User } from './user.models';

@Injectable({ providedIn: 'root' })
export class UsersService {
  private readonly http = inject(HttpClient);
  private readonly baseUrl = `${environment.apiBaseUrl}/users`;
  private readonly adminBaseUrl = `${environment.apiBaseUrl}/admin/users`;

  createUser(request: CreateUserRequest): Observable<void> {
    return this.http.post<void>(this.baseUrl, request);
  }

  listUsers(): Observable<User[]> {
    return this.http.get<User[]>(this.adminBaseUrl);
  }

  updateUser(id: string, request: UpdateUserRequest): Observable<void> {
    return this.http.patch<void>(`${this.adminBaseUrl}/${id}`, request);
  }

  deleteUser(id: string): Observable<void> {
    return this.http.delete<void>(`${this.adminBaseUrl}/${id}`);
  }
}
