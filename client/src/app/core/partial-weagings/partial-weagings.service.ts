import { HttpClient, HttpParams } from '@angular/common/http';
import { Injectable, inject } from '@angular/core';
import { Observable } from 'rxjs';

import { environment } from '../../../environments/environment';
import {
  CreatePartialWeagingRequest,
  PartialWeaging,
  PartialWeagingFilters,
} from './partial-weaging.models';

@Injectable({ providedIn: 'root' })
export class PartialWeagingsService {
  private readonly http = inject(HttpClient);
  private readonly baseUrl = `${environment.apiBaseUrl}/partial-weagings`;

  listPartialWeagings(filters: PartialWeagingFilters = {}): Observable<PartialWeaging[]> {
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

    return this.http.get<PartialWeaging[]>(this.baseUrl, { params });
  }

  createPartialWeaging(request: CreatePartialWeagingRequest): Observable<PartialWeaging> {
    return this.http.post<PartialWeaging>(this.baseUrl, request);
  }
}
