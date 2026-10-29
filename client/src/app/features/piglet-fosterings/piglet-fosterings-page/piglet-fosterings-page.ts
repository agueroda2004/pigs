import { Component, OnInit, computed, inject, signal } from '@angular/core';
import { FormBuilder, ReactiveFormsModule } from '@angular/forms';
import { firstValueFrom } from 'rxjs';

import { AuthService } from '../../../core/auth/auth.service';
import { NotificationService } from '../../../core/notifications/notification.service';
import {
  PigletFostering,
  PigletFosteringFilters,
} from '../../../core/piglet-fosterings/piglet-fostering.models';
import { PigletFosteringsService } from '../../../core/piglet-fosterings/piglet-fosterings.service';
import { DatePicker } from '../../../shared/ui/date-picker/date-picker';
import { toISODate } from '../../../shared/utils/date';
import { DeletePigletFosteringModal } from '../delete-piglet-fostering-modal/delete-piglet-fostering-modal';
import { EditPigletFosteringModal } from '../edit-piglet-fostering-modal/edit-piglet-fostering-modal';
import { PigletFosteringCard } from '../piglet-fostering-card/piglet-fostering-card';
import {
  CreatedFostering,
  RegisterPigletFosteringModal,
} from '../register-piglet-fostering-modal/register-piglet-fostering-modal';

@Component({
  selector: 'app-piglet-fosterings-page',
  imports: [
    ReactiveFormsModule,
    PigletFosteringCard,
    RegisterPigletFosteringModal,
    EditPigletFosteringModal,
    DeletePigletFosteringModal,
    DatePicker,
  ],
  styleUrl: './piglet-fosterings-page.css',
  templateUrl: './piglet-fosterings-page.html',
})
export class PigletFosteringsPage implements OnInit {
  private readonly pigletFosteringsService = inject(PigletFosteringsService);
  private readonly notifications = inject(NotificationService);
  private readonly auth = inject(AuthService);
  private readonly formBuilder = inject(FormBuilder);

  protected readonly isAdmin = this.auth.isAdmin;
  protected readonly fosterings = signal<PigletFostering[]>([]);
  protected readonly appliedFilters = signal<PigletFosteringFilters>({});
  protected readonly loading = signal(false);
  protected readonly error = signal(false);
  protected readonly modalOpen = signal(false);
  protected readonly editOpen = signal(false);
  protected readonly editingFostering = signal<PigletFostering | null>(null);
  protected readonly deleteOpen = signal(false);
  protected readonly deletingFostering = signal<PigletFostering | null>(null);
  protected readonly today = toISODate(new Date());
  protected readonly rangeError = signal<string | null>(null);
  protected readonly page = signal(1);
  protected readonly totalPages = signal(0);
  protected readonly total = signal(0);

  protected readonly hasFilters = computed(() => Object.keys(this.appliedFilters()).length > 0);

  protected readonly filterForm = this.formBuilder.nonNullable.group({
    sow_code: [''],
    from: [''],
    to: [''],
  });

  async ngOnInit(): Promise<void> {
    await this.loadFosterings();
  }

  protected openModal(): void {
    this.editingFostering.set(null);
    this.modalOpen.set(true);
  }

  protected openEditModal(fostering: PigletFostering): void {
    this.editingFostering.set(fostering);
    this.editOpen.set(true);
  }

  protected closeModal(): void {
    this.modalOpen.set(false);
  }

  protected closeEditModal(): void {
    this.editOpen.set(false);
    this.editingFostering.set(null);
  }

  protected async onCreated(codes: CreatedFostering): Promise<void> {
    this.notifications.success(
      `Traslado de "${codes.donor}" a "${codes.receiver}" registrado correctamente`,
    );
    this.page.set(1);
    await this.loadFosterings();
  }

  protected async onUpdated(): Promise<void> {
    this.notifications.success('Traslado actualizado correctamente');
    this.closeEditModal();
    await this.loadFosterings();
  }

  protected openDelete(fostering: PigletFostering): void {
    this.deletingFostering.set(fostering);
    this.deleteOpen.set(true);
  }

  protected closeDelete(): void {
    this.deleteOpen.set(false);
  }

  protected async onDeleted(fostering: PigletFostering): Promise<void> {
    this.deleteOpen.set(false);
    this.deletingFostering.set(null);
    const donor = fostering.donor_sow_code || 'sin identificar';
    const receiver = fostering.receiver_sow_code || 'sin identificar';
    this.notifications.success(`Traslado de "${donor}" a "${receiver}" eliminado correctamente`);
    await this.loadFosterings();
  }

  protected async search(): Promise<void> {
    const { sow_code, from, to } = this.filterForm.getRawValue();
    if (from && to && from > to) {
      this.rangeError.set('La fecha inicial no puede ser posterior a la fecha final');
      return;
    }
    this.rangeError.set(null);

    const filters: PigletFosteringFilters = {};
    const trimmedCode = sow_code.trim();
    if (trimmedCode) {
      filters.sow_code = trimmedCode;
    }
    if (from) {
      filters.from = from;
    }
    if (to) {
      filters.to = to;
    }
    this.appliedFilters.set(filters);
    this.page.set(1);
    await this.loadFosterings();
  }

  protected async clearFilters(): Promise<void> {
    const hadFilters = this.hasFilters();
    this.filterForm.reset();
    this.rangeError.set(null);
    this.appliedFilters.set({});
    if (hadFilters) {
      this.page.set(1);
      await this.loadFosterings();
    }
  }

  protected async previousPage(): Promise<void> {
    if (this.page() > 1) {
      this.page.update((page) => page - 1);
      await this.loadFosterings();
    }
  }

  protected async nextPage(): Promise<void> {
    if (this.page() < this.totalPages()) {
      this.page.update((page) => page + 1);
      await this.loadFosterings();
    }
  }

  protected async loadFosterings(): Promise<void> {
    this.loading.set(true);
    this.error.set(false);
    try {
      const result = await firstValueFrom(
        this.pigletFosteringsService.listPigletFosterings(this.appliedFilters(), this.page()),
      );
      if (result.total_pages > 0 && this.page() > result.total_pages) {
        this.page.set(result.total_pages);
        await this.loadFosterings();
        return;
      }
      this.fosterings.set(result.items);
      this.total.set(result.total);
      this.totalPages.set(result.total_pages);
    } catch {
      this.error.set(true);
    } finally {
      this.loading.set(false);
    }
  }
}
