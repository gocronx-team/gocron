import { describe, it, expect } from 'vitest'
import { extractTimezone, validateCronSpec } from '../cronValidator'

describe('extractTimezone', () => {
  it('extracts CRON_TZ= prefix', () => {
    const result = extractTimezone('CRON_TZ=Asia/Shanghai 0 30 8 * * *')
    expect(result.timezone).toBe('Asia/Shanghai')
    expect(result.spec).toBe('0 30 8 * * *')
  })

  it('extracts TZ= prefix', () => {
    const result = extractTimezone('TZ=America/New_York 0 0 9 * * *')
    expect(result.timezone).toBe('America/New_York')
    expect(result.spec).toBe('0 0 9 * * *')
  })

  it('extracts TZ= prefix with descriptor', () => {
    const result = extractTimezone('CRON_TZ=UTC @daily')
    expect(result.timezone).toBe('UTC')
    expect(result.spec).toBe('@daily')
  })

  it('returns empty timezone when no prefix', () => {
    const result = extractTimezone('0 30 8 * * *')
    expect(result.timezone).toBe('')
    expect(result.spec).toBe('0 30 8 * * *')
  })

  it('returns empty timezone for descriptors without prefix', () => {
    const result = extractTimezone('@daily')
    expect(result.timezone).toBe('')
    expect(result.spec).toBe('@daily')
  })

  it('handles empty string', () => {
    const result = extractTimezone('')
    expect(result.timezone).toBe('')
    expect(result.spec).toBe('')
  })

  it('handles null/undefined', () => {
    expect(extractTimezone(null).timezone).toBe('')
    expect(extractTimezone(undefined).timezone).toBe('')
  })

  it('preserves spec with extra spaces', () => {
    const result = extractTimezone('CRON_TZ=Asia/Tokyo @every 30s')
    expect(result.timezone).toBe('Asia/Tokyo')
    expect(result.spec).toBe('@every 30s')
  })
})

describe('validateCronSpec with timezone prefix', () => {
  it('validates spec with CRON_TZ= prefix', () => {
    const result = validateCronSpec('CRON_TZ=Asia/Shanghai 0 30 8 * * *')
    expect(result.valid).toBe(true)
  })

  it('validates spec with TZ= prefix', () => {
    const result = validateCronSpec('TZ=UTC @daily')
    expect(result.valid).toBe(true)
  })

  it('rejects invalid cron after stripping timezone', () => {
    const result = validateCronSpec('CRON_TZ=Asia/Shanghai invalid')
    expect(result.valid).toBe(false)
  })

  it('validates spec without prefix (backward compatible)', () => {
    expect(validateCronSpec('0 30 8 * * *').valid).toBe(true)
    expect(validateCronSpec('@daily').valid).toBe(true)
    expect(validateCronSpec('@every 30s').valid).toBe(true)
  })

  it('rejects empty spec', () => {
    expect(validateCronSpec('').valid).toBe(false)
    expect(validateCronSpec(null).valid).toBe(false)
  })
})
