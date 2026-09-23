import { HttpErrorResponse } from '@angular/common/http';
import { TestBed } from '@angular/core/testing';
import { of, throwError } from 'rxjs';
import { vi } from 'vitest';

import { NotificationService } from '../../../core/notifications/notification.service';
import { UsersService } from '../../../core/users/users.service';
import { CreateUserModal } from './create-user-modal';

class UsersStub {
  createUser = vi.fn(() => of({}));
}

class NotificationsStub {
  success = vi.fn();
  error = vi.fn();
}

describe('CreateUserModal', () => {
  let stub: UsersStub;
  let notifications: NotificationsStub;

  beforeEach(async () => {
    stub = new UsersStub();
    notifications = new NotificationsStub();
    await TestBed.configureTestingModule({
      imports: [CreateUserModal],
      providers: [
        { provide: UsersService, useValue: stub },
        { provide: NotificationService, useValue: notifications },
      ],
    }).compileComponents();
  });

  // eslint-disable-next-line @typescript-eslint/no-explicit-any
  const create = () => TestBed.createComponent(CreateUserModal).componentInstance as any;

  it('does not submit when the form is empty', async () => {
    const component = create();
    await component.submit();
    expect(stub.createUser).not.toHaveBeenCalled();
  });

  it('does not submit when the passwords do not match', async () => {
    const component = create();
    component.form.setValue({
      name: 'Ana',
      username: 'ana',
      password: 'Password1!',
      confirmPassword: 'Other1!',
      role: 'User',
    });

    await component.submit();
    expect(stub.createUser).not.toHaveBeenCalled();
  });

  it('creates a user and emits the username', async () => {
    const component = create();
    const created = vi.fn();
    component.created.subscribe(created);
    component.form.setValue({
      name: 'Ana',
      username: 'ana',
      password: 'Password1!',
      confirmPassword: 'Password1!',
      role: 'User',
    });

    await component.submit();

    expect(stub.createUser).toHaveBeenCalledWith({
      name: 'Ana',
      username: 'ana',
      password: 'Password1!',
      role: 'User',
    });
    expect(created).toHaveBeenCalledWith('ana');
  });

  it('shows an error toast when the username is taken', async () => {
    stub.createUser = vi.fn(() => throwError(() => new HttpErrorResponse({ status: 409 })));
    const component = create();
    component.form.setValue({
      name: 'Ana',
      username: 'ana',
      password: 'Password1!',
      confirmPassword: 'Password1!',
      role: 'User',
    });

    await component.submit();

    expect(notifications.error).toHaveBeenCalledWith('El nombre de usuario ya está en uso');
  });
});
