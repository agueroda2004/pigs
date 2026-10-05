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
  receiver_sow_id: 'sow-2',
  movement_date: '2026-04-25',
  quantity: 2,
  note: null,
  created_at: '',
  updated_at: '',
  created_by: 'admin',
  updated_by: 'admin',
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
    expect(call.request.params.keys()).toHaveLength(0);
    call.flush([fosteringPayload]);

    await expect(promise).resolves.toHaveLength(1);
  });

  it('lists piglet fosterings with the sow and date range filters', async () => {
    const promise = firstValueFrom(
      service.listPigletFosterings({
        donor_sow_id: 'sow-1',
        receiver_sow_id: 'sow-2',
        from: '2026-04-01',
        to: '2026-04-30',
      }),
    );

    const call = http.expectOne(
      (request) => request.url.endsWith('/piglet-fosterings') && request.method === 'GET',
    );
    expect(call.request.params.get('donor_sow_id')).toBe('sow-1');
    expect(call.request.params.get('receiver_sow_id')).toBe('sow-2');
    expect(call.request.params.get('from')).toBe('2026-04-01');
    expect(call.request.params.get('to')).toBe('2026-04-30');
    call.flush([fosteringPayload]);

    await expect(promise).resolves.toHaveLength(1);
  });

  it('creates a piglet fostering', async () => {
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
    call.flush(fosteringPayload);

    await expect(promise).resolves.toMatchObject({ donor_sow_id: 'sow-1' });
  });
});
