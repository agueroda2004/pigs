import { Component, OnInit, computed, inject, signal } from '@angular/core';
import { FormBuilder, ReactiveFormsModule } from '@angular/forms';
import { firstValueFrom } from 'rxjs';

import { AuthService } from '../../../core/auth/auth.service';
import { NotificationService } from '../../../core/notifications/notification.service';
import { Weaging, WeagingFilters } from '../../../core/weagings/weaging.models';
import { WeagingsService } from '../../../core/weagings/weagings.service';
import { DatePicker } from '../../../shared/ui/date-picker/date-picker';
import { toISODate } from '../../../shared/utils/date';
import { RegisterWeagingModal } from '../register-weaging-modal/register-weaging-modal';
import { DeleteWeagingModal } from '../delete-weaging-modal/delete-weaging-modal';
import { EditWeagingModal } from '../edit-weaging-modal/edit-weaging-modal';
import { WeagingCard } from '../weaging-card/weaging-card';

@Component({
  selector: 'app-weagings-page',
  imports: [
    ReactiveFormsModule,
    WeagingCard,
    RegisterWeagingModal,
    EditWeagingModal,
    DeleteWeagingModal,
    DatePicker,
  ],
  styleUrl: './weagings-page.css',
  templateUrl: './weagings-page.html',
})
export class WeagingsPage implements OnInit {
  private readonly weagingsService = inject(WeagingsService);
  private readonly notifications = inject(NotificationService);
  private readonly auth = inject(AuthService);
  private readonly formBuilder = inject(FormBuilder);

  protected readonly isAdmin = this.auth.isAdmin;
  protected readonly weagings = signal<Weaging[]>([]);
  protected readonly appliedFilters = signal<WeagingFilters>({});
  protected readonly loading = signal(false);
  protected readonly error = signal(false);
  protected readonly modalOpen = signal(false);
  protected readonly editOpen = signal(false);
  protected readonly deleteOpen = signal(false);
  protected readonly editingWeaging = signal<Weaging | null>(null);
  protected readonly selectedWeaging = signal<Weaging | null>(null);
  protected readonly today = toISODate(new Date());
  protected readonly rangeError = signal<string | null>(null);
  protected readonly page = signal(1);
  protected readonly totalPages = signal(0);
  protected readonly total = signal(0);

  protected readonly hasFilters = computed(() => Object.keys(this.appliedFilters()).length > 0);
  protected readonly editingSowCode = computed(() => this.editingWeaging()?.sow_code ?? null);
  protected readonly selectedSowCode = computed(() => this.selectedWeaging()?.sow_code ?? null);

  protected readonly filterForm = this.formBuilder.nonNullable.group({
    sow_code: [''],
    from: [''],
    to: [''],
  });

  async ngOnInit(): Promise<void> {
    await this.loadWeagings();
  }

  protected openModal(): void {
    this.modalOpen.set(true);
  }

  protected closeModal(): void {
    this.modalOpen.set(false);
  }

  protected async onCreated(sowCode: string): Promise<void> {
    this.notifications.success(`Destete de "${sowCode}" registrado correctamente`);
    this.page.set(1);
    await this.loadWeagings();
  }

  protected openEdit(weaging: Weaging): void {
    this.editingWeaging.set(weaging);
    this.editOpen.set(true);
  }

  protected closeEdit(): void {
    this.editOpen.set(false);
    this.editingWeaging.set(null);
  }

  protected async onUpdated(): Promise<void> {
    const sowCode = this.editingSowCode();
    this.editOpen.set(false);
    this.editingWeaging.set(null);
    this.notifications.success(
      `Destete de "${sowCode ?? 'sin identificar'}" actualizado correctamente`,
    );
    await this.loadWeagings();
  }

  protected openDelete(weaging: Weaging): void {
    this.selectedWeaging.set(weaging);
    this.deleteOpen.set(true);
  }

  protected closeDelete(): void {
    this.deleteOpen.set(false);
    this.selectedWeaging.set(null);
  }

  protected async onDeleted(weaging: Weaging): Promise<void> {
    this.deleteOpen.set(false);
    this.selectedWeaging.set(null);
    this.notifications.success(
      `Destete de "${weaging.sow_code || 'sin identificar'}" eliminado correctamente`,
    );
    await this.loadWeagings();
  }

  protected async search(): Promise<void> {
    const { sow_code, from, to } = this.filterForm.getRawValue();
    if (from && to && from > to) {
      this.rangeError.set('La fecha inicial no puede ser posterior a la fecha final');
      return;
    }
    this.rangeError.set(null);

    const filters: WeagingFilters = {};
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
    await this.loadWeagings();
  }

  protected async clearFilters(): Promise<void> {
    const hadFilters = this.hasFilters();
    this.filterForm.reset();
    this.rangeError.set(null);
    this.appliedFilters.set({});
    if (hadFilters) {
      this.page.set(1);
      await this.loadWeagings();
    }
  }

  protected async previousPage(): Promise<void> {
    if (this.page() > 1) {
      this.page.update((page) => page - 1);
      await this.loadWeagings();
    }
  }

  protected async nextPage(): Promise<void> {
    if (this.page() < this.totalPages()) {
      this.page.update((page) => page + 1);
      await this.loadWeagings();
    }
  }

  protected async loadWeagings(): Promise<void> {
    this.loading.set(true);
    this.error.set(false);
    try {
      const result = await firstValueFrom(
        this.weagingsService.listWeagings(this.appliedFilters(), this.page()),
      );
      if (result.total_pages > 0 && this.page() > result.total_pages) {
        this.page.set(result.total_pages);
        await this.loadWeagings();
        return;
      }
      this.weagings.set(result.items);
      this.total.set(result.total);
      this.totalPages.set(result.total_pages);
    } catch {
      this.error.set(true);
    } finally {
      this.loading.set(false);
    }
  }
}
