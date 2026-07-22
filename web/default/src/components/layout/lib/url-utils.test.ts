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
import assert from 'node:assert/strict'
import { describe, test } from 'node:test'
import { normalizeHref } from './url-utils'

describe('normalizeHref', () => {
  test('removes query parameters', () => {
    assert.equal(normalizeHref('/dashboard?page=1'), '/dashboard')
  })

  test('removes trailing slashes', () => {
    assert.equal(normalizeHref('/dashboard/'), '/dashboard')
  })

  test('removes multiple trailing slashes', () => {
    assert.equal(normalizeHref('/dashboard///'), '/dashboard')
  })

  test('removes both query and trailing slashes', () => {
    assert.equal(normalizeHref('/dashboard/?tab=settings'), '/dashboard')
  })

  test('keeps root path as-is', () => {
    assert.equal(normalizeHref('/'), '/')
  })

  test('keeps root path with query as root', () => {
    assert.equal(normalizeHref('/?redirect=/login'), '/')
  })

  test('handles empty string', () => {
    assert.equal(normalizeHref(''), '')
  })

  test('handles plain path without query or slashes', () => {
    assert.equal(normalizeHref('/users'), '/users')
  })
})
