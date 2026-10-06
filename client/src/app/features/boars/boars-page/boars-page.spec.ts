import { WritableSignal, signal } from '@angular/core';
import { TestBed } from '@angular/core/testing';
import { of, throwError } from 'rxjs';
import { vi } from 'vitest';

import { AuthService } from '../../../core/auth/auth.service';
import { Boar } from '../../../core/boars/boar.models';
import { BoarsService } from '../../../core/boars/boars.service';
import { BreedsService } from '../../../core/breeds/breeds.service';
import { NotificationService } from '../../../core/notifications/notification.service';
import { BoarsPage } from './boars-page';

function buildBoar(id: string): Boar {
  return {
    id,
    code: `B-00${id}`,
    location: 'Corral A',
    active: true,
    entry_date: '2026-01-10',
    birth_date: null,
    note: null,
    state: 'Vivo',
    origin: 'Propio',
    breed_id: 'breed-1',
  };
}

class BoarsStub {
  listBoars = vi.fn(() => of([buildBoar('1'), buildBoar('2')]));
}

class BreedsStub {
  listBreedDropdown = vi.fn(() => of([{ id: 'breed-1', name: 'Duroc', active: true }]));
}

class NotificationsStub {
  success = vi.fn();
  error = vi.fn();
}

describe('BoarsPage', () => {
  let stub: BoarsStub;
  let breeds: BreedsStub;
  let notifications: NotificationsStub;
  let isAdmin: WritableSignal<boolean>;

  beforeEach(async () => {
    stub = new BoarsStub();
    breeds = new BreedsStub();
    notifications = new NotificationsStub();
    isAdmin = signal(true);
    await TestBed.configureTestingModule({
      imports: [BoarsPage],
      providers: [
        { provide: BoarsService, useValue: stub },
        { provide: BreedsService, useValue: breeds },
        { provide: NotificationService, useValue: notifications },
        { provide: AuthService, useValue: { isAdmin } },
      ],
    }).compileComponents();
  });

  // eslint-disable-next-line @typescript-eslint/no-explicit-any
  const create = () => TestBed.createComponent(BoarsPage).componentInstance as any;

  it('loads boars on init and renders a card per boar', async () => {
    const fixture = TestBed.createComponent(BoarsPage);
    const component = fixture.componentInstance as any;

    fixture.detectChanges();
    await fixture.whenStable();
    fixture.detectChanges();

    expect(stub.listBoars).toHaveBeenCalled();
    expect(component.boars()).toHaveLength(2);
    expect(fixture.nativeElement.querySelectorAll('app-boar-card')).toHaveLength(2);
  });

  it('shows the error state and retries', async () => {
    stub.listBoars = vi.fn(() => throwError(() => new Error('failed')));
    const component = create();

    await component.loadBoars();
    expect(component.error()).toBe(true);

    stub.listBoars = vi.fn(() => of([buildBoar('1')]));
    await component.loadBoars();

    expect(component.error()).toBe(false);
    expect(component.boars()).toHaveLength(1);
  });

  it('reloads the list after creating a boar', async () => {
    const component = create();
    await component.loadBoars();

    stub.listBoars = vi.fn(() => of([buildBoar('1'), buildBoar('2'), buildBoar('3')]));
    await component.onCreated('B-004');

    expect(notifications.success).toHaveBeenCalledWith('Verraco "B-004" creado correctamente');
    expect(component.modalOpen()).toBe(false);
    expect(component.boars()).toHaveLength(3);
  });

  it('reloads the list after updating a boar', async () => {
    const component = create();
    await component.loadBoars();

    stub.listBoars = vi.fn(() => of([buildBoar('9')]));
    await component.onUpdated();

    expect(notifications.success).toHaveBeenCalledWith('Verraco actualizado correctamente');
    expect(component.editOpen()).toBe(false);
    expect(component.boars()).toHaveLength(1);
  });

  it('opens the delete modal for the selected boar', () => {
    const component = create();
    const boar = buildBoar('9');

    component.openDelete(boar);

    expect(component.deleteOpen()).toBe(true);
    expect(component.deletingBoar()).toEqual(boar);
  });

  it('reloads the list after deleting a boar', async () => {
    const component = create();
    await component.loadBoars();

    stub.listBoars = vi.fn(() => of([buildBoar('2')]));
    await component.onDeleted(buildBoar('9'));

    expect(notifications.success).toHaveBeenCalledWith('Verraco "B-009" eliminado correctamente');
    expect(component.deleteOpen()).toBe(false);
    expect(component.boars()).toHaveLength(1);
  });

  it('hides the create button for non-admins', () => {
    isAdmin.set(false);
    const fixture = TestBed.createComponent(BoarsPage);
    fixture.detectChanges();

    expect(fixture.nativeElement.textContent).not.toContain('Crear verraco');
  });

  it('loads breed options for the filter dropdown', async () => {
    const component = create();

    await component.loadBreeds();

    expect(component.breedOptions()).toEqual([{ value: 'breed-1', label: 'Duroc' }]);
  });

  it('builds the filter options and names from a single dropdown request', async () => {
    breeds.listBreedDropdown = vi.fn(() =>
      of([
        { id: 'breed-1', name: 'Duroc', active: true },
        { id: 'breed-2', name: 'Retired', active: false },
      ]),
    );
    const component = create();

    await component.loadBreeds();

    expect(breeds.listBreedDropdown).toHaveBeenCalledTimes(1);
    expect(breeds.listBreedDropdown).toHaveBeenCalledWith(false);
    expect(component.breedOptions()).toEqual([
      { value: 'breed-1', label: 'Duroc' },
      { value: 'breed-2', label: 'Retired (inactiva)' },
    ]);
    expect(component.breedNames().get('breed-2')).toBe('Retired');
  });

  it('applies the filter values when searching', async () => {
    const component = create();
    component.filterForm.setValue({
      code: '  B-001  ',
      breed_id: 'breed-1',
      origin: 'Externo',
      active: 'false',
    });

    await component.search();

    expect(stub.listBoars).toHaveBeenLastCalledWith({
      code: 'B-001',
      breed_id: 'breed-1',
      origin: 'Externo',
      active: false,
    });
  });

  it('clears the filters and reloads without them', async () => {
    const component = create();
    component.filterForm.setValue({
      code: 'B-001',
      breed_id: 'breed-1',
      origin: 'Externo',
      active: 'true',
    });
    await component.search();

    stub.listBoars = vi.fn(() => of([buildBoar('1')]));
    await component.clearFilters();

    expect(stub.listBoars).toHaveBeenCalledWith({});
    expect(component.filterForm.getRawValue()).toEqual({
      code: '',
      breed_id: '',
      origin: '',
      active: '',
    });
  });

  it('does not reload when clearing without applied filters', async () => {
    const component = create();
    component.filterForm.patchValue({ code: 'typed but not applied' });

    await component.clearFilters();

    expect(stub.listBoars).not.toHaveBeenCalled();
    expect(component.filterForm.getRawValue()).toEqual({
      code: '',
      breed_id: '',
      origin: '',
      active: '',
    });
  });

  it('keeps the applied filters when reloading after a create', async () => {
    const component = create();
    component.filterForm.patchValue({ code: 'B-001' });
    await component.search();

    stub.listBoars = vi.fn(() => of([buildBoar('1')]));
    await component.onCreated('B-005');

    expect(stub.listBoars).toHaveBeenCalledWith({ code: 'B-001' });
  });

  it('shows a filtered empty message when filters match nothing', async () => {
    stub.listBoars = vi.fn(() => of([]));
    const fixture = TestBed.createComponent(BoarsPage);
    const component = fixture.componentInstance as any;
    fixture.detectChanges();
    await fixture.whenStable();

    component.filterForm.patchValue({ code: 'zzz' });
    await component.search();
    fixture.detectChanges();

    expect(fixture.nativeElement.textContent).toContain(
      'No hay verracos que coincidan con los filtros',
    );
  });
});
