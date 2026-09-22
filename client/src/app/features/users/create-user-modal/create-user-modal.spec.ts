import { TestBed } from '@angular/core/testing';
import { of } from 'rxjs';
import { vi } from 'vitest';

import { UsersService } from '../../../core/users/users.service';
import { CreateUserModal } from './create-user-modal';

class UsersStub {
  createUser = vi.fn(() => of({}));
}

describe('CreateUserModal', () => {
  let stub: UsersStub;

  beforeEach(async () => {
    stub = new UsersStub();
    await TestBed.configureTestingModule({
      imports: [CreateUserModal],
      providers: [{ provide: UsersService, useValue: stub }],
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
});
