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
import { SowsService } from '../../../core/sows/sows.service';
import { DatePicker } from '../../../shared/ui/date-picker/date-picker';
import { DropdownOption } from '../../../shared/ui/dropdown/dropdown';
import { SearchDropdown } from '../../../shared/ui/search-dropdown/search-dropdown';
import { toISODate } from '../../../shared/utils/date';
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
    SearchDropdown,
    DatePicker,
  ],
  styleUrl: './piglet-fosterings-page.css',
  templateUrl: './piglet-fosterings-page.html',
})
export class PigletFosteringsPage implements OnInit {
  private readonly pigletFosteringsService = inject(PigletFosteringsService);
  private readonly sowsService = inject(SowsService);
  private readonly notifications = inject(NotificationService);
  private readonly auth = inject(AuthService);
  private readonly formBuilder = inject(FormBuilder);

  protected readonly isAdmin = this.auth.isAdmin;
  protected readonly fosterings = signal<PigletFostering[]>([]);
  protected readonly sowCodes = signal<Map<string, string>>(new Map());
  protected readonly sowOptions = signal<DropdownOption[]>([]);
  protected readonly appliedFilters = signal<PigletFosteringFilters>({});
  protected readonly loading = signal(false);
  protected readonly error = signal(false);
  protected readonly modalOpen = signal(false);
  protected readonly today = toISODate(new Date());
  protected readonly rangeError = signal<string | null>(null);

  protected readonly hasFilters = computed(() => Object.keys(this.appliedFilters()).length > 0);

  protected readonly filterForm = this.formBuilder.nonNullable.group({
    donor_sow_id: [''],
    receiver_sow_id: [''],
    from: [''],
    to: [''],
  });

  async ngOnInit(): Promise<void> {
    await Promise.all([this.loadFosterings(), this.loadLookups()]);
  }

  protected openModal(): void {
    this.modalOpen.set(true);
  }

  protected closeModal(): void {
    this.modalOpen.set(false);
  }

  protected async onCreated(codes: CreatedFostering): Promise<void> {
    this.notifications.success(
      `Traslado de "${codes.donor}" a "${codes.receiver}" registrado correctamente`,
    );
    await this.loadFosterings();
  }

  protected donorSowCode(fostering: PigletFostering): string | null {
    return this.sowCodes().get(fostering.donor_sow_id) ?? null;
  }

  protected receiverSowCode(fostering: PigletFostering): string | null {
    return this.sowCodes().get(fostering.receiver_sow_id) ?? null;
  }

  protected async search(): Promise<void> {
    const { donor_sow_id, receiver_sow_id, from, to } = this.filterForm.getRawValue();
    if (from && to && from > to) {
      this.rangeError.set('La fecha inicial no puede ser posterior a la fecha final');
      return;
    }
    this.rangeError.set(null);

    const filters: PigletFosteringFilters = {};
    if (donor_sow_id) {
      filters.donor_sow_id = donor_sow_id;
    }
    if (receiver_sow_id) {
      filters.receiver_sow_id = receiver_sow_id;
    }
    if (from) {
      filters.from = from;
    }
    if (to) {
      filters.to = to;
    }
    this.appliedFilters.set(filters);
    await this.loadFosterings();
  }

  protected async clearFilters(): Promise<void> {
    const hadFilters = this.hasFilters();
    this.filterForm.reset();
    this.rangeError.set(null);
    this.appliedFilters.set({});
    if (hadFilters) {
      await this.loadFosterings();
    }
  }

  protected async loadFosterings(): Promise<void> {
    this.loading.set(true);
    this.error.set(false);
    try {
      this.fosterings.set(
        await firstValueFrom(
          this.pigletFosteringsService.listPigletFosterings(this.appliedFilters()),
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
