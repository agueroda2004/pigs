import { Component } from '@angular/core';
import { RouterLink } from '@angular/router';

@Component({
  selector: 'app-not-found',
  imports: [RouterLink],
  template: `
    <div
      class="flex min-h-screen flex-col items-center justify-center gap-3 bg-background px-4 text-center"
    >
      <p class="text-5xl font-semibold tracking-tight">404</p>
      <p class="text-sm text-muted-foreground">La página que buscas no existe.</p>
      <a
        routerLink="/"
        class="mt-2 rounded-md bg-primary px-4 py-2 text-sm font-medium text-primary-foreground transition hover:opacity-90"
      >
        Volver al inicio
      </a>
    </div>
  `,
})
export class NotFound {}
