/// <reference types="@testing-library/jest-dom" />
import { render, screen, waitFor } from '@testing-library/svelte/svelte5'
import { afterEach, describe, expect, it, vi } from 'vitest'
import App from './App.svelte'

const listLanguageAvailabilityMock = vi.fn()
const listBundledRuntimesMock = vi.fn()
const runFileExecutableMock = vi.fn()

vi.mock('../wailsjs/go/main/App.js', () => ({
  ListLanguageAvailability: (...args: unknown[]) =>
    listLanguageAvailabilityMock(...args),
  ListBundledRuntimes: (...args: unknown[]) =>
    listBundledRuntimesMock(...args),
  RunFileExecutable: (...args: unknown[]) => runFileExecutableMock(...args)
}))

vi.mock('../wailsjs/go/models', () => ({
  main: {
    Data: class {
      txt = ''
      out = ''
      errout = ''
      lang = ''
      type = ''
    }
  }
}))

afterEach(() => {
  listLanguageAvailabilityMock.mockReset()
  listBundledRuntimesMock.mockReset()
  runFileExecutableMock.mockReset()
  localStorage.clear()
})

describe('App', () => {
  it('renders setup prompt when no executables are found', async () => {
    listLanguageAvailabilityMock.mockResolvedValueOnce({
      system: [],
      bundled: []
    })
    listLanguageAvailabilityMock.mockResolvedValue({ system: [], bundled: [] })
    listBundledRuntimesMock.mockResolvedValueOnce([])
    listBundledRuntimesMock.mockResolvedValue([])

    render(App)

    await waitFor(() =>
      expect(
        screen.getByText('Please Add Golang,Node JS,PHP Installation')
      ).toBeInTheDocument()
    )
  })

  it('lists detected runtimes returned from backend', async () => {
    listLanguageAvailabilityMock.mockResolvedValueOnce({
      system: ['node', 'go'],
      bundled: []
    })
    listLanguageAvailabilityMock.mockResolvedValue({
      system: ['node', 'go'],
      bundled: []
    })
    listBundledRuntimesMock.mockResolvedValueOnce([])
    listBundledRuntimesMock.mockResolvedValue([])

    render(App)

    await waitFor(() => expect(screen.getByDisplayValue('node')).toBeInTheDocument())
    expect(screen.getByDisplayValue('go')).toBeInTheDocument()
  })
})
