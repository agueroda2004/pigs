import { HttpErrorResponse } from '@angular/common/http';
import { ComponentFixture, TestBed } from '@angular/core/testing';
import { of, throwError } from 'rxjs';
import { vi } from 'vitest';

import { Boar } from '../../../core/boars/boar.models';
import { BoarsService } from '../../../core/boars/boars.service';
import { BreedsService } from '../../../core/breeds/breeds.service';
import { NotificationService } from '../../../core/notifications/notification.service';
import { EditBoarModal } from './edit-boar-modal';

function buildBoar(overrides: Partial<Boar> = {}): Boar {
  return {
    id: '1',
    code: 'B-001',
    location: 'Corral A',
    active: true,
    entry_date: '2026-01-10',
    birth_date: '2025-12-01',
    note: 'Nota original',
    state: 'Vivo',
    origin: 'Propio',
    breed_id: 'breed-1',
    ...overrides,
  };
}

class BoarsStub {
  updateBoar = vi.fn((_id: string, _request: Record<string, unknown>) => of(undefined));
}

class BreedsStub {
  listBreedDropdown = vi.fn(() =>
    of([
      { id: 'breed-1', name: 'Duroc', active: true },
      { id: 'breed-2', name: 'Landrace', active: false },
    ]),
  );
}

class NotificationsStub {
  success = vi.fn();
  error = vi.fn();
}

describe('EditBoarModal', () => {
  let stub: BoarsStub;
  let breeds: BreedsStub;
  let notifications: NotificationsStub;

  beforeEach(async () => {
    stub = new BoarsStub();
    breeds = new BreedsStub();
    notifications = new NotificationsStub();
    await TestBed.configureTestingModule({
      imports: [EditBoarModal],
      providers: [
        { provide: BoarsService, useValue: stub },
        { provide: BreedsService, useValue: breeds },
        { provide: NotificationService, useValue: notifications },
      ],
    }).compileComponents();
  });

  it('loads every breed and marks the inactive ones', async () => {
    const fixture = await openWith(buildBoar());
    // eslint-disable-next-line @typescript-eslint/no-explicit-any
    const component = fixture.componentInstance as any;

    expect(breeds.listBreedDropdown).toHaveBeenCalledWith(false);
    expect(component.breedOptions()).toEqual([
      { value: 'breed-1', label: 'Duroc' },
      { value: 'breed-2', label: 'Landrace (inactiva)' },
    ]);
  });

  async function openWith(boar: Boar): Promise<ComponentFixture<EditBoarModal>> {
    const fixture = TestBed.createComponent(EditBoarModal);
    fixture.componentRef.setInput('boar', boar);
    fixture.componentRef.setInput('open', true);
    fixture.detectChanges();
    await fixture.whenStable();
    fixture.detectChanges();
    return fixture;
  }

  it('prefills the form from the boar', async () => {
    const fixture = await openWith(buildBoar());
    // eslint-disable-next-line @typescript-eslint/no-explicit-any
    const component = fixture.componentInstance as any;

    expect(component.form.getRawValue()).toEqual({
      code: 'B-001',
      breed_id: 'breed-1',
      origin: 'Propio',
      location: 'Corral A',
      active: true,
      entry_date: '2026-01-10',
      birth_date: '2025-12-01',
      note: 'Nota original',
    });
  });

  it('sends the origin when it changes', async () => {
    const fixture = await openWith(buildBoar());
    // eslint-disable-next-line @typescript-eslint/no-explicit-any
    const component = fixture.componentInstance as any;
    component.form.patchValue({ origin: 'Externo' });

    await component.submit();

    expect(stub.updateBoar).toHaveBeenCalledWith('1', { origin: 'Externo' });
  });

  it('sends only the changed fields', async () => {
    const fixture = await openWith(buildBoar());
    // eslint-disable-next-line @typescript-eslint/no-explicit-any
    const component = fixture.componentInstance as any;
    component.form.patchValue({ code: 'B-002' });

    await component.submit();

    expect(stub.updateBoar).toHaveBeenCalledWith('1', { code: 'B-002' });
  });

  it('clears nullable fields with an empty string', async () => {
    const fixture = await openWith(buildBoar());
    // eslint-disable-next-line @typescript-eslint/no-explicit-any
    const component = fixture.componentInstance as any;
    component.form.patchValue({ location: '', note: '', birth_date: '' });

    await component.submit();

    expect(stub.updateBoar).toHaveBeenCalledWith('1', {
      location: '',
      note: '',
      birth_date: '',
    });
  });

  it('never sends the state', async () => {
    const fixture = await openWith(buildBoar());
    // eslint-disable-next-line @typescript-eslint/no-explicit-any
    const component = fixture.componentInstance as any;
    component.form.patchValue({ code: 'B-003' });

    await component.submit();

    const request = stub.updateBoar.mock.calls[0][1] as Record<string, unknown>;
    expect(request).not.toHaveProperty('state');
  });

  it('emits when nothing changed', async () => {
    const fixture = await openWith(buildBoar());
    // eslint-disable-next-line @typescript-eslint/no-explicit-any
    const component = fixture.componentInstance as any;
    const updated = vi.fn();
    component.updated.subscribe(updated);

    await component.submit();

    expect(stub.updateBoar).not.toHaveBeenCalled();
    expect(updated).toHaveBeenCalled();
  });

  it('shows an error toast when the boar is missing', async () => {
    stub.updateBoar = vi.fn(() => throwError(() => new HttpErrorResponse({ status: 404 })));
    const fixture = await openWith(buildBoar());
    // eslint-disable-next-line @typescript-eslint/no-explicit-any
    const component = fixture.componentInstance as any;
    component.form.patchValue({ code: 'B-004' });

    await component.submit();

    expect(notifications.error).toHaveBeenCalledWith('El verraco no existe');
  });
});
