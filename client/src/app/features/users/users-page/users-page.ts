import { Component, signal } from '@angular/core';

import { CreateUserModal } from '../create-user-modal/create-user-modal';

@Component({
  selector: 'app-users-page',
  imports: [CreateUserModal],
  styleUrl: './users-page.css',
  templateUrl: './users-page.html',
})
export class UsersPage {
  protected readonly modalOpen = signal(false);
  protected readonly createdUsername = signal<string | null>(null);

  protected openModal(): void {
    this.createdUsername.set(null);
    this.modalOpen.set(true);
  }

  protected closeModal(): void {
    this.modalOpen.set(false);
  }

  protected onCreated(username: string): void {
    this.modalOpen.set(false);
    this.createdUsername.set(username);
  }
}
