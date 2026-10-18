import { HttpErrorResponse } from '@angular/common/http';
import { Component, inject, input, output, signal } from '@angular/core';
import { firstValueFrom } from 'rxjs';

import { Medication } from '../../../core/medications/medication.models';
import { MedicationsService } from '../../../core/medications/medications.service';
import { NotificationService } from '../../../core/notifications/notification.service';
import { Modal } from '../../../shared/ui/modal/modal';

@Component({
  selector: 'app-delete-medication-modal',
  imports: [Modal],
  styleUrl: './delete-medication-modal.css',
  templateUrl: './delete-medication-modal.html',
})
export class DeleteMedicationModal {
  readonly open = input(false);
  readonly medication = input<Medication | null>(null);
  readonly closed = output<void>();
  readonly deleted = output<Medication>();

  private readonly medications = inject(MedicationsService);
  private readonly notifications = inject(NotificationService);

  protected readonly loading = signal(false);

  protected async confirm(): Promise<void> {
    const medication = this.medication();
    if (!medication) {
      return;
    }

    this.loading.set(true);

    try {
      await firstValueFrom(this.medications.deleteMedication(medication.id));
      this.deleted.emit(medication);
    } catch (error) {
      this.notifications.error(this.mapError(error));
    } finally {
      this.loading.set(false);
    }
  }

  private mapError(error: unknown): string {
    if (error instanceof HttpErrorResponse) {
      const serverMessage = (error.error as { error?: string } | null)?.error;
      if (serverMessage) {
        return serverMessage;
      }
      if (error.status === 404) {
        return 'El medicamento no existe';
      }
    }
    return 'No se pudo eliminar el medicamento';
  }
}
