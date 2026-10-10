import { HttpClient, HttpParams } from '@angular/common/http';
import { Injectable, inject } from '@angular/core';
import { Observable } from 'rxjs';

import { environment } from '../../../environments/environment';
import {
  CreateSowRequest,
  SowDropdown,
  SowFilters,
  SowPage,
  SowState,
  UpdateSowRequest,
} from './sow.models';

@Injectable({ providedIn: 'root' })
export class SowsService {
  private readonly http = inject(HttpClient);
  private readonly baseUrl = `${environment.apiBaseUrl}/sows`;

  listSows(filters: SowFilters = {}, page = 1): Observable<SowPage> {
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
    if (filters.state) {
      params = params.set('state', filters.state);
    }
    params = params.set('page', String(page));

    return this.http.get<SowPage>(this.baseUrl, { params });
  }

  listSowDropdown(active?: boolean, states?: SowState[]): Observable<SowDropdown[]> {
    let params = new HttpParams();
    if (active !== undefined) {
      params = params.set('active', String(active));
    }
    for (const state of states ?? []) {
      params = params.append('states', state);
    }

    return this.http.get<SowDropdown[]>(`${this.baseUrl}/dropdown`, { params });
  }

  createSow(request: CreateSowRequest): Observable<void> {
    return this.http.post<void>(this.baseUrl, request);
  }

  updateSow(id: string, request: UpdateSowRequest): Observable<void> {
    return this.http.patch<void>(`${this.baseUrl}/${id}`, request);
  }

  deleteSow(id: string): Observable<void> {
    return this.http.delete<void>(`${this.baseUrl}/${id}`);
  }
}
