import { HttpErrorResponse } from '@angular/common/http';
import { Component, inject, input, output, signal } from '@angular/core';
import { firstValueFrom } from 'rxjs';

import { PigletDeath } from '../../../core/piglet-deaths/piglet-death.models';
import { PigletDeathsService } from '../../../core/piglet-deaths/piglet-deaths.service';
import { NotificationService } from '../../../core/notifications/notification.service';
import { Modal } from '../../../shared/ui/modal/modal';

@Component({
  selector: 'app-delete-piglet-death-modal',
  imports: [Modal],
  styleUrl: './delete-piglet-death-modal.css',
  templateUrl: './delete-piglet-death-modal.html',
})
export class DeletePigletDeathModal {
  readonly open = input(false);
  readonly death = input<PigletDeath | null>(null);
  readonly sowCode = input<string | null>(null);
  readonly closed = output<void>();
  readonly deleted = output<PigletDeath>();

  private readonly pigletDeaths = inject(PigletDeathsService);
  private readonly notifications = inject(NotificationService);

  protected readonly loading = signal(false);

  protected async confirm(): Promise<void> {
    const death = this.death();
    if (!death) {
      return;
    }

    this.loading.set(true);

    try {
      await firstValueFrom(this.pigletDeaths.deletePigletDeath(death.id));
      this.deleted.emit(death);
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
        return 'La muerte no existe';
      }
    }
    return 'No se pudo eliminar la muerte';
  }
}
