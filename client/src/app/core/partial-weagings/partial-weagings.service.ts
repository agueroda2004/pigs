import { HttpClient, HttpParams } from '@angular/common/http';
import { Injectable, inject } from '@angular/core';
import { Observable } from 'rxjs';

import { environment } from '../../../environments/environment';
import {
  CreatePartialWeagingRequest,
  PartialWeagingFilters,
  PartialWeagingPage,
  UpdatePartialWeagingRequest,
} from './partial-weaging.models';

@Injectable({ providedIn: 'root' })
export class PartialWeagingsService {
  private readonly http = inject(HttpClient);
  private readonly baseUrl = `${environment.apiBaseUrl}/partial-weagings`;

  listPartialWeagings(
    filters: PartialWeagingFilters = {},
    page = 1,
  ): Observable<PartialWeagingPage> {
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

    return this.http.get<PartialWeagingPage>(this.baseUrl, { params });
  }

  createPartialWeaging(request: CreatePartialWeagingRequest): Observable<void> {
    return this.http.post<void>(this.baseUrl, request);
  }

  updatePartialWeaging(id: string, request: UpdatePartialWeagingRequest): Observable<void> {
    return this.http.patch<void>(`${this.baseUrl}/${id}`, request);
  }

  deletePartialWeaging(id: string): Observable<void> {
    return this.http.delete<void>(`${this.baseUrl}/${id}`);
  }
}
