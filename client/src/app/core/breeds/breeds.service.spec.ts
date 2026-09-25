import { provideHttpClient } from '@angular/common/http';
import { HttpTestingController, provideHttpClientTesting } from '@angular/common/http/testing';
import { TestBed } from '@angular/core/testing';
import { firstValueFrom } from 'rxjs';

import { BreedsService } from './breeds.service';

const breedPayload = {
  id: '1',
  name: 'Duroc',
  active: true,
  created_at: '',
  updated_at: '',
  created_by: 'admin',
  updated_by: 'admin',
};

describe('BreedsService', () => {
  let service: BreedsService;
  let http: HttpTestingController;

  beforeEach(() => {
    TestBed.configureTestingModule({
      providers: [provideHttpClient(), provideHttpClientTesting()],
    });
    service = TestBed.inject(BreedsService);
    http = TestBed.inject(HttpTestingController);
  });

  afterEach(() => http.verify());

  it('lists breeds', async () => {
    const promise = firstValueFrom(service.listBreeds());

    const call = http.expectOne(
      (request) => request.url.endsWith('/breeds') && request.method === 'GET',
    );
    call.flush([breedPayload]);

    await expect(promise).resolves.toHaveLength(1);
  });

  it('lists the active breed options', async () => {
    const promise = firstValueFrom(service.listBreedOptions());

    const call = http.expectOne(
      (request) => request.url.endsWith('/breeds/options') && request.method === 'GET',
    );
    call.flush([{ id: '1', name: 'Duroc' }]);

    await expect(promise).resolves.toEqual([{ id: '1', name: 'Duroc' }]);
  });

  it('creates a breed', async () => {
    const promise = firstValueFrom(service.createBreed({ name: 'Duroc' }));

    const call = http.expectOne(
      (request) => request.url.endsWith('/breeds') && request.method === 'POST',
    );
    expect(call.request.body).toEqual({ name: 'Duroc' });
    call.flush(breedPayload);

    await expect(promise).resolves.toMatchObject({ name: 'Duroc' });
  });

  it('updates a breed', async () => {
    const payload = { name: 'Landrace', active: false };
    const promise = firstValueFrom(service.updateBreed('7', payload));

    const call = http.expectOne(
      (request) => request.url.endsWith('/breeds/7') && request.method === 'PATCH',
    );
    expect(call.request.body).toEqual(payload);
    call.flush({ ...breedPayload, name: 'Landrace', active: false });

    await expect(promise).resolves.toMatchObject({ name: 'Landrace', active: false });
  });
});
