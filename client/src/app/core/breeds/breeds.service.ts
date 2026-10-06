import { HttpClient, HttpParams } from '@angular/common/http';
import { Injectable, inject } from '@angular/core';
import { Observable } from 'rxjs';

import { environment } from '../../../environments/environment';
import { Breed, BreedFilters, CreateBreedRequest, UpdateBreedRequest } from './breed.models';

@Injectable({ providedIn: 'root' })
export class BreedsService {
  private readonly http = inject(HttpClient);
  private readonly baseUrl = `${environment.apiBaseUrl}/breeds`;

  listBreeds(filters: BreedFilters = {}): Observable<Breed[]> {
    let params = new HttpParams();
    if (filters.name) {
      params = params.set('name', filters.name);
    }
    if (filters.active !== undefined) {
      params = params.set('active', String(filters.active));
    }

    return this.http.get<Breed[]>(this.baseUrl, { params });
  }

  listBreedDropdown(active: boolean): Observable<Breed[]> {
    const params = new HttpParams().set('active', String(active));
    return this.http.get<Breed[]>(`${this.baseUrl}/dropdown`, { params });
  }

  createBreed(request: CreateBreedRequest): Observable<void> {
    return this.http.post<void>(this.baseUrl, request);
  }

  updateBreed(id: string, request: UpdateBreedRequest): Observable<void> {
    return this.http.patch<void>(`${this.baseUrl}/${id}`, request);
  }

  deleteBreed(id: string): Observable<void> {
    return this.http.delete<void>(`${this.baseUrl}/${id}`);
  }
}
