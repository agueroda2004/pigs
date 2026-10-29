import { provideHttpClient } from '@angular/common/http';
import { HttpTestingController, provideHttpClientTesting } from '@angular/common/http/testing';
import { TestBed } from '@angular/core/testing';
import { firstValueFrom } from 'rxjs';

import { PigletFosteringsService } from './piglet-fosterings.service';

const fosteringPayload = {
  id: '1',
  donor_farrowing_id: 'farrowing-1',
  receiver_farrowing_id: 'farrowing-2',
  donor_sow_id: 'sow-1',
  donor_sow_code: 'C-001',
  receiver_sow_id: 'sow-2',
  receiver_sow_code: 'C-002',
  movement_date: '2026-04-25',
  quantity: 2,
  note: null,
};

const pagePayload = {
  items: [fosteringPayload],
  total: 1,
  page: 1,
  page_size: 10,
  total_pages: 1,
};

describe('PigletFosteringsService', () => {
  let service: PigletFosteringsService;
  let http: HttpTestingController;

  beforeEach(() => {
    TestBed.configureTestingModule({
      providers: [provideHttpClient(), provideHttpClientTesting()],
    });
    service = TestBed.inject(PigletFosteringsService);
    http = TestBed.inject(HttpTestingController);
  });

  afterEach(() => http.verify());

  it('lists piglet fosterings', async () => {
    const promise = firstValueFrom(service.listPigletFosterings());

    const call = http.expectOne(
      (request) => request.url.endsWith('/piglet-fosterings') && request.method === 'GET',
    );
    expect(call.request.params.get('page')).toBe('1');
    expect(call.request.params.has('sow_code')).toBe(false);
    call.flush(pagePayload);

    await expect(promise).resolves.toEqual(pagePayload);
  });

  it('lists piglet fosterings with the sow code, date range and page', async () => {
    const promise = firstValueFrom(
      service.listPigletFosterings(
        {
          sow_code: 'C-001',
          from: '2026-04-01',
          to: '2026-04-30',
        },
        2,
      ),
    );

    const call = http.expectOne(
      (request) => request.url.endsWith('/piglet-fosterings') && request.method === 'GET',
    );
    expect(call.request.params.get('sow_code')).toBe('C-001');
    expect(call.request.params.get('from')).toBe('2026-04-01');
    expect(call.request.params.get('to')).toBe('2026-04-30');
    expect(call.request.params.get('page')).toBe('2');
    call.flush(pagePayload);

    await expect(promise).resolves.toEqual(pagePayload);
  });

  it('creates a piglet fostering and returns no body', async () => {
    const payload = {
      donor_sow_id: 'sow-1',
      receiver_sow_id: 'sow-2',
      movement_date: '2026-04-25',
      quantity: 2,
    };
    const promise = firstValueFrom(service.createPigletFostering(payload));

    const call = http.expectOne(
      (request) => request.url.endsWith('/piglet-fosterings') && request.method === 'POST',
    );
    expect(call.request.body).toEqual(payload);
    call.flush(null, { status: 201, statusText: 'Created' });

    await expect(promise).resolves.toBeNull();
  });

  it('updates a piglet fostering and returns no body', async () => {
    const payload = {
      note: 'ajuste',
    };
    const promise = firstValueFrom(service.updatePigletFostering('id-1', payload));

    const call = http.expectOne(
      (request) => request.url.endsWith('/piglet-fosterings/id-1') && request.method === 'PATCH',
    );
    expect(call.request.body).toEqual(payload);
    call.flush(null, { status: 200, statusText: 'OK' });

    await expect(promise).resolves.toBeNull();
  });

  it('deletes a piglet fostering and returns no body', async () => {
    const promise = firstValueFrom(service.deletePigletFostering('id-1'));

    const call = http.expectOne(
      (request) => request.url.endsWith('/piglet-fosterings/id-1') && request.method === 'DELETE',
    );
    expect(call.request.body).toBeNull();
    call.flush(null, { status: 204, statusText: 'No Content' });

    await expect(promise).resolves.toBeNull();
  });
});
