import { HttpErrorResponse } from '@angular/common/http';
import { TestBed } from '@angular/core/testing';
import { of, throwError } from 'rxjs';
import { vi } from 'vitest';

import { NotificationService } from '../../../core/notifications/notification.service';
import { PigletFostering } from '../../../core/piglet-fosterings/piglet-fostering.models';
import { PigletFosteringsService } from '../../../core/piglet-fosterings/piglet-fosterings.service';
import { SowsService } from '../../../core/sows/sows.service';
import { RegisterPigletFosteringModal } from './register-piglet-fostering-modal';

function buildFostering(overrides: Partial<PigletFostering> = {}): PigletFostering {
  return {
    id: '42',
    donor_farrowing_id: 'farrowing-1',
    receiver_farrowing_id: 'farrowing-2',
    donor_sow_id: 'sow-1',
    donor_sow_code: 'C-001',
    receiver_sow_id: 'sow-2',
    receiver_sow_code: 'C-002',
    movement_date: '2026-04-25',
    quantity: 2,
    note: 'camada numerosa',
    ...overrides,
  };
}

class PigletFosteringsStub {
  createPigletFostering = vi.fn((_request: Record<string, unknown>) => of(undefined));
  updatePigletFostering = vi.fn((_id: string, _request: Record<string, unknown>) => of(undefined));
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

describe('RegisterPigletFosteringModal', () => {
  let stub: PigletFosteringsStub;
  let sows: SowsStub;
  let notifications: NotificationsStub;

  beforeEach(async () => {
    stub = new PigletFosteringsStub();
    sows = new SowsStub();
    notifications = new NotificationsStub();
    await TestBed.configureTestingModule({
      imports: [RegisterPigletFosteringModal],
      providers: [
        { provide: PigletFosteringsService, useValue: stub },
        { provide: SowsService, useValue: sows },
        { provide: NotificationService, useValue: notifications },
      ],
    }).compileComponents();
  });

  // eslint-disable-next-line @typescript-eslint/no-explicit-any
  const create = () =>
    TestBed.createComponent(RegisterPigletFosteringModal).componentInstance as any;

  const createEditing = async (fostering: PigletFostering = buildFostering()): Promise<any> => {
    const fixture = TestBed.createComponent(RegisterPigletFosteringModal);
    fixture.componentRef.setInput('fostering', fostering);
    fixture.componentRef.setInput('open', true);
    fixture.detectChanges();
    await fixture.whenStable();
    // eslint-disable-next-line @typescript-eslint/no-explicit-any
    return fixture.componentInstance as any;
  };

  const fill = (component: any): void => {
    component.form.patchValue({
      donor_sow_id: 'sow-1',
      receiver_sow_id: 'sow-2',
      movement_date: '2026-04-25',
      quantity: 2,
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

  it('does not submit when the form is empty', async () => {
    const component = create();
    component.form.reset({
      donor_sow_id: '',
      receiver_sow_id: '',
      movement_date: '',
      quantity: 0,
      note: '',
    });

    await component.submit();

    expect(stub.createPigletFostering).not.toHaveBeenCalled();
  });

  it('creates a piglet fostering and emits both sow codes', async () => {
    const component = create();
    component.loadedSowOptions.set([
      { value: 'sow-1', label: 'C-001' },
      { value: 'sow-2', label: 'C-002' },
    ]);
    const created = vi.fn();
    component.created.subscribe(created);
    fill(component);

    await component.submit();

    expect(stub.createPigletFostering).toHaveBeenCalledWith({
      donor_sow_id: 'sow-1',
      receiver_sow_id: 'sow-2',
      movement_date: '2026-04-25',
      quantity: 2,
    });
    expect(created).toHaveBeenCalledWith({ donor: 'C-001', receiver: 'C-002' });
  });

  it('rejects the same sow on both sides', async () => {
    const component = create();
    fill(component);
    component.form.patchValue({ receiver_sow_id: 'sow-1' });

    await component.submit();

    expect(component.form.hasError('sameSow')).toBe(true);
    expect(stub.createPigletFostering).not.toHaveBeenCalled();
  });

  it('rejects a quantity above the maximum', async () => {
    const component = create();
    fill(component);
    component.form.patchValue({ quantity: 31 });

    await component.submit();

    expect(component.form.controls.quantity.hasError('max')).toBe(true);
    expect(stub.createPigletFostering).not.toHaveBeenCalled();
  });

  it('includes the optional note when provided', async () => {
    const component = create();
    fill(component);
    component.form.patchValue({ note: 'camada numerosa' });

    await component.submit();

    expect(stub.createPigletFostering).toHaveBeenCalledWith(
      expect.objectContaining({ note: 'camada numerosa' }),
    );
  });

  it('shows the server error message when creation fails', async () => {
    stub.createPigletFostering = vi.fn(() =>
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

  it('prefills the form and locks both sows when editing', async () => {
    const component = await createEditing();

    expect(component.editing()).toBe(true);
    expect(component.form.getRawValue()).toEqual({
      donor_sow_id: 'sow-1',
      receiver_sow_id: 'sow-2',
      movement_date: '2026-04-25',
      quantity: 2,
      note: 'camada numerosa',
    });
    expect(component.form.controls.donor_sow_id.disabled).toBe(true);
    expect(component.form.controls.receiver_sow_id.disabled).toBe(true);
    expect(component.sowOptions()).toEqual([
      { value: 'sow-1', label: 'C-001' },
      { value: 'sow-2', label: 'C-002' },
    ]);
  });

  it('updates the fostering without sending the sows', async () => {
    const component = await createEditing();
    const updated = vi.fn();
    component.updated.subscribe(updated);
    component.form.patchValue({ movement_date: '2026-04-26', quantity: 3 });

    await component.submit();

    expect(stub.updatePigletFostering).toHaveBeenCalledWith('42', {
      movement_date: '2026-04-26',
      quantity: 3,
      note: 'camada numerosa',
    });
    expect(stub.createPigletFostering).not.toHaveBeenCalled();
    expect(updated).toHaveBeenCalled();
  });

  it('sends an empty note when the user clears it', async () => {
    const component = await createEditing();

    component.form.patchValue({ note: '' });
    await component.submit();

    expect(stub.updatePigletFostering).toHaveBeenCalledWith(
      '42',
      expect.objectContaining({ note: '' }),
    );
  });

  it('shows the server error message when update fails', async () => {
    stub.updatePigletFostering = vi.fn(() =>
      throwError(
        () =>
          new HttpErrorResponse({
            status: 400,
            error: { error: 'La fecha del traslado debe ser posterior al último destete parcial' },
          }),
      ),
    );
    const component = await createEditing();

    await component.submit();

    expect(notifications.error).toHaveBeenCalledWith(
      'La fecha del traslado debe ser posterior al último destete parcial',
    );
  });
});
