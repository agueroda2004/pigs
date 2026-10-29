import { HttpClient, HttpParams } from '@angular/common/http';
import { Injectable, inject } from '@angular/core';
import { Observable } from 'rxjs';

import { environment } from '../../../environments/environment';
import {
  CreateWeagingRequest,
  UpdateWeagingRequest,
  WeagingFilters,
  WeagingPage,
} from './weaging.models';

@Injectable({ providedIn: 'root' })
export class WeagingsService {
  private readonly http = inject(HttpClient);
  private readonly baseUrl = `${environment.apiBaseUrl}/weagings`;

  listWeagings(filters: WeagingFilters = {}, page = 1): Observable<WeagingPage> {
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

    return this.http.get<WeagingPage>(this.baseUrl, { params });
  }

  createWeaging(request: CreateWeagingRequest): Observable<void> {
    return this.http.post<void>(this.baseUrl, request);
  }

  updateWeaging(id: string, request: UpdateWeagingRequest): Observable<void> {
    return this.http.patch<void>(`${this.baseUrl}/${id}`, request);
  }

  deleteWeaging(id: string): Observable<void> {
    return this.http.delete<void>(`${this.baseUrl}/${id}`);
  }
}
