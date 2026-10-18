import { Component, OnInit, computed, inject, signal } from '@angular/core';
import { FormBuilder, ReactiveFormsModule, Validators } from '@angular/forms';
import { firstValueFrom } from 'rxjs';

import { AuthService } from '../../../core/auth/auth.service';
import { BoarRemoval, BoarRemovalFilters } from '../../../core/boar-removals/boar-removal.models';
import { BoarRemovalsService } from '../../../core/boar-removals/boar-removals.service';
import { NotificationService } from '../../../core/notifications/notification.service';
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
  ],
  styleUrl: './boar-removals-page.css',
  templateUrl: './boar-removals-page.html',
})
export class BoarRemovalsPage implements OnInit {
  private readonly boarRemovalsService = inject(BoarRemovalsService);
  private readonly notifications = inject(NotificationService);
  private readonly auth = inject(AuthService);
  private readonly formBuilder = inject(FormBuilder);

  protected readonly isAdmin = this.auth.isAdmin;
  protected readonly removals = signal<BoarRemoval[]>([]);
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
    code: ['', [Validators.maxLength(50)]],
  });

  async ngOnInit(): Promise<void> {
    await this.loadBoarRemovals();
  }

  protected openModal(): void {
    this.modalOpen.set(true);
  }

  protected closeModal(): void {
    this.modalOpen.set(false);
  }

  protected async onCreated(boarCode: string): Promise<void> {
    this.notifications.success(`Verraco "${boarCode}" removido correctamente`);
    await this.loadBoarRemovals();
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
    await this.loadBoarRemovals();
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
    await this.loadBoarRemovals();
    this.closeDelete();
  }

  protected deletingBoarCode(): string | null {
    const removal = this.deletingRemoval();
    return removal ? this.boarCode(removal) : null;
  }

  protected boarCode(removal: BoarRemoval): string | null {
    return removal.boar_code || null;
  }

  protected async search(): Promise<void> {
    if (this.filterForm.invalid) {
      this.filterForm.markAllAsTouched();
      return;
    }

    const { code } = this.filterForm.getRawValue();
    const filters: BoarRemovalFilters = {};
    const trimmedCode = code.trim();
    if (trimmedCode) {
      filters.boar_code = trimmedCode;
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
}
