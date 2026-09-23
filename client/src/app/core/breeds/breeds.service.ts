import { HttpClient } from '@angular/common/http';
import { Injectable, inject } from '@angular/core';
import { Observable } from 'rxjs';

import { environment } from '../../../environments/environment';
import { Breed, CreateBreedRequest, UpdateBreedRequest } from './breed.models';

@Injectable({ providedIn: 'root' })
export class BreedsService {
  private readonly http = inject(HttpClient);
  private readonly baseUrl = `${environment.apiBaseUrl}/breeds`;

  listBreeds(): Observable<Breed[]> {
    return this.http.get<Breed[]>(this.baseUrl);
  }

  createBreed(request: CreateBreedRequest): Observable<Breed> {
    return this.http.post<Breed>(this.baseUrl, request);
  }

  updateBreed(id: string, request: UpdateBreedRequest): Observable<Breed> {
    return this.http.patch<Breed>(`${this.baseUrl}/${id}`, request);
  }
}
