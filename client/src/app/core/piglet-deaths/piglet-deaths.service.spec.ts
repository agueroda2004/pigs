import { provideHttpClient } from '@angular/common/http';
import { HttpTestingController, provideHttpClientTesting } from '@angular/common/http/testing';
import { TestBed } from '@angular/core/testing';
import { firstValueFrom } from 'rxjs';

import { PigletDeath, PigletDeathPage } from './piglet-death.models';
import { PigletDeathsService } from './piglet-deaths.service';

const deathPayload: PigletDeath = {
  id: '1',
  farrowing_id: 'farrowing-1',
  sow_id: 'sow-1',
  sow_code: 'C-001',
  operator_id: 'operator-1',
  operator_name: 'Operador',
  death_date: '2026-04-25',
  quantity: 2,
  weight: 2.5,
  cause: 'Aplastado',
  turn: 'Mañana',
  note: null,
};

const pagePayload: PigletDeathPage = {
  items: [deathPayload],
  total: 1,
  page: 1,
  page_size: 10,
  total_pages: 1,
};

describe('PigletDeathsService', () => {
  let service: PigletDeathsService;
  let http: HttpTestingController;

  beforeEach(() => {
    TestBed.configureTestingModule({
      providers: [provideHttpClient(), provideHttpClientTesting()],
    });
    service = TestBed.inject(PigletDeathsService);
    http = TestBed.inject(HttpTestingController);
  });

  afterEach(() => http.verify());

  it('lists the first page of piglet deaths', async () => {
    const promise = firstValueFrom(service.listPigletDeaths());

    const call = http.expectOne(
      (request) => request.url.endsWith('/piglet-deaths') && request.method === 'GET',
    );
    expect(call.request.params.get('page')).toBe('1');
    call.flush(pagePayload);

    await expect(promise).resolves.toMatchObject({ total: 1, total_pages: 1 });
  });

  it('lists piglet deaths with the sow code, date range and page filters', async () => {
    const promise = firstValueFrom(
      service.listPigletDeaths({ sow_code: 'C-001', from: '2026-04-01', to: '2026-04-30' }, 2),
    );

    const call = http.expectOne(
      (request) => request.url.endsWith('/piglet-deaths') && request.method === 'GET',
    );
    expect(call.request.params.get('sow_code')).toBe('C-001');
    expect(call.request.params.get('from')).toBe('2026-04-01');
    expect(call.request.params.get('to')).toBe('2026-04-30');
    expect(call.request.params.get('page')).toBe('2');
    call.flush(pagePayload);

    await expect(promise).resolves.toHaveProperty('items');
  });

  it('creates a piglet death', async () => {
    const payload = {
      sow_id: 'sow-1',
      operator_id: 'operator-1',
      death_date: '2026-04-25',
      quantity: 2,
      weight: 2.5,
      cause: 'Aplastado' as const,
      turn: 'Mañana' as const,
    };
    const promise = firstValueFrom(service.createPigletDeath(payload));

    const call = http.expectOne(
      (request) => request.url.endsWith('/piglet-deaths') && request.method === 'POST',
    );
    expect(call.request.body).toEqual(payload);
    call.flush(null, { status: 201, statusText: 'Created' });

    await expect(promise).resolves.toBeNull();
  });

  it('deletes a piglet death', async () => {
    const promise = firstValueFrom(service.deletePigletDeath('death-1'));

    const call = http.expectOne(
      (request) => request.url.endsWith('/piglet-deaths/death-1') && request.method === 'DELETE',
    );
    call.flush(null, { status: 204, statusText: 'No Content' });

    await expect(promise).resolves.toBeNull();
  });

  it('updates a piglet death', async () => {
    const payload = {
      operator_id: 'operator-1',
      weight: 3.5,
      cause: 'Diarrea' as const,
      turn: 'Tarde' as const,
      note: 'nueva nota',
    };
    const promise = firstValueFrom(service.updatePigletDeath('death-1', payload));

    const call = http.expectOne(
      (request) => request.url.endsWith('/piglet-deaths/death-1') && request.method === 'PATCH',
    );
    expect(call.request.body).toEqual(payload);
    call.flush(null, { status: 200, statusText: 'OK' });

    await expect(promise).resolves.toBeNull();
  });
});
