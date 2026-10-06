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
    await component.onUpdated();

    expect(notifications.success).toHaveBeenCalledWith('Raza actualizada correctamente');
    expect(component.editOpen()).toBe(false);
    expect(component.breeds()).toHaveLength(1);
  });

  it('hides the create button for non-admins', () => {
    isAdmin.set(false);
    const fixture = TestBed.createComponent(BreedsPage);
    fixture.detectChanges();

    expect(fixture.nativeElement.textContent).not.toContain('Crear raza');
  });

  it('applies the name and active filters when searching', async () => {
    const component = create();
    component.filterForm.setValue({ name: 'dur', active: 'true' });

    await component.search();

    expect(stub.listBreeds).toHaveBeenLastCalledWith({ name: 'dur', active: true });
  });

  it('clears the filters and reloads without them', async () => {
    const component = create();
    component.filterForm.setValue({ name: 'dur', active: 'true' });
    await component.search();

    stub.listBreeds = vi.fn(() => of([buildBreed('1')]));
    await component.clearFilters();

    expect(stub.listBreeds).toHaveBeenCalledWith({});
    expect(component.filterForm.getRawValue()).toEqual({ name: '', active: '' });
  });

  it('opens the delete modal for the selected breed', () => {
    const component = create();
    const breed = buildBreed('9');

    component.openDelete(breed);

    expect(component.deleteOpen()).toBe(true);
    expect(component.deletingBreed()).toEqual(breed);
  });

  it('reloads the list after deleting a breed', async () => {
    const component = create();
    await component.loadBreeds();

    stub.listBreeds = vi.fn(() => of([buildBreed('2')]));
    await component.onDeleted(buildBreed('9'));

    expect(notifications.success).toHaveBeenCalledWith('Raza "Breed 9" eliminada correctamente');
    expect(component.deleteOpen()).toBe(false);
    expect(component.breeds()).toHaveLength(1);
  });
});
