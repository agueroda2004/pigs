import { HttpErrorResponse } from '@angular/common/http';
import { TestBed } from '@angular/core/testing';
import { of, throwError } from 'rxjs';
import { vi } from 'vitest';

import { BoarRemovalsService } from '../../../core/boar-removals/boar-removals.service';
import { BoarsService } from '../../../core/boars/boars.service';
import { NotificationService } from '../../../core/notifications/notification.service';
import { RegisterBoarRemovalModal } from './register-boar-removal-modal';

class BoarRemovalsStub {
  createBoarRemoval = vi.fn((_request: Record<string, unknown>) => of(void 0));
}

class BoarsStub {
  listBoarDropdown = vi.fn(() => of([{ id: 'boar-1', code: 'B-001', active: true }]));
}

class NotificationsStub {
  success = vi.fn();
  error = vi.fn();
}

describe('RegisterBoarRemovalModal', () => {
  let stub: BoarRemovalsStub;
  let boars: BoarsStub;
  let notifications: NotificationsStub;

  beforeEach(async () => {
    stub = new BoarRemovalsStub();
    boars = new BoarsStub();
    notifications = new NotificationsStub();
    await TestBed.configureTestingModule({
      imports: [RegisterBoarRemovalModal],
      providers: [
        { provide: BoarRemovalsService, useValue: stub },
        { provide: BoarsService, useValue: boars },
        { provide: NotificationService, useValue: notifications },
      ],
    }).compileComponents();
  });

  // eslint-disable-next-line @typescript-eslint/no-explicit-any
  const create = () => TestBed.createComponent(RegisterBoarRemovalModal).componentInstance as any;

  const fill = (component: any, date: string): void => {
    component.form.patchValue({
      boar_id: 'boar-1',
      removal_date: date,
      type: 'Muerte',
      reason: 'Enfermedad',
      note: '',
    });
  };

  it('loads the eligible boar options for selection', async () => {
    const component = create();

    await component.loadOptions();

    expect(boars.listBoarDropdown).toHaveBeenCalledWith(true, 'Vivo');
    expect(component.boarOptions()).toEqual([{ value: 'boar-1', label: 'B-001' }]);
  });

  it('does not submit when the form is empty', async () => {
    const component = create();
    component.form.reset({
      boar_id: '',
      removal_date: '',
      type: '',
      reason: '',
      note: '',
    });

    await component.submit();

    expect(stub.createBoarRemoval).not.toHaveBeenCalled();
  });

  it('creates a removal and emits the boar code', async () => {
    const component = create();
    component.boarOptions.set([{ value: 'boar-1', label: 'B-001' }]);
    const created = vi.fn();
    component.created.subscribe(created);
    fill(component, '2026-01-20');

    await component.submit();

    expect(stub.createBoarRemoval).toHaveBeenCalledWith({
      boar_id: 'boar-1',
      removal_date: '2026-01-20',
      type: 'Muerte',
      reason: 'Enfermedad',
    });
    expect(created).toHaveBeenCalledWith('B-001');
  });

  it('includes the optional note when provided', async () => {
    const component = create();
    component.boarOptions.set([{ value: 'boar-1', label: 'B-001' }]);
    fill(component, '2026-01-20');
    component.form.patchValue({ note: 'baja por enfermedad' });

    await component.submit();

    expect(stub.createBoarRemoval).toHaveBeenCalledWith({
      boar_id: 'boar-1',
      removal_date: '2026-01-20',
      type: 'Muerte',
      reason: 'Enfermedad',
      note: 'baja por enfermedad',
    });
  });

  it('shows the server error message when creation fails', async () => {
    stub.createBoarRemoval = vi.fn(() =>
      throwError(
        () =>
          new HttpErrorResponse({
            status: 409,
            error: { error: 'El verraco no está en un estado válido para la baja' },
          }),
      ),
    );
    const component = create();
    component.boarOptions.set([{ value: 'boar-1', label: 'B-001' }]);
    fill(component, '2026-01-20');

    await component.submit();

    expect(notifications.error).toHaveBeenCalledWith(
      'El verraco no está en un estado válido para la baja',
    );
  });
});
