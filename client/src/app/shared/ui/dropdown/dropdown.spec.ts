import { TestBed } from '@angular/core/testing';
import { vi } from 'vitest';

import { Dropdown } from './dropdown';

const options = [
  { value: '1', label: 'C-001' },
  { value: '2', label: 'C-002' },
];

function mockRects(triggerTop: number, triggerBottom: number, panelHeight: number, viewport = 768) {
  vi.spyOn(window, 'innerHeight', 'get').mockReturnValue(viewport);
  vi.spyOn(Element.prototype, 'getBoundingClientRect').mockImplementation(function (this: Element) {
    if (this.tagName === 'BUTTON') {
      return {
        top: triggerTop,
        bottom: triggerBottom,
        height: triggerBottom - triggerTop,
      } as DOMRect;
    }
    return { top: 0, bottom: 0, height: panelHeight } as DOMRect;
  });
}

describe('Dropdown', () => {
  beforeEach(async () => {
    await TestBed.configureTestingModule({ imports: [Dropdown] }).compileComponents();
  });

  afterEach(() => vi.restoreAllMocks());

  const create = () => {
    const fixture = TestBed.createComponent(Dropdown);
    fixture.componentRef.setInput('options', options);
    fixture.componentRef.setInput('placeholder', 'Seleccionar');
    fixture.detectChanges();
    return fixture;
  };

  // eslint-disable-next-line @typescript-eslint/no-explicit-any
  const toggle = (fixture: any) => {
    (fixture.nativeElement.querySelector('button') as HTMLButtonElement).click();
    fixture.detectChanges();
  };

  it('shows the placeholder when there is no value', () => {
    const fixture = create();

    expect(fixture.nativeElement.textContent).toContain('Seleccionar');
  });

  it('shows the selected label', () => {
    const fixture = create();
    fixture.componentInstance.writeValue('2');
    fixture.detectChanges();

    expect(fixture.nativeElement.textContent).toContain('C-002');
  });

  it('lists every option when opened', () => {
    const fixture = create();
    toggle(fixture);

    expect(fixture.nativeElement.querySelectorAll('li[role="option"]')).toHaveLength(2);
  });

  it('selects an option on click and closes the list', () => {
    const fixture = create();
    const changed = vi.fn();
    fixture.componentInstance.registerOnChange(changed);
    toggle(fixture);

    const first = fixture.nativeElement.querySelector(
      'li[role="option"] button',
    ) as HTMLButtonElement;
    first.click();
    fixture.detectChanges();

    expect(changed).toHaveBeenCalledWith('1');
    expect(fixture.nativeElement.querySelector('ul')).toBeNull();
  });

  it('does not open when disabled', () => {
    const fixture = create();
    fixture.componentInstance.setDisabledState(true);
    fixture.detectChanges();

    toggle(fixture);

    expect(fixture.nativeElement.querySelector('ul')).toBeNull();
  });

  it('opens downwards when the panel fits below the trigger', async () => {
    mockRects(100, 140, 200);
    const fixture = create();
    toggle(fixture);
    await fixture.whenStable();
    fixture.detectChanges();

    const panel = fixture.nativeElement.querySelector('ul') as HTMLElement;
    expect(panel.classList.contains('top-full')).toBe(true);
    expect(panel.classList.contains('bottom-full')).toBe(false);
  });

  it('opens upwards when the panel does not fit below the trigger', async () => {
    mockRects(700, 740, 300);
    const fixture = create();
    toggle(fixture);
    await fixture.whenStable();
    fixture.detectChanges();

    const panel = fixture.nativeElement.querySelector('ul') as HTMLElement;
    expect(panel.classList.contains('bottom-full')).toBe(true);
    expect(panel.classList.contains('top-full')).toBe(false);
  });
});
