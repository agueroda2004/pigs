import { HttpErrorResponse } from '@angular/common/http';
import { TestBed } from '@angular/core/testing';
import { of, throwError } from 'rxjs';
import { vi } from 'vitest';

import { Farrowing } from '../../../core/farrowings/farrowing.models';
import { FarrowingsService } from '../../../core/farrowings/farrowings.service';
import { MedicationsService } from '../../../core/medications/medications.service';
import { NotificationService } from '../../../core/notifications/notification.service';
import { OperatorsService } from '../../../core/operators/operators.service';
import { EditFarrowingModal } from './edit-farrowing-modal';

function buildFarrowing(overrides: Partial<Farrowing> = {}): Farrowing {
  return {
    id: '42',
    sow_id: 'sow-1',
    sow_code: 'C-001',
    service_id: 'service-1',
    farrow_date: '2026-04-20',
    start_time: '22:00',
    end_time: '02:00',
    location: 'Corral 3',
    live_born: 10,
    stillborn: 1,
    mummified: 0,
    current_piglets: 10,
    litter_weight: 15.5,
    stillborn_weight: 1.5,
    is_manipulated: false,
    is_nurse: false,
    nurse_start_date: null,
    note: null,
    operators: [{ id: 'fo-1', operator_id: 'op-1' }],
    medications: [{ id: 'fm-1', medication_id: 'med-1', dose: 2, applied_by: 'op-1' }],
    ...overrides,
  };
}

class FarrowingsStub {
  updateFarrowing = vi.fn((_id: string, _request: unknown) => of(undefined));
}

class OperatorsStub {
  listOperatorDropdown = vi.fn(() => of([{ id: 'op-1', name: 'Ana', active: true }]));
}

class MedicationsStub {
  listMedications = vi.fn(() => of([{ id: 'med-1', name: 'Oxitocina', active: true }]));
}

class NotificationsStub {
  success = vi.fn();
  error = vi.fn();
}

