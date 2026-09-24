import { HttpClient, HttpParams } from '@angular/common/http';
import { Injectable, inject } from '@angular/core';
import { Observable } from 'rxjs';

import { environment } from '../../../environments/environment';
import { CreateSowRequest, Sow, SowFilters, UpdateSowRequest } from './sow.models';

@Injectable({ providedIn: 'root' })
export class SowsService {
  private readonly http = inject(HttpClient);
  private readonly baseUrl = `${environment.apiBaseUrl}/sows`;

  listSows(filters: SowFilters = {}): Observable<Sow[]> {
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

    return this.http.get<Sow[]>(this.baseUrl, { params });
  }

  createSow(request: CreateSowRequest): Observable<Sow> {
    return this.http.post<Sow>(this.baseUrl, request);
  }

  updateSow(id: string, request: UpdateSowRequest): Observable<Sow> {
    return this.http.patch<Sow>(`${this.baseUrl}/${id}`, request);
  }
}
