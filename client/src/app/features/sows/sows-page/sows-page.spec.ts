import { WritableSignal, signal } from '@angular/core';
import { TestBed } from '@angular/core/testing';
import { of, throwError } from 'rxjs';
import { vi } from 'vitest';

import { AuthService } from '../../../core/auth/auth.service';
import { BreedsService } from '../../../core/breeds/breeds.service';
import { NotificationService } from '../../../core/notifications/notification.service';
import { Sow } from '../../../core/sows/sow.models';
import { SowsService } from '../../../core/sows/sows.service';
import { SowsPage } from './sows-page';

function buildSow(id: string): Sow {
  return {
    id,
    code: `C-00${id}`,
    location: 'Corral A',
    active: true,
    entry_date: '2026-01-10',
    birth_date: null,
    note: null,
    state: 'Viva',
    origin: 'Propio',
    parity: 2,
    breed_id: 'breed-1',
  };
}

function pageOf(items: Sow[], total = items.length, totalPages = 1) {
  return { items, total, page: 1, page_size: 20, total_pages: totalPages };
}

class SowsStub {
  listSows = vi.fn((_filters?: unknown, _page?: number) =>
    of(pageOf([buildSow('1'), buildSow('2')])),
  );
}

class BreedsStub {
  listBreedDropdown = vi.fn(() => of([{ id: 'breed-1', name: 'Duroc', active: true }]));
}

class NotificationsStub {
  success = vi.fn();
  error = vi.fn();
}

