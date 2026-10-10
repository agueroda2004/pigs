import { HttpErrorResponse } from '@angular/common/http';
import { Component, inject, input, output, signal } from '@angular/core';
import { firstValueFrom } from 'rxjs';

import { NotificationService } from '../../../core/notifications/notification.service';
import { Service } from '../../../core/services/service.models';
import { ServicesService } from '../../../core/services/services.service';
import { Modal } from '../../../shared/ui/modal/modal';

@Component({
  selector: 'app-delete-service-modal',
  imports: [Modal],
  styleUrl: './delete-service-modal.css',
  templateUrl: './delete-service-modal.html',
})
export class DeleteServiceModal {
  readonly open = input(false);
  readonly service = input<Service | null>(null);
  readonly closed = output<void>();
  readonly deleted = output<Service>();

  private readonly services = inject(ServicesService);
  private readonly notifications = inject(NotificationService);

  protected readonly loading = signal(false);

  protected async confirm(): Promise<void> {
    const service = this.service();
    if (!service) {
      return;
    }

    this.loading.set(true);

    try {
      await firstValueFrom(this.services.deleteService(service.id));
      this.deleted.emit(service);
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
        return 'El servicio no existe';
      }
    }
    return 'No se pudo eliminar el servicio';
  }
}
