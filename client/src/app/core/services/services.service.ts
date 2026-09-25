import { HttpClient, HttpParams } from '@angular/common/http';
import { Injectable, inject } from '@angular/core';
import { Observable } from 'rxjs';

import { environment } from '../../../environments/environment';
import { CreateServiceRequest, Service, ServiceFilters } from './service.models';

@Injectable({ providedIn: 'root' })
export class ServicesService {
  private readonly http = inject(HttpClient);
  private readonly baseUrl = `${environment.apiBaseUrl}/services`;

  listServices(filters: ServiceFilters = {}): Observable<Service[]> {
    let params = new HttpParams();
    if (filters.sow_id) {
      params = params.set('sow_id', filters.sow_id);
    }
    if (filters.state) {
      params = params.set('state', filters.state);
    }

    return this.http.get<Service[]>(this.baseUrl, { params });
  }

  createService(request: CreateServiceRequest): Observable<Service> {
    return this.http.post<Service>(this.baseUrl, request);
  }
}
