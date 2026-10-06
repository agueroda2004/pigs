import { Component, OnInit, computed, inject, signal } from '@angular/core';
import { FormBuilder, ReactiveFormsModule } from '@angular/forms';
import { firstValueFrom } from 'rxjs';

import { AuthService } from '../../../core/auth/auth.service';
import { NotificationService } from '../../../core/notifications/notification.service';
import {
  PartialWeaging,
  PartialWeagingFilters,
} from '../../../core/partial-weagings/partial-weaging.models';
import { PartialWeagingsService } from '../../../core/partial-weagings/partial-weagings.service';
import { SowsService } from '../../../core/sows/sows.service';
import { DatePicker } from '../../../shared/ui/date-picker/date-picker';
import { DropdownOption } from '../../../shared/ui/dropdown/dropdown';
import { SearchDropdown } from '../../../shared/ui/search-dropdown/search-dropdown';
import { toISODate } from '../../../shared/utils/date';
import { PartialWeagingCard } from '../partial-weaging-card/partial-weaging-card';
import { RegisterPartialWeagingModal } from '../register-partial-weaging-modal/register-partial-weaging-modal';

@Component({
  selector: 'app-partial-weagings-page',
  imports: [
    ReactiveFormsModule,
    PartialWeagingCard,
    RegisterPartialWeagingModal,
    SearchDropdown,
    DatePicker,
  ],
  styleUrl: './partial-weagings-page.css',
  templateUrl: './partial-weagings-page.html',
})
export class PartialWeagingsPage implements OnInit {
  private readonly partialWeagingsService = inject(PartialWeagingsService);
  private readonly sowsService = inject(SowsService);
  private readonly notifications = inject(NotificationService);
  private readonly auth = inject(AuthService);
  private readonly formBuilder = inject(FormBuilder);

  protected readonly isAdmin = this.auth.isAdmin;
  protected readonly weagings = signal<PartialWeaging[]>([]);
  protected readonly sowCodes = signal<Map<string, string>>(new Map());
  protected readonly sowOptions = signal<DropdownOption[]>([]);
  protected readonly appliedFilters = signal<PartialWeagingFilters>({});
  protected readonly loading = signal(false);
  protected readonly error = signal(false);
  protected readonly modalOpen = signal(false);
  protected readonly today = toISODate(new Date());
  protected readonly rangeError = signal<string | null>(null);

  protected readonly hasFilters = computed(() => Object.keys(this.appliedFilters()).length > 0);

  protected readonly filterForm = this.formBuilder.nonNullable.group({
    sow_id: [''],
    from: [''],
    to: [''],
  });

  async ngOnInit(): Promise<void> {
    await Promise.all([this.loadWeagings(), this.loadLookups()]);
  }

  protected openModal(): void {
    this.modalOpen.set(true);
  }

  protected closeModal(): void {
    this.modalOpen.set(false);
  }

  protected async onCreated(sowCode: string): Promise<void> {
    this.notifications.success(`Destete parcial de "${sowCode}" registrado correctamente`);
    await this.loadWeagings();
  }

  protected sowCode(weaging: PartialWeaging): string | null {
    return this.sowCodes().get(weaging.sow_id) ?? null;
  }

  protected async search(): Promise<void> {
    const { sow_id, from, to } = this.filterForm.getRawValue();
    if (from && to && from > to) {
      this.rangeError.set('La fecha inicial no puede ser posterior a la fecha final');
      return;
    }
    this.rangeError.set(null);

    const filters: PartialWeagingFilters = {};
    if (sow_id) {
      filters.sow_id = sow_id;
    }
    if (from) {
      filters.from = from;
    }
    if (to) {
      filters.to = to;
    }
    this.appliedFilters.set(filters);
    await this.loadWeagings();
  }

  protected async clearFilters(): Promise<void> {
    const hadFilters = this.hasFilters();
    this.filterForm.reset();
    this.rangeError.set(null);
    this.appliedFilters.set({});
    if (hadFilters) {
      await this.loadWeagings();
    }
  }

  protected async loadWeagings(): Promise<void> {
    this.loading.set(true);
    this.error.set(false);
    try {
      this.weagings.set(
        await firstValueFrom(
          this.partialWeagingsService.listPartialWeagings(this.appliedFilters()),
        ),
      );
    } catch {
      this.error.set(true);
    } finally {
      this.loading.set(false);
    }
  }

  private async loadLookups(): Promise<void> {
    try {
      const sows = await firstValueFrom(this.sowsService.listSows());
      this.sowCodes.set(new Map(sows.map((sow) => [sow.id, sow.code])));
      this.sowOptions.set(sows.map((sow) => ({ value: sow.id, label: sow.code })));
    } catch {
      this.sowCodes.set(new Map());
      this.sowOptions.set([]);
    }
  }
}
