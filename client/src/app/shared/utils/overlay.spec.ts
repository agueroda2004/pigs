import { shouldFlipUp } from './overlay';

function element(rect: { top?: number; bottom?: number; height?: number }): HTMLElement {
  return {
    getBoundingClientRect: () => ({
      top: rect.top ?? 0,
      bottom: rect.bottom ?? 0,
      height: rect.height ?? 0,
    }),
  } as unknown as HTMLElement;
}

describe('shouldFlipUp', () => {
  it('opens downwards when the panel fits below the trigger', () => {
    const trigger = element({ top: 100, bottom: 140 });
    const panel = element({ height: 200 });

    expect(shouldFlipUp(trigger, panel, { viewportHeight: 800 })).toBe(false);
  });

  it('opens upwards when the panel does not fit below and there is more room above', () => {
    const trigger = element({ top: 700, bottom: 740 });
    const panel = element({ height: 300 });

    expect(shouldFlipUp(trigger, panel, { viewportHeight: 800 })).toBe(true);
  });

  it('stays downwards when there is more room below even if the panel does not fit', () => {
    const trigger = element({ top: 300, bottom: 340 });
    const panel = element({ height: 500 });

    expect(shouldFlipUp(trigger, panel, { viewportHeight: 800 })).toBe(false);
  });

  it('stays downwards when the available spaces are equal', () => {
    const trigger = element({ top: 400, bottom: 440 });
    const panel = element({ height: 300 });

    expect(shouldFlipUp(trigger, panel, { viewportHeight: 800 })).toBe(false);
  });

  it('accounts for the configured gap', () => {
    const trigger = element({ top: 500, bottom: 540 });
    const panel = element({ height: 240 });

    expect(shouldFlipUp(trigger, panel, { viewportHeight: 800, gap: 8 })).toBe(false);
    expect(shouldFlipUp(trigger, panel, { viewportHeight: 800, gap: 40 })).toBe(true);
  });
});
