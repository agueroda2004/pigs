import { ComponentFixture, TestBed } from '@angular/core/testing';
import { computed, signal } from '@angular/core';
import { provideRouter } from '@angular/router';

import { AuthUser } from '../../../core/auth/auth.models';
import { AuthService } from '../../../core/auth/auth.service';
import { Shell } from './shell';

function authStub(user: AuthUser) {
  const current = signal(user);

  return {
    user: current.asReadonly(),
    isAdmin: computed(() => current().role === 'Admin'),
    logout: () => Promise.resolve(),
  };
}

describe('Shell', () => {
  const render = async (user: AuthUser): Promise<ComponentFixture<Shell>> => {
    await TestBed.configureTestingModule({
      imports: [Shell],
      providers: [provideRouter([]), { provide: AuthService, useValue: authStub(user) }],
    }).compileComponents();

    const fixture = TestBed.createComponent(Shell);
    fixture.detectChanges();
    return fixture;
  };

  const navigationLinks = (fixture: ComponentFixture<Shell>): HTMLAnchorElement[] =>
    Array.from(fixture.nativeElement.querySelectorAll('aside nav a'));

  const labels = (fixture: ComponentFixture<Shell>): string[] =>
    navigationLinks(fixture).map((link) => link.textContent?.trim() ?? '');

  it('shows only the three regular user links in the requested order', async () => {
    const fixture = await render({ id: '1', username: 'ana', role: 'User' });

    expect(labels(fixture)).toEqual(['Inicio', 'Reportes', 'Historial de cerdas']);
    expect(navigationLinks(fixture).map((link) => link.getAttribute('href'))).toEqual([
      '/home',
      '/reports',
      '/sow-history',
    ]);
  });

  it('keeps the existing administrator navigation', async () => {
    const fixture = await render({ id: '2', username: 'admin', role: 'Admin' });

    expect(labels(fixture)).toEqual([
      'Panel',
      'Razas',
      'Verracos',
      'Remover verracos',
      'Cerdas',
      'Servicios',
      'Abortos',
      'Partos',
      'Muertes de lechones',
      'Traslados de lechones',
      'Destetes parciales',
      'Destetes',
      'Remover cerdas',
      'Medicamentos',
      'Operadores',
      'Usuarios',
    ]);
  });
});
