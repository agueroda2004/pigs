import { Component } from '@angular/core';
import { TestBed } from '@angular/core/testing';
import { vi } from 'vitest';

import { Modal } from './modal';

@Component({
  imports: [Modal],
  template: `<app-modal [open]="true" title="Crear servicio">
    <p class="projected">Contenido proyectado</p>
  </app-modal>`,
})
class HostComponent {}

describe('Modal', () => {
  beforeEach(async () => {
    await TestBed.configureTestingModule({ imports: [Modal, HostComponent] }).compileComponents();
  });

  const create = () => TestBed.createComponent(Modal);

  it('does not render when closed', () => {
    const fixture = create();
    fixture.detectChanges();

    expect(fixture.nativeElement.querySelector('[role="dialog"]')).toBeNull();
  });

  it('renders the title and projected content when open', () => {
    const fixture = TestBed.createComponent(HostComponent);
    fixture.detectChanges();

    const dialog = fixture.nativeElement.querySelector('[role="dialog"]');
    expect(dialog).not.toBeNull();
    expect(fixture.nativeElement.textContent).toContain('Crear servicio');
    expect(fixture.nativeElement.textContent).toContain('Contenido proyectado');
  });

  it('applies the requested size class', () => {
    const fixture = create();
    fixture.componentRef.setInput('open', true);
    fixture.componentRef.setInput('size', 'lg');
    fixture.detectChanges();

    const dialog = fixture.nativeElement.querySelector('[role="dialog"]') as HTMLElement;
    expect(dialog.classList).toContain('max-w-lg');
    expect(dialog.className).toContain('max-h-[90vh]');
  });

  it('scrolls the body while keeping the header fixed', () => {
    const fixture = create();
    fixture.componentRef.setInput('open', true);
    fixture.detectChanges();

    const dialog = fixture.nativeElement.querySelector('[role="dialog"]') as HTMLElement;
    const body = dialog.lastElementChild as HTMLElement;
    expect(body.className).toContain('overflow-y-auto');
  });

  it('emits closed on escape', () => {
    const fixture = create();
    fixture.componentRef.setInput('open', true);
    fixture.detectChanges();
    const closed = vi.fn();
    fixture.componentInstance.closed.subscribe(closed);

    document.dispatchEvent(new KeyboardEvent('keydown', { key: 'Escape' }));

    expect(closed).toHaveBeenCalled();
  });
});
