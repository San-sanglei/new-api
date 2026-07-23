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
/**
 * Application-wide constants
 */

// System Configuration Defaults
export const DEFAULT_SYSTEM_NAME = 'New API'
export const DEFAULT_LOGO = '/logo.png'

// Logo cache-buster: bump this whenever logo assets are replaced to
// force browsers to refetch /logo.png and /logo-white.png.
export const LOGO_CACHE_VERSION = '20260723f'

/**
 * Append a cache-busting query param to a logo URL so updated logo
 * assets bypass the browser disk cache. Preserves any existing path
 * (including -white.png variants) and avoids double-appending ?v=.
 */
export function withLogoCacheBust(src: string): string {
  if (!src) return src
  const sep = src.includes('?') ? '&' : '?'
  return `${src}${sep}v=${LOGO_CACHE_VERSION}`
}

// LocalStorage Keys
export const STORAGE_KEYS = {
  SYSTEM_NAME: 'system_name',
  LOGO: 'logo',
  FOOTER_HTML: 'footer_html',
} as const
