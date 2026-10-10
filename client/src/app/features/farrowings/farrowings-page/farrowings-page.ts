import { Component, OnInit, computed, inject, signal } from '@angular/core';
import { FormBuilder, ReactiveFormsModule } from '@angular/forms';
import { firstValueFrom } from 'rxjs';

import { AuthService } from '../../../core/auth/auth.service';
import { Farrowing, FarrowingFilters } from '../../../core/farrowings/farrowing.models';
import { FarrowingsService } from '../../../core/farrowings/farrowings.service';
import { NotificationService } from '../../../core/notifications/notification.service';
import { DatePicker } from '../../../shared/ui/date-picker/date-picker';
import { toISODate } from '../../../shared/utils/date';
import { DeleteFarrowingModal } from '../delete-farrowing-modal/delete-farrowing-modal';
import { EditFarrowingModal } from '../edit-farrowing-modal/edit-farrowing-modal';
import { FarrowingCard } from '../farrowing-card/farrowing-card';
import { RegisterFarrowingModal } from '../register-farrowing-modal/register-farrowing-modal';

@Component({
  selector: 'app-farrowings-page',
  imports: [
    ReactiveFormsModule,
    FarrowingCard,
    RegisterFarrowingModal,
    EditFarrowingModal,
    DeleteFarrowingModal,
    DatePicker,
  ],
  styleUrl: './farrowings-page.css',
  templateUrl: './farrowings-page.html',
})
export class FarrowingsPage implements OnInit {
  private readonly farrowingsService = inject(FarrowingsService);
  private readonly notifications = inject(NotificationService);
  private readonly auth = inject(AuthService);
  private readonly formBuilder = inject(FormBuilder);

  protected readonly isAdmin = this.auth.isAdmin;
  protected readonly farrowings = signal<Farrowing[]>([]);
  protected readonly appliedFilters = signal<FarrowingFilters>({});
  protected readonly loading = signal(false);
  protected readonly error = signal(false);
  protected readonly modalOpen = signal(false);
  protected readonly editOpen = signal(false);
  protected readonly editingFarrowing = signal<Farrowing | null>(null);
  protected readonly deleteOpen = signal(false);
  protected readonly deletingFarrowing = signal<Farrowing | null>(null);
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
    await this.loadFarrowings();
  }

  protected openModal(): void {
    this.modalOpen.set(true);
  }

  protected closeModal(): void {
    this.modalOpen.set(false);
  }

  protected async onCreated(sowCode: string): Promise<void> {
    this.notifications.success(`Parto de la cerda "${sowCode}" registrado correctamente`);
    this.page.set(1);
    await this.loadFarrowings();
  }

  protected openEdit(farrowing: Farrowing): void {
    this.editingFarrowing.set(farrowing);
    this.editOpen.set(true);
  }

  protected closeEdit(): void {
    this.editOpen.set(false);
  }

  protected async onUpdated(): Promise<void> {
    this.editOpen.set(false);
    this.notifications.success('Parto actualizado correctamente');
    await this.loadFarrowings();
  }

  protected openDelete(farrowing: Farrowing): void {
    this.deletingFarrowing.set(farrowing);
    this.deleteOpen.set(true);
  }

  protected closeDelete(): void {
    this.deleteOpen.set(false);
  }

  protected async onDeleted(farrowing: Farrowing): Promise<void> {
    this.deleteOpen.set(false);
    const sowCode = farrowing.sow_code || 'sin identificar';
    this.notifications.success(`Parto de la cerda "${sowCode}" eliminado correctamente`);
    await this.loadFarrowings();
  }

  protected async search(): Promise<void> {
    const { sow_code, from, to } = this.filterForm.getRawValue();
    if (from && to && from > to) {
      this.rangeError.set('La fecha inicial no puede ser posterior a la fecha final');
      return;
    }
    this.rangeError.set(null);

    const filters: FarrowingFilters = {};
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
    await this.loadFarrowings();
  }

  protected async clearFilters(): Promise<void> {
    const hadFilters = this.hasFilters();
    this.filterForm.reset();
    this.rangeError.set(null);
    this.appliedFilters.set({});
    if (hadFilters) {
      this.page.set(1);
      await this.loadFarrowings();
    }
  }

  protected async previousPage(): Promise<void> {
    if (this.page() > 1) {
      this.page.update((page) => page - 1);
      await this.loadFarrowings();
    }
  }

  protected async nextPage(): Promise<void> {
    if (this.page() < this.totalPages()) {
      this.page.update((page) => page + 1);
      await this.loadFarrowings();
    }
  }

  protected async loadFarrowings(): Promise<void> {
    this.loading.set(true);
    this.error.set(false);
    try {
      const result = await firstValueFrom(
        this.farrowingsService.listFarrowings(this.appliedFilters(), this.page()),
      );
      this.farrowings.set(result.items);
      this.total.set(result.total);
      this.totalPages.set(result.total_pages);
    } catch {
      this.error.set(true);
    } finally {
      this.loading.set(false);
    }
  }
}
