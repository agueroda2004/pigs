import { HttpErrorResponse } from '@angular/common/http';
import { TestBed } from '@angular/core/testing';
import { of, throwError } from 'rxjs';
import { vi } from 'vitest';

import { Operator } from '../../../core/operators/operator.models';
import { OperatorsService } from '../../../core/operators/operators.service';
import { NotificationService } from '../../../core/notifications/notification.service';
import { DeleteOperatorModal } from './delete-operator-modal';

function buildOperator(): Operator {
  return { id: '42', name: 'Juan Pérez', active: true };
}

class OperatorsStub {
  deleteOperator = vi.fn(() => of(undefined));
}

class NotificationsStub {
  success = vi.fn();
  error = vi.fn();
}

describe('DeleteOperatorModal', () => {
  let stub: OperatorsStub;
  let notifications: NotificationsStub;

  beforeEach(async () => {
    stub = new OperatorsStub();
    notifications = new NotificationsStub();
    await TestBed.configureTestingModule({
      imports: [DeleteOperatorModal],
      providers: [
        { provide: OperatorsService, useValue: stub },
        { provide: NotificationService, useValue: notifications },
      ],
    }).compileComponents();
  });

  const create = () => {
    const fixture = TestBed.createComponent(DeleteOperatorModal);
    fixture.componentRef.setInput('operator', buildOperator());
    fixture.componentRef.setInput('open', true);
    fixture.detectChanges();
    // eslint-disable-next-line @typescript-eslint/no-explicit-any
    return fixture.componentInstance as any;
  };

  it('does nothing when there is no operator', async () => {
    const fixture = TestBed.createComponent(DeleteOperatorModal);
    fixture.componentRef.setInput('open', true);
    fixture.detectChanges();

    // eslint-disable-next-line @typescript-eslint/no-explicit-any
    await (fixture.componentInstance as any).confirm();

    expect(stub.deleteOperator).not.toHaveBeenCalled();
  });

  it('deletes the operator and emits it', async () => {
    const component = create();
    const deleted = vi.fn();
    component.deleted.subscribe(deleted);

    await component.confirm();

    expect(stub.deleteOperator).toHaveBeenCalledWith('42');
    expect(deleted).toHaveBeenCalledWith(buildOperator());
  });

  it('shows the server message when the operator has linked records', async () => {
    stub.deleteOperator = vi.fn(() =>
      throwError(
        () =>
          new HttpErrorResponse({
            status: 409,
            error: {
              error: 'No se puede eliminar este operador porque tiene registros enlazados.',
            },
          }),
      ),
    );
    const component = create();

    await component.confirm();

    expect(notifications.error).toHaveBeenCalledWith(
      'No se puede eliminar este operador porque tiene registros enlazados.',
    );
  });

  it('shows a fallback when the operator is missing', async () => {
    stub.deleteOperator = vi.fn(() => throwError(() => new HttpErrorResponse({ status: 404 })));
    const component = create();

    await component.confirm();

    expect(notifications.error).toHaveBeenCalledWith('El operador no existe');
  });
});
