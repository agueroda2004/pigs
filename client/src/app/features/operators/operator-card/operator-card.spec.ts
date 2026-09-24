import { TestBed } from '@angular/core/testing';

import { Operator } from '../../../core/operators/operator.models';
import { OperatorCard } from './operator-card';

function buildOperator(active = true): Operator {
  return {
    id: '1',
    name: 'Juan Pérez',
    active,
    created_at: '2026-01-02T12:00:00',
    updated_at: '2026-01-02T12:00:00',
    created_by: 'admin',
    updated_by: 'admin',
  };
}

describe('OperatorCard', () => {
  beforeEach(async () => {
    await TestBed.configureTestingModule({
      imports: [OperatorCard],
    }).compileComponents();
  });

  it('renders the name, status and created date', () => {
    const fixture = TestBed.createComponent(OperatorCard);
    fixture.componentRef.setInput('operator', buildOperator(true));
    fixture.detectChanges();

    const text = fixture.nativeElement.textContent as string;
    expect(text).toContain('Juan Pérez');
    expect(text).toContain('Activo');
    expect(text).toContain('02/01/2026');
  });

  it('shows the inactive status for inactive operators', () => {
    const fixture = TestBed.createComponent(OperatorCard);
    fixture.componentRef.setInput('operator', buildOperator(false));
    fixture.detectChanges();

    expect(fixture.nativeElement.textContent).toContain('Inactivo');
  });

  it('hides the edit button when editing is not allowed', () => {
    const fixture = TestBed.createComponent(OperatorCard);
    fixture.componentRef.setInput('operator', buildOperator());
    fixture.componentRef.setInput('canEdit', false);
    fixture.detectChanges();

    expect(fixture.nativeElement.querySelector('button')).toBeNull();
  });

  it('emits the operator when the edit button is clicked', () => {
    const operator = buildOperator();
    const fixture = TestBed.createComponent(OperatorCard);
    fixture.componentRef.setInput('operator', operator);
    fixture.componentRef.setInput('canEdit', true);
    fixture.detectChanges();

    const emitted: Operator[] = [];
    fixture.componentInstance.editRequested.subscribe((value) => emitted.push(value));
    (fixture.nativeElement.querySelector('button') as HTMLButtonElement).click();

    expect(emitted).toEqual([operator]);
  });
});
