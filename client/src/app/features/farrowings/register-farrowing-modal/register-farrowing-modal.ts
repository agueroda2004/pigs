import { HttpErrorResponse } from '@angular/common/http';
import { Component, effect, inject, input, output, signal } from '@angular/core';
import { FormBuilder, ReactiveFormsModule, Validators } from '@angular/forms';
import { firstValueFrom } from 'rxjs';

import {
  CreateFarrowingMedicationRequest,
  CreateFarrowingOperatorRequest,
  CreateFarrowingRequest,
} from '../../../core/farrowings/farrowing.models';
import { FarrowingsService } from '../../../core/farrowings/farrowings.service';
import { MedicationsService } from '../../../core/medications/medications.service';
import { NotificationService } from '../../../core/notifications/notification.service';
import { OperatorsService } from '../../../core/operators/operators.service';
import { SowsService } from '../../../core/sows/sows.service';
import { DatePicker } from '../../../shared/ui/date-picker/date-picker';
import { Dropdown, DropdownOption } from '../../../shared/ui/dropdown/dropdown';
import { Modal } from '../../../shared/ui/modal/modal';
import { SearchDropdown } from '../../../shared/ui/search-dropdown/search-dropdown';
import { TimePicker } from '../../../shared/ui/time-picker/time-picker';
import { toISODate } from '../../../shared/utils/date';

@Component({
  selector: 'app-register-farrowing-modal',
  imports: [
    ReactiveFormsModule,
    Modal,
    Dropdown,
    SearchDropdown,
    DatePicker,
    TimePicker,
  ],
  styleUrl: './register-farrowing-modal.css',
  templateUrl: './register-farrowing-modal.html',
})
export class RegisterFarrowingModal {
  readonly open = input(false);
  readonly closed = output<void>();
  readonly created = output<string>();

  private readonly farrowings = inject(FarrowingsService);
  private readonly sows = inject(SowsService);
  private readonly operators = inject(OperatorsService);
  private readonly medications = inject(MedicationsService);
  private readonly notifications = inject(NotificationService);
  private readonly formBuilder = inject(FormBuilder);

  protected readonly loading = signal(false);
  protected readonly sowOptions = signal<DropdownOption[]>([]);
  protected readonly operatorOptions = signal<DropdownOption[]>([]);
  protected readonly medicationOptions = signal<DropdownOption[]>([]);
  protected readonly today = toISODate(new Date());

  protected readonly operatorsArray = this.formBuilder.array<
    ReturnType<RegisterFarrowingModal['newOperatorGroup']>
  >([]);
  protected readonly medicationsArray = this.formBuilder.array<
    ReturnType<RegisterFarrowingModal['newMedicationGroup']>
  >([]);

  protected readonly form = this.formBuilder.nonNullable.group({
    sow_id: ['', [Validators.required]],
    farrow_date: [this.today, [Validators.required]],
    start_time: [''],
    end_time: [''],
    location: ['', [Validators.maxLength(100)]],
    live_born: [null as number | null, [Validators.min(0)]],
    stillborn: [null as number | null, [Validators.min(0)]],
    mummified: [null as number | null, [Validators.min(0)]],
    litter_weight: [null as number | null, [Validators.min(0)]],
    stillborn_weight: [null as number | null, [Validators.min(0)]],
    is_manipulated: [false],
    note: ['', [Validators.maxLength(500)]],
    operators: this.operatorsArray,
    medications: this.medicationsArray,
  });

  constructor() {
    effect(() => {
      if (this.open()) {
        void this.loadLookups();
      }
    });
  }

  protected operatorGroup(index: number) {
    return this.operatorsArray.at(index);
  }

  protected medicationGroup(index: number) {
    return this.medicationsArray.at(index);
  }

  protected addOperator(): void {
    this.operatorsArray.push(this.newOperatorGroup());
  }

