import { HttpErrorResponse } from '@angular/common/http';
import { Component, effect, inject, input, output, signal } from '@angular/core';
import { FormBuilder, ReactiveFormsModule, Validators } from '@angular/forms';
import { firstValueFrom } from 'rxjs';

import { Operator, UpdateOperatorRequest } from '../../../core/operators/operator.models';
import { OperatorsService } from '../../../core/operators/operators.service';
import { NotificationService } from '../../../core/notifications/notification.service';
import { Modal } from '../../../shared/ui/modal/modal';

@Component({
  selector: 'app-edit-operator-modal',
  imports: [ReactiveFormsModule, Modal],
  styleUrl: './edit-operator-modal.css',
  templateUrl: './edit-operator-modal.html',
})
export class EditOperatorModal {
  readonly open = input(false);
  readonly operator = input<Operator | null>(null);
  readonly closed = output<void>();
  readonly updated = output<Operator>();

  private readonly operators = inject(OperatorsService);
  private readonly notifications = inject(NotificationService);
  private readonly formBuilder = inject(FormBuilder);

  protected readonly loading = signal(false);

  protected readonly form = this.formBuilder.nonNullable.group({
    name: ['', [Validators.required, Validators.maxLength(100)]],
    active: [true],
  });

  constructor() {
    effect(() => {
      const current = this.operator();
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

    const current = this.operator();
    if (!current) {
      return;
    }

    const { name, active } = this.form.getRawValue();
    const request: UpdateOperatorRequest = {};
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
      const updated = await firstValueFrom(this.operators.updateOperator(current.id, request));
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
        return 'El nombre del operador ya existe';
      }
      if (error.status === 404) {
        return 'El operador no existe';
      }
      if (error.status === 400) {
        return error.error?.error ?? 'Los datos ingresados no son válidos';
      }
    }
    return 'No se pudo actualizar el operador';
  }
}
