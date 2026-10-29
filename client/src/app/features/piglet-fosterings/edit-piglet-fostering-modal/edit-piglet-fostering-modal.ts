import { HttpErrorResponse } from '@angular/common/http';
import { Component, effect, inject, input, output, signal } from '@angular/core';
import { FormBuilder, ReactiveFormsModule, Validators } from '@angular/forms';
import { firstValueFrom } from 'rxjs';

import { NotificationService } from '../../../core/notifications/notification.service';
import {
  PigletFostering,
  UpdatePigletFosteringRequest,
} from '../../../core/piglet-fosterings/piglet-fostering.models';
import { PigletFosteringsService } from '../../../core/piglet-fosterings/piglet-fosterings.service';
import { Modal } from '../../../shared/ui/modal/modal';

@Component({
  selector: 'app-edit-piglet-fostering-modal',
  imports: [ReactiveFormsModule, Modal],
  styleUrl: './edit-piglet-fostering-modal.css',
  templateUrl: './edit-piglet-fostering-modal.html',
})
export class EditPigletFosteringModal {
  readonly open = input(false);
  readonly fostering = input<PigletFostering | null>(null);
  readonly closed = output<void>();
  readonly updated = output<void>();

  private readonly pigletFosterings = inject(PigletFosteringsService);
  private readonly notifications = inject(NotificationService);
  private readonly formBuilder = inject(FormBuilder);

  protected readonly loading = signal(false);

  protected readonly form = this.formBuilder.nonNullable.group({
    note: ['', [Validators.maxLength(500)]],
  });

  constructor() {
    effect(() => {
      const current = this.fostering();
      if (current && this.open()) {
        this.prefill(current);
      }
    });
  }

  protected async submit(): Promise<void> {
    if (this.form.invalid) {
      this.form.markAllAsTouched();
      return;
    }

    const current = this.fostering();
    if (!current) {
      return;
    }

    const request = this.buildRequest(current);
    if (request.note === undefined) {
      this.closed.emit();
      return;
    }

    this.loading.set(true);

    try {
      await firstValueFrom(this.pigletFosterings.updatePigletFostering(current.id, request));
      this.updated.emit();
    } catch (error) {
      this.notifications.error(this.mapError(error));
    } finally {
      this.loading.set(false);
    }
  }

  // buildRequest sends the note only when it changed, so an untouched form closes
  // without calling the API and an emptied note clears the stored value.
  private buildRequest(current: PigletFostering): UpdatePigletFosteringRequest {
    const trimmedNote = this.form.getRawValue().note.trim();
    if (trimmedNote === (current.note ?? '')) {
      return {};
    }
    return { note: trimmedNote };
  }

  private prefill(current: PigletFostering): void {
    this.form.reset({ note: current.note ?? '' });
  }

  private mapError(error: unknown): string {
    if (error instanceof HttpErrorResponse) {
      const serverMessage = (error.error as { error?: string } | null)?.error;
      if (serverMessage) {
        return serverMessage;
      }
      if (error.status === 404) {
        return 'El traslado no existe';
      }
    }
    return 'No se pudo actualizar el traslado';
  }
}
