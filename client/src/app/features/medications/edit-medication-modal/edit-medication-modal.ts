import { HttpErrorResponse } from '@angular/common/http';
import { Component, effect, inject, input, output, signal } from '@angular/core';
import { FormBuilder, ReactiveFormsModule, Validators } from '@angular/forms';
import { firstValueFrom } from 'rxjs';

import { Medication, UpdateMedicationRequest } from '../../../core/medications/medication.models';
import { MedicationsService } from '../../../core/medications/medications.service';
import { NotificationService } from '../../../core/notifications/notification.service';
import { Modal } from '../../../shared/ui/modal/modal';

@Component({
  selector: 'app-edit-medication-modal',
  imports: [ReactiveFormsModule, Modal],
  styleUrl: './edit-medication-modal.css',
  templateUrl: './edit-medication-modal.html',
})
export class EditMedicationModal {
  readonly open = input(false);
  readonly medication = input<Medication | null>(null);
  readonly closed = output<void>();
  readonly updated = output<Medication>();

  private readonly medications = inject(MedicationsService);
  private readonly notifications = inject(NotificationService);
  private readonly formBuilder = inject(FormBuilder);

  protected readonly loading = signal(false);

  protected readonly form = this.formBuilder.nonNullable.group({
    name: ['', [Validators.required, Validators.maxLength(50)]],
    active: [true],
  });

  constructor() {
    effect(() => {
      const current = this.medication();
      if (current && this.open()) {
        this.form.reset({ name: current.name, active: current.active });
      }
    });
  }

  protected async submit(): Promise<void> {
    if (this.form.invalid) {
      this.form.markAllAsTouched();
      return;
    }

    const current = this.medication();
    if (!current) {
      return;
    }

    const { name, active } = this.form.getRawValue();
    const request: UpdateMedicationRequest = {};
    if (name !== current.name) {
      request.name = name;
    }
    if (active !== current.active) {
      request.active = active;
    }

    if (request.name === undefined && request.active === undefined) {
      this.updated.emit(current);
      return;
    }

    this.loading.set(true);

    try {
      const updated = await firstValueFrom(this.medications.updateMedication(current.id, request));
      this.updated.emit(updated);
    } catch (error) {
      this.notifications.error(this.mapError(error));
    } finally {
      this.loading.set(false);
    }
  }

  private mapError(error: unknown): string {
    if (error instanceof HttpErrorResponse) {
      if (error.status === 409) {
        return 'El nombre del medicamento ya existe';
      }
      if (error.status === 404) {
        return 'El medicamento no existe';
      }
      if (error.status === 400) {
        return error.error?.error ?? 'Los datos ingresados no son válidos';
      }
    }
    return 'No se pudo actualizar el medicamento';
  }
}
