import { HttpErrorResponse } from '@angular/common/http';
import { Component, effect, inject, input, output, signal } from '@angular/core';
import { FormBuilder, ReactiveFormsModule, Validators } from '@angular/forms';
import { firstValueFrom } from 'rxjs';

import { BoarsService } from '../../../core/boars/boars.service';
import { NotificationService } from '../../../core/notifications/notification.service';
import { OperatorsService } from '../../../core/operators/operators.service';
import {
  CreateMountRequest,
  CreateServiceRequest,
  MountType,
} from '../../../core/services/service.models';
import { ServicesService } from '../../../core/services/services.service';
import { SowsService } from '../../../core/sows/sows.service';
import { DatePicker } from '../../../shared/ui/date-picker/date-picker';
import { Dropdown, DropdownOption } from '../../../shared/ui/dropdown/dropdown';
import { Modal } from '../../../shared/ui/modal/modal';
import { SearchDropdown } from '../../../shared/ui/search-dropdown/search-dropdown';
import { toISODate } from '../../../shared/utils/date';

const GESTATION_DAYS = 114;
const MAX_MOUNTS = 3;

@Component({
  selector: 'app-create-service-modal',
  imports: [ReactiveFormsModule, Modal, Dropdown, SearchDropdown, DatePicker],
  styleUrl: './create-service-modal.css',
  templateUrl: './create-service-modal.html',
})
export class CreateServiceModal {
  readonly open = input(false);
  readonly closed = output<void>();
  readonly created = output<string>();

  private readonly services = inject(ServicesService);
  private readonly sows = inject(SowsService);
  private readonly boars = inject(BoarsService);
  private readonly operators = inject(OperatorsService);
  private readonly notifications = inject(NotificationService);
  private readonly formBuilder = inject(FormBuilder);

  protected readonly maxMounts = MAX_MOUNTS;
  protected readonly loading = signal(false);
  protected readonly sowOptions = signal<DropdownOption[]>([]);
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
    sow_id: ['', [Validators.required]],
    location: ['', [Validators.maxLength(100)]],
    note: ['', [Validators.maxLength(500)]],
    mounts: this.mounts,
  });

  constructor() {
    this.mounts.valueChanges.subscribe(() => this.refreshSchedule());

    effect(() => {
      if (this.open()) {
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

    this.loading.set(true);

    const { sow_id, location, note } = this.form.getRawValue();
    const request: CreateServiceRequest = {
      sow_id,
      mounts: this.mounts.getRawValue().map((mount) => {
        const item: CreateMountRequest = {
          boar_id: mount.boar_id,
          operator_id: mount.operator_id,
          mount_date: mount.mount_date,
          type: mount.type as MountType,
        };
        if (mount.note) {
          item.note = mount.note;
        }
        return item;
      }),
    };
    if (location) {
      request.location = location;
    }
    if (note) {
      request.note = note;
    }

    const sowCode = this.sowOptions().find((option) => option.value === sow_id)?.label ?? '';

    try {
      await firstValueFrom(this.services.createService(request));
      this.reset();
      this.created.emit(sowCode);
    } catch (error) {
      this.notifications.error(this.mapError(error));
    } finally {
      this.loading.set(false);
    }
  }

  private newMountGroup() {
    return this.formBuilder.nonNullable.group({
      mount_date: ['', [Validators.required]],
      boar_id: ['', [Validators.required]],
      operator_id: ['', [Validators.required]],
      type: ['Artificial', [Validators.required]],
      note: ['', [Validators.maxLength(500)]],
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

  private reset(): void {
    this.form.controls.sow_id.reset('');
    this.form.controls.location.reset('');
    this.form.controls.note.reset('');
    while (this.mounts.length > 1) {
      this.mounts.removeAt(this.mounts.length - 1);
    }
    this.mounts.at(0).reset({
      mount_date: '',
      boar_id: '',
      operator_id: '',
      type: 'Artificial',
      note: '',
    });
    this.refreshSchedule();
  }

  private async loadOptions(): Promise<void> {
    try {
      const [sows, boars, operators] = await Promise.all([
        firstValueFrom(this.sows.listSowOptions(true)),
        firstValueFrom(this.boars.listBoars()),
        firstValueFrom(this.operators.listOperators()),
      ]);
      this.sowOptions.set(sows.map((sow) => ({ value: sow.id, label: sow.code })));
      this.boarOptions.set(
        boars
          .filter((boar) => boar.active && boar.state === 'Vivo')
          .map((boar) => ({ value: boar.id, label: boar.code })),
      );
      this.operatorOptions.set(
        operators
          .filter((operator) => operator.active)
          .map((operator) => ({ value: operator.id, label: operator.name })),
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
      if (error.status === 409) {
        return 'La cerda o el verraco no están disponibles';
      }
    }
    return 'No se pudo crear el servicio';
  }
}
