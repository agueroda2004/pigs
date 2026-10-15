import { TestBed } from '@angular/core/testing';

import { User } from '../../../core/users/user.models';
import { UserCard } from './user-card';

function buildUser(role: User['role'], active = true): User {
  return {
    id: '1',
    name: 'Ana',
    username: 'ana',
    role,
    active,
  };
}

describe('UserCard', () => {
  beforeEach(async () => {
    await TestBed.configureTestingModule({
      imports: [UserCard],
    }).compileComponents();
  });

  it('renders the name and username', () => {
    const fixture = TestBed.createComponent(UserCard);
    fixture.componentRef.setInput('user', buildUser('User'));
    fixture.detectChanges();

    const text = fixture.nativeElement.textContent as string;
    expect(text).toContain('Ana');
    expect(text).toContain('@ana');
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

  it('shows the active status for active users', () => {
    const fixture = TestBed.createComponent(UserCard);
    fixture.componentRef.setInput('user', buildUser('User', true));
    fixture.detectChanges();

    expect(fixture.nativeElement.textContent).toContain('Activo');
  });

  it('shows the inactive status for inactive users', () => {
    const fixture = TestBed.createComponent(UserCard);
    fixture.componentRef.setInput('user', buildUser('User', false));
    fixture.detectChanges();

    expect(fixture.nativeElement.textContent).toContain('Inactivo');
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

  it('emits the user when the delete button is clicked', () => {
    const user = buildUser('User');
    const fixture = TestBed.createComponent(UserCard);
    fixture.componentRef.setInput('user', user);
    fixture.detectChanges();

    const emitted: User[] = [];
    fixture.componentInstance.deleteRequested.subscribe((value) => emitted.push(value));
    const buttons = fixture.nativeElement.querySelectorAll('button');
    (buttons[1] as HTMLButtonElement).click();

    expect(emitted).toEqual([user]);
  });

  it('disables the delete button for the current user', () => {
    const fixture = TestBed.createComponent(UserCard);
    fixture.componentRef.setInput('user', buildUser('Admin'));
    fixture.componentRef.setInput('currentUserId', '1');
    fixture.detectChanges();

    const buttons = fixture.nativeElement.querySelectorAll('button');
    expect((buttons[1] as HTMLButtonElement).disabled).toBe(true);
  });

  it('enables the delete button for other users', () => {
    const fixture = TestBed.createComponent(UserCard);
    fixture.componentRef.setInput('user', buildUser('Admin'));
    fixture.componentRef.setInput('currentUserId', '999');
    fixture.detectChanges();

    const buttons = fixture.nativeElement.querySelectorAll('button');
    expect((buttons[1] as HTMLButtonElement).disabled).toBe(false);
  });
});
