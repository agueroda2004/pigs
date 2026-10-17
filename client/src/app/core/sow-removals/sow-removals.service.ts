import { HttpClient, HttpParams } from '@angular/common/http';
import { Injectable, inject } from '@angular/core';
import { Observable } from 'rxjs';

import { environment } from '../../../environments/environment';
import {
  CreateSowRemovalRequest,
  SowRemoval,
  SowRemovalFilters,
  UpdateSowRemovalRequest,
} from './sow-removal.models';

@Injectable({ providedIn: 'root' })
export class SowRemovalsService {
  private readonly http = inject(HttpClient);
  private readonly baseUrl = `${environment.apiBaseUrl}/sow-removals`;

  listSowRemovals(filters: SowRemovalFilters = {}): Observable<SowRemoval[]> {
    let params = new HttpParams();
    if (filters.sow_code) {
      params = params.set('sow_code', filters.sow_code);
    }

    return this.http.get<SowRemoval[]>(this.baseUrl, { params });
  }

  createSowRemoval(request: CreateSowRemovalRequest): Observable<SowRemoval> {
    return this.http.post<SowRemoval>(this.baseUrl, request);
  }

  updateSowRemoval(id: string, request: UpdateSowRemovalRequest): Observable<void> {
    return this.http.patch<void>(`${this.baseUrl}/${id}`, request);
  }

  deleteSowRemoval(id: string): Observable<void> {
    return this.http.delete<void>(`${this.baseUrl}/${id}`);
  }
}
