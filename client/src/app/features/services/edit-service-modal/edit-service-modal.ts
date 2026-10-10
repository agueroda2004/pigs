import { HttpErrorResponse } from '@angular/common/http';
import { Component, effect, inject, input, output, signal } from '@angular/core';
import { FormBuilder, ReactiveFormsModule, Validators } from '@angular/forms';
import { firstValueFrom } from 'rxjs';

import { BoarsService } from '../../../core/boars/boars.service';
import { NotificationService } from '../../../core/notifications/notification.service';
import { OperatorsService } from '../../../core/operators/operators.service';
import {
  CreateMountRequest,
  Mount,
  MountType,
  Service,
  UpdateMountRequest,
  UpdateServiceRequest,
} from '../../../core/services/service.models';
import { ServicesService } from '../../../core/services/services.service';
import { DatePicker } from '../../../shared/ui/date-picker/date-picker';
import { Dropdown, DropdownOption } from '../../../shared/ui/dropdown/dropdown';
import { Modal } from '../../../shared/ui/modal/modal';
import { SearchDropdown } from '../../../shared/ui/search-dropdown/search-dropdown';
import { toISODate } from '../../../shared/utils/date';

const GESTATION_DAYS = 114;
const MAX_MOUNTS = 3;

interface OriginalMount {
  id: string;
  boar_id: string;
  operator_id: string;
  mount_date: string;
  type: MountType;
  note: string;
}

@Component({
  selector: 'app-edit-service-modal',
  imports: [ReactiveFormsModule, Modal, Dropdown, SearchDropdown, DatePicker],
  styleUrl: './edit-service-modal.css',
  templateUrl: './edit-service-modal.html',
})
export class EditServiceModal {
  readonly open = input(false);
  readonly service = input<Service | null>(null);
  readonly closed = output<void>();
  readonly updated = output<void>();

  private readonly services = inject(ServicesService);
  private readonly boars = inject(BoarsService);
  private readonly operators = inject(OperatorsService);
  private readonly notifications = inject(NotificationService);
  private readonly formBuilder = inject(FormBuilder);

  protected readonly maxMounts = MAX_MOUNTS;
  protected readonly loading = signal(false);
  protected readonly boarOptions = signal<DropdownOption[]>([]);
  protected readonly operatorOptions = signal<DropdownOption[]>([]);
  protected readonly scheduleError = signal<string | null>(null);
  protected readonly previewFarrowing = signal<string | null>(null);
  protected readonly today = toISODate(new Date());
  protected readonly typeOptions: DropdownOption[] = [
    { value: 'Natural', label: 'Natural' },
    { value: 'Artificial', label: 'Artificial' },
  ];

  protected readonly mounts = this.formBuilder.array([this.newMountGroup()]);

  protected readonly form = this.formBuilder.nonNullable.group({
    location: ['', [Validators.maxLength(100)]],
    note: ['', [Validators.maxLength(500)]],
    mounts: this.mounts,
  });

  private originalMounts: OriginalMount[] = [];

  constructor() {
    this.mounts.valueChanges.subscribe(() => this.refreshSchedule());

    effect(() => {
      const current = this.service();
      if (current && this.open()) {
        this.prefill(current);
        void this.loadOptions();
      }
    });
  }

  protected mountGroup(index: number) {
    return this.mounts.at(index);
  }

  protected maxDateFor(index: number): string {
    if (index === 0) {
      return this.today;
    }
    const previous = this.mounts.at(index - 1).controls.mount_date.value;
    if (!previous) {
      return this.today;
    }
    const limit = this.addDays(previous, 1);
    return limit < this.today ? limit : this.today;
  }

  protected addMount(): void {
    if (this.mounts.length < MAX_MOUNTS) {
      this.mounts.push(this.newMountGroup());
    }
  }

  protected removeMount(index: number): void {
    if (this.mounts.length > 1) {
      this.mounts.removeAt(index);
      this.refreshSchedule();
    }
  }

