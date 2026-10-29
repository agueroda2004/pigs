import { Component, OnInit, computed, inject, signal } from '@angular/core';
import { FormBuilder, ReactiveFormsModule, Validators } from '@angular/forms';
import { firstValueFrom } from 'rxjs';

import { AuthService } from '../../../core/auth/auth.service';
import { NotificationService } from '../../../core/notifications/notification.service';
import {
  PartialWeaging,
  PartialWeagingFilters,
} from '../../../core/partial-weagings/partial-weaging.models';
import { PartialWeagingsService } from '../../../core/partial-weagings/partial-weagings.service';
import { DatePicker } from '../../../shared/ui/date-picker/date-picker';
import { toISODate } from '../../../shared/utils/date';
import { DeletePartialWeagingModal } from '../delete-partial-weaging-modal/delete-partial-weaging-modal';
import { EditPartialWeagingModal } from '../edit-partial-weaging-modal/edit-partial-weaging-modal';
import { PartialWeagingCard } from '../partial-weaging-card/partial-weaging-card';
import { RegisterPartialWeagingModal } from '../register-partial-weaging-modal/register-partial-weaging-modal';

@Component({
  selector: 'app-partial-weagings-page',
  imports: [
    ReactiveFormsModule,
    PartialWeagingCard,
    RegisterPartialWeagingModal,
    EditPartialWeagingModal,
    DeletePartialWeagingModal,
    DatePicker,
  ],
  styleUrl: './partial-weagings-page.css',
  templateUrl: './partial-weagings-page.html',
})
export class PartialWeagingsPage implements OnInit {
  private readonly partialWeagingsService = inject(PartialWeagingsService);
  private readonly notifications = inject(NotificationService);
  private readonly auth = inject(AuthService);
  private readonly formBuilder = inject(FormBuilder);

  protected readonly isAdmin = this.auth.isAdmin;
  protected readonly weagings = signal<PartialWeaging[]>([]);
  protected readonly appliedFilters = signal<PartialWeagingFilters>({});
  protected readonly loading = signal(false);
  protected readonly error = signal(false);
  protected readonly page = signal(1);
  protected readonly total = signal(0);
  protected readonly totalPages = signal(0);
  protected readonly modalOpen = signal(false);
  protected readonly editOpen = signal(false);
  protected readonly deleteOpen = signal(false);
  protected readonly editingWeaging = signal<PartialWeaging | null>(null);
  protected readonly selectedWeaging = signal<PartialWeaging | null>(null);
  protected readonly today = toISODate(new Date());
  protected readonly rangeError = signal<string | null>(null);

  protected readonly hasFilters = computed(() => Object.keys(this.appliedFilters()).length > 0);
  protected readonly selectedSowCode = computed(() => this.selectedWeaging()?.sow_code ?? null);
  protected readonly editingSowCode = computed(() => this.editingWeaging()?.sow_code ?? null);

  protected readonly filterForm = this.formBuilder.nonNullable.group({
    sow_code: ['', [Validators.maxLength(50)]],
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
    this.notifications.success(`Destete parcial de "${sowCode}" registrado correctamente`);
    this.page.set(1);
    await this.loadWeagings();
  }

  protected openEdit(weaging: PartialWeaging): void {
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
      `Destete parcial de "${sowCode ?? 'sin identificar'}" actualizado correctamente`,
    );
    await this.loadWeagings();
  }

  protected openDelete(weaging: PartialWeaging): void {
    this.selectedWeaging.set(weaging);
    this.deleteOpen.set(true);
  }

  protected closeDelete(): void {
    this.deleteOpen.set(false);
    this.selectedWeaging.set(null);
  }

  protected async onDeleted(weaging: PartialWeaging): Promise<void> {
    this.deleteOpen.set(false);
    this.selectedWeaging.set(null);
    this.notifications.success(
      `Destete parcial de "${weaging.sow_code || 'sin identificar'}" eliminado correctamente`,
    );
    await this.loadWeagings();
  }

  protected async search(): Promise<void> {
    if (this.filterForm.controls.sow_code.invalid) {
      this.filterForm.controls.sow_code.markAsTouched();
      return;
    }

    const { sow_code, from, to } = this.filterForm.getRawValue();
    if (from && to && from > to) {
      this.rangeError.set('La fecha inicial no puede ser posterior a la fecha final');
      return;
    }
    this.rangeError.set(null);

    const filters: PartialWeagingFilters = {};
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
        this.partialWeagingsService.listPartialWeagings(this.appliedFilters(), this.page()),
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
