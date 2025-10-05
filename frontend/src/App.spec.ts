/// <reference types="@testing-library/jest-dom" />
import { render, screen, waitFor } from '@testing-library/svelte/svelte5'
import { afterEach, describe, expect, it, vi } from 'vitest'
import App from './App.svelte'

const checkFileExecutableMock = vi.fn()

vi.mock('../wailsjs/go/main/App.js', () => ({
  CheckFileExecutable: (...args: unknown[]) => checkFileExecutableMock(...args),
  RunFileExecutable: vi.fn()
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
  checkFileExecutableMock.mockReset()
  localStorage.clear()
})

describe('App', () => {
  it('renders setup prompt when no executables are found', async () => {
    checkFileExecutableMock.mockResolvedValueOnce([])

    render(App)

    await waitFor(() =>
      expect(
        screen.getByText('Please Add Golang,Node JS,PHP Installation')
      ).toBeInTheDocument()
    )
  })

  it('lists detected runtimes returned from backend', async () => {
    checkFileExecutableMock.mockResolvedValueOnce(['node', 'go'])

    render(App)

    await waitFor(() => expect(screen.getByDisplayValue('node')).toBeInTheDocument())
    expect(screen.getByDisplayValue('go')).toBeInTheDocument()
  })
})
