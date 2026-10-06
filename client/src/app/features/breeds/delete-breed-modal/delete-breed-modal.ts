import { HttpErrorResponse } from '@angular/common/http';
import { Component, inject, input, output, signal } from '@angular/core';
import { firstValueFrom } from 'rxjs';

import { Breed } from '../../../core/breeds/breed.models';
import { BreedsService } from '../../../core/breeds/breeds.service';
import { NotificationService } from '../../../core/notifications/notification.service';
import { Modal } from '../../../shared/ui/modal/modal';

@Component({
  selector: 'app-delete-breed-modal',
  imports: [Modal],
  styleUrl: './delete-breed-modal.css',
  templateUrl: './delete-breed-modal.html',
})
export class DeleteBreedModal {
  readonly open = input(false);
  readonly breed = input<Breed | null>(null);
  readonly closed = output<void>();
  readonly deleted = output<Breed>();

  private readonly breeds = inject(BreedsService);
  private readonly notifications = inject(NotificationService);

  protected readonly loading = signal(false);

  protected async confirm(): Promise<void> {
    const breed = this.breed();
    if (!breed) {
      return;
    }

    this.loading.set(true);

    try {
      await firstValueFrom(this.breeds.deleteBreed(breed.id));
      this.deleted.emit(breed);
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
        return 'La raza no existe';
      }
    }
    return 'No se pudo eliminar la raza';
  }
}