  protected removeOperator(index: number): void {
    this.operatorsArray.removeAt(index);
  }

  protected addMedication(): void {
    this.medicationsArray.push(this.newMedicationGroup());
  }

  protected removeMedication(index: number): void {
    this.medicationsArray.removeAt(index);
  }

  protected async submit(): Promise<void> {
    if (this.form.invalid) {
      this.form.markAllAsTouched();
      return;
    }

    this.loading.set(true);

    const value = this.form.getRawValue();
    const request: CreateFarrowingRequest = {
      sow_id: value.sow_id,
      farrow_date: value.farrow_date,
      live_born: value.live_born ?? 0,
      stillborn: value.stillborn ?? 0,
      mummified: value.mummified ?? 0,
      is_manipulated: value.is_manipulated,
      operators: value.operators.map((operator): CreateFarrowingOperatorRequest => ({
        operator_id: operator.operator_id,
      })),
      medications: value.medications.map(
        (medication): CreateFarrowingMedicationRequest => ({
          medication_id: medication.medication_id,
          dose: medication.dose,
          applied_by: medication.applied_by,
        }),
      ),
    };
    if (value.start_time) {
      request.start_time = value.start_time;
    }
    if (value.end_time) {
      request.end_time = value.end_time;
    }
    if (value.location) {
      request.location = value.location;
    }
    if (value.litter_weight !== null) {
      request.litter_weight = value.litter_weight;
    }
    if (value.stillborn_weight !== null) {
      request.stillborn_weight = value.stillborn_weight;
    }
    if (value.note) {
      request.note = value.note;
    }

    const sowCode = this.sowOptions().find((option) => option.value === value.sow_id)?.label ?? '';

    try {
      await firstValueFrom(this.farrowings.createFarrowing(request));
      this.reset();
      this.created.emit(sowCode);
    } catch (error) {
      this.notifications.error(this.mapError(error));
    } finally {
      this.loading.set(false);
    }
  }

  private newOperatorGroup() {
    return this.formBuilder.nonNullable.group({
      operator_id: ['', [Validators.required]],
    });
  }

  private newMedicationGroup() {
    return this.formBuilder.nonNullable.group({
      medication_id: ['', [Validators.required]],
      dose: [1, [Validators.required, Validators.min(0.01)]],
      applied_by: ['', [Validators.required]],
    });
  }

  private reset(): void {
    this.form.reset({
      sow_id: '',
      farrow_date: this.today,
      start_time: '',
      end_time: '',
      location: '',
      live_born: null,
      stillborn: null,
      mummified: null,
      litter_weight: null,
      stillborn_weight: null,
      is_manipulated: false,
      note: '',
    });
    this.operatorsArray.clear();
    this.medicationsArray.clear();
  }

  private async loadLookups(): Promise<void> {
    try {
      const [sows, operators, medications] = await Promise.all([
        firstValueFrom(this.sows.listSowDropdown(true, ['Gestando'])),
        firstValueFrom(this.operators.listOperators()),
        firstValueFrom(this.medications.listMedicationOptions()),
      ]);
      this.sowOptions.set(sows.map((sow) => ({ value: sow.id, label: sow.code })));
      this.operatorOptions.set(
        operators
          .filter((operator) => operator.active)
          .map((operator) => ({ value: operator.id, label: operator.name })),
      );
      this.medicationOptions.set(
        medications.map((medication) => ({ value: medication.id, label: medication.name })),
      );
    } catch {
      this.notifications.error('No se pudieron cargar los datos del parto');
    }
  }

  private mapError(error: unknown): string {
    if (error instanceof HttpErrorResponse) {
      const serverMessage = (error.error as { error?: string } | null)?.error;
      if (serverMessage) {
        return serverMessage;
      }
      if (error.status === 409) {
        return 'La cerda o el servicio no están en un estado válido para el parto';
      }
    }
    return 'No se pudo registrar el parto';
  }
}
