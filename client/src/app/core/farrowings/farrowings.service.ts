import { HttpClient, HttpParams } from '@angular/common/http';
import { Injectable, inject } from '@angular/core';
import { Observable } from 'rxjs';

import { environment } from '../../../environments/environment';
import {
  CreateFarrowingRequest,
  FarrowingFilters,
  FarrowingPage,
  UpdateFarrowingRequest,
} from './farrowing.models';

@Injectable({ providedIn: 'root' })
export class FarrowingsService {
  private readonly http = inject(HttpClient);
  private readonly baseUrl = `${environment.apiBaseUrl}/farrowings`;

  listFarrowings(filters: FarrowingFilters = {}, page = 1): Observable<FarrowingPage> {
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

    return this.http.get<FarrowingPage>(this.baseUrl, { params });
  }

  createFarrowing(request: CreateFarrowingRequest): Observable<void> {
    return this.http.post<void>(this.baseUrl, request);
  }

  updateFarrowing(id: string, request: UpdateFarrowingRequest): Observable<void> {
    return this.http.patch<void>(`${this.baseUrl}/${id}`, request);
  }

  deleteFarrowing(id: string): Observable<void> {
    return this.http.delete<void>(`${this.baseUrl}/${id}`);
  }
}
