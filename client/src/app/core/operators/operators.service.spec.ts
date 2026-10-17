import { provideHttpClient } from '@angular/common/http';
import { HttpTestingController, provideHttpClientTesting } from '@angular/common/http/testing';
import { TestBed } from '@angular/core/testing';
import { firstValueFrom } from 'rxjs';

import { OperatorsService } from './operators.service';

const operatorPayload = {
  id: '1',
  name: 'Juan Pérez',
  active: true,
};

describe('OperatorsService', () => {
  let service: OperatorsService;
  let http: HttpTestingController;

  beforeEach(() => {
    TestBed.configureTestingModule({
      providers: [provideHttpClient(), provideHttpClientTesting()],
    });
    service = TestBed.inject(OperatorsService);
    http = TestBed.inject(HttpTestingController);
  });

  afterEach(() => http.verify());

  it('lists operators', async () => {
    const promise = firstValueFrom(service.listOperators());

    const call = http.expectOne(
      (request) => request.url.endsWith('/operators') && request.method === 'GET',
    );
    call.flush([operatorPayload]);

    await expect(promise).resolves.toHaveLength(1);
  });

  it('lists the operator dropdown with the active filter', async () => {
    const promise = firstValueFrom(service.listOperatorDropdown(true));

    const call = http.expectOne(
      (request) => request.url.endsWith('/operators/dropdown') && request.method === 'GET',
    );
    expect(call.request.params.get('active')).toBe('true');
    call.flush([{ id: '1', name: 'Juan Pérez', active: true }]);

    await expect(promise).resolves.toHaveLength(1);
  });

  it('lists the operator dropdown without an active filter', async () => {
    const promise = firstValueFrom(service.listOperatorDropdown());

    const call = http.expectOne(
      (request) => request.url.endsWith('/operators/dropdown') && request.method === 'GET',
    );
    expect(call.request.params.has('active')).toBe(false);
    call.flush([]);

    await expect(promise).resolves.toHaveLength(0);
  });

  it('creates an operator', async () => {
    const promise = firstValueFrom(service.createOperator({ name: 'Juan Pérez' }));

    const call = http.expectOne(
      (request) => request.url.endsWith('/operators') && request.method === 'POST',
    );
    expect(call.request.body).toEqual({ name: 'Juan Pérez' });
    call.flush(null);

    await expect(promise).resolves.toBeNull();
  });

  it('updates an operator', async () => {
    const payload = { name: 'María López', active: false };
    const promise = firstValueFrom(service.updateOperator('7', payload));

    const call = http.expectOne(
      (request) => request.url.endsWith('/operators/7') && request.method === 'PATCH',
    );
    expect(call.request.body).toEqual(payload);
    call.flush(null);

    await expect(promise).resolves.toBeNull();
  });

  it('deletes an operator', async () => {
    const promise = firstValueFrom(service.deleteOperator('7'));

    const call = http.expectOne(
      (request) => request.url.endsWith('/operators/7') && request.method === 'DELETE',
    );
    call.flush(null);

    await expect(promise).resolves.toBeNull();
  });
});
