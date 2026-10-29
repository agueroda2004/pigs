import { HttpErrorResponse } from '@angular/common/http';
import { TestBed } from '@angular/core/testing';
import { of, throwError } from 'rxjs';
import { vi } from 'vitest';

import { NotificationService } from '../../../core/notifications/notification.service';
import { PigletFostering } from '../../../core/piglet-fosterings/piglet-fostering.models';
import { PigletFosteringsService } from '../../../core/piglet-fosterings/piglet-fosterings.service';
import { EditPigletFosteringModal } from './edit-piglet-fostering-modal';

function buildFostering(overrides: Partial<PigletFostering> = {}): PigletFostering {
  return {
    id: '42',
    donor_farrowing_id: 'farrowing-1',
    receiver_farrowing_id: 'farrowing-2',
    donor_sow_id: 'sow-1',
    donor_sow_code: 'C-001',
    receiver_sow_id: 'sow-2',
    receiver_sow_code: 'C-002',
    movement_date: '2026-04-25',
    quantity: 2,
    note: 'camada numerosa',
    ...overrides,
  };
}

class PigletFosteringsStub {
  updatePigletFostering = vi.fn((_id: string, _request: Record<string, unknown>) => of(undefined));
}

class NotificationsStub {
  success = vi.fn();
  error = vi.fn();
}

describe('EditPigletFosteringModal', () => {
  let stub: PigletFosteringsStub;
  let notifications: NotificationsStub;

  beforeEach(async () => {
    stub = new PigletFosteringsStub();
    notifications = new NotificationsStub();
    await TestBed.configureTestingModule({
      imports: [EditPigletFosteringModal],
      providers: [
        { provide: PigletFosteringsService, useValue: stub },
        { provide: NotificationService, useValue: notifications },
      ],
    }).compileComponents();
  });

  const create = async (fostering: PigletFostering = buildFostering()) => {
    const fixture = TestBed.createComponent(EditPigletFosteringModal);
    fixture.componentRef.setInput('fostering', fostering);
    fixture.componentRef.setInput('open', true);
    fixture.detectChanges();
    await fixture.whenStable();
    // eslint-disable-next-line @typescript-eslint/no-explicit-any
    return { fixture, component: fixture.componentInstance as any };
  };

  it('prefills the note and renders the read-only summary', async () => {
    const { fixture, component } = await create();

    expect(component.form.getRawValue()).toEqual({ note: 'camada numerosa' });

    const text = fixture.nativeElement.textContent.replace(/\s+/g, ' ');
    expect(text).toContain('C-001');
    expect(text).toContain('C-002');
    expect(text).toContain('2026-04-25');
    expect(text).toContain('2');
    expect(text).toContain(
      'La fecha, la cantidad y las cerdas donante y receptora no se pueden editar para mantener la integridad de los datos.',
    );
  });

  it('does not expose the locked fields as editable controls', async () => {
    const { fixture, component } = await create();

    expect(component.form.contains('movement_date')).toBe(false);
    expect(component.form.contains('quantity')).toBe(false);
    expect(fixture.nativeElement.querySelector('app-date-picker')).toBeNull();
    expect(fixture.nativeElement.querySelector('input[type="number"]')).toBeNull();
  });

  it('sends only the changed note', async () => {
    const { component } = await create();
    const updated = vi.fn();
    component.updated.subscribe(updated);
    component.form.patchValue({ note: 'nota ajustada' });

    await component.submit();

    expect(stub.updatePigletFostering).toHaveBeenCalledWith('42', { note: 'nota ajustada' });
    expect(updated).toHaveBeenCalled();
  });

  it('clears the note when the field is emptied', async () => {
    const { component } = await create();
    component.form.patchValue({ note: '' });

    await component.submit();

    expect(stub.updatePigletFostering).toHaveBeenCalledWith('42', { note: '' });
  });

  it('closes without calling the API when nothing changed', async () => {
    const { component } = await create();
    const closed = vi.fn();
    component.closed.subscribe(closed);

    await component.submit();

    expect(stub.updatePigletFostering).not.toHaveBeenCalled();
    expect(closed).toHaveBeenCalled();
  });

  it('does not submit when the note is too long', async () => {
    const { component } = await create();
    component.form.patchValue({ note: 'a'.repeat(501) });

    await component.submit();

    expect(stub.updatePigletFostering).not.toHaveBeenCalled();
  });

  it('shows the server message when the update fails', async () => {
    stub.updatePigletFostering = vi.fn(() =>
      throwError(
        () =>
          new HttpErrorResponse({
            status: 400,
            error: { error: 'La nota debe tener como máximo 500 caracteres' },
          }),
      ),
    );
    const { component } = await create();
    component.form.patchValue({ note: 'nota' });

    await component.submit();

    expect(notifications.error).toHaveBeenCalledWith(
      'La nota debe tener como máximo 500 caracteres',
    );
  });

  it('shows a fallback when the fostering is missing', async () => {
    stub.updatePigletFostering = vi.fn(() =>
      throwError(() => new HttpErrorResponse({ status: 404 })),
    );
    const { component } = await create();
    component.form.patchValue({ note: 'nota' });

    await component.submit();

    expect(notifications.error).toHaveBeenCalledWith('El traslado no existe');
  });

  it('shows a generic fallback for unexpected errors', async () => {
    stub.updatePigletFostering = vi.fn(() => throwError(() => new Error('failed')));
    const { component } = await create();
    component.form.patchValue({ note: 'nota' });

    await component.submit();

    expect(notifications.error).toHaveBeenCalledWith('No se pudo actualizar el traslado');
  });
});