  protected async submit(): Promise<void> {
    this.refreshSchedule();
    if (this.form.invalid || this.scheduleError()) {
      this.form.markAllAsTouched();
      return;
    }

    const current = this.service();
    if (!current) {
      return;
    }

    const { location, note } = this.form.getRawValue();
    const create: CreateMountRequest[] = [];
    const update: UpdateMountRequest[] = [];
    const keptIDs = new Set<string>();

    for (const mount of this.mounts.getRawValue()) {
      if (mount.id) {
        keptIDs.add(mount.id);
        if (this.mountChanged(mount.id, mount)) {
          update.push({
            id: mount.id,
            boar_id: mount.boar_id,
            operator_id: mount.operator_id,
            mount_date: mount.mount_date,
            type: mount.type as MountType,
            note: mount.note,
          });
        }
        continue;
      }
      create.push({
        boar_id: mount.boar_id,
        operator_id: mount.operator_id,
        mount_date: mount.mount_date,
        type: mount.type as MountType,
        note: mount.note,
      });
    }

    const deleted = this.originalMounts.map((mount) => mount.id).filter((id) => !keptIDs.has(id));

    const changed =
      location !== (current.location ?? '') ||
      note !== (current.note ?? '') ||
      create.length > 0 ||
      update.length > 0 ||
      deleted.length > 0;

    if (!changed) {
      this.updated.emit();
      return;
    }

    const request: UpdateServiceRequest = {
      location,
      note,
      mounts: { create, update, delete: deleted },
    };

    this.loading.set(true);

    try {
      await firstValueFrom(this.services.updateService(current.id, request));
      this.updated.emit();
    } catch (error) {
      this.notifications.error(this.mapError(error));
    } finally {
      this.loading.set(false);
    }
  }

  private mountChanged(
    id: string,
    mount: { boar_id: string; operator_id: string; mount_date: string; type: string; note: string },
  ): boolean {
    const original = this.originalMounts.find((item) => item.id === id);
    if (!original) {
      return true;
    }
    return (
      mount.boar_id !== original.boar_id ||
      mount.operator_id !== original.operator_id ||
      mount.mount_date !== original.mount_date ||
      mount.type !== original.type ||
      mount.note !== original.note
    );
  }

  private prefill(current: Service): void {
    this.form.controls.location.reset(current.location ?? '');
    this.form.controls.note.reset(current.note ?? '');
    this.originalMounts = current.mounts.map((mount) => ({
      id: mount.id,
      boar_id: mount.boar_id,
      operator_id: mount.operator_id,
      mount_date: mount.mount_date,
      type: mount.type,
      note: mount.note ?? '',
    }));

    while (this.mounts.length > 0) {
      this.mounts.removeAt(0);
    }
    for (const mount of current.mounts) {
      this.mounts.push(this.newMountGroup(mount));
    }
    if (this.mounts.length === 0) {
      this.mounts.push(this.newMountGroup());
    }
    this.refreshSchedule();
  }

  private newMountGroup(mount?: Mount) {
    return this.formBuilder.nonNullable.group({
      id: [mount?.id ?? ''],
      mount_date: [mount?.mount_date ?? '', [Validators.required]],
      boar_id: [mount?.boar_id ?? '', [Validators.required]],
      operator_id: [mount?.operator_id ?? '', [Validators.required]],
      type: [mount?.type ?? 'Artificial', [Validators.required]],
      note: [mount?.note ?? '', [Validators.maxLength(500)]],
    });
  }

  private refreshSchedule(): void {
    let error: string | null = null;
    let previous: string | null = null;
    let last: string | null = null;

    for (const control of this.mounts.controls) {
      const date = control.controls.mount_date.value;
      if (!date) {
        previous = null;
        continue;
      }
      if (date > this.today) {
        error = 'La fecha de monta no puede ser futura';
        break;
      }
      if (previous) {
        const difference = (Date.parse(date) - Date.parse(previous)) / 86_400_000;
        if (difference < 0) {
          error = 'Las fechas de monta deben estar en orden ascendente';
          break;
        }
        if (difference > 1) {
          error = 'Las montas no pueden tener más de 24 horas de diferencia';
          break;
        }
      }
      previous = date;
      last = date;
    }

    this.scheduleError.set(error);
    this.previewFarrowing.set(last ? this.addDays(last, GESTATION_DAYS) : null);
  }

  private addDays(date: string, days: number): string {
    const value = new Date(`${date}T00:00:00`);
    value.setDate(value.getDate() + days);
    return toISODate(value);
  }

  private async loadOptions(): Promise<void> {
    try {
      const [boars, operators] = await Promise.all([
        firstValueFrom(this.boars.listBoarDropdown()),
        firstValueFrom(this.operators.listOperatorDropdown()),
      ]);
      this.boarOptions.set(
        boars.map((boar) => ({
          value: boar.id,
          label: boar.active ? boar.code : `${boar.code} (inactivo)`,
        })),
      );
      this.operatorOptions.set(
        operators.map((operator) => ({
          value: operator.id,
          label: operator.active ? operator.name : `${operator.name} (inactivo)`,
        })),
      );
    } catch {
      this.notifications.error('No se pudieron cargar los datos del servicio');
    }
  }

  private mapError(error: unknown): string {
    if (error instanceof HttpErrorResponse) {
      const serverMessage = (error.error as { error?: string } | null)?.error;
      if (serverMessage) {
        return serverMessage;
      }
      if (error.status === 404) {
        return 'El servicio no existe';
      }
      if (error.status === 409) {
        return 'Solo se pueden editar servicios en estado Confirmado';
      }
    }
    return 'No se pudo actualizar el servicio';
  }
}
