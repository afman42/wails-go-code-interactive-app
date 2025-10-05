import '@testing-library/jest-dom/vitest'
import '@testing-library/svelte/vitest'

if (typeof Range !== 'undefined' && !Range.prototype.getClientRects) {
  Range.prototype.getClientRects = () => ({
    length: 0,
    item: () => null,
    [Symbol.iterator]: function* () {}
  }) as any
}
