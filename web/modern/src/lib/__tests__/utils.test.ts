import { describe, it, expect } from 'vitest'
import { toDateTimeLocal, fromDateTimeLocal } from '@/lib/utils'

describe('datetime-local helpers', () => {
  it('round-trips epoch seconds via datetime-local', () => {
    const now = Math.floor(Date.now() / 1000)
    const str = toDateTimeLocal(now)
    const back = fromDateTimeLocal(str)
    // minutes precision due to formatting
    expect(Math.abs(back - now)).toBeLessThanOrEqual(60)
  })

  it('handles empty input', () => {
    expect(fromDateTimeLocal('')).toBe(0)
    expect(toDateTimeLocal(0)).toBe('')
  })
})
