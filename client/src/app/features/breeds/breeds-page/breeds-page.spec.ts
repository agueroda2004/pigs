import { WritableSignal, signal } from '@angular/core';
import { TestBed } from '@angular/core/testing';
import { of, throwError } from 'rxjs';
import { vi } from 'vitest';

import { AuthService } from '../../../core/auth/auth.service';
import { Breed } from '../../../core/breeds/breed.models';
import { BreedsService } from '../../../core/breeds/breeds.service';
import { NotificationService } from '../../../core/notifications/notification.service';
import { BreedsPage } from './breeds-page';

function buildBreed(id: string): Breed {
  return {
    id,
    name: `Breed ${id}`,
    active: true,
    created_at: '2026-01-02T12:00:00',
    updated_at: '2026-01-02T12:00:00',
    created_by: 'admin',
    updated_by: 'admin',
  };
}

class BreedsStub {
  listBreeds = vi.fn(() => of([buildBreed('1'), buildBreed('2')]));
}

class NotificationsStub {
  success = vi.fn();
  error = vi.fn();
}

describe('BreedsPage', () => {
  let stub: BreedsStub;
  let notifications: NotificationsStub;
  let isAdmin: WritableSignal<boolean>;

  beforeEach(async () => {
    stub = new BreedsStub();
    notifications = new NotificationsStub();
    isAdmin = signal(true);
    await TestBed.configureTestingModule({
      imports: [BreedsPage],
      providers: [
        { provide: BreedsService, useValue: stub },
        { provide: NotificationService, useValue: notifications },
        { provide: AuthService, useValue: { isAdmin } },
      ],
    }).compileComponents();
  });

  // eslint-disable-next-line @typescript-eslint/no-explicit-any
  const create = () => TestBed.createComponent(BreedsPage).componentInstance as any;

  it('loads breeds on init and renders a card per breed', async () => {
    const fixture = TestBed.createComponent(BreedsPage);
    const component = fixture.componentInstance as any;

    fixture.detectChanges();
    await fixture.whenStable();
    fixture.detectChanges();

    expect(stub.listBreeds).toHaveBeenCalled();
    expect(component.breeds()).toHaveLength(2);
    expect(fixture.nativeElement.querySelectorAll('app-breed-card')).toHaveLength(2);
  });

  it('shows the error state and retries', async () => {
    stub.listBreeds = vi.fn(() => throwError(() => new Error('failed')));
    const component = create();

    await component.loadBreeds();
    expect(component.error()).toBe(true);

    stub.listBreeds = vi.fn(() => of([buildBreed('1')]));
    await component.loadBreeds();

    expect(component.error()).toBe(false);
    expect(component.breeds()).toHaveLength(1);
  });

  it('reloads the list after creating a breed', async () => {
    const component = create();
    await component.loadBreeds();

    stub.listBreeds = vi.fn(() => of([buildBreed('1'), buildBreed('2'), buildBreed('3')]));
    await component.onCreated('Duroc');

    expect(notifications.success).toHaveBeenCalledWith('Raza "Duroc" creada correctamente');
    expect(component.modalOpen()).toBe(false);
    expect(component.breeds()).toHaveLength(3);
  });

  it('reloads the list after updating a breed', async () => {
    const component = create();
    await component.loadBreeds();

    stub.listBreeds = vi.fn(() => of([buildBreed('9')]));
    await component.onUpdated(buildBreed('9'));

    expect(notifications.success).toHaveBeenCalledWith('Raza "Breed 9" actualizada correctamente');
    expect(component.editOpen()).toBe(false);
    expect(component.breeds()).toHaveLength(1);
  });

  it('hides the create button for non-admins', () => {
    isAdmin.set(false);
    const fixture = TestBed.createComponent(BreedsPage);
    fixture.detectChanges();

    expect(fixture.nativeElement.textContent).not.toContain('Crear raza');
  });
});
