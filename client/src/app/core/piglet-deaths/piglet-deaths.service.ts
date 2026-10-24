import { HttpClient, HttpParams } from '@angular/common/http';
import { Injectable, inject } from '@angular/core';
import { Observable } from 'rxjs';

import { environment } from '../../../environments/environment';
import {
  CreatePigletDeathRequest,
  PigletDeath,
  PigletDeathFilters,
  PigletDeathPage,
  UpdatePigletDeathRequest,
} from './piglet-death.models';

@Injectable({ providedIn: 'root' })
export class PigletDeathsService {
  private readonly http = inject(HttpClient);
  private readonly baseUrl = `${environment.apiBaseUrl}/piglet-deaths`;

  listPigletDeaths(filters: PigletDeathFilters = {}, page = 1): Observable<PigletDeathPage> {
    let params = new HttpParams();
    if (filters.sow_code) {
      params = params.set('sow_code', filters.sow_code);
    }
    if (filters.from) {
      params = params.set('from', filters.from);
    }
    if (filters.to) {
      params = params.set('to', filters.to);
    }
    params = params.set('page', String(page));

    return this.http.get<PigletDeathPage>(this.baseUrl, { params });
  }

  createPigletDeath(request: CreatePigletDeathRequest): Observable<void> {
    return this.http.post<void>(this.baseUrl, request);
  }

  deletePigletDeath(id: string): Observable<void> {
    return this.http.delete<void>(`${this.baseUrl}/${id}`);
  }

  updatePigletDeath(id: string, request: UpdatePigletDeathRequest): Observable<void> {
    return this.http.patch<void>(`${this.baseUrl}/${id}`, request);
  }
}
