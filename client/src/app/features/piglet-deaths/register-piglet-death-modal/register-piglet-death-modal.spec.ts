import { HttpErrorResponse } from '@angular/common/http';
import { TestBed } from '@angular/core/testing';
import { of, throwError } from 'rxjs';
import { vi } from 'vitest';

import { NotificationService } from '../../../core/notifications/notification.service';
import { OperatorsService } from '../../../core/operators/operators.service';
import { PigletDeathsService } from '../../../core/piglet-deaths/piglet-deaths.service';
import { SowsService } from '../../../core/sows/sows.service';
import { RegisterPigletDeathModal } from './register-piglet-death-modal';

class PigletDeathsStub {
  createPigletDeath = vi.fn((_request: Record<string, unknown>) => of({}));
}

class SowsStub {
  listSows = vi.fn(() => of([{ id: 'sow-1', code: 'C-001' }]));
}

class OperatorsStub {
  listOperators = vi.fn(() => of([{ id: 'operator-1', name: 'Operador', active: true }]));
}

class NotificationsStub {
  success = vi.fn();
  error = vi.fn();
}

describe('RegisterPigletDeathModal', () => {
  let stub: PigletDeathsStub;
  let sows: SowsStub;
  let operators: OperatorsStub;
  let notifications: NotificationsStub;

  beforeEach(async () => {
    stub = new PigletDeathsStub();
    sows = new SowsStub();
    operators = new OperatorsStub();
    notifications = new NotificationsStub();
    await TestBed.configureTestingModule({
      imports: [RegisterPigletDeathModal],
      providers: [
        { provide: PigletDeathsService, useValue: stub },
        { provide: SowsService, useValue: sows },
        { provide: OperatorsService, useValue: operators },
        { provide: NotificationService, useValue: notifications },
      ],
    }).compileComponents();
  });

  // eslint-disable-next-line @typescript-eslint/no-explicit-any
  const create = () => TestBed.createComponent(RegisterPigletDeathModal).componentInstance as any;

  const fill = (component: any): void => {
    component.form.patchValue({
      sow_id: 'sow-1',
      operator_id: 'operator-1',
      death_date: '2026-04-25',
      quantity: 2,
      weight: 2.5,
      cause: 'Aplastado',
      turn: 'Mañana',
      note: '',
    });
  };

  it('loads the lactating sows and active operators for selection', async () => {
    const component = create();

    await component.loadOptions();

    expect(sows.listSows).toHaveBeenCalledWith({ state: 'Lactando' });
    expect(component.sowOptions()).toEqual([{ value: 'sow-1', label: 'C-001' }]);
    expect(component.operatorOptions()).toEqual([{ value: 'operator-1', label: 'Operador' }]);
  });

  it('does not submit when the form is empty', async () => {
    const component = create();
    component.form.reset({
      sow_id: '',
      operator_id: '',
      death_date: '',
      quantity: 0,
      cause: '',
      turn: '',
      note: '',
    });

    await component.submit();

    expect(stub.createPigletDeath).not.toHaveBeenCalled();
  });

  it('creates a piglet death and emits the sow code', async () => {
    const component = create();
    component.sowOptions.set([{ value: 'sow-1', label: 'C-001' }]);
    const created = vi.fn();
    component.created.subscribe(created);
    fill(component);

    await component.submit();

    expect(stub.createPigletDeath).toHaveBeenCalledWith({
      sow_id: 'sow-1',
      operator_id: 'operator-1',
      death_date: '2026-04-25',
      quantity: 2,
      weight: 2.5,
      cause: 'Aplastado',
      turn: 'Mañana',
    });
    expect(created).toHaveBeenCalledWith('C-001');
  });

  it('omits the weight when it is empty', async () => {
    const component = create();
    component.sowOptions.set([{ value: 'sow-1', label: 'C-001' }]);
    fill(component);
    component.form.patchValue({ weight: null });

    await component.submit();

    expect(stub.createPigletDeath).toHaveBeenCalledWith(
      expect.not.objectContaining({ weight: expect.anything() }),
    );
  });

  it('includes the optional note when provided', async () => {
    const component = create();
    component.sowOptions.set([{ value: 'sow-1', label: 'C-001' }]);
    fill(component);
    component.form.patchValue({ note: 'muerte durante la noche' });

    await component.submit();

    expect(stub.createPigletDeath).toHaveBeenCalledWith(
      expect.objectContaining({ note: 'muerte durante la noche' }),
    );
  });

  it('shows the server error message when creation fails', async () => {
    stub.createPigletDeath = vi.fn(() =>
      throwError(
        () =>
          new HttpErrorResponse({
            status: 409,
            error: { error: 'La cerda no está lactando' },
          }),
      ),
    );
    const component = create();
    component.sowOptions.set([{ value: 'sow-1', label: 'C-001' }]);
    fill(component);

    await component.submit();

    expect(notifications.error).toHaveBeenCalledWith('La cerda no está lactando');
  });
});
