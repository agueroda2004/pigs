import { TestBed } from '@angular/core/testing';

import { Service } from '../../../core/services/service.models';
import { ServiceCard } from './service-card';

function buildMount(number: number): Service['mounts'][number] {
  return {
    id: `m${number}`,
    service_id: '1',
    boar_id: `boar-${number}`,
    boar_code: `V-00${number}`,
    operator_id: `operator-${number}`,
    operator_name: `Operador ${number}`,
    mount_number: number,
    mount_date: `2025-01-1${number}`,
    type: 'Artificial',
    note: null,
  };
}

function buildService(overrides: Partial<Service> = {}): Service {
  return {
    id: '1',
    sow_id: 'sow-1',
    sow_code: 'C-001',
    expected_farrowing_date: '2025-05-04',
    note: null,
    state: 'Confirmado',
    location: 'Nave 1',
    mounts: [
      {
        id: 'm1',
        service_id: '1',
        boar_id: 'boar-1',
        boar_code: 'V-001',
        operator_id: 'operator-1',
        operator_name: 'Ana',
        mount_number: 1,
        mount_date: '2025-01-10',
        type: 'Artificial',
        note: null,
      },
    ],
    ...overrides,
  };
}

describe('ServiceCard', () => {
  beforeEach(async () => {
    await TestBed.configureTestingModule({ imports: [ServiceCard] }).compileComponents();
  });

  it('renders the sow code, state, farrowing date and mounts', () => {
    const fixture = TestBed.createComponent(ServiceCard);
    fixture.componentRef.setInput('service', buildService());
    fixture.detectChanges();

    const text = fixture.nativeElement.textContent as string;
    expect(text).toContain('C-001');
    expect(text).toContain('Confirmado');
    expect(text).toContain('2025-05-04');
    expect(text).toContain('Nave 1');
    expect(text).toContain('Monta 1');
    expect(text).toContain('V-001');
    expect(text).toContain('Ana');
    expect(text).toContain('Artificial');
  });

  it('maps the state to its badge class', () => {
    const fixture = TestBed.createComponent(ServiceCard);
    fixture.componentRef.setInput('service', buildService({ state: 'Fallido' }));
    fixture.detectChanges();

    expect(fixture.nativeElement.querySelector('.text-danger')).not.toBeNull();
  });

  it('shows placeholders when the resolved names are missing', () => {
    const fixture = TestBed.createComponent(ServiceCard);
    fixture.componentRef.setInput(
      'service',
      buildService({
        sow_code: '',
        mounts: [
          {
            id: 'm1',
            service_id: '1',
            boar_id: 'boar-1',
            boar_code: '',
            operator_id: 'operator-1',
            operator_name: '',
            mount_number: 1,
            mount_date: '2025-01-10',
            type: 'Artificial',
            note: null,
          },
        ],
      }),
    );
    fixture.detectChanges();

    const text = fixture.nativeElement.textContent as string;
    expect(text).toContain('sin identificar');
    expect(text).toContain('Verraco');
    expect(text).toContain('Operador');
  });

  it('hides the actions when editing is not allowed', () => {
    const fixture = TestBed.createComponent(ServiceCard);
    fixture.componentRef.setInput('service', buildService());
    fixture.componentRef.setInput('canEdit', false);
    fixture.detectChanges();

    expect(fixture.nativeElement.querySelectorAll('button')).toHaveLength(0);
  });

  it('hides the actions for non-confirmed services', () => {
    const fixture = TestBed.createComponent(ServiceCard);
    fixture.componentRef.setInput('service', buildService({ state: 'Terminado' }));
    fixture.componentRef.setInput('canEdit', true);
    fixture.detectChanges();

    expect(fixture.nativeElement.querySelectorAll('button')).toHaveLength(0);
  });

  it('emits the service when the edit button is clicked', () => {
    const service = buildService();
    const fixture = TestBed.createComponent(ServiceCard);
    fixture.componentRef.setInput('service', service);
    fixture.componentRef.setInput('canEdit', true);
    fixture.detectChanges();

    const emitted: Service[] = [];
    fixture.componentInstance.editRequested.subscribe((value) => emitted.push(value));
    (fixture.nativeElement.querySelectorAll('button')[0] as HTMLButtonElement).click();

    expect(emitted).toEqual([service]);
  });

  it('emits the service when the delete button is clicked', () => {
    const service = buildService();
    const fixture = TestBed.createComponent(ServiceCard);
    fixture.componentRef.setInput('service', service);
    fixture.componentRef.setInput('canEdit', true);
    fixture.detectChanges();

    const emitted: Service[] = [];
    fixture.componentInstance.deleteRequested.subscribe((value) => emitted.push(value));
    (fixture.nativeElement.querySelectorAll('button')[1] as HTMLButtonElement).click();

    expect(emitted).toEqual([service]);
  });

  it('previews only the first mount and offers a toggle when there are several', () => {
    const fixture = TestBed.createComponent(ServiceCard);
    fixture.componentRef.setInput(
      'service',
      buildService({ mounts: [buildMount(1), buildMount(2)] }),
    );
    fixture.detectChanges();

    const text = fixture.nativeElement.textContent as string;
    expect(text).toContain('Monta 1');
    expect(text).not.toContain('Monta 2');
    expect(findToggleButton(fixture.nativeElement)?.textContent).toContain('Ver todas (1)');
  });

  it('reveals the remaining mounts when the toggle is clicked', () => {
    const fixture = TestBed.createComponent(ServiceCard);
    fixture.componentRef.setInput(
      'service',
      buildService({ mounts: [buildMount(1), buildMount(2)] }),
    );
    fixture.detectChanges();

    (findToggleButton(fixture.nativeElement) as HTMLButtonElement).click();
    fixture.detectChanges();

    const text = fixture.nativeElement.textContent as string;
    expect(text).toContain('Monta 2');
    expect(findToggleButton(fixture.nativeElement)?.textContent).toContain('Ocultar');
  });

  it('shows the mounts toggle even when editing is not allowed', () => {
    const fixture = TestBed.createComponent(ServiceCard);
    fixture.componentRef.setInput(
      'service',
      buildService({ mounts: [buildMount(1), buildMount(2)] }),
    );
    fixture.componentRef.setInput('canEdit', false);
    fixture.detectChanges();

    expect(findToggleButton(fixture.nativeElement)).not.toBeNull();
  });

  it('does not show the mounts toggle for a single mount', () => {
    const fixture = TestBed.createComponent(ServiceCard);
    fixture.componentRef.setInput('service', buildService());
    fixture.componentRef.setInput('canEdit', false);
    fixture.detectChanges();

    expect(findToggleButton(fixture.nativeElement)).toBeNull();
  });
});

function findToggleButton(nativeElement: HTMLElement): HTMLButtonElement | null {
  const buttons = Array.from(nativeElement.querySelectorAll('button'));
  return (
    (buttons.find((button) => /Ver todas|Ocultar/.test(button.textContent ?? '')) as
      HTMLButtonElement | undefined) ?? null
  );
}
