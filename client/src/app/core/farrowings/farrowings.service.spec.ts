import { provideHttpClient } from '@angular/common/http';
import { HttpTestingController, provideHttpClientTesting } from '@angular/common/http/testing';
import { TestBed } from '@angular/core/testing';
import { firstValueFrom } from 'rxjs';

import { FarrowingsService } from './farrowings.service';

const farrowingPayload = {
  id: '1',
  sow_id: 'sow-1',
  sow_code: 'C-001',
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
};

const pagePayload = {
  items: [farrowingPayload],
  total: 1,
  page: 1,
  page_size: 10,
  total_pages: 1,
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

  it('lists farrowings with the default page', async () => {
    const promise = firstValueFrom(service.listFarrowings());

    const call = http.expectOne(
      (request) => request.url.endsWith('/farrowings') && request.method === 'GET',
    );
    expect(call.request.params.get('page')).toBe('1');
    call.flush(pagePayload);

    await expect(promise).resolves.toMatchObject({ items: [farrowingPayload], total: 1 });
  });

  it('lists farrowings with the sow code filter and a page', async () => {
    const promise = firstValueFrom(service.listFarrowings({ sow_code: 'C-001' }, 2));

    const call = http.expectOne(
      (request) => request.url.endsWith('/farrowings') && request.method === 'GET',
    );
    expect(call.request.params.get('sow_code')).toBe('C-001');
    expect(call.request.params.get('page')).toBe('2');
    call.flush(pagePayload);

    await expect(promise).resolves.toMatchObject({ items: [farrowingPayload] });
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
    call.flush(pagePayload);

    await expect(promise).resolves.toMatchObject({ total: 1 });
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
    call.flush(null);

    await expect(promise).resolves.toBeNull();
  });

  it('updates a farrowing', async () => {
    const payload = { live_born: 12, clear_litter_weight: true };
    const promise = firstValueFrom(service.updateFarrowing('7', payload));

    const call = http.expectOne(
      (request) => request.url.endsWith('/farrowings/7') && request.method === 'PATCH',
    );
    expect(call.request.body).toEqual(payload);
    call.flush(null);

    await expect(promise).resolves.toBeNull();
  });

  it('deletes a farrowing', async () => {
    const promise = firstValueFrom(service.deleteFarrowing('7'));

    const call = http.expectOne(
      (request) => request.url.endsWith('/farrowings/7') && request.method === 'DELETE',
    );
    call.flush(null);

    await expect(promise).resolves.toBeNull();
  });
});
