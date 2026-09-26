import { Component, OnInit, computed, inject, signal } from '@angular/core';
import { FormBuilder, ReactiveFormsModule } from '@angular/forms';
import { firstValueFrom } from 'rxjs';

import { Abortion, AbortionFilters } from '../../../core/abortions/abortion.models';
import { AbortionsService } from '../../../core/abortions/abortions.service';
import { AuthService } from '../../../core/auth/auth.service';
import { NotificationService } from '../../../core/notifications/notification.service';
import { SowsService } from '../../../core/sows/sows.service';
import { DropdownOption } from '../../../shared/ui/dropdown/dropdown';
import { SearchDropdown } from '../../../shared/ui/search-dropdown/search-dropdown';
import { AbortionCard } from '../abortion-card/abortion-card';
import { RegisterAbortionModal } from '../register-abortion-modal/register-abortion-modal';

@Component({
  selector: 'app-abortions-page',
  imports: [ReactiveFormsModule, AbortionCard, RegisterAbortionModal, SearchDropdown],
  styleUrl: './abortions-page.css',
  templateUrl: './abortions-page.html',
})
export class AbortionsPage implements OnInit {
  private readonly abortionsService = inject(AbortionsService);
  private readonly sowsService = inject(SowsService);
  private readonly notifications = inject(NotificationService);
  private readonly auth = inject(AuthService);
  private readonly formBuilder = inject(FormBuilder);

  protected readonly isAdmin = this.auth.isAdmin;
  protected readonly abortions = signal<Abortion[]>([]);
  protected readonly sowCodes = signal<Map<string, string>>(new Map());
  protected readonly sowOptions = signal<DropdownOption[]>([]);
  protected readonly appliedFilters = signal<AbortionFilters>({});
  protected readonly loading = signal(false);
  protected readonly error = signal(false);
  protected readonly modalOpen = signal(false);

  protected readonly hasFilters = computed(() => Object.keys(this.appliedFilters()).length > 0);

  protected readonly filterForm = this.formBuilder.nonNullable.group({
    sow_id: [''],
  });

  async ngOnInit(): Promise<void> {
    await Promise.all([this.loadAbortions(), this.loadLookups()]);
  }

  protected openModal(): void {
    this.modalOpen.set(true);
  }

  protected closeModal(): void {
    this.modalOpen.set(false);
  }

  protected async onCreated(sowCode: string): Promise<void> {
    this.notifications.success(`Aborto de la cerda "${sowCode}" registrado correctamente`);
    await Promise.all([this.loadAbortions(), this.loadLookups()]);
  }

  protected sowCode(abortion: Abortion): string | null {
    return this.sowCodes().get(abortion.sow_id) ?? null;
  }

  protected async search(): Promise<void> {
    const { sow_id } = this.filterForm.getRawValue();
    const filters: AbortionFilters = {};
    if (sow_id) {
      filters.sow_id = sow_id;
    }
    this.appliedFilters.set(filters);
    await this.loadAbortions();
  }

  protected async clearFilters(): Promise<void> {
    const hadFilters = this.hasFilters();
    this.filterForm.reset();
    this.appliedFilters.set({});
    if (hadFilters) {
      await this.loadAbortions();
    }
  }

  protected async loadAbortions(): Promise<void> {
    this.loading.set(true);
    this.error.set(false);
    try {
      this.abortions.set(
        await firstValueFrom(this.abortionsService.listAbortions(this.appliedFilters())),
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
