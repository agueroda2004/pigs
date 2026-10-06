import { provideHttpClient } from '@angular/common/http';
import { HttpTestingController, provideHttpClientTesting } from '@angular/common/http/testing';
import { TestBed } from '@angular/core/testing';
import { firstValueFrom } from 'rxjs';

import { PartialWeagingsService } from './partial-weagings.service';

const weagingPayload = {
  id: '1',
  farrowing_id: 'farrowing-1',
  sow_id: 'sow-1',
  weaging_date: '2026-04-25',
  quantity: 2,
  total_weight: 42.5,
  type: 'Normal',
  note: null,
  created_at: '',
  updated_at: '',
  created_by: 'admin',
  updated_by: 'admin',
};

describe('PartialWeagingsService', () => {
  let service: PartialWeagingsService;
  let http: HttpTestingController;

  beforeEach(() => {
    TestBed.configureTestingModule({
      providers: [provideHttpClient(), provideHttpClientTesting()],
    });
    service = TestBed.inject(PartialWeagingsService);
    http = TestBed.inject(HttpTestingController);
  });

  afterEach(() => http.verify());

  it('lists partial weagings', async () => {
    const promise = firstValueFrom(service.listPartialWeagings());

    const call = http.expectOne(
      (request) => request.url.endsWith('/partial-weagings') && request.method === 'GET',
    );
    expect(call.request.params.keys()).toHaveLength(0);
    call.flush([weagingPayload]);

    await expect(promise).resolves.toHaveLength(1);
  });

  it('lists partial weagings with the sow and date range filters', async () => {
    const promise = firstValueFrom(
      service.listPartialWeagings({
        sow_id: 'sow-1',
        from: '2026-04-01',
        to: '2026-04-30',
      }),
    );

    const call = http.expectOne(
      (request) => request.url.endsWith('/partial-weagings') && request.method === 'GET',
    );
    expect(call.request.params.get('sow_id')).toBe('sow-1');
    expect(call.request.params.get('from')).toBe('2026-04-01');
    expect(call.request.params.get('to')).toBe('2026-04-30');
    call.flush([weagingPayload]);

    await expect(promise).resolves.toHaveLength(1);
  });

  it('creates a partial weaging', async () => {
    const payload = {
      sow_id: 'sow-1',
      weaging_date: '2026-04-25',
      quantity: 2,
      type: 'Normal' as const,
    };
    const promise = firstValueFrom(service.createPartialWeaging(payload));

    const call = http.expectOne(
      (request) => request.url.endsWith('/partial-weagings') && request.method === 'POST',
    );
    expect(call.request.body).toEqual(payload);
    call.flush(weagingPayload);

    await expect(promise).resolves.toMatchObject({ sow_id: 'sow-1' });
  });
});
