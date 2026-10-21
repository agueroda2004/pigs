import { HttpErrorResponse } from '@angular/common/http';
import { Component, effect, inject, input, output, signal } from '@angular/core';
import { FormBuilder, ReactiveFormsModule, Validators } from '@angular/forms';
import { firstValueFrom } from 'rxjs';

import {
  CreateFarrowingMedicationRequest,
  CreateFarrowingOperatorRequest,
  Farrowing,
  FarrowingMedicationLink,
  FarrowingOperatorLink,
  UpdateFarrowingRequest,
} from '../../../core/farrowings/farrowing.models';
import { FarrowingsService } from '../../../core/farrowings/farrowings.service';
import { MedicationsService } from '../../../core/medications/medications.service';
import { NotificationService } from '../../../core/notifications/notification.service';
import { OperatorsService } from '../../../core/operators/operators.service';
import { DatePicker } from '../../../shared/ui/date-picker/date-picker';
import { Dropdown, DropdownOption } from '../../../shared/ui/dropdown/dropdown';
import { Modal } from '../../../shared/ui/modal/modal';
import { SearchDropdown } from '../../../shared/ui/search-dropdown/search-dropdown';
import { TimePicker } from '../../../shared/ui/time-picker/time-picker';
import { toISODate } from '../../../shared/utils/date';

@Component({
  selector: 'app-edit-farrowing-modal',
  imports: [ReactiveFormsModule, Modal, Dropdown, SearchDropdown, DatePicker, TimePicker],
  styleUrl: './edit-farrowing-modal.css',
  templateUrl: './edit-farrowing-modal.html',
})
export class EditFarrowingModal {
  readonly open = input(false);
  readonly farrowing = input<Farrowing | null>(null);
  readonly closed = output<void>();
  readonly updated = output<void>();

  private readonly farrowings = inject(FarrowingsService);
  private readonly operators = inject(OperatorsService);
  private readonly medications = inject(MedicationsService);
  private readonly notifications = inject(NotificationService);
  private readonly formBuilder = inject(FormBuilder);

  protected readonly loading = signal(false);
  protected readonly operatorOptions = signal<DropdownOption[]>([]);
  protected readonly medicationOptions = signal<DropdownOption[]>([]);
  protected readonly today = toISODate(new Date());

  protected readonly operatorsArray = this.formBuilder.array<
    ReturnType<EditFarrowingModal['newOperatorGroup']>
  >([]);
  protected readonly medicationsArray = this.formBuilder.array<
    ReturnType<EditFarrowingModal['newMedicationGroup']>
  >([]);

  protected readonly form = this.formBuilder.nonNullable.group({
    farrow_date: ['', [Validators.required]],
    start_time: [''],
    end_time: [''],
    location: ['', [Validators.maxLength(100)]],
    live_born: [null as number | null, [Validators.min(0)]],
    stillborn: [null as number | null, [Validators.min(0)]],
    mummified: [null as number | null, [Validators.min(0)]],
    litter_weight: [
      null as number | null,
      [Validators.required, Validators.min(0), Validators.max(30)],
    ],
    stillborn_weight: [
      null as number | null,
      [Validators.required, Validators.min(0), Validators.max(30)],
    ],
    is_manipulated: [false],
    note: ['', [Validators.maxLength(500)]],
    operators: this.operatorsArray,
    medications: this.medicationsArray,
  });

  constructor() {
    this.lockImmutableFields();
    effect(() => {
      const current = this.farrowing();
      if (current && this.open()) {
        this.prefill(current);
        this.lockImmutableFields();
        void this.loadOptions();
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

    const current = this.farrowing();
    if (!current) {
      return;
    }

    const value = this.form.getRawValue();
    if (value.litter_weight === null || value.stillborn_weight === null) {
      this.form.markAllAsTouched();
      return;
    }

    const request: UpdateFarrowingRequest = {
      start_time: value.start_time,
      end_time: value.end_time,
      location: value.location,
      litter_weight: value.litter_weight,
      stillborn_weight: value.stillborn_weight,
      is_manipulated: value.is_manipulated,
      note: value.note,
      operators: value.operators.map((operator): CreateFarrowingOperatorRequest => ({
        operator_id: operator.operator_id,
      })),
      medications: value.medications.map((medication): CreateFarrowingMedicationRequest => ({
        medication_id: medication.medication_id,
        dose: medication.dose,
        applied_by: medication.applied_by,
      })),
    };

    this.loading.set(true);
    try {
      await firstValueFrom(this.farrowings.updateFarrowing(current.id, request));
      this.updated.emit();
    } catch (error) {
      this.notifications.error(this.mapError(error));
    } finally {
      this.loading.set(false);
    }
  }

  private newOperatorGroup(operator?: FarrowingOperatorLink) {
    return this.formBuilder.nonNullable.group({
      operator_id: [operator?.operator_id ?? '', [Validators.required]],
    });
  }

  private newMedicationGroup(medication?: FarrowingMedicationLink) {
    return this.formBuilder.nonNullable.group({
      medication_id: [medication?.medication_id ?? '', [Validators.required]],
      dose: [
        medication?.dose ?? 1,
        [Validators.required, Validators.min(0.01), Validators.max(30)],
      ],
      applied_by: [medication?.applied_by ?? '', [Validators.required]],
    });
  }

  // Disables the farrow date and the piglet counts, which are fixed after creation.
  // The disabled values stay visible in the form but are never sent on update.
  private lockImmutableFields(): void {
    for (const name of ['farrow_date', 'live_born', 'stillborn', 'mummified'] as const) {
      this.form.controls[name].disable({ emitEvent: false });
    }
  }

  private prefill(current: Farrowing): void {
    this.form.reset({
      farrow_date: current.farrow_date,
      start_time: current.start_time ?? '',
      end_time: current.end_time ?? '',
      location: current.location ?? '',
      live_born: current.live_born,
      stillborn: current.stillborn,
      mummified: current.mummified,
      litter_weight: current.litter_weight,
      stillborn_weight: current.stillborn_weight,
      is_manipulated: current.is_manipulated,
      note: current.note ?? '',
    });

    this.operatorsArray.clear();
    for (const operator of current.operators) {
      this.operatorsArray.push(this.newOperatorGroup(operator));
    }
    this.medicationsArray.clear();
    for (const medication of current.medications) {
      this.medicationsArray.push(this.newMedicationGroup(medication));
    }
  }

  private async loadOptions(): Promise<void> {
    try {
      const [operators, medications] = await Promise.all([
        firstValueFrom(this.operators.listOperatorDropdown()),
        firstValueFrom(this.medications.listMedications()),
      ]);
      this.operatorOptions.set(
        operators.map((operator) => ({
          value: operator.id,
          label: operator.active ? operator.name : `${operator.name} (inactivo)`,
        })),
      );
      this.medicationOptions.set(
        medications.map((medication) => ({
          value: medication.id,
          label: medication.active ? medication.name : `${medication.name} (inactivo)`,
        })),
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
      if (error.status === 404) {
        return 'El parto no existe';
      }
      if (error.status === 409) {
        return 'No se pudo actualizar el parto';
      }
    }
    return 'No se pudo actualizar el parto';
  }
}
