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
    created_at: '2026-01-02T12:00:00',
    updated_at: '2026-01-02T12:00:00',
    created_by: 'admin',
    updated_by: 'admin',
    ...overrides,
  };
}

class SowsStub {
  updateSow = vi.fn((_id: string, _request: Record<string, unknown>) => of(buildSow()));
}

class BreedsStub {
  listBreeds = vi.fn(() => of([{ id: 'breed-1', name: 'Duroc' }]));
}

class NotificationsStub {
  success = vi.fn();
  error = vi.fn();
}

describe('EditSowModal', () => {
  let stub: SowsStub;
  let notifications: NotificationsStub;

  beforeEach(async () => {
    stub = new SowsStub();
    notifications = new NotificationsStub();
    await TestBed.configureTestingModule({
      imports: [EditSowModal],
      providers: [
        { provide: SowsService, useValue: stub },
        { provide: BreedsService, useValue: new BreedsStub() },
        { provide: NotificationService, useValue: notifications },
      ],
    }).compileComponents();
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
      location: 'Corral A',
      active: true,
      entry_date: '2026-01-10',
      birth_date: '2025-12-01',
      note: 'Nota original',
    });
  });

  it('shows the parity read-only', async () => {
    const fixture = await openWith(buildSow({ parity: 5 }));
    // eslint-disable-next-line @typescript-eslint/no-explicit-any
    const component = fixture.componentInstance as any;

    expect(component.parity()).toBe(5);
    expect(fixture.nativeElement.textContent).toContain('Paridad');
  });

  it('sends the origin when it changes', async () => {
    const fixture = await openWith(buildSow());
    // eslint-disable-next-line @typescript-eslint/no-explicit-any
    const component = fixture.componentInstance as any;
    component.form.patchValue({ origin: 'Externo' });

    await component.submit();

    expect(stub.updateSow).toHaveBeenCalledWith('1', { origin: 'Externo' });
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

  it('never sends the state or parity', async () => {
    const fixture = await openWith(buildSow());
    // eslint-disable-next-line @typescript-eslint/no-explicit-any
    const component = fixture.componentInstance as any;
    component.form.patchValue({ code: 'C-003' });

    await component.submit();

    const request = stub.updateSow.mock.calls[0][1] as Record<string, unknown>;
    expect(request).not.toHaveProperty('state');
    expect(request).not.toHaveProperty('parity');
  });

  it('emits the current sow when nothing changed', async () => {
    const sow = buildSow();
    const fixture = await openWith(sow);
    // eslint-disable-next-line @typescript-eslint/no-explicit-any
    const component = fixture.componentInstance as any;
    const updated = vi.fn();
    component.updated.subscribe(updated);

    await component.submit();

    expect(stub.updateSow).not.toHaveBeenCalled();
    expect(updated).toHaveBeenCalledWith(sow);
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
