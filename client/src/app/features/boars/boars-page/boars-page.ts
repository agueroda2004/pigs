import { Component, OnInit, computed, inject, signal } from '@angular/core';
import { FormBuilder, ReactiveFormsModule } from '@angular/forms';
import { firstValueFrom } from 'rxjs';

import { AuthService } from '../../../core/auth/auth.service';
import { Boar, BoarFilters, BoarOrigin } from '../../../core/boars/boar.models';
import { BoarsService } from '../../../core/boars/boars.service';
import { BreedsService } from '../../../core/breeds/breeds.service';
import { NotificationService } from '../../../core/notifications/notification.service';
import { Dropdown, DropdownOption } from '../../../shared/ui/dropdown/dropdown';
import { BoarCard } from '../boar-card/boar-card';
import { CreateBoarModal } from '../create-boar-modal/create-boar-modal';
import { EditBoarModal } from '../edit-boar-modal/edit-boar-modal';

@Component({
  selector: 'app-boars-page',
  imports: [ReactiveFormsModule, BoarCard, CreateBoarModal, EditBoarModal, Dropdown],
  styleUrl: './boars-page.css',
  templateUrl: './boars-page.html',
})
export class BoarsPage implements OnInit {
  private readonly boarsService = inject(BoarsService);
  private readonly breedsService = inject(BreedsService);
  private readonly notifications = inject(NotificationService);
  private readonly auth = inject(AuthService);
  private readonly formBuilder = inject(FormBuilder);

  protected readonly isAdmin = this.auth.isAdmin;
  protected readonly boars = signal<Boar[]>([]);
  protected readonly breedNames = signal<Map<string, string>>(new Map());
  protected readonly breedOptions = signal<DropdownOption[]>([]);
  protected readonly appliedFilters = signal<BoarFilters>({});
  protected readonly loading = signal(false);
  protected readonly error = signal(false);
  protected readonly modalOpen = signal(false);
  protected readonly editOpen = signal(false);
  protected readonly editingBoar = signal<Boar | null>(null);

  protected readonly originOptions: DropdownOption[] = [
    { value: '', label: 'Todos' },
    { value: 'Propio', label: 'Propio' },
    { value: 'Externo', label: 'Externo' },
  ];

  protected readonly activeOptions: DropdownOption[] = [
    { value: '', label: 'Todos' },
    { value: 'true', label: 'Activos' },
    { value: 'false', label: 'Inactivos' },
  ];

  protected readonly hasFilters = computed(() => Object.keys(this.appliedFilters()).length > 0);

  protected readonly filterForm = this.formBuilder.nonNullable.group({
    code: [''],
    breed_id: [''],
    origin: [''],
    active: [''],
  });

  async ngOnInit(): Promise<void> {
    await Promise.all([this.loadBoars(), this.loadBreeds()]);
  }

  protected openModal(): void {
    this.modalOpen.set(true);
  }

  protected closeModal(): void {
    this.modalOpen.set(false);
  }

  protected async onCreated(code: string): Promise<void> {
    this.modalOpen.set(false);
    this.notifications.success(`Verraco "${code}" creado correctamente`);
    await this.loadBoars();
  }

  protected openEdit(boar: Boar): void {
    this.editingBoar.set(boar);
    this.editOpen.set(true);
  }

  protected closeEdit(): void {
    this.editOpen.set(false);
  }

  protected async onUpdated(boar: Boar): Promise<void> {
    this.editOpen.set(false);
    this.notifications.success(`Verraco "${boar.code}" actualizado correctamente`);
    await this.loadBoars();
  }

  protected breedName(boar: Boar): string | null {
    return this.breedNames().get(boar.breed_id) ?? null;
  }

  protected async search(): Promise<void> {
    const { code, breed_id, origin, active } = this.filterForm.getRawValue();
    const filters: BoarFilters = {};
    const trimmedCode = code.trim();
    if (trimmedCode) {
      filters.code = trimmedCode;
    }
    if (breed_id) {
      filters.breed_id = breed_id;
    }
    if (origin) {
      filters.origin = origin as BoarOrigin;
    }
    if (active) {
      filters.active = active === 'true';
    }
    this.appliedFilters.set(filters);
    await this.loadBoars();
  }

  protected async clearFilters(): Promise<void> {
    const hadFilters = this.hasFilters();
    this.filterForm.reset();
    this.appliedFilters.set({});
    if (hadFilters) {
      await this.loadBoars();
    }
  }

  protected async loadBoars(): Promise<void> {
    this.loading.set(true);
    this.error.set(false);
    try {
      this.boars.set(await firstValueFrom(this.boarsService.listBoars(this.appliedFilters())));
    } catch {
      this.error.set(true);
    } finally {
      this.loading.set(false);
    }
  }

  private async loadBreeds(): Promise<void> {
    try {
      const breeds = await firstValueFrom(this.breedsService.listBreeds());
      this.breedNames.set(new Map(breeds.map((breed) => [breed.id, breed.name])));
      this.breedOptions.set(breeds.map((breed) => ({ value: breed.id, label: breed.name })));
    } catch {
      this.breedNames.set(new Map());
      this.breedOptions.set([]);
    }
  }
}
