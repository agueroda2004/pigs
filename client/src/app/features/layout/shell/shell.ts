import { Component, computed, inject, signal } from '@angular/core';
import { Router, RouterLink, RouterLinkActive, RouterOutlet } from '@angular/router';

import { landingUrl } from '../../../core/auth/auth.navigation';
import { AuthService } from '../../../core/auth/auth.service';
import { ThemeToggle } from '../../../shared/ui/theme-toggle/theme-toggle';

interface SidebarItem {
  label: string;
  route: string;
}

interface SidebarGroup {
  label: string;
  items: SidebarItem[];
}

@Component({
  selector: 'app-shell',
  imports: [RouterOutlet, RouterLink, RouterLinkActive, ThemeToggle],
  styleUrl: './shell.css',
  templateUrl: './shell.html',
})
export class Shell {
  private readonly auth = inject(AuthService);
  private readonly router = inject(Router);

  protected readonly user = this.auth.user;
  protected readonly isAdmin = this.auth.isAdmin;
  protected readonly homeLink = computed(() => landingUrl(this.auth.isAdmin()));
  protected readonly sidebarOpen = signal(false);

  protected readonly adminGroups = computed<SidebarGroup[]>(() => [
    {
      label: 'General',
      items: [
        { label: 'Panel', route: this.homeLink() },
        { label: 'Reportes', route: '/reports' },
        { label: 'Historial de cerdas', route: '/sow-history' },
      ],
    },
    {
      label: 'Ganado',
      items: [
        { label: 'Verracos', route: '/boars' },
        { label: 'Cerdas', route: '/sows' },
        { label: 'Remover verracos', route: '/boar-removals' },
        { label: 'Remover cerdas', route: '/sow-removals' },
      ],
    },
    {
      label: 'Reproducción',
      items: [
        { label: 'Servicios', route: '/services' },
        { label: 'Abortos', route: '/abortions' },
        { label: 'Partos', route: '/farrowings' },
      ],
    },
    {
      label: 'Lechones',
      items: [
        { label: 'Muertes de lechones', route: '/piglet-deaths' },
        { label: 'Traslados de lechones', route: '/piglet-fosterings' },
        { label: 'Destetes parciales', route: '/partial-weagings' },
        { label: 'Destetes', route: '/weagings' },
      ],
    },
    {
      label: 'Administración',
      items: [
        { label: 'Razas', route: '/breeds' },
        { label: 'Medicamentos', route: '/medications' },
        { label: 'Operadores', route: '/operators' },
        { label: 'Usuarios', route: '/users' },
      ],
    },
  ]);

  protected readonly expandedGroups = signal<Set<string>>(
    new Set(this.adminGroups().map((group) => group.label)),
  );

  protected toggleGroup(label: string): void {
    this.expandedGroups.update((groups) => {
      const next = new Set(groups);
      if (next.has(label)) {
        next.delete(label);
      } else {
        next.add(label);
      }
      return next;
    });
  }

  protected isGroupExpanded(label: string): boolean {
    return this.expandedGroups().has(label);
  }

  protected toggleSidebar(): void {
    this.sidebarOpen.update((isOpen) => !isOpen);
  }

  protected closeSidebar(): void {
    this.sidebarOpen.set(false);
  }

  protected async logout(): Promise<void> {
    await this.auth.logout();
    await this.router.navigate(['/login']);
  }
}
