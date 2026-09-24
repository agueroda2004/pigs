import { Component, OnInit, computed, inject, signal } from '@angular/core';
import { FormBuilder, ReactiveFormsModule } from '@angular/forms';
import { firstValueFrom } from 'rxjs';

import { AuthService } from '../../../core/auth/auth.service';
import { BreedsService } from '../../../core/breeds/breeds.service';
import { NotificationService } from '../../../core/notifications/notification.service';
import { Sow, SowFilters, SowOrigin, SowState } from '../../../core/sows/sow.models';
import { SowsService } from '../../../core/sows/sows.service';
import { Dropdown, DropdownOption } from '../../../shared/ui/dropdown/dropdown';
import { CreateSowModal } from '../create-sow-modal/create-sow-modal';
import { EditSowModal } from '../edit-sow-modal/edit-sow-modal';
import { SowCard } from '../sow-card/sow-card';
import { SOW_STATE_LABELS } from '../sow-state';

@Component({
  selector: 'app-sows-page',
  imports: [ReactiveFormsModule, SowCard, CreateSowModal, EditSowModal, Dropdown],
  styleUrl: './sows-page.css',
  templateUrl: './sows-page.html',
})
export class SowsPage implements OnInit {
  private readonly sowsService = inject(SowsService);
  private readonly breedsService = inject(BreedsService);
  private readonly notifications = inject(NotificationService);
  private readonly auth = inject(AuthService);
  private readonly formBuilder = inject(FormBuilder);

  protected readonly isAdmin = this.auth.isAdmin;
  protected readonly sows = signal<Sow[]>([]);
  protected readonly breedNames = signal<Map<string, string>>(new Map());
  protected readonly breedOptions = signal<DropdownOption[]>([]);
  protected readonly appliedFilters = signal<SowFilters>({});
  protected readonly loading = signal(false);
  protected readonly error = signal(false);
  protected readonly modalOpen = signal(false);
  protected readonly editOpen = signal(false);
  protected readonly editingSow = signal<Sow | null>(null);

  protected readonly originOptions: DropdownOption[] = [
    { value: '', label: 'Todos' },
    { value: 'Propio', label: 'Propio' },
    { value: 'Externo', label: 'Externo' },
  ];

  protected readonly stateOptions: DropdownOption[] = [
    { value: '', label: 'Todos' },
    ...(Object.keys(SOW_STATE_LABELS) as SowState[]).map((state) => ({
      value: state,
      label: SOW_STATE_LABELS[state],
    })),
  ];

  protected readonly activeOptions: DropdownOption[] = [
    { value: '', label: 'Todos' },
    { value: 'true', label: 'Activas' },
    { value: 'false', label: 'Inactivas' },
  ];

  protected readonly hasFilters = computed(() => Object.keys(this.appliedFilters()).length > 0);

  protected readonly filterForm = this.formBuilder.nonNullable.group({
    code: [''],
    breed_id: [''],
    origin: [''],
    state: [''],
    active: [''],
  });

  async ngOnInit(): Promise<void> {
    await Promise.all([this.loadSows(), this.loadBreeds()]);
  }

  protected openModal(): void {
    this.modalOpen.set(true);
  }

  protected closeModal(): void {
    this.modalOpen.set(false);
  }

  protected async onCreated(code: string): Promise<void> {
    this.modalOpen.set(false);
    this.notifications.success(`Cerda "${code}" creada correctamente`);
    await this.loadSows();
  }

  protected openEdit(sow: Sow): void {
    this.editingSow.set(sow);
    this.editOpen.set(true);
  }

  protected closeEdit(): void {
    this.editOpen.set(false);
  }

  protected async onUpdated(sow: Sow): Promise<void> {
    this.editOpen.set(false);
    this.notifications.success(`Cerda "${sow.code}" actualizada correctamente`);
    await this.loadSows();
  }

  protected breedName(sow: Sow): string | null {
    return this.breedNames().get(sow.breed_id) ?? null;
  }

  protected async search(): Promise<void> {
    const { code, breed_id, origin, state, active } = this.filterForm.getRawValue();
    const filters: SowFilters = {};
    const trimmedCode = code.trim();
    if (trimmedCode) {
      filters.code = trimmedCode;
    }
    if (breed_id) {
      filters.breed_id = breed_id;
    }
    if (origin) {
      filters.origin = origin as SowOrigin;
    }
    if (state) {
      filters.state = state as SowState;
    }
    if (active) {
      filters.active = active === 'true';
    }
    this.appliedFilters.set(filters);
    await this.loadSows();
  }

  protected async clearFilters(): Promise<void> {
    const hadFilters = this.hasFilters();
    this.filterForm.reset();
    this.appliedFilters.set({});
    if (hadFilters) {
      await this.loadSows();
    }
  }

  protected async loadSows(): Promise<void> {
    this.loading.set(true);
    this.error.set(false);
    try {
      this.sows.set(await firstValueFrom(this.sowsService.listSows(this.appliedFilters())));
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