describe('EditFarrowingModal', () => {
  let stub: FarrowingsStub;
  let notifications: NotificationsStub;

  beforeEach(async () => {
    stub = new FarrowingsStub();
    notifications = new NotificationsStub();
    await TestBed.configureTestingModule({
      imports: [EditFarrowingModal],
      providers: [
        { provide: FarrowingsService, useValue: stub },
        { provide: OperatorsService, useValue: new OperatorsStub() },
        { provide: MedicationsService, useValue: new MedicationsStub() },
        { provide: NotificationService, useValue: notifications },
      ],
    }).compileComponents();
  });

  const create = async (farrowing: Farrowing = buildFarrowing()): Promise<any> => {
    const fixture = TestBed.createComponent(EditFarrowingModal);
    fixture.componentRef.setInput('farrowing', farrowing);
    fixture.componentRef.setInput('open', true);
    fixture.detectChanges();
    await fixture.whenStable();
    // eslint-disable-next-line @typescript-eslint/no-explicit-any
    return fixture.componentInstance as any;
  };

  it('prefills the form, operators and medications from the farrowing', async () => {
    const component = await create();

    expect(component.form.getRawValue()).toMatchObject({
      farrow_date: '2026-04-20',
      start_time: '22:00',
      end_time: '02:00',
      location: 'Corral 3',
      live_born: 10,
      stillborn: 1,
      litter_weight: 15.5,
    });
    expect(component.operatorsArray.length).toBe(1);
    expect(component.operatorsArray.at(0).getRawValue()).toEqual({ operator_id: 'op-1' });
    expect(component.medicationsArray.at(0).getRawValue()).toEqual({
      medication_id: 'med-1',
      dose: 2,
      applied_by: 'op-1',
    });
  });

  it('shows the locked sow code and its helper text', async () => {
    const fixture = TestBed.createComponent(EditFarrowingModal);
    fixture.componentRef.setInput('farrowing', buildFarrowing());
    fixture.componentRef.setInput('open', true);
    fixture.detectChanges();
    await fixture.whenStable();

    const text = fixture.nativeElement.textContent as string;
    expect(text).toContain('C-001');
    expect(text).toContain('El código de la cerda no se puede cambiar en el parto.');
  });

  it('labels inactive operators and medications', async () => {
    const operators = TestBed.inject(OperatorsService) as unknown as OperatorsStub;
    const medications = TestBed.inject(MedicationsService) as unknown as MedicationsStub;
    operators.listOperatorDropdown = vi.fn(() => of([{ id: 'op-1', name: 'Ana', active: false }]));
    medications.listMedications = vi.fn(() =>
      of([{ id: 'med-1', name: 'Oxitocina', active: false }]),
    );
    const component = await create();

    await component.loadOptions();

    expect(component.operatorOptions()).toEqual([{ value: 'op-1', label: 'Ana (inactivo)' }]);
    expect(component.medicationOptions()).toEqual([
      { value: 'med-1', label: 'Oxitocina (inactivo)' },
    ]);
  });

  it('keeps the date and the piglet counts disabled', async () => {
    const component = await create();

    for (const name of ['farrow_date', 'live_born', 'stillborn', 'mummified'] as const) {
      expect(component.form.controls[name].disabled).toBe(true);
    }
  });

  it('sends the update request without the locked fields and emits when saved', async () => {
    const component = await create();
    const updated = vi.fn();
    component.updated.subscribe(updated);
    component.form.patchValue({ location: 'Corral 9', stillborn_weight: 2.5 });

    await component.submit();

    expect(stub.updateFarrowing).toHaveBeenCalledWith(
      '42',
      expect.objectContaining({
        location: 'Corral 9',
        litter_weight: 15.5,
        stillborn_weight: 2.5,
        operators: [{ operator_id: 'op-1' }],
        medications: [{ medication_id: 'med-1', dose: 2, applied_by: 'op-1' }],
      }),
    );
    // eslint-disable-next-line @typescript-eslint/no-explicit-any
    const request = stub.updateFarrowing.mock.calls[0][1] as any;
    for (const locked of ['farrow_date', 'live_born', 'stillborn', 'mummified']) {
      expect(request).not.toHaveProperty(locked);
    }
    expect(request).not.toHaveProperty('clear_litter_weight');
    expect(request).not.toHaveProperty('clear_stillborn_weight');
    expect(updated).toHaveBeenCalled();
  });

  it('requires both weights', async () => {
    for (const field of ['litter_weight', 'stillborn_weight'] as const) {
      stub.updateFarrowing.mockClear();
      const component = await create();
      component.form.controls[field].setValue(null);

      await component.submit();

      expect(component.form.controls[field].hasError('required')).toBe(true);
      expect(stub.updateFarrowing).not.toHaveBeenCalled();
    }
  });

  it('rejects weights greater than 30', async () => {
    for (const field of ['litter_weight', 'stillborn_weight'] as const) {
      stub.updateFarrowing.mockClear();
      const component = await create();
      component.form.controls[field].setValue(31);

      await component.submit();

      expect(component.form.controls[field].hasError('max')).toBe(true);
      expect(stub.updateFarrowing).not.toHaveBeenCalled();
    }
  });

  it('rejects a medication dose greater than 30', async () => {
    const component = await create();
    component.medicationsArray.at(0).controls.dose.setValue(31);

    await component.submit();

    expect(component.medicationsArray.at(0).controls.dose.hasError('max')).toBe(true);
    expect(stub.updateFarrowing).not.toHaveBeenCalled();
  });

  it('does not submit when the form is invalid', async () => {
    const component = await create();
    component.form.controls.litter_weight.setValue(null);

    await component.submit();

    expect(stub.updateFarrowing).not.toHaveBeenCalled();
  });

  it('shows the server message when the update fails', async () => {
    stub.updateFarrowing = vi.fn(() =>
      throwError(
        () =>
          new HttpErrorResponse({
            status: 400,
            error: {
              error: 'La fecha y las cantidades de lechones del parto no se pueden editar',
            },
          }),
      ),
    );
    const component = await create();

    await component.submit();

    expect(notifications.error).toHaveBeenCalledWith(
      'La fecha y las cantidades de lechones del parto no se pueden editar',
    );
  });
});
