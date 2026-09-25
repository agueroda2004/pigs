import { DatePipe } from '@angular/common';
import { Component, computed, input, output } from '@angular/core';

import { User } from '../../../core/users/user.models';

@Component({
  selector: 'app-user-card',
  imports: [DatePipe],
  styleUrl: './user-card.css',
  templateUrl: './user-card.html',
})
export class UserCard {
  readonly user = input.required<User>();
  readonly editRequested = output<User>();

  protected readonly roleLabel = computed(() =>
    this.user().role === 'Admin' ? 'Administrador' : 'Usuario',
  );

  protected readonly roleClass = computed(() =>
    this.user().role === 'Admin' ? 'bg-primary/10 text-primary' : 'bg-muted text-muted-foreground',
  );

  protected readonly statusLabel = computed(() => (this.user().active ? 'Activo' : 'Inactivo'));

  protected readonly statusClass = computed(() =>
    this.user().active ? 'bg-success/10 text-success' : 'bg-muted text-muted-foreground',
  );
}
