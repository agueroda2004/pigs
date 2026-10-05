import { provideHttpClient } from '@angular/common/http';
import { HttpTestingController, provideHttpClientTesting } from '@angular/common/http/testing';
import { TestBed } from '@angular/core/testing';
import { firstValueFrom } from 'rxjs';

import { PigletDeathsService } from './piglet-deaths.service';

const deathPayload = {
  id: '1',
  farrowing_id: 'farrowing-1',
  sow_id: 'sow-1',
  operator_id: 'operator-1',
  operator_name: 'Operador',
  death_date: '2026-04-25',
  quantity: 2,
  weight: 2.5,
  cause: 'Aplastado',
  turn: 'Mañana',
  note: null,
  created_at: '',
  updated_at: '',
  created_by: 'admin',
  updated_by: 'admin',
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

  it('lists piglet deaths', async () => {
    const promise = firstValueFrom(service.listPigletDeaths());

    const call = http.expectOne(
      (request) => request.url.endsWith('/piglet-deaths') && request.method === 'GET',
    );
    expect(call.request.params.keys()).toHaveLength(0);
    call.flush([deathPayload]);

    await expect(promise).resolves.toHaveLength(1);
  });

  it('lists piglet deaths with the sow and date range filters', async () => {
    const promise = firstValueFrom(
      service.listPigletDeaths({ sow_id: 'sow-1', from: '2026-04-01', to: '2026-04-30' }),
    );

    const call = http.expectOne(
      (request) => request.url.endsWith('/piglet-deaths') && request.method === 'GET',
    );
    expect(call.request.params.get('sow_id')).toBe('sow-1');
    expect(call.request.params.get('from')).toBe('2026-04-01');
    expect(call.request.params.get('to')).toBe('2026-04-30');
    call.flush([deathPayload]);

    await expect(promise).resolves.toHaveLength(1);
  });

  it('creates a piglet death', async () => {
    const payload = {
      sow_id: 'sow-1',
      operator_id: 'operator-1',
      death_date: '2026-04-25',
      quantity: 2,
      cause: 'Aplastado' as const,
      turn: 'Mañana' as const,
    };
    const promise = firstValueFrom(service.createPigletDeath(payload));

    const call = http.expectOne(
      (request) => request.url.endsWith('/piglet-deaths') && request.method === 'POST',
    );
    expect(call.request.body).toEqual(payload);
    call.flush(deathPayload);

    await expect(promise).resolves.toMatchObject({ operator_name: 'Operador' });
  });
});
