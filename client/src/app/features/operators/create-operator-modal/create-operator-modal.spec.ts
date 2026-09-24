import { HttpErrorResponse } from '@angular/common/http';
import { TestBed } from '@angular/core/testing';
import { of, throwError } from 'rxjs';
import { vi } from 'vitest';

import { OperatorsService } from '../../../core/operators/operators.service';
import { NotificationService } from '../../../core/notifications/notification.service';
import { CreateOperatorModal } from './create-operator-modal';

class OperatorsStub {
  createOperator = vi.fn(() => of({}));
}

class NotificationsStub {
  success = vi.fn();
  error = vi.fn();
}

describe('CreateOperatorModal', () => {
  let stub: OperatorsStub;
  let notifications: NotificationsStub;

  beforeEach(async () => {
    stub = new OperatorsStub();
    notifications = new NotificationsStub();
    await TestBed.configureTestingModule({
      imports: [CreateOperatorModal],
      providers: [
        { provide: OperatorsService, useValue: stub },
        { provide: NotificationService, useValue: notifications },
      ],
    }).compileComponents();
  });

  // eslint-disable-next-line @typescript-eslint/no-explicit-any
  const create = () => TestBed.createComponent(CreateOperatorModal).componentInstance as any;

  it('does not submit when the form is empty', async () => {
    const component = create();
    await component.submit();
    expect(stub.createOperator).not.toHaveBeenCalled();
  });

  it('creates an operator and emits the name', async () => {
    const component = create();
    const created = vi.fn();
    component.created.subscribe(created);
    component.form.setValue({ name: 'Juan Pérez' });

    await component.submit();

    expect(stub.createOperator).toHaveBeenCalledWith({ name: 'Juan Pérez' });
    expect(created).toHaveBeenCalledWith('Juan Pérez');
  });

  it('shows an error toast when the name is taken', async () => {
    stub.createOperator = vi.fn(() => throwError(() => new HttpErrorResponse({ status: 409 })));
    const component = create();
    component.form.setValue({ name: 'Juan Pérez' });

    await component.submit();

    expect(notifications.error).toHaveBeenCalledWith('El nombre del operador ya existe');
  });
});
