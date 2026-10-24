import { Component, OnInit, computed, inject, signal } from '@angular/core';
import { FormBuilder, ReactiveFormsModule } from '@angular/forms';
import { firstValueFrom } from 'rxjs';

import { AuthService } from '../../../core/auth/auth.service';
import { NotificationService } from '../../../core/notifications/notification.service';
import { PigletDeath, PigletDeathFilters } from '../../../core/piglet-deaths/piglet-death.models';
import { PigletDeathsService } from '../../../core/piglet-deaths/piglet-deaths.service';
import { DatePicker } from '../../../shared/ui/date-picker/date-picker';
import { toISODate } from '../../../shared/utils/date';
import { DeletePigletDeathModal } from '../delete-piglet-death-modal/delete-piglet-death-modal';
import { EditPigletDeathModal } from '../edit-piglet-death-modal/edit-piglet-death-modal';
import { PigletDeathCard } from '../piglet-death-card/piglet-death-card';
import { RegisterPigletDeathModal } from '../register-piglet-death-modal/register-piglet-death-modal';

@Component({
  selector: 'app-piglet-deaths-page',
  imports: [
    ReactiveFormsModule,
    PigletDeathCard,
    RegisterPigletDeathModal,
    EditPigletDeathModal,
    DeletePigletDeathModal,
    DatePicker,
  ],
  styleUrl: './piglet-deaths-page.css',
  templateUrl: './piglet-deaths-page.html',
})
export class PigletDeathsPage implements OnInit {
  private readonly pigletDeathsService = inject(PigletDeathsService);
  private readonly notifications = inject(NotificationService);
  private readonly auth = inject(AuthService);
  private readonly formBuilder = inject(FormBuilder);

  protected readonly isAdmin = this.auth.isAdmin;
  protected readonly deaths = signal<PigletDeath[]>([]);
  protected readonly appliedFilters = signal<PigletDeathFilters>({});
  protected readonly loading = signal(false);
  protected readonly error = signal(false);
  protected readonly modalOpen = signal(false);
  protected readonly editOpen = signal(false);
  protected readonly editingDeath = signal<PigletDeath | null>(null);
  protected readonly deleteOpen = signal(false);
  protected readonly deletingDeath = signal<PigletDeath | null>(null);
  protected readonly today = toISODate(new Date());
  protected readonly rangeError = signal<string | null>(null);
  protected readonly page = signal(1);
  protected readonly totalPages = signal(0);
  protected readonly total = signal(0);

  protected readonly hasFilters = computed(() => Object.keys(this.appliedFilters()).length > 0);
  protected readonly deletingSowCode = computed(() => this.deletingDeath()?.sow_code ?? null);
  protected readonly editingSowCode = computed(() => this.editingDeath()?.sow_code ?? null);

  protected readonly filterForm = this.formBuilder.nonNullable.group({
    sow_code: [''],
    from: [''],
    to: [''],
  });

  async ngOnInit(): Promise<void> {
    await this.loadDeaths();
  }

  protected openModal(): void {
    this.modalOpen.set(true);
  }

  protected closeModal(): void {
    this.modalOpen.set(false);
  }

  protected async onCreated(sowCode: string): Promise<void> {
    this.notifications.success(`Muerte de la cerda "${sowCode}" registrada correctamente`);
    this.page.set(1);
    await this.loadDeaths();
  }

  protected openDelete(death: PigletDeath): void {
    this.deletingDeath.set(death);
    this.deleteOpen.set(true);
  }

  protected closeDelete(): void {
    this.deleteOpen.set(false);
  }

  protected openEdit(death: PigletDeath): void {
    this.editingDeath.set(death);
    this.editOpen.set(true);
  }

  protected closeEdit(): void {
    this.editOpen.set(false);
  }

  protected async onUpdated(): Promise<void> {
    this.editOpen.set(false);
    this.notifications.success('Muerte actualizada correctamente');
    await this.loadDeaths();
  }

  protected async onDeleted(death: PigletDeath): Promise<void> {
    this.deleteOpen.set(false);
    const sowCode = death.sow_code || 'sin identificar';
    this.notifications.success(`Muerte de la cerda "${sowCode}" eliminada correctamente`);
    await this.loadDeaths();
  }

  protected async search(): Promise<void> {
    const { sow_code, from, to } = this.filterForm.getRawValue();
    if (from && to && from > to) {
      this.rangeError.set('La fecha inicial no puede ser posterior a la fecha final');
      return;
    }
    this.rangeError.set(null);

    const filters: PigletDeathFilters = {};
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
    await this.loadDeaths();
  }

  protected async clearFilters(): Promise<void> {
    const hadFilters = this.hasFilters();
    this.filterForm.reset();
    this.rangeError.set(null);
    this.appliedFilters.set({});
    if (hadFilters) {
      this.page.set(1);
      await this.loadDeaths();
    }
  }

  protected async previousPage(): Promise<void> {
    if (this.page() > 1) {
      this.page.update((page) => page - 1);
      await this.loadDeaths();
    }
  }

  protected async nextPage(): Promise<void> {
    if (this.page() < this.totalPages()) {
      this.page.update((page) => page + 1);
      await this.loadDeaths();
    }
  }

  protected async loadDeaths(): Promise<void> {
    this.loading.set(true);
    this.error.set(false);
    try {
      const result = await firstValueFrom(
        this.pigletDeathsService.listPigletDeaths(this.appliedFilters(), this.page()),
      );
      if (result.total_pages > 0 && this.page() > result.total_pages) {
        this.page.set(result.total_pages);
        await this.loadDeaths();
        return;
      }
      this.deaths.set(result.items);
      this.total.set(result.total);
      this.totalPages.set(result.total_pages);
    } catch {
      this.error.set(true);
    } finally {
      this.loading.set(false);
    }
  }
}
