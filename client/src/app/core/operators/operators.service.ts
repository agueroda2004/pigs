import { HttpClient } from '@angular/common/http';
import { Injectable, inject } from '@angular/core';
import { Observable } from 'rxjs';

import { environment } from '../../../environments/environment';
import { CreateOperatorRequest, Operator, UpdateOperatorRequest } from './operator.models';

@Injectable({ providedIn: 'root' })
export class OperatorsService {
  private readonly http = inject(HttpClient);
  private readonly baseUrl = `${environment.apiBaseUrl}/operators`;

  listOperators(): Observable<Operator[]> {
    return this.http.get<Operator[]>(this.baseUrl);
  }

  createOperator(request: CreateOperatorRequest): Observable<Operator> {
    return this.http.post<Operator>(this.baseUrl, request);
  }

  updateOperator(id: string, request: UpdateOperatorRequest): Observable<Operator> {
    return this.http.patch<Operator>(`${this.baseUrl}/${id}`, request);
  }
}
