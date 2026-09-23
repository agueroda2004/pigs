import { Component, OnInit, inject, signal } from '@angular/core';
import { firstValueFrom } from 'rxjs';

import { AuthService } from '../../../core/auth/auth.service';
import { Breed } from '../../../core/breeds/breed.models';
import { BreedsService } from '../../../core/breeds/breeds.service';
import { NotificationService } from '../../../core/notifications/notification.service';
import { BreedCard } from '../breed-card/breed-card';
import { CreateBreedModal } from '../create-breed-modal/create-breed-modal';
import { EditBreedModal } from '../edit-breed-modal/edit-breed-modal';

@Component({
  selector: 'app-breeds-page',
  imports: [BreedCard, CreateBreedModal, EditBreedModal],
  styleUrl: './breeds-page.css',
  templateUrl: './breeds-page.html',
})
export class BreedsPage implements OnInit {
  private readonly breedsService = inject(BreedsService);
  private readonly notifications = inject(NotificationService);
  private readonly auth = inject(AuthService);

  protected readonly isAdmin = this.auth.isAdmin;
  protected readonly modalOpen = signal(false);
  protected readonly editOpen = signal(false);
  protected readonly editingBreed = signal<Breed | null>(null);
  protected readonly breeds = signal<Breed[]>([]);
  protected readonly loading = signal(false);
  protected readonly error = signal(false);

  async ngOnInit(): Promise<void> {
    await this.loadBreeds();
  }

  protected openModal(): void {
    this.modalOpen.set(true);
  }

  protected closeModal(): void {
    this.modalOpen.set(false);
  }

  protected async onCreated(name: string): Promise<void> {
    this.modalOpen.set(false);
    this.notifications.success(`Raza "${name}" creada correctamente`);
    await this.loadBreeds();
  }

  protected openEdit(breed: Breed): void {
    this.editingBreed.set(breed);
    this.editOpen.set(true);
  }

  protected closeEdit(): void {
    this.editOpen.set(false);
  }

  protected async onUpdated(breed: Breed): Promise<void> {
    this.editOpen.set(false);
    this.notifications.success(`Raza "${breed.name}" actualizada correctamente`);
    await this.loadBreeds();
  }

  protected async loadBreeds(): Promise<void> {
    this.loading.set(true);
    this.error.set(false);
    try {
      this.breeds.set(await firstValueFrom(this.breedsService.listBreeds()));
    } catch {
      this.error.set(true);
    } finally {
      this.loading.set(false);
    }
  }
}
