import { HttpErrorResponse } from '@angular/common/http';
import { TestBed } from '@angular/core/testing';
import { of, throwError } from 'rxjs';
import { vi } from 'vitest';

import { AbortionsService } from '../../../core/abortions/abortions.service';
import { NotificationService } from '../../../core/notifications/notification.service';
import { SowsService } from '../../../core/sows/sows.service';
import { RegisterAbortionModal } from './register-abortion-modal';

class AbortionsStub {
  createAbortion = vi.fn((_request: Record<string, unknown>) => of({}));
}

class SowsStub {
  listSows = vi.fn(() => of([{ id: 'sow-1', code: 'C-001' }]));
}

class NotificationsStub {
  success = vi.fn();
  error = vi.fn();
}

describe('RegisterAbortionModal', () => {
  let stub: AbortionsStub;
  let sows: SowsStub;
  let notifications: NotificationsStub;

  beforeEach(async () => {
    stub = new AbortionsStub();
    sows = new SowsStub();
    notifications = new NotificationsStub();
    await TestBed.configureTestingModule({
      imports: [RegisterAbortionModal],
      providers: [
        { provide: AbortionsService, useValue: stub },
        { provide: SowsService, useValue: sows },
        { provide: NotificationService, useValue: notifications },
      ],
    }).compileComponents();
  });

  // eslint-disable-next-line @typescript-eslint/no-explicit-any
  const create = () => TestBed.createComponent(RegisterAbortionModal).componentInstance as any;

  const fill = (component: any, date: string): void => {
    component.form.patchValue({
      sow_id: 'sow-1',
      abortion_date: date,
      cause: 'Infeccioso',
      note: '',
    });
  };

  it('loads the gestating sow options for selection', async () => {
    const component = create();

    await component.loadOptions();

    expect(sows.listSows).toHaveBeenCalledWith({ state: 'Gestando' });
    expect(component.sowOptions()).toEqual([{ value: 'sow-1', label: 'C-001' }]);
  });

  it('does not submit when the form is empty', async () => {
    const component = create();
    component.form.reset({ sow_id: '', abortion_date: '', cause: '', note: '' });

    await component.submit();

    expect(stub.createAbortion).not.toHaveBeenCalled();
  });

  it('creates an abortion and emits the sow code', async () => {
    const component = create();
    component.sowOptions.set([{ value: 'sow-1', label: 'C-001' }]);
    const created = vi.fn();
    component.created.subscribe(created);
    fill(component, '2026-01-15');

    await component.submit();

    expect(stub.createAbortion).toHaveBeenCalledWith({
      sow_id: 'sow-1',
      abortion_date: '2026-01-15',
      cause: 'Infeccioso',
    });
    expect(created).toHaveBeenCalledWith('C-001');
  });

  it('includes the optional note when provided', async () => {
    const component = create();
    component.sowOptions.set([{ value: 'sow-1', label: 'C-001' }]);
    fill(component, '2026-01-15');
    component.form.patchValue({ note: 'aborto espontáneo' });

    await component.submit();

    expect(stub.createAbortion).toHaveBeenCalledWith({
      sow_id: 'sow-1',
      abortion_date: '2026-01-15',
      cause: 'Infeccioso',
      note: 'aborto espontáneo',
    });
  });

  it('shows the server error message when creation fails', async () => {
    stub.createAbortion = vi.fn(() =>
      throwError(
        () =>
          new HttpErrorResponse({
            status: 409,
            error: { error: 'La cerda no está gestando' },
          }),
      ),
    );
    const component = create();
    component.sowOptions.set([{ value: 'sow-1', label: 'C-001' }]);
    fill(component, '2026-01-15');

    await component.submit();

    expect(notifications.error).toHaveBeenCalledWith('La cerda no está gestando');
  });
});
