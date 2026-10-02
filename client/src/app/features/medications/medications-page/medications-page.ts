import { Component, OnInit, computed, inject, signal } from '@angular/core';
import { FormBuilder, ReactiveFormsModule } from '@angular/forms';
import { firstValueFrom } from 'rxjs';

import { AuthService } from '../../../core/auth/auth.service';
import { Medication, MedicationFilters } from '../../../core/medications/medication.models';
import { MedicationsService } from '../../../core/medications/medications.service';
import { NotificationService } from '../../../core/notifications/notification.service';
import { Dropdown, DropdownOption } from '../../../shared/ui/dropdown/dropdown';
import { CreateMedicationModal } from '../create-medication-modal/create-medication-modal';
import { EditMedicationModal } from '../edit-medication-modal/edit-medication-modal';
import { MedicationCard } from '../medication-card/medication-card';

@Component({
  selector: 'app-medications-page',
  imports: [ReactiveFormsModule, MedicationCard, CreateMedicationModal, EditMedicationModal, Dropdown],
  styleUrl: './medications-page.css',
  templateUrl: './medications-page.html',
})
export class MedicationsPage implements OnInit {
  private readonly medicationsService = inject(MedicationsService);
  private readonly notifications = inject(NotificationService);
  private readonly auth = inject(AuthService);
  private readonly formBuilder = inject(FormBuilder);

  protected readonly isAdmin = this.auth.isAdmin;
  protected readonly medications = signal<Medication[]>([]);
  protected readonly appliedFilters = signal<MedicationFilters>({});
  protected readonly loading = signal(false);
  protected readonly error = signal(false);
  protected readonly modalOpen = signal(false);
  protected readonly editOpen = signal(false);
  protected readonly editingMedication = signal<Medication | null>(null);

  protected readonly activeOptions: DropdownOption[] = [
    { value: '', label: 'Todos' },
    { value: 'true', label: 'Activos' },
    { value: 'false', label: 'Inactivos' },
  ];

  protected readonly hasFilters = computed(() => Object.keys(this.appliedFilters()).length > 0);

  protected readonly filterForm = this.formBuilder.nonNullable.group({
    name: [''],
    active: [''],
  });

  async ngOnInit(): Promise<void> {
    await this.loadMedications();
  }

  protected openModal(): void {
    this.modalOpen.set(true);
  }

  protected closeModal(): void {
    this.modalOpen.set(false);
  }

  protected async onCreated(name: string): Promise<void> {
    this.modalOpen.set(false);
    this.notifications.success(`Medicamento "${name}" creado correctamente`);
    await this.loadMedications();
  }

  protected openEdit(medication: Medication): void {
    this.editingMedication.set(medication);
    this.editOpen.set(true);
  }

  protected closeEdit(): void {
    this.editOpen.set(false);
  }

  protected async onUpdated(medication: Medication): Promise<void> {
    this.editOpen.set(false);
    this.notifications.success(`Medicamento "${medication.name}" actualizado correctamente`);
    await this.loadMedications();
  }

  protected async search(): Promise<void> {
    const { name, active } = this.filterForm.getRawValue();
    const filters: MedicationFilters = {};
    const trimmedName = name.trim();
    if (trimmedName) {
      filters.name = trimmedName;
    }
    if (active) {
      filters.active = active === 'true';
    }
    this.appliedFilters.set(filters);
    await this.loadMedications();
  }

  protected async clearFilters(): Promise<void> {
    const hadFilters = this.hasFilters();
    this.filterForm.reset();
    this.appliedFilters.set({});
    if (hadFilters) {
      await this.loadMedications();
    }
  }

  protected async loadMedications(): Promise<void> {
    this.loading.set(true);
    this.error.set(false);
    try {
      this.medications.set(
        await firstValueFrom(this.medicationsService.listMedications(this.appliedFilters())),
      );
    } catch {
      this.error.set(true);
    } finally {
      this.loading.set(false);
    }
  }
}
