import { TestBed } from '@angular/core/testing';

import { User } from '../../../core/users/user.models';
import { UserCard } from './user-card';

function buildUser(role: User['role']): User {
  return {
    id: '1',
    name: 'Ana',
    username: 'ana',
    role,
    created_at: '2026-01-02T12:00:00',
    updated_at: '2026-01-02T12:00:00',
  };
}

describe('UserCard', () => {
  beforeEach(async () => {
    await TestBed.configureTestingModule({
      imports: [UserCard],
    }).compileComponents();
  });

  it('renders the name, username and created date', () => {
    const fixture = TestBed.createComponent(UserCard);
    fixture.componentRef.setInput('user', buildUser('User'));
    fixture.detectChanges();

    const text = fixture.nativeElement.textContent as string;
    expect(text).toContain('Ana');
    expect(text).toContain('@ana');
    expect(text).toContain('02/01/2026');
  });

  it('shows the admin badge for admin users', () => {
    const fixture = TestBed.createComponent(UserCard);
    fixture.componentRef.setInput('user', buildUser('Admin'));
    fixture.detectChanges();

    expect(fixture.nativeElement.textContent).toContain('Administrador');
  });

  it('shows the user badge for regular users', () => {
    const fixture = TestBed.createComponent(UserCard);
    fixture.componentRef.setInput('user', buildUser('User'));
    fixture.detectChanges();

    expect(fixture.nativeElement.textContent).toContain('Usuario');
  });

  it('emits the user when the edit button is clicked', () => {
    const user = buildUser('User');
    const fixture = TestBed.createComponent(UserCard);
    fixture.componentRef.setInput('user', user);
    fixture.detectChanges();

    const emitted: User[] = [];
    fixture.componentInstance.editRequested.subscribe((value) => emitted.push(value));
    (fixture.nativeElement.querySelector('button') as HTMLButtonElement).click();

    expect(emitted).toEqual([user]);
  });
});
