import { HttpErrorResponse } from '@angular/common/http';
import { TestBed } from '@angular/core/testing';
import { of, throwError } from 'rxjs';
import { vi } from 'vitest';

import { NotificationService } from '../../../core/notifications/notification.service';
import {
  PartialWeaging,
  UpdatePartialWeagingRequest,
} from '../../../core/partial-weagings/partial-weaging.models';
import { PartialWeagingsService } from '../../../core/partial-weagings/partial-weagings.service';
import { EditPartialWeagingModal } from './edit-partial-weaging-modal';

function buildWeaging(overrides: Partial<PartialWeaging> = {}): PartialWeaging {
  return {
    id: '1',
    farrowing_id: 'farrowing-1',
    sow_id: 'sow-1',
    sow_code: 'C-001',
    weaging_date: '2026-04-25',
    quantity: 2,
    total_weight: 42.5,
    type: 'Normal',
    note: null,
    ...overrides,
  };
}

class PartialWeagingsStub {
  updatePartialWeaging = vi.fn((_id: string, _request: UpdatePartialWeagingRequest) => of(void 0));
}

class NotificationsStub {
  success = vi.fn();
  error = vi.fn();
}

describe('EditPartialWeagingModal', () => {
  let stub: PartialWeagingsStub;
  let notifications: NotificationsStub;

  beforeEach(async () => {
    stub = new PartialWeagingsStub();
    notifications = new NotificationsStub();
    await TestBed.configureTestingModule({
      imports: [EditPartialWeagingModal],
      providers: [
        { provide: PartialWeagingsService, useValue: stub },
        { provide: NotificationService, useValue: notifications },
      ],
    }).compileComponents();
  });

  const create = async (weaging: PartialWeaging = buildWeaging()) => {
    const fixture = TestBed.createComponent(EditPartialWeagingModal);
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
      note: 'nota',
    });

    const text = fixture.nativeElement.textContent.replace(/\s+/g, ' ');
    expect(text).toContain('C-001');
    expect(text).toContain('2026-04-25');
    expect(text).toContain('Normal');
    expect(text).toContain('Cantidad');
    expect(text).toContain('2');
    expect(text).toContain(
      'La cerda, la fecha, el tipo y la cantidad no se pueden editar para mantener la integridad de los datos.',
    );
  });

  it('does not expose the date as an editable field', async () => {
    const { fixture, component } = await create();

    expect(component.form.contains('weaging_date')).toBe(false);
    expect(fixture.nativeElement.querySelector('app-date-picker')).toBeNull();
  });

  it('sends only the changed weight', async () => {
    const { component } = await create();
    component.form.patchValue({ total_weight: 50 });

    await component.submit();

    expect(stub.updatePartialWeaging).toHaveBeenCalledWith('1', { total_weight: 50 });
  });

  it('clears the weight when the field is emptied', async () => {
    const { component } = await create();
    component.form.patchValue({ total_weight: null });

    await component.submit();

    expect(stub.updatePartialWeaging).toHaveBeenCalledWith('1', {
      clear_total_weight: true,
    });
  });

  it('clears the note when the field is emptied', async () => {
    const { component } = await create(buildWeaging({ note: 'nota' }));
    component.form.patchValue({ note: '' });

    await component.submit();

    expect(stub.updatePartialWeaging).toHaveBeenCalledWith('1', { note: '' });
  });

  it('closes without calling the API when nothing changed', async () => {
    const { component } = await create();
    const closed = vi.fn();
    component.closed.subscribe(closed);

    await component.submit();

    expect(stub.updatePartialWeaging).not.toHaveBeenCalled();
    expect(closed).toHaveBeenCalled();
  });

  it('does not submit when the weight is above 200 kg', async () => {
    const { component } = await create();
    component.form.patchValue({ total_weight: 250 });

    await component.submit();

    expect(stub.updatePartialWeaging).not.toHaveBeenCalled();
  });

  it('shows the server message when the update fails', async () => {
    stub.updatePartialWeaging = vi.fn(() =>
      throwError(
        () =>
          new HttpErrorResponse({
            status: 400,
            error: { error: 'La nota no es válida' },
          }),
      ),
    );
    const { component } = await create();
    component.form.patchValue({ note: 'nota' });

    await component.submit();

    expect(notifications.error).toHaveBeenCalledWith('La nota no es válida');
  });

  it('shows a fallback when the weaging is missing', async () => {
    stub.updatePartialWeaging = vi.fn(() =>
      throwError(() => new HttpErrorResponse({ status: 404 })),
    );
    const { component } = await create();
    component.form.patchValue({ total_weight: 50 });

    await component.submit();

    expect(notifications.error).toHaveBeenCalledWith('El destete parcial no existe');
  });

  it('shows a generic fallback for unexpected errors', async () => {
    stub.updatePartialWeaging = vi.fn(() => throwError(() => new Error('failed')));
    const { component } = await create();
    component.form.patchValue({ total_weight: 50 });

    await component.submit();

    expect(notifications.error).toHaveBeenCalledWith('No se pudo actualizar el destete parcial');
  });
});
