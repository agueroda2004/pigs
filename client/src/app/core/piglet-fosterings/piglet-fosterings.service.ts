import { HttpClient, HttpParams } from '@angular/common/http';
import { Injectable, inject } from '@angular/core';
import { Observable } from 'rxjs';

import { environment } from '../../../environments/environment';
import {
  CreatePigletFosteringRequest,
  PigletFostering,
  PigletFosteringFilters,
} from './piglet-fostering.models';

@Injectable({ providedIn: 'root' })
export class PigletFosteringsService {
  private readonly http = inject(HttpClient);
  private readonly baseUrl = `${environment.apiBaseUrl}/piglet-fosterings`;

  listPigletFosterings(filters: PigletFosteringFilters = {}): Observable<PigletFostering[]> {
    let params = new HttpParams();
    if (filters.donor_sow_id) {
      params = params.set('donor_sow_id', filters.donor_sow_id);
    }
    if (filters.receiver_sow_id) {
      params = params.set('receiver_sow_id', filters.receiver_sow_id);
    }
    if (filters.from) {
      params = params.set('from', filters.from);
    }
    if (filters.to) {
      params = params.set('to', filters.to);
    }

    return this.http.get<PigletFostering[]>(this.baseUrl, { params });
  }

  createPigletFostering(request: CreatePigletFosteringRequest): Observable<PigletFostering> {
    return this.http.post<PigletFostering>(this.baseUrl, request);
  }
}
