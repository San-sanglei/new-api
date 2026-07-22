import { Component, type ReactNode } from 'react'
import { VChart as VChartOriginal, type VChartProps } from '@visactor/react-vchart'

interface ErrorBoundaryState {
  hasError: boolean
}

class VChartErrorBoundary extends Component<
  { children: ReactNode },
  ErrorBoundaryState
> {
  state: ErrorBoundaryState = { hasError: false }

  static getDerivedStateFromError(): ErrorBoundaryState {
    return { hasError: true }
  }

  render() {
    if (this.state.hasError) {
      return null
    }
    return this.props.children
  }
}

/**
 * Safe wrapper around `@visactor/react-vchart` that catches crashes
 * caused by React StrictMode's double-render (the internal VChart
 * instance is destroyed on unmount but still referenced on remount,
 * leading to "Cannot read properties of null (reading 'spec')").
 */
export function SafeVChart(props: VChartProps) {
  if (!props.spec) return null

  return (
    <VChartErrorBoundary>
      <VChartOriginal {...props} />
    </VChartErrorBoundary>
  )
}
