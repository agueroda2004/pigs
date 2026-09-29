import { WritableSignal, signal } from '@angular/core';
import { TestBed } from '@angular/core/testing';
import { of, throwError } from 'rxjs';
import { vi } from 'vitest';

import { AuthService } from '../../../core/auth/auth.service';
import { BoarRemoval } from '../../../core/boar-removals/boar-removal.models';
import { BoarRemovalsService } from '../../../core/boar-removals/boar-removals.service';
import { BoarsService } from '../../../core/boars/boars.service';
import { NotificationService } from '../../../core/notifications/notification.service';
import { BoarRemovalsPage } from './boar-removals-page';

function buildRemoval(id: string): BoarRemoval {
  return {
    id,
    boar_id: 'boar-1',
    removal_date: '2026-01-20',
    type: 'Muerte',
    reason: 'Enfermedad',
    note: null,
    last_state: 'Vivo',
    created_at: '2026-01-20T12:00:00',
    updated_at: '2026-01-20T12:00:00',
    created_by: 'admin',
    updated_by: 'admin',
  };
}

class BoarRemovalsStub {
  listBoarRemovals = vi.fn(() => of([buildRemoval('1'), buildRemoval('2')]));
  createBoarRemoval = vi.fn(() => of(buildRemoval('3')));
  updateBoarRemoval = vi.fn(() => of(buildRemoval('1')));
  deleteBoarRemoval = vi.fn(() => of(void 0));
}

class BoarsStub {
  listBoars = vi.fn(() => of([{ id: 'boar-1', code: 'B-001' }]));
}

class NotificationsStub {
  success = vi.fn();
  error = vi.fn();
}

describe('BoarRemovalsPage', () => {
  let stub: BoarRemovalsStub;
  let notifications: NotificationsStub;
  let isAdmin: WritableSignal<boolean>;

  beforeEach(async () => {
    stub = new BoarRemovalsStub();
    notifications = new NotificationsStub();
    isAdmin = signal(true);
    await TestBed.configureTestingModule({
      imports: [BoarRemovalsPage],
      providers: [
        { provide: BoarRemovalsService, useValue: stub },
        { provide: BoarsService, useValue: new BoarsStub() },
        { provide: NotificationService, useValue: notifications },
        { provide: AuthService, useValue: { isAdmin } },
      ],
    }).compileComponents();
  });

  // eslint-disable-next-line @typescript-eslint/no-explicit-any
  const create = () => TestBed.createComponent(BoarRemovalsPage).componentInstance as any;

  it('loads removals on init and renders a card per removal', async () => {
    const fixture = TestBed.createComponent(BoarRemovalsPage);

    fixture.detectChanges();
    await fixture.whenStable();
    fixture.detectChanges();

    expect(stub.listBoarRemovals).toHaveBeenCalled();
    expect(fixture.nativeElement.querySelectorAll('app-boar-removal-card')).toHaveLength(2);
  });

  it('resolves the boar code from the lookup', async () => {
    const component = create();
    await component.loadLookups();

    expect(component.boarCode(buildRemoval('1'))).toBe('B-001');
  });

  it('shows the error state and retries', async () => {
    stub.listBoarRemovals = vi.fn(() => throwError(() => new Error('failed')));
    const component = create();

    await component.loadBoarRemovals();
    expect(component.error()).toBe(true);

    stub.listBoarRemovals = vi.fn(() => of([buildRemoval('1')]));
    await component.loadBoarRemovals();

    expect(component.error()).toBe(false);
    expect(component.removals()).toHaveLength(1);
  });

  it('reloads the list after registering a removal', async () => {
    const component = create();
    await component.loadBoarRemovals();

    stub.listBoarRemovals = vi.fn(() =>
      of([buildRemoval('1'), buildRemoval('2'), buildRemoval('3')]),
    );
    await component.onCreated('B-001');

    expect(notifications.success).toHaveBeenCalledWith('Verraco "B-001" removido correctamente');
    expect(component.removals()).toHaveLength(3);
  });

  it('hides the register button for non-admins', () => {
    isAdmin.set(false);
    const fixture = TestBed.createComponent(BoarRemovalsPage);
    fixture.detectChanges();

    expect(fixture.nativeElement.textContent).not.toContain('Registrar baja');
  });

  it('opens the edit modal with the selected removal', async () => {
    const component = create();
    await component.loadLookups();
    const removal = buildRemoval('1');

    component.openEdit(removal);

    expect(component.editingRemoval()).toBe(removal);
    expect(component.editModalOpen()).toBe(true);
    expect(component.editingBoarCode()).toBe('B-001');
  });

  it('reloads the list and closes the edit modal after updating', async () => {
    const component = create();
    await component.loadBoarRemovals();
    await component.loadLookups();
    component.openEdit(buildRemoval('1'));

    await component.onUpdated(buildRemoval('1'));

    expect(notifications.success).toHaveBeenCalled();
    expect(component.editModalOpen()).toBe(false);
  });

  it('hides the edit button for non-admins', async () => {
    isAdmin.set(false);
    const fixture = TestBed.createComponent(BoarRemovalsPage);
    fixture.detectChanges();
    await fixture.whenStable();
    fixture.detectChanges();

    expect(fixture.nativeElement.textContent).not.toContain('Editar');
  });

  it('opens the delete modal with the selected removal', async () => {
    const component = create();
    await component.loadLookups();
    const removal = buildRemoval('1');

    component.openDelete(removal);

    expect(component.deletingRemoval()).toBe(removal);
    expect(component.deleteModalOpen()).toBe(true);
    expect(component.deletingBoarCode()).toBe('B-001');
  });

  it('reloads the list and closes the delete modal after deleting', async () => {
    const component = create();
    await component.loadBoarRemovals();
    await component.loadLookups();
    component.openDelete(buildRemoval('1'));
    stub.listBoarRemovals = vi.fn(() => of([buildRemoval('2')]));

    await component.onDeleted();

    expect(notifications.success).toHaveBeenCalled();
    expect(component.deleteModalOpen()).toBe(false);
    expect(component.removals()).toHaveLength(1);
  });

  it('hides the delete button for non-admins', async () => {
    isAdmin.set(false);
    const fixture = TestBed.createComponent(BoarRemovalsPage);
    fixture.detectChanges();
    await fixture.whenStable();
    fixture.detectChanges();

    expect(fixture.nativeElement.textContent).not.toContain('Eliminar');
  });

  it('applies the boar filter when searching', async () => {
    const component = create();
    component.filterForm.setValue({ boar_id: 'boar-1' });

    await component.search();

    expect(stub.listBoarRemovals).toHaveBeenLastCalledWith({ boar_id: 'boar-1' });
  });

  it('clears the filters and reloads without them', async () => {
    const component = create();
    component.filterForm.setValue({ boar_id: 'boar-1' });
    await component.search();

    stub.listBoarRemovals = vi.fn(() => of([buildRemoval('1')]));
    await component.clearFilters();

    expect(stub.listBoarRemovals).toHaveBeenCalledWith({});
    expect(component.filterForm.getRawValue()).toEqual({ boar_id: '' });
  });
});
