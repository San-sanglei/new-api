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
import { type ReactNode } from 'react'
import { cn } from '@/lib/utils'

function Skeleton({ className, ...props }: React.ComponentProps<'div'>) {
  return (
    <div
      data-slot='skeleton'
      className={cn('bg-muted animate-pulse rounded-md', className)}
      {...props}
    />
  )
}

/** Full-width shimmer bar — use for line placeholders */
function SkeletonBar({
  className,
  width,
  ...props
}: React.ComponentProps<'div'> & { width?: string }) {
  return (
    <Skeleton
      className={cn('h-4', className)}
      style={width ? { width } : undefined}
      {...props}
    />
  )
}

/** Card-shaped skeleton block */
function SkeletonCard({ className, ...props }: React.ComponentProps<'div'>) {
  return (
    <div
      className={cn(
        'bg-card overflow-hidden rounded-2xl border shadow-xs',
        className
      )}
      {...props}
    >
      <div className='border-b px-4 py-3 sm:px-5'>
        <Skeleton className='h-4 w-1/3' />
      </div>
      <div className='space-y-3 p-4 sm:p-5'>
        <Skeleton className='h-3 w-full' />
        <Skeleton className='h-3 w-4/5' />
        <Skeleton className='h-3 w-3/5' />
      </div>
    </div>
  )
}

/** Table-shaped skeleton block */
function SkeletonTable({
  className,
  rows = 5,
  cols = 4,
  ...props
}: React.ComponentProps<'div'> & { rows?: number; cols?: number }) {
  return (
    <div className={cn('overflow-hidden rounded-2xl border', className)} {...props}>
      <div className='bg-muted/20 flex h-10 items-center gap-4 border-b px-4 sm:px-5'>
        {Array.from({ length: cols }).map((_, i) => (
          <Skeleton key={i} className='h-3.5 w-16 sm:w-24' />
        ))}
      </div>
      {Array.from({ length: rows }).map((_, r) => (
        <div
          key={r}
          className='flex items-center gap-4 border-b px-4 py-3 last:border-0 sm:px-5'
        >
          {Array.from({ length: cols }).map((_, c) => (
            <Skeleton
              key={c}
              className='h-3.5'
              style={{
                width: `${50 + Math.sin(r * 7 + c * 11) * 30}%`,
              }}
            />
          ))}
        </div>
      ))}
    </div>
  )
}

/** Form-shaped skeleton block */
function SkeletonForm({ className, ...props }: React.ComponentProps<'div'>) {
  return (
    <div
      className={cn('space-y-6', className)}
      {...props}
    >
      {Array.from({ length: 3 }).map((_, i) => (
        <div key={i} className='space-y-2'>
          <Skeleton className='h-3.5 w-24' />
          <Skeleton className='h-9 w-full rounded-lg' />
          <Skeleton className='h-3 w-3/4' />
        </div>
      ))}
      <div className='flex gap-3 pt-2'>
        <Skeleton className='h-9 w-24 rounded-lg' />
        <Skeleton className='h-9 w-24 rounded-lg' />
      </div>
    </div>
  )
}

/** Dashboard stats row skeleton */
function SkeletonStats({
  className,
  count = 4,
  ...props
}: React.ComponentProps<'div'> & { count?: number }) {
  return (
    <div
      className={cn('grid grid-cols-1 gap-4 sm:grid-cols-2 xl:grid-cols-4', className)}
      {...props}
    >
      {Array.from({ length: count }).map((_, i) => (
        <div
          key={i}
          className='bg-card space-y-3 rounded-2xl border p-4 shadow-xs'
        >
          <div className='flex items-center gap-3'>
            <Skeleton className='size-8 rounded-lg' />
            <div className='flex-1 space-y-1.5'>
              <Skeleton className='h-3 w-20' />
              <Skeleton className='h-5 w-28' />
            </div>
          </div>
          <Skeleton className='h-3 w-16' />
        </div>
      ))}
    </div>
  )
}

/**
 * Full-page section skeleton that mirrors SectionPageLayout structure.
 * Use as a quick loading fallback in routes.
 */
function SectionSkeleton({
  className,
  children,
  ...props
}: React.ComponentProps<'div'> & { children?: ReactNode }) {
  if (children) {
    return (
      <div className={cn('opacity-50 pointer-events-none select-none', className)} {...props}>
        {children}
      </div>
    )
  }

  return (
    <div className={cn('flex min-h-0 flex-1 flex-col', className)} {...props}>
      <div className='flex shrink-0 items-center justify-between px-3 pt-3 pb-2.5 sm:px-4 sm:pt-5 sm:pb-3'>
        <Skeleton className='h-5 w-40 sm:h-6' />
        <div className='flex gap-2'>
          <Skeleton className='h-8 w-20 rounded-lg' />
          <Skeleton className='h-8 w-20 rounded-lg' />
        </div>
      </div>
      <div className='flex-1 space-y-4 overflow-auto px-3 pb-3 sm:px-4 sm:pb-4'>
        <SkeletonTable />
      </div>
    </div>
  )
}

export {
  Skeleton,
  SkeletonBar,
  SkeletonCard,
  SkeletonTable,
  SkeletonForm,
  SkeletonStats,
  SectionSkeleton,
}
