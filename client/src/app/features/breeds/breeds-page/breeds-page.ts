import { Component, OnInit, computed, inject, signal } from '@angular/core';
import { FormBuilder, ReactiveFormsModule } from '@angular/forms';
import { firstValueFrom } from 'rxjs';

import { AuthService } from '../../../core/auth/auth.service';
import { Breed, BreedFilters } from '../../../core/breeds/breed.models';
import { BreedsService } from '../../../core/breeds/breeds.service';
import { NotificationService } from '../../../core/notifications/notification.service';
import { Dropdown, DropdownOption } from '../../../shared/ui/dropdown/dropdown';
import { BreedCard } from '../breed-card/breed-card';
import { CreateBreedModal } from '../create-breed-modal/create-breed-modal';
import { DeleteBreedModal } from '../delete-breed-modal/delete-breed-modal';
import { EditBreedModal } from '../edit-breed-modal/edit-breed-modal';

@Component({
  selector: 'app-breeds-page',
  imports: [
    ReactiveFormsModule,
    BreedCard,
    CreateBreedModal,
    EditBreedModal,
    DeleteBreedModal,
    Dropdown,
  ],
  styleUrl: './breeds-page.css',
  templateUrl: './breeds-page.html',
})
export class BreedsPage implements OnInit {
  private readonly breedsService = inject(BreedsService);
  private readonly notifications = inject(NotificationService);
  private readonly auth = inject(AuthService);
  private readonly formBuilder = inject(FormBuilder);

  protected readonly isAdmin = this.auth.isAdmin;
  protected readonly modalOpen = signal(false);
  protected readonly editOpen = signal(false);
  protected readonly editingBreed = signal<Breed | null>(null);
  protected readonly deleteOpen = signal(false);
  protected readonly deletingBreed = signal<Breed | null>(null);
  protected readonly breeds = signal<Breed[]>([]);
  protected readonly appliedFilters = signal<BreedFilters>({});
  protected readonly loading = signal(false);
  protected readonly error = signal(false);

  protected readonly activeOptions: DropdownOption[] = [
    { value: '', label: 'Todas' },
    { value: 'true', label: 'Activas' },
    { value: 'false', label: 'Inactivas' },
  ];

  protected readonly hasFilters = computed(() => Object.keys(this.appliedFilters()).length > 0);

  protected readonly filterForm = this.formBuilder.nonNullable.group({
    name: [''],
    active: [''],
  });

  async ngOnInit(): Promise<void> {
    await this.loadBreeds();
  }

  protected openModal(): void {
    this.modalOpen.set(true);
  }

  protected closeModal(): void {
    this.modalOpen.set(false);
  }

  protected async onCreated(name: string): Promise<void> {
    this.modalOpen.set(false);
    this.notifications.success(`Raza "${name}" creada correctamente`);
    await this.loadBreeds();
  }

  protected openEdit(breed: Breed): void {
    this.editingBreed.set(breed);
    this.editOpen.set(true);
  }

  protected closeEdit(): void {
    this.editOpen.set(false);
  }

  protected async onUpdated(): Promise<void> {
    this.editOpen.set(false);
    this.notifications.success('Raza actualizada correctamente');
    await this.loadBreeds();
  }

  protected openDelete(breed: Breed): void {
    this.deletingBreed.set(breed);
    this.deleteOpen.set(true);
  }

  protected closeDelete(): void {
    this.deleteOpen.set(false);
  }

  protected async onDeleted(breed: Breed): Promise<void> {
    this.deleteOpen.set(false);
    this.notifications.success(`Raza "${breed.name}" eliminada correctamente`);
    await this.loadBreeds();
  }

  protected async search(): Promise<void> {
    const { name, active } = this.filterForm.getRawValue();
    const filters: BreedFilters = {};
    const trimmedName = name.trim();
    if (trimmedName) {
      filters.name = trimmedName;
    }
    if (active) {
      filters.active = active === 'true';
    }
    this.appliedFilters.set(filters);
    await this.loadBreeds();
  }

  protected async clearFilters(): Promise<void> {
    const hadFilters = this.hasFilters();
    this.filterForm.reset();
    this.appliedFilters.set({});
    if (hadFilters) {
      await this.loadBreeds();
    }
  }

  protected async loadBreeds(): Promise<void> {
    this.loading.set(true);
    this.error.set(false);
    try {
      this.breeds.set(await firstValueFrom(this.breedsService.listBreeds(this.appliedFilters())));
    } catch {
      this.error.set(true);
    } finally {
      this.loading.set(false);
    }
  }
}
