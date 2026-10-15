import { HttpErrorResponse } from '@angular/common/http';
import { TestBed } from '@angular/core/testing';
import { of, throwError } from 'rxjs';
import { vi } from 'vitest';

import { NotificationService } from '../../../core/notifications/notification.service';
import { User } from '../../../core/users/user.models';
import { UsersService } from '../../../core/users/users.service';
import { DeleteUserModal } from './delete-user-modal';

function buildUser(): User {
  return {
    id: '42',
    name: 'Ana',
    username: 'ana',
    role: 'User',
    active: true,
  };
}

class UsersStub {
  deleteUser = vi.fn(() => of(undefined));
}

class NotificationsStub {
  success = vi.fn();
  error = vi.fn();
}

describe('DeleteUserModal', () => {
  let stub: UsersStub;
  let notifications: NotificationsStub;

  beforeEach(async () => {
    stub = new UsersStub();
    notifications = new NotificationsStub();
    await TestBed.configureTestingModule({
      imports: [DeleteUserModal],
      providers: [
        { provide: UsersService, useValue: stub },
        { provide: NotificationService, useValue: notifications },
      ],
    }).compileComponents();
  });

  const create = () => {
    const fixture = TestBed.createComponent(DeleteUserModal);
    fixture.componentRef.setInput('user', buildUser());
    fixture.componentRef.setInput('open', true);
    fixture.detectChanges();
    // eslint-disable-next-line @typescript-eslint/no-explicit-any
    return fixture.componentInstance as any;
  };

  it('does nothing when there is no user', async () => {
    const fixture = TestBed.createComponent(DeleteUserModal);
    fixture.componentRef.setInput('open', true);
    fixture.detectChanges();

    // eslint-disable-next-line @typescript-eslint/no-explicit-any
    await (fixture.componentInstance as any).confirm();

    expect(stub.deleteUser).not.toHaveBeenCalled();
  });

  it('deletes the user and emits the username', async () => {
    const component = create();
    const deleted = vi.fn();
    component.deleted.subscribe(deleted);

    await component.confirm();

    expect(stub.deleteUser).toHaveBeenCalledWith('42');
    expect(deleted).toHaveBeenCalledWith('ana');
  });

  it('shows the server message when the user has linked records', async () => {
    stub.deleteUser = vi.fn(() =>
      throwError(
        () =>
          new HttpErrorResponse({
            status: 409,
            error: {
              error: 'No se puede eliminar este usuario porque tiene registros enlazados.',
            },
          }),
      ),
    );
    const component = create();

    await component.confirm();

    expect(notifications.error).toHaveBeenCalledWith(
      'No se puede eliminar este usuario porque tiene registros enlazados.',
    );
  });

  it('shows a fallback when the user is missing', async () => {
    stub.deleteUser = vi.fn(() => throwError(() => new HttpErrorResponse({ status: 404 })));
    const component = create();

    await component.confirm();

    expect(notifications.error).toHaveBeenCalledWith('El usuario no existe');
  });
});
