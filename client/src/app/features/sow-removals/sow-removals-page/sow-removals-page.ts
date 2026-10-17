import { Component, OnInit, computed, inject, signal } from '@angular/core';
import { FormBuilder, ReactiveFormsModule } from '@angular/forms';
import { firstValueFrom } from 'rxjs';

import { AuthService } from '../../../core/auth/auth.service';
import { NotificationService } from '../../../core/notifications/notification.service';
import { SowRemoval, SowRemovalFilters } from '../../../core/sow-removals/sow-removal.models';
import { SowRemovalsService } from '../../../core/sow-removals/sow-removals.service';
import { DeleteSowRemovalModal } from '../delete-sow-removal-modal/delete-sow-removal-modal';
import { EditSowRemovalModal } from '../edit-sow-removal-modal/edit-sow-removal-modal';
import { RegisterSowRemovalModal } from '../register-sow-removal-modal/register-sow-removal-modal';
import { SowRemovalCard } from '../sow-removal-card/sow-removal-card';

@Component({
  selector: 'app-sow-removals-page',
  imports: [
    ReactiveFormsModule,
    SowRemovalCard,
    RegisterSowRemovalModal,
    EditSowRemovalModal,
    DeleteSowRemovalModal,
  ],
  styleUrl: './sow-removals-page.css',
  templateUrl: './sow-removals-page.html',
})
export class SowRemovalsPage implements OnInit {
  private readonly sowRemovalsService = inject(SowRemovalsService);
  private readonly notifications = inject(NotificationService);
  private readonly auth = inject(AuthService);
  private readonly formBuilder = inject(FormBuilder);

  protected readonly isAdmin = this.auth.isAdmin;
  protected readonly removals = signal<SowRemoval[]>([]);
  protected readonly appliedFilters = signal<SowRemovalFilters>({});
  protected readonly loading = signal(false);
  protected readonly error = signal(false);
  protected readonly modalOpen = signal(false);
  protected readonly editModalOpen = signal(false);
  protected readonly editingRemoval = signal<SowRemoval | null>(null);
  protected readonly deleteModalOpen = signal(false);
  protected readonly deletingRemoval = signal<SowRemoval | null>(null);

  protected readonly hasFilters = computed(() => Object.keys(this.appliedFilters()).length > 0);

  protected readonly filterForm = this.formBuilder.nonNullable.group({
    sow_code: [''],
  });

  async ngOnInit(): Promise<void> {
    await this.loadSowRemovals();
  }

  protected openModal(): void {
    this.modalOpen.set(true);
  }

  protected closeModal(): void {
    this.modalOpen.set(false);
  }

  protected async onCreated(sowCode: string): Promise<void> {
    this.notifications.success(`Cerda "${sowCode}" removida correctamente`);
    await this.loadSowRemovals();
  }

  protected openEdit(removal: SowRemoval): void {
    this.editingRemoval.set(removal);
    this.editModalOpen.set(true);
  }

  protected closeEdit(): void {
    this.editModalOpen.set(false);
  }

  protected async onUpdated(removal: SowRemoval): Promise<void> {
    this.notifications.success(`Baja de "${this.sowCode(removal) ?? 'la cerda'}" actualizada`);
    await this.loadSowRemovals();
    this.closeEdit();
  }

  protected editingSowCode(): string | null {
    const removal = this.editingRemoval();
    return removal ? this.sowCode(removal) : null;
  }

  protected openDelete(removal: SowRemoval): void {
    this.deletingRemoval.set(removal);
    this.deleteModalOpen.set(true);
  }

  protected closeDelete(): void {
    this.deleteModalOpen.set(false);
  }

  protected async onDeleted(): Promise<void> {
    const removal = this.deletingRemoval();
    this.notifications.success(
      `Baja de "${removal ? (this.sowCode(removal) ?? 'la cerda') : 'la cerda'}" eliminada`,
    );
    await this.loadSowRemovals();
    this.closeDelete();
  }

  protected deletingSowCode(): string | null {
    const removal = this.deletingRemoval();
    return removal ? this.sowCode(removal) : null;
  }

  protected sowCode(removal: SowRemoval): string | null {
    return removal.sow_code || null;
  }

  protected async search(): Promise<void> {
    const { sow_code } = this.filterForm.getRawValue();
    const filters: SowRemovalFilters = {};
    if (sow_code) {
      filters.sow_code = sow_code;
    }
    this.appliedFilters.set(filters);
    await this.loadSowRemovals();
  }

  protected async clearFilters(): Promise<void> {
    const hadFilters = this.hasFilters();
    this.filterForm.reset();
    this.appliedFilters.set({});
    if (hadFilters) {
      await this.loadSowRemovals();
    }
  }

  protected async loadSowRemovals(): Promise<void> {
    this.loading.set(true);
    this.error.set(false);
    try {
      this.removals.set(
        await firstValueFrom(this.sowRemovalsService.listSowRemovals(this.appliedFilters())),
      );
    } catch {
      this.error.set(true);
    } finally {
      this.loading.set(false);
    }
  }
}
