import { provideHttpClient } from '@angular/common/http';
import { HttpTestingController, provideHttpClientTesting } from '@angular/common/http/testing';
import { TestBed } from '@angular/core/testing';
import { firstValueFrom } from 'rxjs';

import { WeagingsService } from './weagings.service';

const weagingPayload = {
  id: '1',
  farrowing_id: 'farrowing-1',
  sow_id: 'sow-1',
  weaging_date: '2026-04-25',
  quantity: 8,
  total_weight: 120.5,
  destination: 'Nave 2',
  note: null,
  created_at: '',
  updated_at: '',
  created_by: 'admin',
  updated_by: 'admin',
};

describe('WeagingsService', () => {
  let service: WeagingsService;
  let http: HttpTestingController;

  beforeEach(() => {
    TestBed.configureTestingModule({
      providers: [provideHttpClient(), provideHttpClientTesting()],
    });
    service = TestBed.inject(WeagingsService);
    http = TestBed.inject(HttpTestingController);
  });

  afterEach(() => http.verify());

  it('lists weagings', async () => {
    const promise = firstValueFrom(service.listWeagings());

    const call = http.expectOne(
      (request) => request.url.endsWith('/weagings') && request.method === 'GET',
    );
    expect(call.request.params.keys()).toHaveLength(0);
    call.flush([weagingPayload]);

    await expect(promise).resolves.toHaveLength(1);
  });

  it('lists weagings with the sow and date range filters', async () => {
    const promise = firstValueFrom(
      service.listWeagings({ sow_id: 'sow-1', from: '2026-04-01', to: '2026-04-30' }),
    );

    const call = http.expectOne(
      (request) => request.url.endsWith('/weagings') && request.method === 'GET',
    );
    expect(call.request.params.get('sow_id')).toBe('sow-1');
    expect(call.request.params.get('from')).toBe('2026-04-01');
    expect(call.request.params.get('to')).toBe('2026-04-30');
    call.flush([weagingPayload]);

    await expect(promise).resolves.toHaveLength(1);
  });

  it('creates a weaging', async () => {
    const payload = {
      sow_id: 'sow-1',
      weaging_date: '2026-04-25',
      quantity: 8,
    };
    const promise = firstValueFrom(service.createWeaging(payload));

    const call = http.expectOne(
      (request) => request.url.endsWith('/weagings') && request.method === 'POST',
    );
    expect(call.request.body).toEqual(payload);
    call.flush(weagingPayload);

    await expect(promise).resolves.toMatchObject({ sow_id: 'sow-1' });
  });
});
