import { Component, OnInit, computed, inject, signal } from '@angular/core';
import { FormBuilder, ReactiveFormsModule } from '@angular/forms';
import { firstValueFrom } from 'rxjs';

import { AuthService } from '../../../core/auth/auth.service';
import { NotificationService } from '../../../core/notifications/notification.service';
import { SowRemoval, SowRemovalFilters } from '../../../core/sow-removals/sow-removal.models';
import { SowRemovalsService } from '../../../core/sow-removals/sow-removals.service';
import { SowsService } from '../../../core/sows/sows.service';
import { DropdownOption } from '../../../shared/ui/dropdown/dropdown';
import { SearchDropdown } from '../../../shared/ui/search-dropdown/search-dropdown';
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
    SearchDropdown,
  ],
  styleUrl: './sow-removals-page.css',
  templateUrl: './sow-removals-page.html',
})
export class SowRemovalsPage implements OnInit {
  private readonly sowRemovalsService = inject(SowRemovalsService);
  private readonly sowsService = inject(SowsService);
  private readonly notifications = inject(NotificationService);
  private readonly auth = inject(AuthService);
  private readonly formBuilder = inject(FormBuilder);

  protected readonly isAdmin = this.auth.isAdmin;
  protected readonly removals = signal<SowRemoval[]>([]);
  protected readonly sowCodes = signal<Map<string, string>>(new Map());
  protected readonly sowOptions = signal<DropdownOption[]>([]);
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
    sow_id: [''],
  });

  async ngOnInit(): Promise<void> {
    await Promise.all([this.loadSowRemovals(), this.loadLookups()]);
  }

  protected openModal(): void {
    this.modalOpen.set(true);
  }

  protected closeModal(): void {
    this.modalOpen.set(false);
  }

  protected async onCreated(sowCode: string): Promise<void> {
    this.notifications.success(`Cerda "${sowCode}" removida correctamente`);
    await Promise.all([this.loadSowRemovals(), this.loadLookups()]);
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
    await Promise.all([this.loadSowRemovals(), this.loadLookups()]);
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
    await Promise.all([this.loadSowRemovals(), this.loadLookups()]);
    this.closeDelete();
  }

  protected deletingSowCode(): string | null {
    const removal = this.deletingRemoval();
    return removal ? this.sowCode(removal) : null;
  }

  protected sowCode(removal: SowRemoval): string | null {
    return this.sowCodes().get(removal.sow_id) ?? null;
  }

  protected async search(): Promise<void> {
    const { sow_id } = this.filterForm.getRawValue();
    const filters: SowRemovalFilters = {};
    if (sow_id) {
      filters.sow_id = sow_id;
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

  private async loadLookups(): Promise<void> {
    try {
      const sows = await firstValueFrom(this.sowsService.listSowDropdown());
      this.sowCodes.set(new Map(sows.map((sow) => [sow.id, sow.code])));
      this.sowOptions.set(sows.map((sow) => ({ value: sow.id, label: sow.code })));
    } catch {
      this.sowCodes.set(new Map());
      this.sowOptions.set([]);
    }
  }
}
