import { provideHttpClient } from '@angular/common/http';
import { HttpTestingController, provideHttpClientTesting } from '@angular/common/http/testing';
import { TestBed } from '@angular/core/testing';
import { firstValueFrom } from 'rxjs';

import { PartialWeaging, PartialWeagingPage } from './partial-weaging.models';
import { PartialWeagingsService } from './partial-weagings.service';

const weagingPayload: PartialWeaging = {
  id: '1',
  farrowing_id: 'farrowing-1',
  sow_id: 'sow-1',
  sow_code: 'C-001',
  weaging_date: '2026-04-25',
  quantity: 2,
  total_weight: 42.5,
  type: 'Normal',
  note: null,
};

const pagePayload: PartialWeagingPage = {
  items: [weagingPayload],
  total: 1,
  page: 1,
  page_size: 10,
  total_pages: 1,
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

  it('lists the first page of partial weagings', async () => {
    const promise = firstValueFrom(service.listPartialWeagings());

    const call = http.expectOne(
      (request) => request.url.endsWith('/partial-weagings') && request.method === 'GET',
    );
    expect(call.request.params.get('page')).toBe('1');
    call.flush(pagePayload);

    await expect(promise).resolves.toMatchObject({ total: 1, total_pages: 1 });
  });

  it('lists partial weagings with the sow code, date range and page filters', async () => {
    const promise = firstValueFrom(
      service.listPartialWeagings(
        {
          sow_code: 'C-001',
          from: '2026-04-01',
          to: '2026-04-30',
        },
        2,
      ),
    );

    const call = http.expectOne(
      (request) => request.url.endsWith('/partial-weagings') && request.method === 'GET',
    );
    expect(call.request.params.get('sow_code')).toBe('C-001');
    expect(call.request.params.get('from')).toBe('2026-04-01');
    expect(call.request.params.get('to')).toBe('2026-04-30');
    expect(call.request.params.get('page')).toBe('2');
    call.flush(pagePayload);

    await expect(promise).resolves.toMatchObject({ items: [weagingPayload] });
  });

  it('creates a partial weaging without expecting a body', async () => {
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
    call.flush(null);

    await expect(promise).resolves.toBeNull();
  });

  it('updates a partial weaging without expecting a body', async () => {
    const payload = {
      total_weight: 50,
      note: 'nota',
    };
    const promise = firstValueFrom(service.updatePartialWeaging('42', payload));

    const call = http.expectOne(
      (request) => request.url.endsWith('/partial-weagings/42') && request.method === 'PATCH',
    );
    expect(call.request.body).toEqual(payload);
    call.flush(null);

    await expect(promise).resolves.toBeNull();
  });

  it('deletes a partial weaging without expecting a body', async () => {
    const promise = firstValueFrom(service.deletePartialWeaging('42'));

    const call = http.expectOne(
      (request) => request.url.endsWith('/partial-weagings/42') && request.method === 'DELETE',
    );
    call.flush(null);

    await expect(promise).resolves.toBeNull();
  });
});
