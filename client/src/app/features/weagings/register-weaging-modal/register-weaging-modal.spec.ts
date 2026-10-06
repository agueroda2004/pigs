import { HttpErrorResponse } from '@angular/common/http';
import { TestBed } from '@angular/core/testing';
import { of, throwError } from 'rxjs';
import { vi } from 'vitest';

import { NotificationService } from '../../../core/notifications/notification.service';
import { SowsService } from '../../../core/sows/sows.service';
import { WeagingsService } from '../../../core/weagings/weagings.service';
import { RegisterWeagingModal } from './register-weaging-modal';

class WeagingsStub {
  createWeaging = vi.fn((_request: Record<string, unknown>) => of({}));
}

class SowsStub {
  listSows = vi.fn(() =>
    of([
      { id: 'sow-1', code: 'C-001' },
      { id: 'sow-2', code: 'C-002' },
    ]),
  );
}

class NotificationsStub {
  success = vi.fn();
  error = vi.fn();
}

describe('RegisterWeagingModal', () => {
  let stub: WeagingsStub;
  let sows: SowsStub;
  let notifications: NotificationsStub;

  beforeEach(async () => {
    stub = new WeagingsStub();
    sows = new SowsStub();
    notifications = new NotificationsStub();
    await TestBed.configureTestingModule({
      imports: [RegisterWeagingModal],
      providers: [
        { provide: WeagingsService, useValue: stub },
        { provide: SowsService, useValue: sows },
        { provide: NotificationService, useValue: notifications },
      ],
    }).compileComponents();
  });

  // eslint-disable-next-line @typescript-eslint/no-explicit-any
  const create = () => TestBed.createComponent(RegisterWeagingModal).componentInstance as any;

  const fill = (component: any): void => {
    component.form.patchValue({
      sow_id: 'sow-1',
      weaging_date: '2026-04-25',
      quantity: 8,
      total_weight: null,
      destination: '',
      note: '',
    });
  };

  it('loads the lactating sows for selection', async () => {
    const component = create();

    await component.loadOptions();

    expect(sows.listSows).toHaveBeenCalledWith({ state: 'Lactando' });
    expect(component.sowOptions()).toEqual([
      { value: 'sow-1', label: 'C-001' },
      { value: 'sow-2', label: 'C-002' },
    ]);
  });

  it('does not submit when the form is empty', async () => {
    const component = create();
    component.form.reset({
      sow_id: '',
      weaging_date: '',
      quantity: 0,
      total_weight: null,
      destination: '',
      note: '',
    });

    await component.submit();

    expect(stub.createWeaging).not.toHaveBeenCalled();
  });

  it('creates a weaging and emits the sow code', async () => {
    const component = create();
    component.sowOptions.set([{ value: 'sow-1', label: 'C-001' }]);
    const created = vi.fn();
    component.created.subscribe(created);
    fill(component);

    await component.submit();

    expect(stub.createWeaging).toHaveBeenCalledWith({
      sow_id: 'sow-1',
      weaging_date: '2026-04-25',
      quantity: 8,
    });
    expect(created).toHaveBeenCalledWith('C-001');
  });

  it('includes the optional weight, destination and note when provided', async () => {
    const component = create();
    fill(component);
    component.form.patchValue({
      total_weight: 120.5,
      destination: 'Nave 2',
      note: 'camada numerosa',
    });

    await component.submit();

    expect(stub.createWeaging).toHaveBeenCalledWith(
      expect.objectContaining({
        total_weight: 120.5,
        destination: 'Nave 2',
        note: 'camada numerosa',
      }),
    );
  });

  it('shows the server error message when creation fails', async () => {
    stub.createWeaging = vi.fn(() =>
      throwError(
        () =>
          new HttpErrorResponse({
            status: 409,
            error: { error: 'La cantidad del destete debe coincidir con los lechones actuales' },
          }),
      ),
    );
    const component = create();
    fill(component);

    await component.submit();

    expect(notifications.error).toHaveBeenCalledWith(
      'La cantidad del destete debe coincidir con los lechones actuales',
    );
  });
});
