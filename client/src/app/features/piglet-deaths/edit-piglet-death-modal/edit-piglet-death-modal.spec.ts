import { HttpErrorResponse } from '@angular/common/http';
import { TestBed } from '@angular/core/testing';
import { of, throwError } from 'rxjs';
import { vi } from 'vitest';

import { NotificationService } from '../../../core/notifications/notification.service';
import { OperatorsService } from '../../../core/operators/operators.service';
import { PigletDeath } from '../../../core/piglet-deaths/piglet-death.models';
import { PigletDeathsService } from '../../../core/piglet-deaths/piglet-deaths.service';
import { EditPigletDeathModal } from './edit-piglet-death-modal';

function buildDeath(overrides: Partial<PigletDeath> = {}): PigletDeath {
  return {
    id: '42',
    farrowing_id: 'farrowing-1',
    sow_id: 'sow-1',
    sow_code: 'C-001',
    operator_id: 'op-1',
    operator_name: 'Ana',
    death_date: '2026-04-25',
    quantity: 2,
    weight: 2.5,
    cause: 'Aplastado',
    turn: 'Mañana',
    note: null,
    ...overrides,
  };
}

class PigletDeathsStub {
  updatePigletDeath = vi.fn((_id: string, _request: unknown) => of(undefined));
}

class OperatorsStub {
  listOperatorDropdown = vi.fn(() => of([{ id: 'op-1', name: 'Ana', active: true }]));
}

class NotificationsStub {
  success = vi.fn();
  error = vi.fn();
}

