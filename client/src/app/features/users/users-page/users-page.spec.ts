import { TestBed } from '@angular/core/testing';
import { of, throwError } from 'rxjs';
import { vi } from 'vitest';

import { NotificationService } from '../../../core/notifications/notification.service';
import { User } from '../../../core/users/user.models';
import { UsersService } from '../../../core/users/users.service';
import { UsersPage } from './users-page';

function buildUser(id: string): User {
  return {
    id,
    name: `User ${id}`,
    username: id,
    role: 'User',
    created_at: '2026-01-02T03:04:05Z',
    updated_at: '2026-01-02T03:04:05Z',
  };
}

class UsersStub {
  listUsers = vi.fn(() => of([buildUser('1'), buildUser('2')]));
}

class NotificationsStub {
  success = vi.fn();
  error = vi.fn();
}

describe('UsersPage', () => {
  let stub: UsersStub;
  let notifications: NotificationsStub;

  beforeEach(async () => {
    stub = new UsersStub();
    notifications = new NotificationsStub();
    await TestBed.configureTestingModule({
      imports: [UsersPage],
      providers: [
        { provide: UsersService, useValue: stub },
        { provide: NotificationService, useValue: notifications },
      ],
    }).compileComponents();
  });

  // eslint-disable-next-line @typescript-eslint/no-explicit-any
  const create = () => TestBed.createComponent(UsersPage).componentInstance as any;

  it('loads users on init and renders a card per user', async () => {
    const fixture = TestBed.createComponent(UsersPage);
    const component = fixture.componentInstance as any;

    fixture.detectChanges();
    await fixture.whenStable();
    fixture.detectChanges();

    expect(stub.listUsers).toHaveBeenCalled();
    expect(component.users()).toHaveLength(2);
    expect(fixture.nativeElement.querySelectorAll('app-user-card')).toHaveLength(2);
  });

  it('shows the error state and retries', async () => {
    stub.listUsers = vi.fn(() => throwError(() => new Error('failed')));
    const component = create();

    await component.loadUsers();
    expect(component.error()).toBe(true);

    stub.listUsers = vi.fn(() => of([buildUser('1')]));
    await component.loadUsers();

    expect(component.error()).toBe(false);
    expect(component.users()).toHaveLength(1);
  });

  it('reloads the list after creating a user', async () => {
    const component = create();
    await component.loadUsers();

    stub.listUsers = vi.fn(() => of([buildUser('1'), buildUser('2'), buildUser('3')]));
    await component.onCreated('ana');

    expect(notifications.success).toHaveBeenCalledWith('Usuario "ana" creado correctamente');
    expect(component.modalOpen()).toBe(false);
    expect(component.users()).toHaveLength(3);
  });

  it('opens the edit modal for the selected user', () => {
    const component = create();
    const user = buildUser('1');

    component.openEdit(user);

    expect(component.editOpen()).toBe(true);
    expect(component.editingUser()).toBe(user);
  });

  it('reloads the list after updating a user', async () => {
    const component = create();
    await component.loadUsers();

    stub.listUsers = vi.fn(() => of([buildUser('9')]));
    await component.onUpdated(buildUser('9'));

    expect(notifications.success).toHaveBeenCalledWith('Usuario "9" actualizado correctamente');
    expect(component.editOpen()).toBe(false);
    expect(component.users()).toHaveLength(1);
  });
});
