import { Component, OnInit, computed, inject, signal } from '@angular/core';
import { FormBuilder, ReactiveFormsModule } from '@angular/forms';
import { firstValueFrom } from 'rxjs';

import { AuthService } from '../../../core/auth/auth.service';
import { BoarRemoval, BoarRemovalFilters } from '../../../core/boar-removals/boar-removal.models';
import { BoarRemovalsService } from '../../../core/boar-removals/boar-removals.service';
import { BoarsService } from '../../../core/boars/boars.service';
import { NotificationService } from '../../../core/notifications/notification.service';
import { DropdownOption } from '../../../shared/ui/dropdown/dropdown';
import { SearchDropdown } from '../../../shared/ui/search-dropdown/search-dropdown';
import { BoarRemovalCard } from '../boar-removal-card/boar-removal-card';
import { DeleteBoarRemovalModal } from '../delete-boar-removal-modal/delete-boar-removal-modal';
import { EditBoarRemovalModal } from '../edit-boar-removal-modal/edit-boar-removal-modal';
import { RegisterBoarRemovalModal } from '../register-boar-removal-modal/register-boar-removal-modal';

@Component({
  selector: 'app-boar-removals-page',
  imports: [
    ReactiveFormsModule,
    BoarRemovalCard,
    RegisterBoarRemovalModal,
    EditBoarRemovalModal,
    DeleteBoarRemovalModal,
    SearchDropdown,
  ],
  styleUrl: './boar-removals-page.css',
  templateUrl: './boar-removals-page.html',
})
export class BoarRemovalsPage implements OnInit {
  private readonly boarRemovalsService = inject(BoarRemovalsService);
  private readonly boarsService = inject(BoarsService);
  private readonly notifications = inject(NotificationService);
  private readonly auth = inject(AuthService);
  private readonly formBuilder = inject(FormBuilder);

  protected readonly isAdmin = this.auth.isAdmin;
  protected readonly removals = signal<BoarRemoval[]>([]);
  protected readonly boarCodes = signal<Map<string, string>>(new Map());
  protected readonly boarOptions = signal<DropdownOption[]>([]);
  protected readonly appliedFilters = signal<BoarRemovalFilters>({});
  protected readonly loading = signal(false);
  protected readonly error = signal(false);
  protected readonly modalOpen = signal(false);
  protected readonly editModalOpen = signal(false);
  protected readonly editingRemoval = signal<BoarRemoval | null>(null);
  protected readonly deleteModalOpen = signal(false);
  protected readonly deletingRemoval = signal<BoarRemoval | null>(null);

  protected readonly hasFilters = computed(() => Object.keys(this.appliedFilters()).length > 0);

  protected readonly filterForm = this.formBuilder.nonNullable.group({
    boar_id: [''],
  });

  async ngOnInit(): Promise<void> {
    await Promise.all([this.loadBoarRemovals(), this.loadLookups()]);
  }

  protected openModal(): void {
    this.modalOpen.set(true);
  }

  protected closeModal(): void {
    this.modalOpen.set(false);
  }

  protected async onCreated(boarCode: string): Promise<void> {
    this.notifications.success(`Verraco "${boarCode}" removido correctamente`);
    await Promise.all([this.loadBoarRemovals(), this.loadLookups()]);
  }

  protected openEdit(removal: BoarRemoval): void {
    this.editingRemoval.set(removal);
    this.editModalOpen.set(true);
  }

  protected closeEdit(): void {
    this.editModalOpen.set(false);
  }

  protected async onUpdated(removal: BoarRemoval): Promise<void> {
    this.notifications.success(`Baja de "${this.boarCode(removal) ?? 'el verraco'}" actualizada`);
    await Promise.all([this.loadBoarRemovals(), this.loadLookups()]);
    this.closeEdit();
  }

  protected editingBoarCode(): string | null {
    const removal = this.editingRemoval();
    return removal ? this.boarCode(removal) : null;
  }

  protected openDelete(removal: BoarRemoval): void {
    this.deletingRemoval.set(removal);
    this.deleteModalOpen.set(true);
  }

  protected closeDelete(): void {
    this.deleteModalOpen.set(false);
  }

  protected async onDeleted(): Promise<void> {
    const removal = this.deletingRemoval();
    this.notifications.success(
      `Baja de "${removal ? (this.boarCode(removal) ?? 'el verraco') : 'el verraco'}" eliminada`,
    );
    await Promise.all([this.loadBoarRemovals(), this.loadLookups()]);
    this.closeDelete();
  }

  protected deletingBoarCode(): string | null {
    const removal = this.deletingRemoval();
    return removal ? this.boarCode(removal) : null;
  }

  protected boarCode(removal: BoarRemoval): string | null {
    return this.boarCodes().get(removal.boar_id) ?? null;
  }

  protected async search(): Promise<void> {
    const { boar_id } = this.filterForm.getRawValue();
    const filters: BoarRemovalFilters = {};
    if (boar_id) {
      filters.boar_id = boar_id;
    }
    this.appliedFilters.set(filters);
    await this.loadBoarRemovals();
  }

  protected async clearFilters(): Promise<void> {
    const hadFilters = this.hasFilters();
    this.filterForm.reset();
    this.appliedFilters.set({});
    if (hadFilters) {
      await this.loadBoarRemovals();
    }
  }

  protected async loadBoarRemovals(): Promise<void> {
    this.loading.set(true);
    this.error.set(false);
    try {
      this.removals.set(
        await firstValueFrom(this.boarRemovalsService.listBoarRemovals(this.appliedFilters())),
      );
    } catch {
      this.error.set(true);
    } finally {
      this.loading.set(false);
    }
  }

  private async loadLookups(): Promise<void> {
    try {
      const boars = await firstValueFrom(this.boarsService.listBoarDropdown(false));
      this.boarCodes.set(new Map(boars.map((boar) => [boar.id, boar.code])));
      this.boarOptions.set(boars.map((boar) => ({ value: boar.id, label: boar.code })));
    } catch {
      this.boarCodes.set(new Map());
      this.boarOptions.set([]);
    }
  }
}
