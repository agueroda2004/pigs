import { HttpClient, HttpParams } from '@angular/common/http';
import { Injectable, inject } from '@angular/core';
import { Observable } from 'rxjs';

import { environment } from '../../../environments/environment';
import {
  CreateMedicationRequest,
  Medication,
  MedicationFilters,
  MedicationOption,
  UpdateMedicationRequest,
} from './medication.models';

@Injectable({ providedIn: 'root' })
export class MedicationsService {
  private readonly http = inject(HttpClient);
  private readonly baseUrl = `${environment.apiBaseUrl}/medications`;

  listMedications(filters: MedicationFilters = {}): Observable<Medication[]> {
    let params = new HttpParams();
    if (filters.name) {
      params = params.set('name', filters.name);
    }
    if (filters.active !== undefined) {
      params = params.set('active', String(filters.active));
    }

    return this.http.get<Medication[]>(this.baseUrl, { params });
  }

  listMedicationOptions(includeInactive = false): Observable<MedicationOption[]> {
    const params = new HttpParams().set('include_inactive', String(includeInactive));

    return this.http.get<MedicationOption[]>(`${this.baseUrl}/options`, { params });
  }

  createMedication(request: CreateMedicationRequest): Observable<Medication> {
    return this.http.post<Medication>(this.baseUrl, request);
  }

  updateMedication(id: string, request: UpdateMedicationRequest): Observable<Medication> {
    return this.http.patch<Medication>(`${this.baseUrl}/${id}`, request);
  }
}
