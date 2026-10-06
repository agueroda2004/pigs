import { HttpClient, HttpParams } from '@angular/common/http';
import { Injectable, inject } from '@angular/core';
import { Observable } from 'rxjs';

import { environment } from '../../../environments/environment';
import { CreateWeagingRequest, Weaging, WeagingFilters } from './weaging.models';

@Injectable({ providedIn: 'root' })
export class WeagingsService {
  private readonly http = inject(HttpClient);
  private readonly baseUrl = `${environment.apiBaseUrl}/weagings`;

  listWeagings(filters: WeagingFilters = {}): Observable<Weaging[]> {
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

    return this.http.get<Weaging[]>(this.baseUrl, { params });
  }

  createWeaging(request: CreateWeagingRequest): Observable<Weaging> {
    return this.http.post<Weaging>(this.baseUrl, request);
  }
}
