import { HttpErrorResponse } from '@angular/common/http';
import { TestBed } from '@angular/core/testing';
import { of, throwError } from 'rxjs';
import { vi } from 'vitest';

import { BoarsService } from '../../../core/boars/boars.service';
import { NotificationService } from '../../../core/notifications/notification.service';
import { OperatorsService } from '../../../core/operators/operators.service';
import { ServicesService } from '../../../core/services/services.service';
import { SowsService } from '../../../core/sows/sows.service';
import { CreateServiceModal } from './create-service-modal';

class ServicesStub {
  createService = vi.fn((_request: Record<string, unknown>) => of({}));
}

class SowsStub {
  listSows = vi.fn(() => of([]));
}

class BoarsStub {
  listBoars = vi.fn(() => of([]));
}

class OperatorsStub {
  listOperators = vi.fn(() => of([]));
}

class NotificationsStub {
  success = vi.fn();
  error = vi.fn();
}

describe('CreateServiceModal', () => {
  let stub: ServicesStub;
  let notifications: NotificationsStub;

  beforeEach(async () => {
    stub = new ServicesStub();
    notifications = new NotificationsStub();
    await TestBed.configureTestingModule({
      imports: [CreateServiceModal],
      providers: [
        { provide: ServicesService, useValue: stub },
        { provide: SowsService, useValue: new SowsStub() },
        { provide: BoarsService, useValue: new BoarsStub() },
        { provide: OperatorsService, useValue: new OperatorsStub() },
        { provide: NotificationService, useValue: notifications },
      ],
    }).compileComponents();
  });

  // eslint-disable-next-line @typescript-eslint/no-explicit-any
  const create = () => TestBed.createComponent(CreateServiceModal).componentInstance as any;

  const fillFirstMount = (component: any, date: string): void => {
    component.form.patchValue({ sow_id: 'sow-1', location: '', note: '' });
    component.mounts.at(0).patchValue({
      mount_date: date,
      boar_id: 'boar-1',
      operator_id: 'operator-1',
      type: 'Artificial',
      note: '',
    });
  };

  it('does not submit when the form is empty', async () => {
    const component = create();

    await component.submit();

    expect(stub.createService).not.toHaveBeenCalled();
  });

  it('creates a service and emits the sow code', async () => {
    const component = create();
    component.sowOptions.set([{ value: 'sow-1', label: 'C-001' }]);
    const created = vi.fn();
    component.created.subscribe(created);
    fillFirstMount(component, '2025-01-10');

    await component.submit();

    expect(stub.createService).toHaveBeenCalledWith({
      sow_id: 'sow-1',
      mounts: [
        {
          boar_id: 'boar-1',
          operator_id: 'operator-1',
          mount_date: '2025-01-10',
          type: 'Artificial',
        },
      ],
    });
    expect(created).toHaveBeenCalledWith('C-001');
  });

  it('includes optional fields when provided', async () => {
    const component = create();
    component.sowOptions.set([{ value: 'sow-1', label: 'C-001' }]);
    fillFirstMount(component, '2025-01-10');
    component.form.patchValue({ location: 'Nave 1', note: 'ok' });
    component.mounts.at(0).patchValue({ note: 'monta' });

    await component.submit();

    expect(stub.createService).toHaveBeenCalledWith({
      sow_id: 'sow-1',
      location: 'Nave 1',
      note: 'ok',
      mounts: [
        {
          boar_id: 'boar-1',
          operator_id: 'operator-1',
          mount_date: '2025-01-10',
          type: 'Artificial',
          note: 'monta',
        },
      ],
    });
  });

  it('rejects out-of-order mount dates', async () => {
    const component = create();
    component.addMount();
    fillFirstMount(component, '2025-01-11');
    component.mounts.at(1).patchValue({
      mount_date: '2025-01-10',
      boar_id: 'boar-1',
      operator_id: 'operator-1',
      type: 'Artificial',
      note: '',
    });

    await component.submit();

    expect(component.scheduleError()).toContain('orden ascendente');
    expect(stub.createService).not.toHaveBeenCalled();
  });

  it('rejects a mount gap larger than 24 hours', async () => {
    const component = create();
    component.addMount();
    fillFirstMount(component, '2025-01-10');
    component.mounts.at(1).patchValue({
      mount_date: '2025-01-12',
      boar_id: 'boar-1',
      operator_id: 'operator-1',
      type: 'Artificial',
      note: '',
    });

    await component.submit();

    expect(component.scheduleError()).toContain('24 horas');
    expect(stub.createService).not.toHaveBeenCalled();
  });

  it('rejects a future mount date', async () => {
    const component = create();
    fillFirstMount(component, '2099-01-01');

    await component.submit();

    expect(component.scheduleError()).toContain('futura');
    expect(stub.createService).not.toHaveBeenCalled();
  });

  it('does not allow more than three mounts', () => {
    const component = create();

    component.addMount();
    component.addMount();
    component.addMount();

    expect(component.mounts.length).toBe(3);
  });

  it('shows the server error message when creation fails', async () => {
    stub.createService = vi.fn(() =>
      throwError(
        () =>
          new HttpErrorResponse({
            status: 409,
            error: { error: 'La cerda no está en un estado válido para el servicio' },
          }),
      ),
    );
    const component = create();
    component.sowOptions.set([{ value: 'sow-1', label: 'C-001' }]);
    fillFirstMount(component, '2025-01-10');

    await component.submit();

    expect(notifications.error).toHaveBeenCalledWith(
      'La cerda no está en un estado válido para el servicio',
    );
  });
});
