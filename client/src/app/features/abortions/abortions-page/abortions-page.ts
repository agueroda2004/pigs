import { Component, OnInit, computed, inject, signal } from '@angular/core';
import { FormBuilder, ReactiveFormsModule } from '@angular/forms';
import { firstValueFrom } from 'rxjs';

import { Abortion, AbortionFilters } from '../../../core/abortions/abortion.models';
import { AbortionsService } from '../../../core/abortions/abortions.service';
import { AuthService } from '../../../core/auth/auth.service';
import { NotificationService } from '../../../core/notifications/notification.service';
import { AbortionCard } from '../abortion-card/abortion-card';
import { DeleteAbortionModal } from '../delete-abortion-modal/delete-abortion-modal';
import { EditAbortionModal } from '../edit-abortion-modal/edit-abortion-modal';
import { RegisterAbortionModal } from '../register-abortion-modal/register-abortion-modal';

@Component({
  selector: 'app-abortions-page',
  imports: [
    ReactiveFormsModule,
    AbortionCard,
    DeleteAbortionModal,
    EditAbortionModal,
    RegisterAbortionModal,
  ],
  styleUrl: './abortions-page.css',
  templateUrl: './abortions-page.html',
})
export class AbortionsPage implements OnInit {
  private readonly abortionsService = inject(AbortionsService);
  private readonly notifications = inject(NotificationService);
  private readonly auth = inject(AuthService);
  private readonly formBuilder = inject(FormBuilder);

  protected readonly isAdmin = this.auth.isAdmin;
  protected readonly abortions = signal<Abortion[]>([]);
  protected readonly appliedFilters = signal<AbortionFilters>({});
  protected readonly loading = signal(false);
  protected readonly error = signal(false);
  protected readonly modalOpen = signal(false);
  protected readonly editOpen = signal(false);
  protected readonly editingAbortion = signal<Abortion | null>(null);
  protected readonly deleteOpen = signal(false);
  protected readonly deletingAbortion = signal<Abortion | null>(null);
  protected readonly page = signal(1);
  protected readonly totalPages = signal(0);
  protected readonly total = signal(0);

  protected readonly hasFilters = computed(() => Object.keys(this.appliedFilters()).length > 0);

  protected readonly filterForm = this.formBuilder.nonNullable.group({
    sow_code: [''],
  });

  async ngOnInit(): Promise<void> {
    await this.loadAbortions();
  }

  protected openModal(): void {
    this.modalOpen.set(true);
  }

  protected closeModal(): void {
    this.modalOpen.set(false);
  }

  protected async onCreated(sowCode: string): Promise<void> {
    this.notifications.success(`Aborto de la cerda "${sowCode}" registrado correctamente`);
    this.page.set(1);
    await this.loadAbortions();
  }

  protected openEdit(abortion: Abortion): void {
    this.editingAbortion.set(abortion);
    this.editOpen.set(true);
  }

  protected closeEdit(): void {
    this.editOpen.set(false);
  }

  protected async onUpdated(): Promise<void> {
    this.editOpen.set(false);
    this.notifications.success('Aborto actualizado correctamente');
    await this.loadAbortions();
  }

  protected openDelete(abortion: Abortion): void {
    this.deletingAbortion.set(abortion);
    this.deleteOpen.set(true);
  }

  protected closeDelete(): void {
    this.deleteOpen.set(false);
  }

  protected async onDeleted(abortion: Abortion): Promise<void> {
    this.deleteOpen.set(false);
    const sowCode = abortion.sow_code || 'sin identificar';
    this.notifications.success(`Aborto de la cerda "${sowCode}" eliminado correctamente`);
    await this.loadAbortions();
  }

  protected async search(): Promise<void> {
    const { sow_code } = this.filterForm.getRawValue();
    const filters: AbortionFilters = {};
    if (sow_code) {
      filters.sow_code = sow_code;
    }
    this.appliedFilters.set(filters);
    this.page.set(1);
    await this.loadAbortions();
  }

  protected async clearFilters(): Promise<void> {
    const hadFilters = this.hasFilters();
    this.filterForm.reset();
    this.appliedFilters.set({});
    if (hadFilters) {
      this.page.set(1);
      await this.loadAbortions();
    }
  }

  protected async previousPage(): Promise<void> {
    if (this.page() > 1) {
      this.page.update((page) => page - 1);
      await this.loadAbortions();
    }
  }

  protected async nextPage(): Promise<void> {
    if (this.page() < this.totalPages()) {
      this.page.update((page) => page + 1);
      await this.loadAbortions();
    }
  }

  protected async loadAbortions(): Promise<void> {
    this.loading.set(true);
    this.error.set(false);
    try {
      const result = await firstValueFrom(
        this.abortionsService.listAbortions(this.appliedFilters(), this.page()),
      );
      this.abortions.set(result.items);
      this.total.set(result.total);
      this.totalPages.set(result.total_pages);
    } catch {
      this.error.set(true);
    } finally {
      this.loading.set(false);
    }
  }
}