describe('SowsPage', () => {
  let stub: SowsStub;
  let breeds: BreedsStub;
  let notifications: NotificationsStub;
  let isAdmin: WritableSignal<boolean>;

  beforeEach(async () => {
    stub = new SowsStub();
    breeds = new BreedsStub();
    notifications = new NotificationsStub();
    isAdmin = signal(true);
    await TestBed.configureTestingModule({
      imports: [SowsPage],
      providers: [
        { provide: SowsService, useValue: stub },
        { provide: BreedsService, useValue: breeds },
        { provide: NotificationService, useValue: notifications },
        { provide: AuthService, useValue: { isAdmin } },
      ],
    }).compileComponents();
  });

  // eslint-disable-next-line @typescript-eslint/no-explicit-any
  const create = () => TestBed.createComponent(SowsPage).componentInstance as any;

  it('loads sows on init and renders a card per sow', async () => {
    const fixture = TestBed.createComponent(SowsPage);
    // eslint-disable-next-line @typescript-eslint/no-explicit-any
    const component = fixture.componentInstance as any;

    fixture.detectChanges();
    await fixture.whenStable();
    fixture.detectChanges();

    expect(stub.listSows).toHaveBeenCalled();
    expect(component.sows()).toHaveLength(2);
    expect(fixture.nativeElement.querySelectorAll('app-sow-card')).toHaveLength(2);
  });

  it('shows the error state and retries', async () => {
    stub.listSows.mockReturnValue(throwError(() => new Error('failed')));
    const component = create();

    await component.loadSows();
    expect(component.error()).toBe(true);

    stub.listSows.mockReturnValue(of(pageOf([buildSow('1')])));
    await component.loadSows();

    expect(component.error()).toBe(false);
    expect(component.sows()).toHaveLength(1);
  });

  it('reloads the list after creating a sow', async () => {
    const component = create();
    await component.loadSows();

    stub.listSows.mockReturnValue(of(pageOf([buildSow('1'), buildSow('2'), buildSow('3')])));
    await component.onCreated('C-004');

    expect(notifications.success).toHaveBeenCalledWith('Cerda "C-004" creada correctamente');
    expect(component.modalOpen()).toBe(false);
    expect(component.sows()).toHaveLength(3);
    expect(stub.listSows).toHaveBeenCalledWith({}, 1);
  });

  it('reloads the list after updating a sow', async () => {
    const component = create();
    await component.loadSows();

    stub.listSows.mockReturnValue(of(pageOf([buildSow('9')])));
    await component.onUpdated('C-009');

    expect(notifications.success).toHaveBeenCalledWith('Cerda "C-009" actualizada correctamente');
    expect(component.editOpen()).toBe(false);
    expect(component.sows()).toHaveLength(1);
  });

  it('opens the delete modal for the selected sow', () => {
    const component = create();
    const sow = buildSow('9');

    component.openDelete(sow);

    expect(component.deleteOpen()).toBe(true);
    expect(component.deletingSow()).toEqual(sow);
  });

  it('reloads the list after deleting a sow', async () => {
    const component = create();
    await component.loadSows();

    stub.listSows.mockReturnValue(of(pageOf([buildSow('2')])));
    await component.onDeleted(buildSow('9'));

    expect(notifications.success).toHaveBeenCalledWith('Cerda "C-009" eliminada correctamente');
    expect(component.deleteOpen()).toBe(false);
    expect(component.sows()).toHaveLength(1);
  });

  it('hides the create button for non-admins', () => {
    isAdmin.set(false);
    const fixture = TestBed.createComponent(SowsPage);
    fixture.detectChanges();

    expect(fixture.nativeElement.textContent).not.toContain('Crear cerda');
  });

  it('loads breed options for the filter dropdown', async () => {
    const component = create();

    await component.loadBreeds();

    expect(component.breedOptions()).toEqual([{ value: 'breed-1', label: 'Duroc' }]);
  });

  it('builds the filter options and names from the dropdown breeds', async () => {
    breeds.listBreedDropdown = vi.fn(() =>
      of([
        { id: 'breed-1', name: 'Duroc', active: true },
        { id: 'breed-2', name: 'Retired', active: false },
      ]),
    );
    const component = create();

    await component.loadBreeds();

    expect(breeds.listBreedDropdown).toHaveBeenCalledWith(false);
    expect(component.breedOptions()).toEqual([
      { value: 'breed-1', label: 'Duroc' },
      { value: 'breed-2', label: 'Retired (inactiva)' },
    ]);
    expect(component.breedNames().get('breed-2')).toBe('Retired');
  });

  it('offers every state in the filter dropdown', () => {
    const component = create();
    const values = component.stateOptions.map((option: { value: string }) => option.value);

    expect(values).toContain('Gestando');
    expect(values).toContain('Destetada');
    expect(values).toHaveLength(9);
  });

  it('applies the filter values when searching', async () => {
    const component = create();
    component.filterForm.setValue({
      code: '  C-001  ',
      breed_id: 'breed-1',
      origin: 'Externo',
      state: 'Gestando',
      active: 'false',
    });

    await component.search();

    expect(stub.listSows).toHaveBeenLastCalledWith(
      {
        code: 'C-001',
        breed_id: 'breed-1',
        origin: 'Externo',
        state: 'Gestando',
        active: false,
      },
      1,
    );
  });

  it('clears the filters and reloads without them', async () => {
    const component = create();
    component.filterForm.setValue({
      code: 'C-001',
      breed_id: 'breed-1',
      origin: 'Externo',
      state: 'Gestando',
      active: 'true',
    });
    await component.search();

    stub.listSows.mockReturnValue(of(pageOf([buildSow('1')])));
    await component.clearFilters();

    expect(stub.listSows).toHaveBeenCalledWith({}, 1);
    expect(component.filterForm.getRawValue()).toEqual({
      code: '',
      breed_id: '',
      origin: '',
      state: '',
      active: '',
    });
  });

  it('does not reload when clearing without applied filters', async () => {
    const component = create();
    component.filterForm.patchValue({ code: 'typed but not applied' });

    await component.clearFilters();

    expect(stub.listSows).not.toHaveBeenCalled();
    expect(component.filterForm.getRawValue()).toEqual({
      code: '',
      breed_id: '',
      origin: '',
      state: '',
      active: '',
    });
  });

  it('keeps the applied filters when reloading after a create', async () => {
    const component = create();
    component.filterForm.patchValue({ code: 'C-001' });
    await component.search();

    stub.listSows.mockReturnValue(of(pageOf([buildSow('1')])));
    await component.onCreated('C-005');

    expect(stub.listSows).toHaveBeenCalledWith({ code: 'C-001' }, 1);
  });

  it('paginates forward and backward', async () => {
    stub.listSows.mockImplementation((_filters, page) =>
      of(pageOf([buildSow(String(page))], 40, 2)),
    );
    const component = create();
    await component.loadSows();

    expect(component.page()).toBe(1);
    expect(component.totalPages()).toBe(2);

    await component.nextPage();
    expect(component.page()).toBe(2);
    expect(stub.listSows).toHaveBeenLastCalledWith({}, 2);

    await component.previousPage();
    expect(component.page()).toBe(1);
    expect(stub.listSows).toHaveBeenLastCalledWith({}, 1);
  });

  it('shows a filtered empty message when filters match nothing', async () => {
    stub.listSows.mockReturnValue(of(pageOf([], 0, 0)));
    const fixture = TestBed.createComponent(SowsPage);
    // eslint-disable-next-line @typescript-eslint/no-explicit-any
    const component = fixture.componentInstance as any;
    fixture.detectChanges();
    await fixture.whenStable();

    component.filterForm.patchValue({ code: 'zzz' });
    await component.search();
    fixture.detectChanges();

    expect(fixture.nativeElement.textContent).toContain(
      'No hay cerdas que coincidan con los filtros',
    );
  });
});
