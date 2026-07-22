/*
Copyright (C) 2023-2026 QuantumNous

This program is free software: you can redistribute it and/or modify
it under the terms of the GNU Affero General Public License as
published by the Free Software Foundation, either version 3 of the
License, or (at your option) any later version.

This program is distributed in the hope that it will be useful,
but WITHOUT ANY WARRANTY; without even the implied warranty of
MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE. See the
GNU Affero General Public License for more details.

You should have received a copy of the GNU Affero General Public License
along with this program. If not, see <https://www.gnu.org/licenses/>.

For commercial licensing, please contact support@quantumnous.com
*/
import {
  createContext,
  useCallback,
  useContext,
  useEffect,
  useMemo,
  useState,
} from 'react'
import { getCookie, setCookie, removeCookie } from '@/lib/cookies'

type Theme = 'dark' | 'light' | 'system'
type ResolvedTheme = Exclude<Theme, 'system'>

const DEFAULT_THEME = 'dark'
const THEME_STORAGE_KEY = 'vite-ui-theme'
// Legacy cookie name kept in sync so server-side reads (if any) still work
const THEME_COOKIE_NAME = 'vite-ui-theme'
const THEME_COOKIE_MAX_AGE = 60 * 60 * 24 * 365 // 1 year
const THEMES = new Set<Theme>(['dark', 'light', 'system'])

type ThemeProviderProps = {
  children: React.ReactNode
  defaultTheme?: Theme
  storageKey?: string
}

type ThemeProviderState = {
  defaultTheme: Theme
  resolvedTheme: ResolvedTheme
  theme: Theme
  setTheme: (theme: Theme) => void
  resetTheme: () => void
}

const initialState: ThemeProviderState = {
  defaultTheme: DEFAULT_THEME,
  resolvedTheme: 'light',
  theme: DEFAULT_THEME,
  setTheme: () => null,
  resetTheme: () => null,
}

const ThemeContext = createContext<ThemeProviderState>(initialState)

function getSystemTheme(): ResolvedTheme {
  if (typeof window === 'undefined') return 'light'
  return window.matchMedia('(prefers-color-scheme: dark)').matches
    ? 'dark'
    : 'light'
}

function resolveTheme(theme: Theme): ResolvedTheme {
  return theme === 'system' ? getSystemTheme() : theme
}

/**
 * Read the persisted theme preference.
 *
 * Primary source is `localStorage` (per the project requirement). We also
 * fall back to the legacy cookie so existing users keep their preference
 * after the migration. Returns the fallback when nothing valid is stored.
 */
function getStoredTheme(storageKey: string, fallback: Theme): Theme {
  if (typeof window !== 'undefined') {
    try {
      const lsTheme = window.localStorage.getItem(storageKey) as Theme | null
      if (lsTheme && THEMES.has(lsTheme)) return lsTheme
    } catch {
      // localStorage may be unavailable (private mode / disabled) — fall through to cookie
    }
  }

  const cookieTheme = getCookie(THEME_COOKIE_NAME) as Theme | undefined
  if (cookieTheme && THEMES.has(cookieTheme)) return cookieTheme

  return fallback
}

export function ThemeProvider({
  children,
  defaultTheme = DEFAULT_THEME,
  storageKey = THEME_STORAGE_KEY,
  ...props
}: ThemeProviderProps) {
  const [theme, _setTheme] = useState<Theme>(() =>
    getStoredTheme(storageKey, defaultTheme)
  )
  const [resolvedTheme, setResolvedTheme] = useState<ResolvedTheme>(() =>
    resolveTheme(getStoredTheme(storageKey, defaultTheme))
  )

  useEffect(() => {
    const root = window.document.documentElement
    const mediaQuery = window.matchMedia('(prefers-color-scheme: dark)')

    const applyTheme = () => {
      const nextResolvedTheme = theme === 'system' ? getSystemTheme() : theme
      root.classList.remove('light', 'dark')
      root.classList.add(nextResolvedTheme)
      root.style.colorScheme = nextResolvedTheme
      setResolvedTheme(nextResolvedTheme)
    }

    applyTheme()

    // Enable smooth color transitions only AFTER the first paint. This keeps
    // the initial load (handled by the inline script in index.html) flash-free
    // while making subsequent user-initiated switches animate smoothly.
    const rafId = window.requestAnimationFrame(() => {
      root.classList.add('theme-transition')
    })

    mediaQuery.addEventListener('change', applyTheme)

    return () => {
      window.cancelAnimationFrame(rafId)
      mediaQuery.removeEventListener('change', applyTheme)
    }
  }, [theme])

  const setTheme = useCallback(
    (theme: Theme) => {
      // Persist to localStorage (primary) and mirror to the legacy cookie
      // so any server-side / pre-hydration reads stay consistent.
      try {
        window.localStorage.setItem(storageKey, theme)
      } catch {
        /* localStorage unavailable — cookie below still carries the value */
      }
      setCookie(THEME_COOKIE_NAME, theme, THEME_COOKIE_MAX_AGE)
      _setTheme(theme)
    },
    [storageKey]
  )

  const resetTheme = useCallback(() => {
    try {
      window.localStorage.removeItem(storageKey)
    } catch {
      /* empty */
    }
    removeCookie(THEME_COOKIE_NAME)
    _setTheme(defaultTheme)
  }, [defaultTheme, storageKey])

  const contextValue = useMemo(
    () => ({
      defaultTheme,
      resolvedTheme,
      resetTheme,
      theme,
      setTheme,
    }),
    [defaultTheme, resolvedTheme, resetTheme, theme, setTheme]
  )

  return (
    <ThemeContext value={contextValue} {...props}>
      {children}
    </ThemeContext>
  )
}

// eslint-disable-next-line react-refresh/only-export-components
export const useTheme = () => {
  const context = useContext(ThemeContext)

  if (!context) throw new Error('useTheme must be used within a ThemeProvider')

  return context
}
