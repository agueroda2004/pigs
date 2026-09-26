import { HttpClient, HttpParams } from '@angular/common/http';
import { Injectable, inject } from '@angular/core';
import { Observable } from 'rxjs';

import { environment } from '../../../environments/environment';
import { Abortion, AbortionFilters, CreateAbortionRequest } from './abortion.models';

@Injectable({ providedIn: 'root' })
export class AbortionsService {
  private readonly http = inject(HttpClient);
  private readonly baseUrl = `${environment.apiBaseUrl}/abortions`;

  listAbortions(filters: AbortionFilters = {}): Observable<Abortion[]> {
    let params = new HttpParams();
    if (filters.sow_id) {
      params = params.set('sow_id', filters.sow_id);
    }

    return this.http.get<Abortion[]>(this.baseUrl, { params });
  }

  createAbortion(request: CreateAbortionRequest): Observable<Abortion> {
    return this.http.post<Abortion>(this.baseUrl, request);
  }
}
