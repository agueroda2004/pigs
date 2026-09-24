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
    created_at: '2026-01-02T12:00:00',
    updated_at: '2026-01-02T12:00:00',
    created_by: 'admin',
    updated_by: 'admin',
  };
}

class SowsStub {
  listSows = vi.fn(() => of([buildSow('1'), buildSow('2')]));
}

class BreedsStub {
  listBreeds = vi.fn(() => of([{ id: 'breed-1', name: 'Duroc' }]));
}

class NotificationsStub {
  success = vi.fn();
  error = vi.fn();
}

describe('SowsPage', () => {
  let stub: SowsStub;
  let notifications: NotificationsStub;
  let isAdmin: WritableSignal<boolean>;

  beforeEach(async () => {
    stub = new SowsStub();
    notifications = new NotificationsStub();
    isAdmin = signal(true);
    await TestBed.configureTestingModule({
      imports: [SowsPage],
      providers: [
        { provide: SowsService, useValue: stub },
        { provide: BreedsService, useValue: new BreedsStub() },
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
    stub.listSows = vi.fn(() => throwError(() => new Error('failed')));
    const component = create();

    await component.loadSows();
    expect(component.error()).toBe(true);

    stub.listSows = vi.fn(() => of([buildSow('1')]));
    await component.loadSows();

    expect(component.error()).toBe(false);
    expect(component.sows()).toHaveLength(1);
  });

  it('reloads the list after creating a sow', async () => {
    const component = create();
    await component.loadSows();

    stub.listSows = vi.fn(() => of([buildSow('1'), buildSow('2'), buildSow('3')]));
    await component.onCreated('C-004');

    expect(notifications.success).toHaveBeenCalledWith('Cerda "C-004" creada correctamente');
    expect(component.modalOpen()).toBe(false);
    expect(component.sows()).toHaveLength(3);
  });

  it('reloads the list after updating a sow', async () => {
    const component = create();
    await component.loadSows();

    stub.listSows = vi.fn(() => of([buildSow('9')]));
    await component.onUpdated(buildSow('9'));

    expect(notifications.success).toHaveBeenCalledWith('Cerda "C-009" actualizada correctamente');
    expect(component.editOpen()).toBe(false);
    expect(component.sows()).toHaveLength(1);
  });

  it('hides the create button for non-admins', () => {
    isAdmin.set(false);
    const fixture = TestBed.createComponent(SowsPage);
    fixture.detectChanges();

    expect(fixture.nativeElement.textContent).not.toContain('Crear cerda');
  });

  it('loads breed options for the filter dropdown', async () => {
    const fixture = TestBed.createComponent(SowsPage);
    // eslint-disable-next-line @typescript-eslint/no-explicit-any
    const component = fixture.componentInstance as any;

    fixture.detectChanges();
    await fixture.whenStable();

    expect(component.breedOptions()).toEqual([{ value: 'breed-1', label: 'Duroc' }]);
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

    expect(stub.listSows).toHaveBeenLastCalledWith({
      code: 'C-001',
      breed_id: 'breed-1',
      origin: 'Externo',
      state: 'Gestando',
      active: false,
    });
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

    stub.listSows = vi.fn(() => of([buildSow('1')]));
    await component.clearFilters();

    expect(stub.listSows).toHaveBeenCalledWith({});
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

    stub.listSows = vi.fn(() => of([buildSow('1')]));
    await component.onCreated('C-005');

    expect(stub.listSows).toHaveBeenCalledWith({ code: 'C-001' });
  });

  it('shows a filtered empty message when filters match nothing', async () => {
    stub.listSows = vi.fn(() => of([]));
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
