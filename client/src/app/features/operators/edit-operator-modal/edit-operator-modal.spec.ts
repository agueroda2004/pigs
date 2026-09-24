import { HttpErrorResponse } from '@angular/common/http';
import { TestBed } from '@angular/core/testing';
import { of, throwError } from 'rxjs';
import { vi } from 'vitest';

import { Operator } from '../../../core/operators/operator.models';
import { OperatorsService } from '../../../core/operators/operators.service';
import { NotificationService } from '../../../core/notifications/notification.service';
import { EditOperatorModal } from './edit-operator-modal';

function buildOperator(): Operator {
  return {
    id: '42',
    name: 'Juan Pérez',
    active: true,
    created_at: '2026-01-02T12:00:00',
    updated_at: '2026-01-02T12:00:00',
    created_by: 'admin',
    updated_by: 'admin',
  };
}

class OperatorsStub {
  updateOperator = vi.fn(() => of(buildOperator()));
}

class NotificationsStub {
  success = vi.fn();
  error = vi.fn();
}

describe('EditOperatorModal', () => {
  let stub: OperatorsStub;
  let notifications: NotificationsStub;

  beforeEach(async () => {
    stub = new OperatorsStub();
    notifications = new NotificationsStub();
    await TestBed.configureTestingModule({
      imports: [EditOperatorModal],
      providers: [
        { provide: OperatorsService, useValue: stub },
        { provide: NotificationService, useValue: notifications },
      ],
    }).compileComponents();
  });

  // eslint-disable-next-line @typescript-eslint/no-explicit-any
  const create = () => {
    const fixture = TestBed.createComponent(EditOperatorModal);
    fixture.componentRef.setInput('operator', buildOperator());
    fixture.componentRef.setInput('open', true);
    fixture.detectChanges();
    return fixture.componentInstance as any;
  };

  it('prefills the form from the selected operator', () => {
    const component = create();

    expect(component.form.getRawValue()).toEqual({ name: 'Juan Pérez', active: true });
  });

  it('updates the operator with changed fields and emits the result', async () => {
    const component = create();
    const updated = vi.fn();
    component.updated.subscribe(updated);
    component.form.controls.name.setValue('María López');
    component.form.controls.active.setValue(false);

    await component.submit();

    expect(stub.updateOperator).toHaveBeenCalledWith('42', {
      name: 'María López',
      active: false,
    });
    expect(updated).toHaveBeenCalled();
  });

  it('does not call the API when nothing changed', async () => {
    const component = create();
    const updated = vi.fn();
    component.updated.subscribe(updated);

    await component.submit();

    expect(stub.updateOperator).not.toHaveBeenCalled();
    expect(updated).toHaveBeenCalled();
  });

  it('shows an error toast when the name is taken', async () => {
    stub.updateOperator = vi.fn(() => throwError(() => new HttpErrorResponse({ status: 409 })));
    const component = create();
    component.form.controls.name.setValue('María López');

    await component.submit();

    expect(notifications.error).toHaveBeenCalledWith('El nombre del operador ya existe');
  });
});
