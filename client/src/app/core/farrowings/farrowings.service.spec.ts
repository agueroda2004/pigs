import { provideHttpClient } from '@angular/common/http';
import { HttpTestingController, provideHttpClientTesting } from '@angular/common/http/testing';
import { TestBed } from '@angular/core/testing';
import { firstValueFrom } from 'rxjs';

import { FarrowingsService } from './farrowings.service';

const farrowingPayload = {
  id: '1',
  sow_id: 'sow-1',
  service_id: 'service-1',
  farrow_date: '2026-04-20',
  start_time: '22:00',
  end_time: '02:00',
  location: 'Corral 3',
  live_born: 10,
  stillborn: 1,
  mummified: 0,
  current_piglets: 10,
  litter_weight: 15.5,
  stillborn_weight: null,
  is_manipulated: false,
  is_nurse: false,
  nurse_start_date: null,
  note: null,
  operators: [],
  medications: [],
  created_at: '',
  updated_at: '',
  created_by: 'admin',
  updated_by: 'admin',
};

describe('FarrowingsService', () => {
  let service: FarrowingsService;
  let http: HttpTestingController;

  beforeEach(() => {
    TestBed.configureTestingModule({
      providers: [provideHttpClient(), provideHttpClientTesting()],
    });
    service = TestBed.inject(FarrowingsService);
    http = TestBed.inject(HttpTestingController);
  });

  afterEach(() => http.verify());

  it('lists farrowings', async () => {
    const promise = firstValueFrom(service.listFarrowings());

    const call = http.expectOne(
      (request) => request.url.endsWith('/farrowings') && request.method === 'GET',
    );
    expect(call.request.params.keys()).toHaveLength(0);
    call.flush([farrowingPayload]);

    await expect(promise).resolves.toHaveLength(1);
  });

  it('lists farrowings with the sow and service filters', async () => {
    const promise = firstValueFrom(
      service.listFarrowings({ sow_id: 'sow-1', service_id: 'service-1' }),
    );

    const call = http.expectOne(
      (request) => request.url.endsWith('/farrowings') && request.method === 'GET',
    );
    expect(call.request.params.get('sow_id')).toBe('sow-1');
    expect(call.request.params.get('service_id')).toBe('service-1');
    call.flush([farrowingPayload]);

    await expect(promise).resolves.toHaveLength(1);
  });

  it('lists farrowings with the date range filters', async () => {
    const promise = firstValueFrom(
      service.listFarrowings({ from: '2026-04-01', to: '2026-04-30' }),
    );

    const call = http.expectOne(
      (request) => request.url.endsWith('/farrowings') && request.method === 'GET',
    );
    expect(call.request.params.get('from')).toBe('2026-04-01');
    expect(call.request.params.get('to')).toBe('2026-04-30');
    call.flush([farrowingPayload]);

    await expect(promise).resolves.toHaveLength(1);
  });

  it('creates a farrowing', async () => {
    const payload = {
      sow_id: 'sow-1',
      farrow_date: '2026-04-20',
      live_born: 10,
      stillborn: 1,
      mummified: 0,
      is_manipulated: false,
      operators: [],
      medications: [],
    };
    const promise = firstValueFrom(service.createFarrowing(payload));

    const call = http.expectOne(
      (request) => request.url.endsWith('/farrowings') && request.method === 'POST',
    );
    expect(call.request.body).toEqual(payload);
    call.flush(farrowingPayload);

    await expect(promise).resolves.toMatchObject({ sow_id: 'sow-1' });
  });
});
