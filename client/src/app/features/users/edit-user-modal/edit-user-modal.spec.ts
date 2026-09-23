import { HttpErrorResponse } from '@angular/common/http';
import { TestBed } from '@angular/core/testing';
import { of, throwError } from 'rxjs';
import { vi } from 'vitest';

import { NotificationService } from '../../../core/notifications/notification.service';
import { User } from '../../../core/users/user.models';
import { UsersService } from '../../../core/users/users.service';
import { EditUserModal } from './edit-user-modal';

function buildUser(): User {
  return {
    id: '42',
    name: 'Ana',
    username: 'ana',
    role: 'User',
    created_at: '2026-01-02T12:00:00',
    updated_at: '2026-01-02T12:00:00',
  };
}

class UsersStub {
  updateUser = vi.fn(() => of(buildUser()));
}

class NotificationsStub {
  success = vi.fn();
  error = vi.fn();
}

describe('EditUserModal', () => {
  let stub: UsersStub;
  let notifications: NotificationsStub;

  beforeEach(async () => {
    stub = new UsersStub();
    notifications = new NotificationsStub();
    await TestBed.configureTestingModule({
      imports: [EditUserModal],
      providers: [
        { provide: UsersService, useValue: stub },
        { provide: NotificationService, useValue: notifications },
      ],
    }).compileComponents();
  });

  // eslint-disable-next-line @typescript-eslint/no-explicit-any
  const create = () => {
    const fixture = TestBed.createComponent(EditUserModal);
    fixture.componentRef.setInput('user', buildUser());
    fixture.componentRef.setInput('open', true);
    fixture.detectChanges();
    return fixture.componentInstance as any;
  };

  it('prefills the form from the selected user', () => {
    const component = create();

    expect(component.form.getRawValue()).toMatchObject({
      name: 'Ana',
      username: 'ana',
      role: 'User',
      password: '',
    });
  });

  it('does not submit when required fields are empty', async () => {
    const component = create();
    component.form.controls.name.setValue('');

    await component.submit();

    expect(stub.updateUser).not.toHaveBeenCalled();
  });

  it('updates the user without a password and emits the result', async () => {
    const component = create();
    const updated = vi.fn();
    component.updated.subscribe(updated);
    component.form.controls.name.setValue('Ana Updated');
    component.form.controls.role.setValue('Admin');

    await component.submit();

    expect(stub.updateUser).toHaveBeenCalledWith('42', {
      name: 'Ana Updated',
      username: 'ana',
      role: 'Admin',
    });
    expect(updated).toHaveBeenCalled();
  });

  it('sends the password when one is provided', async () => {
    const component = create();
    component.form.controls.password.setValue('Password1!');
    component.form.controls.confirmPassword.setValue('Password1!');

    await component.submit();

    expect(stub.updateUser).toHaveBeenCalledWith(
      '42',
      expect.objectContaining({ password: 'Password1!' }),
    );
  });

  it('shows an error toast when the username is taken', async () => {
    stub.updateUser = vi.fn(() => throwError(() => new HttpErrorResponse({ status: 409 })));
    const component = create();

    await component.submit();

    expect(notifications.error).toHaveBeenCalledWith('El nombre de usuario ya está en uso');
  });
});
