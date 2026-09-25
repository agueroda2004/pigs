import { HttpClient } from '@angular/common/http';
import { Injectable, computed, inject, signal } from '@angular/core';
import { firstValueFrom } from 'rxjs';

import { environment } from '../../../environments/environment';
import { AuthUser, LoginCredentials } from './auth.models';

@Injectable({ providedIn: 'root' })
export class AuthService {
  private readonly http = inject(HttpClient);
  private readonly baseUrl = `${environment.apiBaseUrl}/auth`;

  private readonly userSignal = signal<AuthUser | null>(null);
  private refreshRequest: Promise<void> | null = null;

  readonly user = this.userSignal.asReadonly();
  readonly isAuthenticated = computed(() => this.userSignal() !== null);
  readonly isAdmin = computed(() => this.userSignal()?.role === 'Admin');

  async login(credentials: LoginCredentials): Promise<void> {
    await firstValueFrom(this.http.post<void>(`${this.baseUrl}/login`, credentials));
    await this.loadCurrentUser();
  }

  async logout(): Promise<void> {
    try {
      await firstValueFrom(this.http.post<void>(`${this.baseUrl}/logout`, {}));
    } finally {
      this.userSignal.set(null);
    }
  }

  refresh(): Promise<void> {
    if (!this.refreshRequest) {
      this.refreshRequest = firstValueFrom(
        this.http.post<void>(`${this.baseUrl}/refresh`, {}),
      ).finally(() => {
        this.refreshRequest = null;
      });
    }
    return this.refreshRequest;
  }

  async loadCurrentUser(): Promise<AuthUser | null> {
    try {
      const user = await firstValueFrom(this.http.get<AuthUser>(`${this.baseUrl}/me`));
      this.userSignal.set(user);
      return user;
    } catch {
      this.userSignal.set(null);
      return null;
    }
  }

  clear(): void {
    this.userSignal.set(null);
  }
}
