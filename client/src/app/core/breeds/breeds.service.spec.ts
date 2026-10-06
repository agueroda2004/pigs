import { provideHttpClient } from '@angular/common/http';
import { HttpTestingController, provideHttpClientTesting } from '@angular/common/http/testing';
import { TestBed } from '@angular/core/testing';
import { firstValueFrom } from 'rxjs';

import { BreedsService } from './breeds.service';

const breedPayload = {
  id: '1',
  name: 'Duroc',
  active: true,
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
    expect(call.request.params.keys()).toHaveLength(0);
    call.flush([breedPayload]);

    await expect(promise).resolves.toHaveLength(1);
  });

  it('lists breeds with the name and active filters', async () => {
    const promise = firstValueFrom(service.listBreeds({ name: 'dur', active: true }));

    const call = http.expectOne(
      (request) => request.url.endsWith('/breeds') && request.method === 'GET',
    );
    expect(call.request.params.get('name')).toBe('dur');
    expect(call.request.params.get('active')).toBe('true');
    call.flush([breedPayload]);

    await expect(promise).resolves.toHaveLength(1);
  });

  it('lists the breed dropdown with the active query', async () => {
    const promise = firstValueFrom(service.listBreedDropdown(false));

    const call = http.expectOne(
      (request) => request.url.endsWith('/breeds/dropdown') && request.method === 'GET',
    );
    expect(call.request.params.get('active')).toBe('false');
    call.flush([breedPayload]);

    await expect(promise).resolves.toEqual([breedPayload]);
  });

  it('creates a breed', async () => {
    const promise = firstValueFrom(service.createBreed({ name: 'Duroc' }));

    const call = http.expectOne(
      (request) => request.url.endsWith('/breeds') && request.method === 'POST',
    );
    expect(call.request.body).toEqual({ name: 'Duroc' });
    call.flush(null);

    await expect(promise).resolves.toBeNull();
  });

  it('updates a breed', async () => {
    const payload = { name: 'Landrace', active: false };
    const promise = firstValueFrom(service.updateBreed('7', payload));

    const call = http.expectOne(
      (request) => request.url.endsWith('/breeds/7') && request.method === 'PATCH',
    );
    expect(call.request.body).toEqual(payload);
    call.flush(null);

    await expect(promise).resolves.toBeNull();
  });

  it('deletes a breed', async () => {
    const promise = firstValueFrom(service.deleteBreed('7'));

    const call = http.expectOne(
      (request) => request.url.endsWith('/breeds/7') && request.method === 'DELETE',
    );
    call.flush(null);

    await expect(promise).resolves.toBeNull();
  });
});
