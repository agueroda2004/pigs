import { HttpClient, HttpParams } from '@angular/common/http';
import { Injectable, inject } from '@angular/core';
import { Observable } from 'rxjs';

import { environment } from '../../../environments/environment';
import { CreateFarrowingRequest, Farrowing, FarrowingFilters } from './farrowing.models';

@Injectable({ providedIn: 'root' })
export class FarrowingsService {
  private readonly http = inject(HttpClient);
  private readonly baseUrl = `${environment.apiBaseUrl}/farrowings`;

  listFarrowings(filters: FarrowingFilters = {}): Observable<Farrowing[]> {
    let params = new HttpParams();
    if (filters.sow_id) {
      params = params.set('sow_id', filters.sow_id);
    }
    if (filters.service_id) {
      params = params.set('service_id', filters.service_id);
    }
    if (filters.from) {
      params = params.set('from', filters.from);
    }
    if (filters.to) {
      params = params.set('to', filters.to);
    }

    return this.http.get<Farrowing[]>(this.baseUrl, { params });
  }

  createFarrowing(request: CreateFarrowingRequest): Observable<Farrowing> {
    return this.http.post<Farrowing>(this.baseUrl, request);
  }
}
