import { HttpErrorResponse } from '@angular/common/http';
import { ComponentFixture, TestBed } from '@angular/core/testing';
import { of, throwError } from 'rxjs';
import { vi } from 'vitest';

import { BreedsService } from '../../../core/breeds/breeds.service';
import { NotificationService } from '../../../core/notifications/notification.service';
import { Sow } from '../../../core/sows/sow.models';
import { SowsService } from '../../../core/sows/sows.service';
import { EditSowModal } from './edit-sow-modal';

function buildSow(overrides: Partial<Sow> = {}): Sow {
  return {
    id: '1',
    code: 'C-001',
    location: 'Corral A',
    active: true,
    entry_date: '2026-01-10',
    birth_date: '2025-12-01',
    note: 'Nota original',
    state: 'Viva',
    origin: 'Propio',
    parity: 3,
    breed_id: 'breed-1',
    ...overrides,
  };
}

class SowsStub {
  updateSow = vi.fn((_id: string, _request: Record<string, unknown>) => of(undefined));
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

describe('EditSowModal', () => {
  let stub: SowsStub;
  let breeds: BreedsStub;
  let notifications: NotificationsStub;

  beforeEach(async () => {
    stub = new SowsStub();
    breeds = new BreedsStub();
    notifications = new NotificationsStub();
    await TestBed.configureTestingModule({
      imports: [EditSowModal],
      providers: [
        { provide: SowsService, useValue: stub },
        { provide: BreedsService, useValue: breeds },
        { provide: NotificationService, useValue: notifications },
      ],
    }).compileComponents();
  });

  it('loads every breed and marks the inactive ones', async () => {
    const fixture = await openWith(buildSow());
    // eslint-disable-next-line @typescript-eslint/no-explicit-any
    const component = fixture.componentInstance as any;

    expect(breeds.listBreedDropdown).toHaveBeenCalledWith(false);
    expect(component.breedOptions()).toEqual([
      { value: 'breed-1', label: 'Duroc' },
      { value: 'breed-2', label: 'Landrace (inactiva)' },
    ]);
  });

  async function openWith(sow: Sow): Promise<ComponentFixture<EditSowModal>> {
    const fixture = TestBed.createComponent(EditSowModal);
    fixture.componentRef.setInput('sow', sow);
    fixture.componentRef.setInput('open', true);
    fixture.detectChanges();
    await fixture.whenStable();
    fixture.detectChanges();
    return fixture;
  }

  it('prefills the form from the sow', async () => {
    const fixture = await openWith(buildSow());
    // eslint-disable-next-line @typescript-eslint/no-explicit-any
    const component = fixture.componentInstance as any;

    expect(component.form.getRawValue()).toEqual({
      code: 'C-001',
      breed_id: 'breed-1',
      origin: 'Propio',
      parity: 3,
      location: 'Corral A',
      entry_date: '2026-01-10',
      birth_date: '2025-12-01',
      note: 'Nota original',
    });
  });

  it('enables parity when the sow is alive', async () => {
    const fixture = await openWith(buildSow({ state: 'Viva' }));
    // eslint-disable-next-line @typescript-eslint/no-explicit-any
    const component = fixture.componentInstance as any;

    expect(component.canEditParity()).toBe(true);
    expect(component.form.controls.parity.enabled).toBe(true);
  });

  it('disables parity when the sow is not alive', async () => {
    const fixture = await openWith(buildSow({ state: 'Gestando' }));
    // eslint-disable-next-line @typescript-eslint/no-explicit-any
    const component = fixture.componentInstance as any;

    expect(component.canEditParity()).toBe(false);
    expect(component.form.controls.parity.disabled).toBe(true);
  });

  it('disables the dates when the sow is inactive', async () => {
    const fixture = await openWith(buildSow({ active: false }));
    // eslint-disable-next-line @typescript-eslint/no-explicit-any
    const component = fixture.componentInstance as any;

    expect(component.canEditDates()).toBe(false);
    expect(component.form.controls.entry_date.disabled).toBe(true);
    expect(component.form.controls.birth_date.disabled).toBe(true);
  });

  it('sends the origin when it changes', async () => {
    const fixture = await openWith(buildSow());
    // eslint-disable-next-line @typescript-eslint/no-explicit-any
    const component = fixture.componentInstance as any;
    component.form.patchValue({ origin: 'Externo' });

    await component.submit();

    expect(stub.updateSow).toHaveBeenCalledWith('1', { origin: 'Externo' });
  });

  it('sends the parity when it changes on an alive sow', async () => {
    const fixture = await openWith(buildSow({ state: 'Viva' }));
    // eslint-disable-next-line @typescript-eslint/no-explicit-any
    const component = fixture.componentInstance as any;
    component.form.patchValue({ parity: 5 });

    await component.submit();

    expect(stub.updateSow).toHaveBeenCalledWith('1', { parity: 5 });
  });

  it('sends only the changed fields', async () => {
    const fixture = await openWith(buildSow());
    // eslint-disable-next-line @typescript-eslint/no-explicit-any
    const component = fixture.componentInstance as any;
    component.form.patchValue({ code: 'C-002' });

    await component.submit();

    expect(stub.updateSow).toHaveBeenCalledWith('1', { code: 'C-002' });
  });

  it('clears nullable fields with an empty string', async () => {
    const fixture = await openWith(buildSow());
    // eslint-disable-next-line @typescript-eslint/no-explicit-any
    const component = fixture.componentInstance as any;
    component.form.patchValue({ location: '', note: '', birth_date: '' });

    await component.submit();

    expect(stub.updateSow).toHaveBeenCalledWith('1', {
      location: '',
      note: '',
      birth_date: '',
    });
  });

  it('never sends the state or the active flag', async () => {
    const fixture = await openWith(buildSow());
    // eslint-disable-next-line @typescript-eslint/no-explicit-any
    const component = fixture.componentInstance as any;
    component.form.patchValue({ code: 'C-003' });

    await component.submit();

    const request = stub.updateSow.mock.calls[0][1] as Record<string, unknown>;
    expect(request).not.toHaveProperty('state');
    expect(request).not.toHaveProperty('active');
  });

  it('emits the current code when nothing changed', async () => {
    const fixture = await openWith(buildSow());
    // eslint-disable-next-line @typescript-eslint/no-explicit-any
    const component = fixture.componentInstance as any;
    const updated = vi.fn();
    component.updated.subscribe(updated);

    await component.submit();

    expect(stub.updateSow).not.toHaveBeenCalled();
    expect(updated).toHaveBeenCalledWith('C-001');
  });

  it('emits the resulting code after a successful update', async () => {
    const fixture = await openWith(buildSow());
    // eslint-disable-next-line @typescript-eslint/no-explicit-any
    const component = fixture.componentInstance as any;
    const updated = vi.fn();
    component.updated.subscribe(updated);
    component.form.patchValue({ code: 'C-002' });

    await component.submit();

    expect(updated).toHaveBeenCalledWith('C-002');
  });

  it('shows an error toast when the sow is missing', async () => {
    stub.updateSow = vi.fn(() => throwError(() => new HttpErrorResponse({ status: 404 })));
    const fixture = await openWith(buildSow());
    // eslint-disable-next-line @typescript-eslint/no-explicit-any
    const component = fixture.componentInstance as any;
    component.form.patchValue({ code: 'C-004' });

    await component.submit();

    expect(notifications.error).toHaveBeenCalledWith('La cerda no existe');
  });
});
