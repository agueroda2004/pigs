import { HttpErrorResponse } from '@angular/common/http';
import { TestBed } from '@angular/core/testing';
import { of, throwError } from 'rxjs';
import { vi } from 'vitest';

import { NotificationService } from '../../../core/notifications/notification.service';
import { SowRemovalsService } from '../../../core/sow-removals/sow-removals.service';
import { SowsService } from '../../../core/sows/sows.service';
import { RegisterSowRemovalModal } from './register-sow-removal-modal';

class SowRemovalsStub {
  createSowRemoval = vi.fn((_request: Record<string, unknown>) => of({}));
}

class SowsStub {
  listSowOptions = vi.fn(() => of([{ id: 'sow-1', code: 'C-001' }]));
}

class NotificationsStub {
  success = vi.fn();
  error = vi.fn();
}

describe('RegisterSowRemovalModal', () => {
  let stub: SowRemovalsStub;
  let sows: SowsStub;
  let notifications: NotificationsStub;

  beforeEach(async () => {
    stub = new SowRemovalsStub();
    sows = new SowsStub();
    notifications = new NotificationsStub();
    await TestBed.configureTestingModule({
      imports: [RegisterSowRemovalModal],
      providers: [
        { provide: SowRemovalsService, useValue: stub },
        { provide: SowsService, useValue: sows },
        { provide: NotificationService, useValue: notifications },
      ],
    }).compileComponents();
  });

  // eslint-disable-next-line @typescript-eslint/no-explicit-any
  const create = () => TestBed.createComponent(RegisterSowRemovalModal).componentInstance as any;

  const fill = (component: any, date: string): void => {
    component.form.patchValue({
      sow_id: 'sow-1',
      removal_date: date,
      type: 'Muerte',
      reason: 'Enfermedad',
      note: '',
    });
  };

  it('loads the eligible sow options for selection', async () => {
    const component = create();

    await component.loadOptions();

    expect(sows.listSowOptions).toHaveBeenCalledWith(true);
    expect(component.sowOptions()).toEqual([{ value: 'sow-1', label: 'C-001' }]);
  });

  it('does not submit when the form is empty', async () => {
    const component = create();
    component.form.reset({
      sow_id: '',
      removal_date: '',
      type: '',
      reason: '',
      note: '',
    });

    await component.submit();

    expect(stub.createSowRemoval).not.toHaveBeenCalled();
  });

  it('creates a removal and emits the sow code', async () => {
    const component = create();
    component.sowOptions.set([{ value: 'sow-1', label: 'C-001' }]);
    const created = vi.fn();
    component.created.subscribe(created);
    fill(component, '2026-01-20');

    await component.submit();

    expect(stub.createSowRemoval).toHaveBeenCalledWith({
      sow_id: 'sow-1',
      removal_date: '2026-01-20',
      type: 'Muerte',
      reason: 'Enfermedad',
    });
    expect(created).toHaveBeenCalledWith('C-001');
  });

  it('includes the optional note when provided', async () => {
    const component = create();
    component.sowOptions.set([{ value: 'sow-1', label: 'C-001' }]);
    fill(component, '2026-01-20');
    component.form.patchValue({ note: 'baja por enfermedad' });

    await component.submit();

    expect(stub.createSowRemoval).toHaveBeenCalledWith({
      sow_id: 'sow-1',
      removal_date: '2026-01-20',
      type: 'Muerte',
      reason: 'Enfermedad',
      note: 'baja por enfermedad',
    });
  });

  it('shows the server error message when creation fails', async () => {
    stub.createSowRemoval = vi.fn(() =>
      throwError(
        () =>
          new HttpErrorResponse({
            status: 409,
            error: { error: 'La cerda no está en un estado válido para la baja' },
          }),
      ),
    );
    const component = create();
    component.sowOptions.set([{ value: 'sow-1', label: 'C-001' }]);
    fill(component, '2026-01-20');

    await component.submit();

    expect(notifications.error).toHaveBeenCalledWith(
      'La cerda no está en un estado válido para la baja',
    );
  });
});
