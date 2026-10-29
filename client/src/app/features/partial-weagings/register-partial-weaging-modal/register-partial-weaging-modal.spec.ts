import { HttpErrorResponse } from '@angular/common/http';
import { TestBed } from '@angular/core/testing';
import { of, throwError } from 'rxjs';
import { vi } from 'vitest';

import { NotificationService } from '../../../core/notifications/notification.service';
import { PartialWeagingsService } from '../../../core/partial-weagings/partial-weagings.service';
import { SowsService } from '../../../core/sows/sows.service';
import { RegisterPartialWeagingModal } from './register-partial-weaging-modal';

class PartialWeagingsStub {
  createPartialWeaging = vi.fn((_request: Record<string, unknown>) => of(void 0));
}

class SowsStub {
  listSowDropdown = vi.fn(() =>
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

describe('RegisterPartialWeagingModal', () => {
  let stub: PartialWeagingsStub;
  let sows: SowsStub;
  let notifications: NotificationsStub;

  beforeEach(async () => {
    stub = new PartialWeagingsStub();
    sows = new SowsStub();
    notifications = new NotificationsStub();
    await TestBed.configureTestingModule({
      imports: [RegisterPartialWeagingModal],
      providers: [
        { provide: PartialWeagingsService, useValue: stub },
        { provide: SowsService, useValue: sows },
        { provide: NotificationService, useValue: notifications },
      ],
    }).compileComponents();
  });

  // eslint-disable-next-line @typescript-eslint/no-explicit-any
  const create = () =>
    TestBed.createComponent(RegisterPartialWeagingModal).componentInstance as any;

  const fill = (component: any): void => {
    component.form.patchValue({
      sow_id: 'sow-1',
      weaging_date: '2026-04-25',
      quantity: 2,
      type: 'Normal',
      total_weight: null,
      note: '',
    });
  };

  it('loads the lactating sows for selection', async () => {
    const component = create();

    await component.loadOptions();

    expect(sows.listSowDropdown).toHaveBeenCalledWith(true, ['Lactando']);
    expect(component.sowOptions()).toEqual([
      { value: 'sow-1', label: 'C-001' },
      { value: 'sow-2', label: 'C-002' },
    ]);
  });

  it('does not submit when the quantity is above 30', async () => {
    const component = create();
    fill(component);
    component.form.patchValue({ quantity: 31 });

    await component.submit();

    expect(stub.createPartialWeaging).not.toHaveBeenCalled();
    expect(component.form.controls.quantity.hasError('max')).toBe(true);
  });

  it('does not submit when the weight is above 200', async () => {
    const component = create();
    fill(component);
    component.form.patchValue({ total_weight: 200.5 });

    await component.submit();

    expect(stub.createPartialWeaging).not.toHaveBeenCalled();
    expect(component.form.controls.total_weight.hasError('max')).toBe(true);
  });

  it('accepts the maximum quantity and weight', async () => {
    const component = create();
    fill(component);
    component.form.patchValue({ quantity: 30, total_weight: 200 });

    await component.submit();

    expect(stub.createPartialWeaging).toHaveBeenCalledWith(
      expect.objectContaining({ quantity: 30, total_weight: 200 }),
    );
  });

  it('does not submit when the form is empty', async () => {
    const component = create();
    component.form.reset({
      sow_id: '',
      weaging_date: '',
      quantity: 0,
      type: '',
      total_weight: null,
      note: '',
    });

    await component.submit();

    expect(stub.createPartialWeaging).not.toHaveBeenCalled();
  });

  it('creates a partial weaging and emits the sow code', async () => {
    const component = create();
    component.sowOptions.set([{ value: 'sow-1', label: 'C-001' }]);
    const created = vi.fn();
    component.created.subscribe(created);
    fill(component);

    await component.submit();

    expect(stub.createPartialWeaging).toHaveBeenCalledWith({
      sow_id: 'sow-1',
      weaging_date: '2026-04-25',
      quantity: 2,
      type: 'Normal',
    });
    expect(created).toHaveBeenCalledWith('C-001');
  });

  it('includes the optional weight and note when provided', async () => {
    const component = create();
    fill(component);
    component.form.patchValue({ total_weight: 42.5, note: 'camada numerosa' });

    await component.submit();

    expect(stub.createPartialWeaging).toHaveBeenCalledWith(
      expect.objectContaining({ total_weight: 42.5, note: 'camada numerosa' }),
    );
  });

  it('shows the server error message when creation fails', async () => {
    stub.createPartialWeaging = vi.fn(() =>
      throwError(
        () =>
          new HttpErrorResponse({
            status: 409,
            error: { error: 'La cerda no está lactando' },
          }),
      ),
    );
    const component = create();
    fill(component);

    await component.submit();

    expect(notifications.error).toHaveBeenCalledWith('La cerda no está lactando');
  });

  it('shows the server message when the piglet quantity does not match', async () => {
    stub.createPartialWeaging = vi.fn(() =>
      throwError(
        () =>
          new HttpErrorResponse({
            status: 400,
            error: {
              error: 'La cantidad de lechones debe ser igual a los lechones actuales del parto',
            },
          }),
      ),
    );
    const component = create();
    fill(component);

    await component.submit();

    expect(notifications.error).toHaveBeenCalledWith(
      'La cantidad de lechones debe ser igual a los lechones actuales del parto',
    );
  });

  it('emits the created event when the response has no body', async () => {
    stub.createPartialWeaging = vi.fn(() => of(void 0));
    const component = create();
    component.sowOptions.set([{ value: 'sow-1', label: 'C-001' }]);
    const created = vi.fn();
    component.created.subscribe(created);
    fill(component);

    await component.submit();

    expect(created).toHaveBeenCalledWith('C-001');
  });
});
