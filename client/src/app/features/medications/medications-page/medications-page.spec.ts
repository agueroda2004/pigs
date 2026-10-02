import { WritableSignal, signal } from '@angular/core';
import { TestBed } from '@angular/core/testing';
import { of, throwError } from 'rxjs';
import { vi } from 'vitest';

import { AuthService } from '../../../core/auth/auth.service';
import { Medication } from '../../../core/medications/medication.models';
import { MedicationsService } from '../../../core/medications/medications.service';
import { NotificationService } from '../../../core/notifications/notification.service';
import { MedicationsPage } from './medications-page';

function buildMedication(id: string): Medication {
  return {
    id,
    name: `Medicamento ${id}`,
    active: true,
    created_at: '2026-01-02T12:00:00',
    updated_at: '2026-01-02T12:00:00',
    created_by: 'admin',
    updated_by: 'admin',
  };
}

class MedicationsStub {
  listMedications = vi.fn(() => of([buildMedication('1'), buildMedication('2')]));
}

class NotificationsStub {
  success = vi.fn();
  error = vi.fn();
}

describe('MedicationsPage', () => {
  let stub: MedicationsStub;
  let notifications: NotificationsStub;
  let isAdmin: WritableSignal<boolean>;

  beforeEach(async () => {
    stub = new MedicationsStub();
    notifications = new NotificationsStub();
    isAdmin = signal(true);
    await TestBed.configureTestingModule({
      imports: [MedicationsPage],
      providers: [
        { provide: MedicationsService, useValue: stub },
        { provide: NotificationService, useValue: notifications },
        { provide: AuthService, useValue: { isAdmin } },
      ],
    }).compileComponents();
  });

  // eslint-disable-next-line @typescript-eslint/no-explicit-any
  const create = () => TestBed.createComponent(MedicationsPage).componentInstance as any;

  it('loads medications on init and renders a card per medication', async () => {
    const fixture = TestBed.createComponent(MedicationsPage);
    // eslint-disable-next-line @typescript-eslint/no-explicit-any
    const component = fixture.componentInstance as any;

    fixture.detectChanges();
    await fixture.whenStable();
    fixture.detectChanges();

    expect(stub.listMedications).toHaveBeenCalled();
    expect(component.medications()).toHaveLength(2);
    expect(fixture.nativeElement.querySelectorAll('app-medication-card')).toHaveLength(2);
  });

  it('shows the error state and retries', async () => {
    stub.listMedications = vi.fn(() => throwError(() => new Error('failed')));
    const component = create();

    await component.loadMedications();
    expect(component.error()).toBe(true);

    stub.listMedications = vi.fn(() => of([buildMedication('1')]));
    await component.loadMedications();

    expect(component.error()).toBe(false);
    expect(component.medications()).toHaveLength(1);
  });

  it('reloads the list after creating a medication', async () => {
    const component = create();
    await component.loadMedications();

    stub.listMedications = vi.fn(() =>
      of([buildMedication('1'), buildMedication('2'), buildMedication('3')]),
    );
    await component.onCreated('Ivermectina');

    expect(notifications.success).toHaveBeenCalledWith(
      'Medicamento "Ivermectina" creado correctamente',
    );
    expect(component.modalOpen()).toBe(false);
    expect(component.medications()).toHaveLength(3);
  });

  it('reloads the list after updating a medication', async () => {
    const component = create();
    await component.loadMedications();

    stub.listMedications = vi.fn(() => of([buildMedication('9')]));
    await component.onUpdated(buildMedication('9'));

    expect(notifications.success).toHaveBeenCalledWith(
      'Medicamento "Medicamento 9" actualizado correctamente',
    );
    expect(component.editOpen()).toBe(false);
    expect(component.medications()).toHaveLength(1);
  });

  it('hides the create button for non-admins', () => {
    isAdmin.set(false);
    const fixture = TestBed.createComponent(MedicationsPage);
    fixture.detectChanges();

    expect(fixture.nativeElement.textContent).not.toContain('Crear medicamento');
  });

  it('applies the filter values when searching', async () => {
    const component = create();
    component.filterForm.setValue({ name: '  Iver  ', active: 'false' });

    await component.search();

    expect(stub.listMedications).toHaveBeenLastCalledWith({ name: 'Iver', active: false });
  });

  it('clears the filters and reloads without them', async () => {
    const component = create();
    component.filterForm.setValue({ name: 'Iver', active: 'true' });
    await component.search();

    stub.listMedications = vi.fn(() => of([buildMedication('1')]));
    await component.clearFilters();

    expect(stub.listMedications).toHaveBeenCalledWith({});
    expect(component.filterForm.getRawValue()).toEqual({ name: '', active: '' });
  });

  it('does not reload when clearing without applied filters', async () => {
    const component = create();
    component.filterForm.patchValue({ name: 'typed but not applied' });

    await component.clearFilters();

    expect(stub.listMedications).not.toHaveBeenCalled();
    expect(component.filterForm.getRawValue()).toEqual({ name: '', active: '' });
  });

  it('keeps the applied filters when reloading after a create', async () => {
    const component = create();
    component.filterForm.patchValue({ name: 'Iver' });
    await component.search();

    stub.listMedications = vi.fn(() => of([buildMedication('1')]));
    await component.onCreated('Ivermectina');

    expect(stub.listMedications).toHaveBeenCalledWith({ name: 'Iver' });
  });

  it('shows a filtered empty message when filters match nothing', async () => {
    stub.listMedications = vi.fn(() => of([]));
    const fixture = TestBed.createComponent(MedicationsPage);
    // eslint-disable-next-line @typescript-eslint/no-explicit-any
    const component = fixture.componentInstance as any;
    fixture.detectChanges();
    await fixture.whenStable();

    component.filterForm.patchValue({ name: 'zzz' });
    await component.search();
    fixture.detectChanges();

    expect(fixture.nativeElement.textContent).toContain(
      'No hay medicamentos que coincidan con los filtros',
    );
  });
});
