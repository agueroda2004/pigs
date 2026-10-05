import { HttpClient, HttpParams } from '@angular/common/http';
import { Injectable, inject } from '@angular/core';
import { Observable } from 'rxjs';

import { environment } from '../../../environments/environment';
import {
  CreatePigletDeathRequest,
  PigletDeath,
  PigletDeathFilters,
} from './piglet-death.models';

@Injectable({ providedIn: 'root' })
export class PigletDeathsService {
  private readonly http = inject(HttpClient);
  private readonly baseUrl = `${environment.apiBaseUrl}/piglet-deaths`;

  listPigletDeaths(filters: PigletDeathFilters = {}): Observable<PigletDeath[]> {
    let params = new HttpParams();
    if (filters.sow_id) {
      params = params.set('sow_id', filters.sow_id);
    }
    if (filters.from) {
      params = params.set('from', filters.from);
    }
    if (filters.to) {
      params = params.set('to', filters.to);
    }

    return this.http.get<PigletDeath[]>(this.baseUrl, { params });
  }

  createPigletDeath(request: CreatePigletDeathRequest): Observable<PigletDeath> {
    return this.http.post<PigletDeath>(this.baseUrl, request);
  }
}