describe('EditPigletDeathModal', () => {
  let stub: PigletDeathsStub;
  let notifications: NotificationsStub;

  beforeEach(async () => {
    stub = new PigletDeathsStub();
    notifications = new NotificationsStub();
    await TestBed.configureTestingModule({
      imports: [EditPigletDeathModal],
      providers: [
        { provide: PigletDeathsService, useValue: stub },
        { provide: OperatorsService, useValue: new OperatorsStub() },
        { provide: NotificationService, useValue: notifications },
      ],
    }).compileComponents();
  });

  const create = async (death: PigletDeath = buildDeath()): Promise<any> => {
    const fixture = TestBed.createComponent(EditPigletDeathModal);
    fixture.componentRef.setInput('death', death);
    fixture.componentRef.setInput('sowCode', 'C-001');
    fixture.componentRef.setInput('open', true);
    fixture.detectChanges();
    await fixture.whenStable();
    // eslint-disable-next-line @typescript-eslint/no-explicit-any
    return fixture.componentInstance as any;
  };

  it('prefills the form and disables the locked fields', async () => {
    const fixture = TestBed.createComponent(EditPigletDeathModal);
    fixture.componentRef.setInput('death', buildDeath());
    fixture.componentRef.setInput('sowCode', 'C-001');
    fixture.componentRef.setInput('open', true);
    fixture.detectChanges();
    await fixture.whenStable();

    // eslint-disable-next-line @typescript-eslint/no-explicit-any
    const component = fixture.componentInstance as any;
    expect(component.form.getRawValue()).toEqual({
      sow_id: 'sow-1',
      operator_id: 'op-1',
      death_date: '2026-04-25',
      quantity: 2,
      weight: 2.5,
      cause: 'Aplastado',
      turn: 'Mañana',
      note: '',
    });

    for (const locked of ['sow_id', 'death_date', 'quantity'] as const) {
      expect(component.form.controls[locked].disabled).toBe(true);
    }
    expect(component.form.controls.weight.disabled).toBe(false);
    expect(component.form.controls.operator_id.disabled).toBe(false);

    const text = fixture.nativeElement.textContent as string;
    expect(text).toContain('C-001');
    expect(text).toContain('La cerda no se puede cambiar en la muerte.');
    expect(text).toContain('La fecha de la muerte no se puede cambiar.');
    expect(text).toContain('La cantidad de lechones no se puede cambiar.');
  });

  it('sends the update without the locked fields and emits when saved', async () => {
    const component = await create();
    const updated = vi.fn();
    component.updated.subscribe(updated);
    component.form.patchValue({ weight: 3, cause: 'Diarrea', turn: 'Tarde', note: 'nueva' });

    await component.submit();

    expect(stub.updatePigletDeath).toHaveBeenCalledWith(
      '42',
      expect.objectContaining({
        operator_id: 'op-1',
        weight: 3,
        cause: 'Diarrea',
        turn: 'Tarde',
        note: 'nueva',
      }),
    );
    // eslint-disable-next-line @typescript-eslint/no-explicit-any
    const request = stub.updatePigletDeath.mock.calls[0][1] as any;
    for (const locked of ['death_date', 'quantity', 'sow_id']) {
      expect(request).not.toHaveProperty(locked);
    }
    expect(updated).toHaveBeenCalled();
  });

  it('clears the weight when the stored weight is removed', async () => {
    const component = await create();
    component.form.controls.weight.setValue(null);

    await component.submit();

    expect(stub.updatePigletDeath).toHaveBeenCalledTimes(1);
    const request = stub.updatePigletDeath.mock.calls[0][1] as Record<string, unknown>;
    expect(request).toMatchObject({ clear_weight: true });
    expect(request).not.toHaveProperty('weight');
  });

  it('does not send the weight when it was already empty', async () => {
    const component = await create(buildDeath({ weight: null }));

    await component.submit();

    expect(stub.updatePigletDeath).toHaveBeenCalledTimes(1);
    const request = stub.updatePigletDeath.mock.calls[0][1] as Record<string, unknown>;
    expect(request).not.toHaveProperty('weight');
    expect(request).not.toHaveProperty('clear_weight');
  });

  it('offers Sin Turno as a turn option', async () => {
    const component = await create();

    expect(component.turnOptions).toContainEqual({ value: 'Sin Turno', label: 'Sin Turno' });
  });

  it('offers Desaparecido as a cause option', async () => {
    const component = await create();

    expect(component.causeOptions).toContainEqual({ value: 'Desaparecido', label: 'Desaparecido' });
  });

  it('sends the Desaparecido cause', async () => {
    const component = await create();
    component.form.patchValue({ cause: 'Desaparecido' });

    await component.submit();

    expect(stub.updatePigletDeath).toHaveBeenCalledWith(
      '42',
      expect.objectContaining({ cause: 'Desaparecido' }),
    );
  });

  it('rejects a weight greater than 20 kg', async () => {
    const component = await create();
    component.form.controls.weight.setValue(20.01);

    await component.submit();

    expect(component.form.controls.weight.hasError('max')).toBe(true);
    expect(stub.updatePigletDeath).not.toHaveBeenCalled();
  });

  it('rejects a note longer than 500 characters', async () => {
    const component = await create();
    component.form.controls.note.setValue('a'.repeat(501));

    await component.submit();

    expect(component.form.controls.note.hasError('maxlength')).toBe(true);
    expect(stub.updatePigletDeath).not.toHaveBeenCalled();
  });

  it('shows the server message when the update fails', async () => {
    stub.updatePigletDeath = vi.fn(() =>
      throwError(
        () =>
          new HttpErrorResponse({
            status: 400,
            error: {
              error: 'La fecha, la cantidad y la cerda de la muerte no se pueden editar',
            },
          }),
      ),
    );
    const component = await create();

    await component.submit();

    expect(notifications.error).toHaveBeenCalledWith(
      'La fecha, la cantidad y la cerda de la muerte no se pueden editar',
    );
  });

  it('shows a fallback when the death is missing', async () => {
    stub.updatePigletDeath = vi.fn(() => throwError(() => new HttpErrorResponse({ status: 404 })));
    const component = await create();

    await component.submit();

    expect(notifications.error).toHaveBeenCalledWith('La muerte no existe');
  });
});
