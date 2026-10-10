import { HttpClient, HttpParams } from '@angular/common/http';
import { Injectable, inject } from '@angular/core';
import { Observable } from 'rxjs';

import { environment } from '../../../environments/environment';
import {
  AbortionFilters,
  AbortionPage,
  CreateAbortionRequest,
  UpdateAbortionRequest,
} from './abortion.models';

@Injectable({ providedIn: 'root' })
export class AbortionsService {
  private readonly http = inject(HttpClient);
  private readonly baseUrl = `${environment.apiBaseUrl}/abortions`;

  listAbortions(filters: AbortionFilters = {}, page = 1): Observable<AbortionPage> {
    let params = new HttpParams();
    if (filters.sow_code) {
      params = params.set('sow_code', filters.sow_code);
    }
    params = params.set('page', String(page));

    return this.http.get<AbortionPage>(this.baseUrl, { params });
  }

  createAbortion(request: CreateAbortionRequest): Observable<void> {
    return this.http.post<void>(this.baseUrl, request);
  }

  updateAbortion(id: string, request: UpdateAbortionRequest): Observable<void> {
    return this.http.patch<void>(`${this.baseUrl}/${id}`, request);
  }

  deleteAbortion(id: string): Observable<void> {
    return this.http.delete<void>(`${this.baseUrl}/${id}`);
  }
}
