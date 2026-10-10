import { HttpErrorResponse } from '@angular/common/http';
import { TestBed } from '@angular/core/testing';
import { of, throwError } from 'rxjs';
import { vi } from 'vitest';

import { FarrowingsService } from '../../../core/farrowings/farrowings.service';
import { MedicationsService } from '../../../core/medications/medications.service';
import { NotificationService } from '../../../core/notifications/notification.service';
import { OperatorsService } from '../../../core/operators/operators.service';
import { SowsService } from '../../../core/sows/sows.service';
import { RegisterFarrowingModal } from './register-farrowing-modal';

class FarrowingsStub {
  createFarrowing = vi.fn((_request: Record<string, unknown>) => of({}));
}

class SowsStub {
  listSowDropdown = vi.fn(() => of([{ id: 'sow-1', code: 'C-001' }]));
}

class OperatorsStub {
  listOperators = vi.fn(() => of([{ id: 'op-1', name: 'Ana', active: true }]));
}

class MedicationsStub {
  listMedicationOptions = vi.fn(() => of([{ id: 'med-1', name: 'Oxitocina' }]));
}

class NotificationsStub {
  success = vi.fn();
  error = vi.fn();
}

describe('RegisterFarrowingModal', () => {
  let stub: FarrowingsStub;
  let notifications: NotificationsStub;

  beforeEach(async () => {
    stub = new FarrowingsStub();
    notifications = new NotificationsStub();
    await TestBed.configureTestingModule({
      imports: [RegisterFarrowingModal],
      providers: [
        { provide: FarrowingsService, useValue: stub },
        { provide: SowsService, useValue: new SowsStub() },
        { provide: OperatorsService, useValue: new OperatorsStub() },
        { provide: MedicationsService, useValue: new MedicationsStub() },
        { provide: NotificationService, useValue: notifications },
      ],
    }).compileComponents();
  });

  // eslint-disable-next-line @typescript-eslint/no-explicit-any
  const create = () => TestBed.createComponent(RegisterFarrowingModal).componentInstance as any;

  const fill = (component: any, date = '2026-04-20'): void => {
    component.form.patchValue({
      sow_id: 'sow-1',
      farrow_date: date,
      live_born: 10,
      stillborn: 1,
      mummified: 0,
    });
  };

  it('loads only gestating sow options for selection', async () => {
    const component = create();

    await component.loadLookups();

    const sows = TestBed.inject(SowsService);
    expect(sows.listSowDropdown).toHaveBeenCalledWith(true, ['Gestando']);
    expect(component.sowOptions()).toEqual([{ value: 'sow-1', label: 'C-001' }]);
  });

  it('does not submit when the form is invalid', async () => {
    const component = create();
    component.form.reset({ sow_id: '', farrow_date: '' });

    await component.submit();

    expect(stub.createFarrowing).not.toHaveBeenCalled();
  });

  it('creates a farrowing with counts and emits the sow code', async () => {
    const component = create();
    component.sowOptions.set([{ value: 'sow-1', label: 'C-001' }]);
    const created = vi.fn();
    component.created.subscribe(created);
    fill(component);

    await component.submit();

    expect(stub.createFarrowing).toHaveBeenCalledWith(
      expect.objectContaining({
        sow_id: 'sow-1',
        farrow_date: '2026-04-20',
        live_born: 10,
        stillborn: 1,
        mummified: 0,
        operators: [],
        medications: [],
      }),
    );
    expect(created).toHaveBeenCalledWith('C-001');
  });

  it('sends zero for counts left empty', async () => {
    const component = create();
    component.form.patchValue({ sow_id: 'sow-1', farrow_date: '2026-04-20' });

    await component.submit();

    expect(stub.createFarrowing).toHaveBeenCalledWith(
      expect.objectContaining({ live_born: 0, stillborn: 0, mummified: 0 }),
    );
  });

  it('includes the optional times, weights, location and note when provided', async () => {
    const component = create();
    fill(component);
    component.form.patchValue({
      start_time: '22:00',
      end_time: '02:00',
      location: 'Corral 3',
      litter_weight: 15.5,
      stillborn_weight: 2.5,
      note: 'Parto sin complicaciones',
    });

    await component.submit();

    expect(stub.createFarrowing).toHaveBeenCalledWith(
      expect.objectContaining({
        start_time: '22:00',
        end_time: '02:00',
        location: 'Corral 3',
        litter_weight: 15.5,
        stillborn_weight: 2.5,
        note: 'Parto sin complicaciones',
      }),
    );
  });

  it('includes operator and medication rows', async () => {
    const component = create();
    fill(component);
    component.addOperator();
    component.operatorsArray.at(0).patchValue({ operator_id: 'op-1' });
    component.addMedication();
    component.medicationsArray.at(0).patchValue({
      medication_id: 'med-1',
      dose: 2,
      applied_by: 'op-1',
    });

    await component.submit();

    expect(stub.createFarrowing).toHaveBeenCalledWith(
      expect.objectContaining({
        operators: [{ operator_id: 'op-1' }],
        medications: [{ medication_id: 'med-1', dose: 2, applied_by: 'op-1' }],
      }),
    );
  });

  it('accepts a decimal medication dose', async () => {
    const component = create();
    fill(component);
    component.addMedication();
    component.medicationsArray.at(0).patchValue({
      medication_id: 'med-1',
      dose: 0.5,
      applied_by: 'op-1',
    });

    await component.submit();

    expect(stub.createFarrowing).toHaveBeenCalledWith(
      expect.objectContaining({
        medications: [{ medication_id: 'med-1', dose: 0.5, applied_by: 'op-1' }],
      }),
    );
  });

  it('shows the server error message when creation fails', async () => {
    stub.createFarrowing = vi.fn(() =>
      throwError(
        () =>
          new HttpErrorResponse({
            status: 409,
            error: { error: 'La cerda no está gestando' },
          }),
      ),
    );
    const component = create();
    fill(component);

    await component.submit();

    expect(notifications.error).toHaveBeenCalledWith('La cerda no está gestando');
  });
});
