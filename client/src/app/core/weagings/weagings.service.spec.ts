import { provideHttpClient } from '@angular/common/http';
import { HttpTestingController, provideHttpClientTesting } from '@angular/common/http/testing';
import { TestBed } from '@angular/core/testing';
import { firstValueFrom } from 'rxjs';

import { WeagingPage } from './weaging.models';
import { WeagingsService } from './weagings.service';

const weagingPayload = {
  id: '1',
  farrowing_id: 'farrowing-1',
  sow_id: 'sow-1',
  sow_code: 'C-001',
  weaging_date: '2026-04-25',
  quantity: 8,
  total_weight: 120.5,
  destination: 'Nave 2',
  note: null,
};

function buildPage(overrides: Partial<WeagingPage> = {}): WeagingPage {
  return {
    items: [weagingPayload],
    total: 1,
    page: 1,
    page_size: 10,
    total_pages: 1,
    ...overrides,
  };
}

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

  it('lists the first page of weagings', async () => {
    const promise = firstValueFrom(service.listWeagings());

    const call = http.expectOne(
      (request) => request.url.endsWith('/weagings') && request.method === 'GET',
    );
    expect(call.request.params.keys()).toEqual(['page']);
    expect(call.request.params.get('page')).toBe('1');
    call.flush(buildPage());

    await expect(promise).resolves.toMatchObject({ items: [weagingPayload], total_pages: 1 });
  });

  it('lists weagings with the sow code, date range and page filters', async () => {
    const promise = firstValueFrom(
      service.listWeagings({ sow_code: 'C-001', from: '2026-04-01', to: '2026-04-30' }, 2),
    );

    const call = http.expectOne(
      (request) => request.url.endsWith('/weagings') && request.method === 'GET',
    );
    expect(call.request.params.get('sow_code')).toBe('C-001');
    expect(call.request.params.get('from')).toBe('2026-04-01');
    expect(call.request.params.get('to')).toBe('2026-04-30');
    expect(call.request.params.get('page')).toBe('2');
    call.flush(buildPage({ total: 11, page: 2, total_pages: 2 }));

    await expect(promise).resolves.toMatchObject({ page: 2, total: 11, total_pages: 2 });
  });

  it('creates a weaging without expecting a response body', async () => {
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
    call.flush(null);

    await expect(promise).resolves.toBeNull();
  });

  it('updates a weaging', async () => {
    const payload = {
      total_weight: 130.5,
      destination: 'Nave 3',
      note: 'nueva nota',
    };
    const promise = firstValueFrom(service.updateWeaging('1', payload));

    const call = http.expectOne(
      (request) => request.url.endsWith('/weagings/1') && request.method === 'PATCH',
    );
    expect(call.request.body).toEqual(payload);
    call.flush(null);

    await expect(promise).resolves.toBeNull();
  });

  it('deletes a weaging', async () => {
    const promise = firstValueFrom(service.deleteWeaging('1'));

    const call = http.expectOne(
      (request) => request.url.endsWith('/weagings/1') && request.method === 'DELETE',
    );
    call.flush(null);

    await expect(promise).resolves.toBeNull();
  });
});
