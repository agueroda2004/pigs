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
    created_at: '2026-01-02T12:00:00',
    updated_at: '2026-01-02T12:00:00',
    created_by: 'admin',
    updated_by: 'admin',
  };
}

class BoarsStub {
  listBoars = vi.fn(() => of([buildBoar('1'), buildBoar('2')]));
}

class BreedsStub {
  listBreeds = vi.fn(() => of([{ id: 'breed-1', name: 'Duroc' }]));
  listBreedOptions = vi.fn(() => of([{ id: 'breed-1', name: 'Duroc' }]));
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
    await component.onUpdated(buildBoar('9'));

    expect(notifications.success).toHaveBeenCalledWith('Verraco "B-009" actualizado correctamente');
    expect(component.editOpen()).toBe(false);
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

  it('uses only active breeds for the filter while keeping names for every breed', async () => {
    breeds.listBreeds = vi.fn(() =>
      of([
        { id: 'breed-1', name: 'Duroc' },
        { id: 'breed-2', name: 'Retired' },
      ]),
    );
    breeds.listBreedOptions = vi.fn(() => of([{ id: 'breed-1', name: 'Duroc' }]));
    const component = create();

    await component.loadBreeds();

    expect(component.breedOptions()).toEqual([{ value: 'breed-1', label: 'Duroc' }]);
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
