import { useState, useEffect } from 'react'

interface BreakpointState {
  isMobile: boolean
  isTablet: boolean
  isDesktop: boolean
  isLarge: boolean
  currentBreakpoint: 'mobile' | 'tablet' | 'desktop' | 'large'
  width: number
  height: number
}

export function useResponsive(): BreakpointState {
  const [state, setState] = useState<BreakpointState>(() => {
    // Initialize with current window size if available
    if (typeof window !== 'undefined') {
      const width = window.innerWidth
      const height = window.innerHeight
      return {
        width,
        height,
        isMobile: width < 768,
        isTablet: width >= 768 && width < 1024,
        isDesktop: width >= 1024 && width < 1280,
        isLarge: width >= 1280,
        currentBreakpoint: width < 768 ? 'mobile' : 
                          width < 1024 ? 'tablet' : 
                          width < 1280 ? 'desktop' : 'large'
      }
    }
    
    // Server-side rendering fallback
    return {
      width: 0,
      height: 0,
      isMobile: false,
      isTablet: false,
      isDesktop: true,
      isLarge: false,
      currentBreakpoint: 'desktop'
    }
  })

  useEffect(() => {
    const updateState = () => {
      const width = window.innerWidth
      const height = window.innerHeight
      
      const newState: BreakpointState = {
        width,
        height,
        isMobile: width < 768,
        isTablet: width >= 768 && width < 1024,
        isDesktop: width >= 1024 && width < 1280,
        isLarge: width >= 1280,
        currentBreakpoint: width < 768 ? 'mobile' : 
                          width < 1024 ? 'tablet' : 
                          width < 1280 ? 'desktop' : 'large'
      }
      
      setState(newState)
    }

    // Update on mount
    updateState()

    // Add event listener with debouncing
    let timeoutId: NodeJS.Timeout
    const debouncedUpdate = () => {
      clearTimeout(timeoutId)
      timeoutId = setTimeout(updateState, 100)
    }

    window.addEventListener('resize', debouncedUpdate)
    
    return () => {
      window.removeEventListener('resize', debouncedUpdate)
      clearTimeout(timeoutId)
    }
  }, [])

  return state
}

// Additional utility hooks for specific breakpoints
export function useIsMobile(): boolean {
  const { isMobile } = useResponsive()
  return isMobile
}

export function useIsTablet(): boolean {
  const { isTablet } = useResponsive()
  return isTablet
}

export function useIsDesktop(): boolean {
  const { isDesktop, isLarge } = useResponsive()
  return isDesktop || isLarge
}

// Hook for media query matching
export function useMediaQuery(query: string): boolean {
  const [matches, setMatches] = useState(() => {
    if (typeof window !== 'undefined') {
      return window.matchMedia(query).matches
    }
    return false
  })

  useEffect(() => {
    const mediaQuery = window.matchMedia(query)
    const handler = (event: MediaQueryListEvent) => {
      setMatches(event.matches)
    }

    mediaQuery.addEventListener('change', handler)
    setMatches(mediaQuery.matches)

    return () => mediaQuery.removeEventListener('change', handler)
  }, [query])

  return matches
}

// Predefined media query hooks
export function useIsTouchDevice(): boolean {
  return useMediaQuery('(hover: none) and (pointer: coarse)')
}

export function usePrefersReducedMotion(): boolean {
  return useMediaQuery('(prefers-reduced-motion: reduce)')
}

export function usePrefersDarkMode(): boolean {
  return useMediaQuery('(prefers-color-scheme: dark)')
}

export function useIsLandscape(): boolean {
  return useMediaQuery('(orientation: landscape)')
}

export function useIsPortrait(): boolean {
  return useMediaQuery('(orientation: portrait)')
}
