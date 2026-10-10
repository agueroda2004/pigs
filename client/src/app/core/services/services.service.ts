import { HttpClient, HttpParams } from '@angular/common/http';
import { Injectable, inject } from '@angular/core';
import { Observable } from 'rxjs';

import { environment } from '../../../environments/environment';
import {
  CreateServiceRequest,
  ServiceFilters,
  ServicePage,
  UpdateServiceRequest,
} from './service.models';

@Injectable({ providedIn: 'root' })
export class ServicesService {
  private readonly http = inject(HttpClient);
  private readonly baseUrl = `${environment.apiBaseUrl}/services`;

  listServices(filters: ServiceFilters = {}, page = 1): Observable<ServicePage> {
    let params = new HttpParams();
    if (filters.sow_code) {
      params = params.set('sow_code', filters.sow_code);
    }
    if (filters.state) {
      params = params.set('state', filters.state);
    }
    params = params.set('page', String(page));

    return this.http.get<ServicePage>(this.baseUrl, { params });
  }

  createService(request: CreateServiceRequest): Observable<void> {
    return this.http.post<void>(this.baseUrl, request);
  }

  updateService(id: string, request: UpdateServiceRequest): Observable<void> {
    return this.http.patch<void>(`${this.baseUrl}/${id}`, request);
  }

  deleteService(id: string): Observable<void> {
    return this.http.delete<void>(`${this.baseUrl}/${id}`);
  }
}
