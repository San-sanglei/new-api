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
import { useNavigate, useRouter } from '@tanstack/react-router'
import { useTranslation } from 'react-i18next'
import { cn } from '@/lib/utils'
import { Button } from '@/components/ui/button'
import { PublicLayout } from '@/components/layout'

const FEEDBACK_URL = 'https://github.com/QuantumNous/new-api/issues'

type GeneralErrorProps = React.HTMLAttributes<HTMLDivElement> & {
  minimal?: boolean
  error?: unknown
}

function getHttpStatus(error: unknown): number | undefined {
  if (typeof error !== 'object' || error === null) return undefined
  const response = (error as Record<string, unknown>).response
  if (typeof response !== 'object' || response === null) return undefined
  const status = (response as Record<string, unknown>).status
  return typeof status === 'number' ? status : undefined
}

export function GeneralError({
  className,
  minimal = false,
  error,
}: GeneralErrorProps) {
  const { t } = useTranslation()
  const navigate = useNavigate()
  const { history } = useRouter()
  const status = getHttpStatus(error)
  const isRateLimited = status === 429
  const title = isRateLimited
    ? t('Too many requests')
    : `${t('Oops! Something went wrong')} ${`:')`}`
  const description = isRateLimited
    ? t('Please wait a moment before trying again.')
    : t('Please try again later.')

  // Debug: log the actual error to help diagnose 500 pages
  if (import.meta.env.DEV && error) {
    // eslint-disable-next-line no-console
    console.error('[GeneralError] error object:', error)
  }

  const content = (
    <div className='m-auto flex h-full w-full flex-col items-center justify-center gap-2'>
      {!minimal && (
        <h1 className='text-[7rem] leading-tight font-bold'>
          {status ?? 500}
        </h1>
      )}
      <span className='font-medium'>{title}</span>
      <p className='text-muted-foreground text-center'>
        {t('We apologize for the inconvenience.')} <br /> {description}
      </p>
      {import.meta.env.DEV && error && (
        <pre className='text-muted-foreground mt-2 max-w-4xl overflow-auto rounded bg-muted p-3 text-left text-xs'>
          {typeof error === 'object' && error !== null
            ? `${(error as Error).message || ''}\n\n${(error as Error).stack || ''}\n\n${JSON.stringify(error, Object.getOwnPropertyNames(error), 2)}`
            : String(error)}
        </pre>
      )}
      {!minimal && (
        <p className='text-muted-foreground text-center text-sm'>
          {t('If this keeps happening, please report it on GitHub Issues.')}
        </p>
      )}
      {!minimal && (
        <div className='mt-6 flex flex-wrap justify-center gap-4'>
          <Button variant='outline' onClick={() => history.go(-1)}>
            {t('Go Back')}
          </Button>
          <Button
            variant='outline'
            render={
              <a
                href={FEEDBACK_URL}
                target='_blank'
                rel='noopener noreferrer'
              />
            }
          >
            {t('Report an issue')}
          </Button>
          <Button onClick={() => navigate({ to: '/' })}>
            {t('Back to Home')}
          </Button>
        </div>
      )}
    </div>
  )

  // minimal 模式用于其他布局内嵌展示，不加 PublicLayout
  if (minimal) {
    return (
      <div className={cn('h-svh w-full', className)}>
        {content}
      </div>
    )
  }

  return (
    <PublicLayout showMainContainer={false}>
      <div className={cn('flex h-[calc(100svh-4rem)] w-full flex-col', className)}>
        {content}
      </div>
    </PublicLayout>
  )
}
