import { Component, OnInit, computed, inject, signal } from '@angular/core';
import { FormBuilder, ReactiveFormsModule } from '@angular/forms';
import { firstValueFrom } from 'rxjs';

import { AuthService } from '../../../core/auth/auth.service';
import { BoarsService } from '../../../core/boars/boars.service';
import { NotificationService } from '../../../core/notifications/notification.service';
import { OperatorsService } from '../../../core/operators/operators.service';
import { Service, ServiceFilters, ServiceState } from '../../../core/services/service.models';
import { ServicesService } from '../../../core/services/services.service';
import { SowsService } from '../../../core/sows/sows.service';
import { Dropdown, DropdownOption } from '../../../shared/ui/dropdown/dropdown';
import { SearchDropdown } from '../../../shared/ui/search-dropdown/search-dropdown';
import { CreateServiceModal } from '../create-service-modal/create-service-modal';
import { ServiceCard } from '../service-card/service-card';
import { SERVICE_STATE_LABELS } from '../service-state';

@Component({
  selector: 'app-services-page',
  imports: [ReactiveFormsModule, ServiceCard, CreateServiceModal, Dropdown, SearchDropdown],
  styleUrl: './services-page.css',
  templateUrl: './services-page.html',
})
export class ServicesPage implements OnInit {
  private readonly servicesService = inject(ServicesService);
  private readonly sowsService = inject(SowsService);
  private readonly boarsService = inject(BoarsService);
  private readonly operatorsService = inject(OperatorsService);
  private readonly notifications = inject(NotificationService);
  private readonly auth = inject(AuthService);
  private readonly formBuilder = inject(FormBuilder);

  protected readonly isAdmin = this.auth.isAdmin;
  protected readonly services = signal<Service[]>([]);
  protected readonly sowCodes = signal<Map<string, string>>(new Map());
  protected readonly boarCodes = signal<Map<string, string>>(new Map());
  protected readonly operatorNames = signal<Map<string, string>>(new Map());
  protected readonly sowOptions = signal<DropdownOption[]>([]);
  protected readonly appliedFilters = signal<ServiceFilters>({});
  protected readonly loading = signal(false);
  protected readonly error = signal(false);
  protected readonly modalOpen = signal(false);

  protected readonly stateOptions: DropdownOption[] = [
    { value: '', label: 'Todos' },
    ...(Object.keys(SERVICE_STATE_LABELS) as ServiceState[]).map((state) => ({
      value: state,
      label: SERVICE_STATE_LABELS[state],
    })),
  ];

  protected readonly hasFilters = computed(() => Object.keys(this.appliedFilters()).length > 0);

  protected readonly filterForm = this.formBuilder.nonNullable.group({
    sow_id: [''],
    state: [''],
  });

  async ngOnInit(): Promise<void> {
    await Promise.all([this.loadServices(), this.loadLookups()]);
  }

  protected openModal(): void {
    this.modalOpen.set(true);
  }

  protected closeModal(): void {
    this.modalOpen.set(false);
  }

  protected async onCreated(sowCode: string): Promise<void> {
    this.notifications.success(`Servicio para la cerda "${sowCode}" creado correctamente`);
    await Promise.all([this.loadServices(), this.loadLookups()]);
  }

  protected sowCode(service: Service): string | null {
    return this.sowCodes().get(service.sow_id) ?? null;
  }

  protected async search(): Promise<void> {
    const { sow_id, state } = this.filterForm.getRawValue();
    const filters: ServiceFilters = {};
    if (sow_id) {
      filters.sow_id = sow_id;
    }
    if (state) {
      filters.state = state as ServiceState;
    }
    this.appliedFilters.set(filters);
    await this.loadServices();
  }

  protected async clearFilters(): Promise<void> {
    const hadFilters = this.hasFilters();
    this.filterForm.reset();
    this.appliedFilters.set({});
    if (hadFilters) {
      await this.loadServices();
    }
  }

  protected async loadServices(): Promise<void> {
    this.loading.set(true);
    this.error.set(false);
    try {
      this.services.set(
        await firstValueFrom(this.servicesService.listServices(this.appliedFilters())),
      );
    } catch {
      this.error.set(true);
    } finally {
      this.loading.set(false);
    }
  }

  private async loadLookups(): Promise<void> {
    try {
      const [sows, sowOptions, boars, operators] = await Promise.all([
        firstValueFrom(this.sowsService.listSows()),
        firstValueFrom(this.sowsService.listSowOptions(true)),
        firstValueFrom(this.boarsService.listBoars()),
        firstValueFrom(this.operatorsService.listOperators()),
      ]);
      this.sowCodes.set(new Map(sows.map((sow) => [sow.id, sow.code])));
      this.boarCodes.set(new Map(boars.map((boar) => [boar.id, boar.code])));
      this.operatorNames.set(new Map(operators.map((operator) => [operator.id, operator.name])));
      this.sowOptions.set(sowOptions.map((option) => ({ value: option.id, label: option.code })));
    } catch {
      this.sowCodes.set(new Map());
      this.boarCodes.set(new Map());
      this.operatorNames.set(new Map());
      this.sowOptions.set([]);
    }
  }
}
