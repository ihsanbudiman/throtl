import { api } from '@/lib/api'

test('smoke: basic arithmetic', () => {
  expect(1 + 1).toBe(2)
})

test('smoke: @/ path alias resolves', () => {
  expect(api).toBeDefined()
  expect(typeof api.checkSetup).toBe('function')
})
