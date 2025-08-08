import { describe, it, expect } from 'vitest'
import { barColor } from './DashboardPage'

describe('Dashboard helpers', () => {
  it('barColor wraps palette', () => {
    const a = barColor(0)
    const b = barColor(15)
    const c = barColor(16)
    expect(a).toBeTruthy()
  expect(b).toBeTruthy()
  // 15 % 15 = 0
  expect(b).toBe(a)
  // 16 % 15 = 1
  expect(c).not.toBe(a)
  })
})
