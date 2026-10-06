import { HttpClient, HttpParams } from '@angular/common/http';
import { Injectable, inject } from '@angular/core';
import { Observable } from 'rxjs';

import { environment } from '../../../environments/environment';
import {
  Boar,
  BoarDropdown,
  BoarFilters,
  BoarState,
  CreateBoarRequest,
  UpdateBoarRequest,
} from './boar.models';

@Injectable({ providedIn: 'root' })
export class BoarsService {
  private readonly http = inject(HttpClient);
  private readonly baseUrl = `${environment.apiBaseUrl}/boars`;

  listBoars(filters: BoarFilters = {}): Observable<Boar[]> {
    let params = new HttpParams();
    if (filters.code) {
      params = params.set('code', filters.code);
    }
    if (filters.breed_id) {
      params = params.set('breed_id', filters.breed_id);
    }
    if (filters.origin) {
      params = params.set('origin', filters.origin);
    }
    if (filters.active !== undefined) {
      params = params.set('active', String(filters.active));
    }

    return this.http.get<Boar[]>(this.baseUrl, { params });
  }

  createBoar(request: CreateBoarRequest): Observable<void> {
    return this.http.post<void>(this.baseUrl, request);
  }

  listBoarDropdown(active: boolean, state?: BoarState): Observable<BoarDropdown[]> {
    let params = new HttpParams().set('active', String(active));
    if (state) {
      params = params.set('state', state);
    }

    return this.http.get<BoarDropdown[]>(`${this.baseUrl}/dropdown`, { params });
  }

  updateBoar(id: string, request: UpdateBoarRequest): Observable<void> {
    return this.http.patch<void>(`${this.baseUrl}/${id}`, request);
  }

  deleteBoar(id: string): Observable<void> {
    return this.http.delete<void>(`${this.baseUrl}/${id}`);
  }
}
