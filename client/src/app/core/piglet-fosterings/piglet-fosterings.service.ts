import { HttpClient, HttpParams } from '@angular/common/http';
import { Injectable, inject } from '@angular/core';
import { Observable } from 'rxjs';

import { environment } from '../../../environments/environment';
import {
  CreatePigletFosteringRequest,
  PigletFosteringFilters,
  PigletFosteringPage,
  UpdatePigletFosteringRequest,
} from './piglet-fostering.models';

@Injectable({ providedIn: 'root' })
export class PigletFosteringsService {
  private readonly http = inject(HttpClient);
  private readonly baseUrl = `${environment.apiBaseUrl}/piglet-fosterings`;

  listPigletFosterings(
    filters: PigletFosteringFilters = {},
    page = 1,
  ): Observable<PigletFosteringPage> {
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

    return this.http.get<PigletFosteringPage>(this.baseUrl, { params });
  }

  // createPigletFostering registers a fostering and resolves once the server
  // answers with its status code, since the endpoint returns no body.
  createPigletFostering(request: CreatePigletFosteringRequest): Observable<void> {
    return this.http.post<void>(this.baseUrl, request);
  }

  // updatePigletFostering applies the editable fields of a fostering and resolves
  // once the server answers with its status code, since the endpoint returns no body.
  updatePigletFostering(id: string, request: UpdatePigletFosteringRequest): Observable<void> {
    return this.http.patch<void>(`${this.baseUrl}/${id}`, request);
  }

  // deletePigletFostering removes a fostering and resolves once the server answers
  // with its status code, since the endpoint returns no body.
  deletePigletFostering(id: string): Observable<void> {
    return this.http.delete<void>(`${this.baseUrl}/${id}`);
  }
}
