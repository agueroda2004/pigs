import { HttpClient, HttpParams } from '@angular/common/http';
import { Injectable, inject } from '@angular/core';
import { Observable } from 'rxjs';

import { environment } from '../../../environments/environment';
import {
  BoarRemoval,
  BoarRemovalFilters,
  CreateBoarRemovalRequest,
  UpdateBoarRemovalRequest,
} from './boar-removal.models';

@Injectable({ providedIn: 'root' })
export class BoarRemovalsService {
  private readonly http = inject(HttpClient);
  private readonly baseUrl = `${environment.apiBaseUrl}/boar-removals`;

  listBoarRemovals(filters: BoarRemovalFilters = {}): Observable<BoarRemoval[]> {
    let params = new HttpParams();
    if (filters.boar_id) {
      params = params.set('boar_id', filters.boar_id);
    }

    return this.http.get<BoarRemoval[]>(this.baseUrl, { params });
  }

  createBoarRemoval(request: CreateBoarRemovalRequest): Observable<void> {
    return this.http.post<void>(this.baseUrl, request);
  }

  updateBoarRemoval(id: string, request: UpdateBoarRemovalRequest): Observable<void> {
    return this.http.patch<void>(`${this.baseUrl}/${id}`, request);
  }

  deleteBoarRemoval(id: string): Observable<void> {
    return this.http.delete<void>(`${this.baseUrl}/${id}`);
  }
}
