import { HttpErrorResponse } from '@angular/common/http';
import { TestBed } from '@angular/core/testing';
import { of, throwError } from 'rxjs';
import { vi } from 'vitest';

import { NotificationService } from '../../../core/notifications/notification.service';
import { UpdateWeagingRequest, Weaging } from '../../../core/weagings/weaging.models';
import { WeagingsService } from '../../../core/weagings/weagings.service';
import { EditWeagingModal } from './edit-weaging-modal';

function buildWeaging(overrides: Partial<Weaging> = {}): Weaging {
  return {
    id: '1',
    farrowing_id: 'farrowing-1',
    sow_id: 'sow-1',
    sow_code: 'C-001',
    weaging_date: '2026-04-25',
    quantity: 8,
    total_weight: 42.5,
    destination: 'Nave 2',
    note: null,
    ...overrides,
  };
}

class WeagingsStub {
  updateWeaging = vi.fn((_id: string, _request: UpdateWeagingRequest) => of(void 0));
}

class NotificationsStub {
  success = vi.fn();
  error = vi.fn();
}

describe('EditWeagingModal', () => {
  let stub: WeagingsStub;
  let notifications: NotificationsStub;

  beforeEach(async () => {
    stub = new WeagingsStub();
    notifications = new NotificationsStub();
    await TestBed.configureTestingModule({
      imports: [EditWeagingModal],
      providers: [
        { provide: WeagingsService, useValue: stub },
        { provide: NotificationService, useValue: notifications },
      ],
    }).compileComponents();
  });

  const create = async (weaging: Weaging = buildWeaging()) => {
    const fixture = TestBed.createComponent(EditWeagingModal);
    fixture.componentRef.setInput('weaging', weaging);
    fixture.componentRef.setInput('sowCode', 'C-001');
    fixture.componentRef.setInput('open', true);
    fixture.detectChanges();
    await fixture.whenStable();
    // eslint-disable-next-line @typescript-eslint/no-explicit-any
    return { fixture, component: fixture.componentInstance as any };
  };

  it('prefills the form and renders the read-only summary', async () => {
    const { fixture, component } = await create(buildWeaging({ note: 'nota' }));

    expect(component.form.getRawValue()).toEqual({
      total_weight: 42.5,
      destination: 'Nave 2',
      note: 'nota',
    });

    const text = fixture.nativeElement.textContent.replace(/\s+/g, ' ');
    expect(text).toContain('C-001');
    expect(text).toContain('2026-04-25');
    expect(text).toContain('Cantidad');
    expect(text).toContain('8');
    expect(text).toContain(
      'La cerda, la fecha y la cantidad no se pueden editar para mantener la integridad de los datos.',
    );
  });

  it('does not expose the immutable fields as form controls', async () => {
    const { fixture, component } = await create();

    expect(component.form.contains('weaging_date')).toBe(false);
    expect(component.form.contains('quantity')).toBe(false);
    expect(component.form.contains('sow_id')).toBe(false);
    expect(fixture.nativeElement.querySelector('app-date-picker')).toBeNull();
  });

  it('sends only the changed weight', async () => {
    const { component } = await create();
    component.form.patchValue({ total_weight: 50 });

    await component.submit();

    expect(stub.updateWeaging).toHaveBeenCalledWith('1', { total_weight: 50 });
  });

  it('sends only the changed destination and note', async () => {
    const { component } = await create();
    component.form.patchValue({ destination: '  Nave 3  ', note: '  nueva nota  ' });

    await component.submit();

    expect(stub.updateWeaging).toHaveBeenCalledWith('1', {
      destination: 'Nave 3',
      note: 'nueva nota',
    });
  });

  it('clears the weight when the field is emptied', async () => {
    const { component } = await create();
    component.form.patchValue({ total_weight: null });

    await component.submit();

    expect(stub.updateWeaging).toHaveBeenCalledWith('1', { clear_total_weight: true });
  });

  it('clears the destination when the field is emptied', async () => {
    const { component } = await create();
    component.form.patchValue({ destination: '   ' });

    await component.submit();

    expect(stub.updateWeaging).toHaveBeenCalledWith('1', { destination: '' });
  });

  it('clears the note when the field is emptied', async () => {
    const { component } = await create(buildWeaging({ note: 'nota' }));
    component.form.patchValue({ note: '' });

    await component.submit();

    expect(stub.updateWeaging).toHaveBeenCalledWith('1', { note: '' });
  });

  it('closes without calling the API when nothing changed', async () => {
    const { component } = await create();
    const closed = vi.fn();
    component.closed.subscribe(closed);

    await component.submit();

    expect(stub.updateWeaging).not.toHaveBeenCalled();
    expect(closed).toHaveBeenCalled();
  });

  it('does not submit when the weight is above 200 kg', async () => {
    const { component } = await create();
    component.form.patchValue({ total_weight: 250 });

    await component.submit();

    expect(stub.updateWeaging).not.toHaveBeenCalled();
  });

  it('does not submit when the destination is too long', async () => {
    const { component } = await create();
    component.form.patchValue({ destination: 'a'.repeat(101) });

    await component.submit();

    expect(stub.updateWeaging).not.toHaveBeenCalled();
  });

  it('does not submit when the note is too long', async () => {
    const { component } = await create();
    component.form.patchValue({ note: 'a'.repeat(501) });

    await component.submit();

    expect(stub.updateWeaging).not.toHaveBeenCalled();
  });

  it('shows the server message when the update fails', async () => {
    stub.updateWeaging = vi.fn(() =>
      throwError(
        () =>
          new HttpErrorResponse({
            status: 400,
            error: { error: 'El destino no es válido' },
          }),
      ),
    );
    const { component } = await create();
    component.form.patchValue({ destination: 'Nave 3' });

    await component.submit();

    expect(notifications.error).toHaveBeenCalledWith('El destino no es válido');
  });

  it('shows a fallback when the weaging is missing', async () => {
    stub.updateWeaging = vi.fn(() => throwError(() => new HttpErrorResponse({ status: 404 })));
    const { component } = await create();
    component.form.patchValue({ total_weight: 50 });

    await component.submit();

    expect(notifications.error).toHaveBeenCalledWith('El destete no existe');
  });

  it('shows a generic fallback for unexpected errors', async () => {
    stub.updateWeaging = vi.fn(() => throwError(() => new Error('failed')));
    const { component } = await create();
    component.form.patchValue({ total_weight: 50 });

    await component.submit();

    expect(notifications.error).toHaveBeenCalledWith('No se pudo actualizar el destete');
  });
});
