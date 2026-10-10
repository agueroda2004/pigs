import { Component, OnInit, computed, inject, signal } from '@angular/core';
import { FormBuilder, ReactiveFormsModule } from '@angular/forms';
import { firstValueFrom } from 'rxjs';

import { AuthService } from '../../../core/auth/auth.service';
import { NotificationService } from '../../../core/notifications/notification.service';
import { Service, ServiceFilters, ServiceState } from '../../../core/services/service.models';
import { ServicesService } from '../../../core/services/services.service';
import { Dropdown, DropdownOption } from '../../../shared/ui/dropdown/dropdown';
import { CreateServiceModal } from '../create-service-modal/create-service-modal';
import { DeleteServiceModal } from '../delete-service-modal/delete-service-modal';
import { EditServiceModal } from '../edit-service-modal/edit-service-modal';
import { ServiceCard } from '../service-card/service-card';
import { SERVICE_STATE_LABELS } from '../service-state';

@Component({
  selector: 'app-services-page',
  imports: [
    ReactiveFormsModule,
    ServiceCard,
    CreateServiceModal,
    EditServiceModal,
    DeleteServiceModal,
    Dropdown,
  ],
  styleUrl: './services-page.css',
  templateUrl: './services-page.html',
})
export class ServicesPage implements OnInit {
  private readonly servicesService = inject(ServicesService);
  private readonly notifications = inject(NotificationService);
  private readonly auth = inject(AuthService);
  private readonly formBuilder = inject(FormBuilder);

  protected readonly isAdmin = this.auth.isAdmin;
  protected readonly services = signal<Service[]>([]);
  protected readonly appliedFilters = signal<ServiceFilters>({});
  protected readonly loading = signal(false);
  protected readonly error = signal(false);
  protected readonly modalOpen = signal(false);
  protected readonly editOpen = signal(false);
  protected readonly editingService = signal<Service | null>(null);
  protected readonly deleteOpen = signal(false);
  protected readonly deletingService = signal<Service | null>(null);
  protected readonly page = signal(1);
  protected readonly totalPages = signal(0);
  protected readonly total = signal(0);

  protected readonly stateOptions: DropdownOption[] = [
    { value: '', label: 'Todos' },
    ...(Object.keys(SERVICE_STATE_LABELS) as ServiceState[]).map((state) => ({
      value: state,
      label: SERVICE_STATE_LABELS[state],
    })),
  ];

  protected readonly hasFilters = computed(() => Object.keys(this.appliedFilters()).length > 0);

  protected readonly filterForm = this.formBuilder.nonNullable.group({
    sow_code: [''],
    state: [''],
  });

  async ngOnInit(): Promise<void> {
    await this.loadServices();
  }

  protected openModal(): void {
    this.modalOpen.set(true);
  }

  protected closeModal(): void {
    this.modalOpen.set(false);
  }

  protected async onCreated(sowCode: string): Promise<void> {
    this.notifications.success(`Servicio para la cerda "${sowCode}" creado correctamente`);
    this.page.set(1);
    await this.loadServices();
  }

  protected openEdit(service: Service): void {
    this.editingService.set(service);
    this.editOpen.set(true);
  }

  protected closeEdit(): void {
    this.editOpen.set(false);
  }

  protected async onUpdated(): Promise<void> {
    this.editOpen.set(false);
    this.notifications.success('Servicio actualizado correctamente');
    await this.loadServices();
  }

  protected openDelete(service: Service): void {
    this.deletingService.set(service);
    this.deleteOpen.set(true);
  }

  protected closeDelete(): void {
    this.deleteOpen.set(false);
  }

  protected async onDeleted(service: Service): Promise<void> {
    this.deleteOpen.set(false);
    const sowCode = service.sow_code || 'sin identificar';
    this.notifications.success(`Servicio de la cerda "${sowCode}" eliminado correctamente`);
    await this.loadServices();
  }

  protected async previousPage(): Promise<void> {
    if (this.page() > 1) {
      this.page.update((page) => page - 1);
      await this.loadServices();
    }
  }

  protected async nextPage(): Promise<void> {
    if (this.page() < this.totalPages()) {
      this.page.update((page) => page + 1);
      await this.loadServices();
    }
  }

  protected async search(): Promise<void> {
    const { sow_code, state } = this.filterForm.getRawValue();
    const filters: ServiceFilters = {};
    if (sow_code) {
      filters.sow_code = sow_code;
    }
    if (state) {
      filters.state = state as ServiceState;
    }
    this.appliedFilters.set(filters);
    this.page.set(1);
    await this.loadServices();
  }

  protected async clearFilters(): Promise<void> {
    const hadFilters = this.hasFilters();
    this.filterForm.reset();
    this.appliedFilters.set({});
    if (hadFilters) {
      this.page.set(1);
      await this.loadServices();
    }
  }

  protected async loadServices(): Promise<void> {
    this.loading.set(true);
    this.error.set(false);
    try {
      const result = await firstValueFrom(
        this.servicesService.listServices(this.appliedFilters(), this.page()),
      );
      this.services.set(result.items);
      this.total.set(result.total);
      this.totalPages.set(result.total_pages);
    } catch {
      this.error.set(true);
    } finally {
      this.loading.set(false);
    }
  }
}
