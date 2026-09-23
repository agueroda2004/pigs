import { HttpClient, HttpParams } from '@angular/common/http';
import { Injectable, inject } from '@angular/core';
import { Observable } from 'rxjs';

import { environment } from '../../../environments/environment';
import { Boar, BoarFilters, CreateBoarRequest, UpdateBoarRequest } from './boar.models';

@Injectable({ providedIn: 'root' })
export class BoarsService {
  private readonly http = inject(HttpClient);
  private readonly baseUrl = `${environment.apiBaseUrl}/boars`;

  listBoars(filters: BoarFilters = {}): Observable<Boar[]> {
    let params = new HttpParams();
    if (filters.code) {
      params = params.set('code', filters.code);
    }
    if (filters.breed_id) {
      params = params.set('breed_id', filters.breed_id);
    }
    if (filters.origin) {
      params = params.set('origin', filters.origin);
    }
    if (filters.active !== undefined) {
      params = params.set('active', String(filters.active));
    }

    return this.http.get<Boar[]>(this.baseUrl, { params });
  }

  createBoar(request: CreateBoarRequest): Observable<Boar> {
    return this.http.post<Boar>(this.baseUrl, request);
  }

  updateBoar(id: string, request: UpdateBoarRequest): Observable<Boar> {
    return this.http.patch<Boar>(`${this.baseUrl}/${id}`, request);
  }
}
