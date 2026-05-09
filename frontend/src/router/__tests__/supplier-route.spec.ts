import { describe, expect, it, vi } from 'vitest'

const authStore = vi.hoisted(() => ({
  checkAuth: vi.fn(),
  isAuthenticated: false,
  isAdmin: false,
  isSupplier: false,
  hasSupplierProfile: false,
  isSimpleMode: false,
  hasPendingAuthSession: false,
}))

const appStore = vi.hoisted(() => ({
  siteName: 'Sub2API',
  backendModeEnabled: false,
  cachedPublicSettings: null as null | Record<string, unknown>,
}))

vi.mock('@/stores/auth', () => ({
  useAuthStore: () => authStore,
}))

vi.mock('@/stores/app', () => ({
  useAppStore: () => appStore,
}))

vi.mock('@/stores/adminSettings', () => ({
  useAdminSettingsStore: () => ({
    customMenuItems: [],
  }),
}))

vi.mock('@/composables/useNavigationLoading', () => ({
  useNavigationLoadingState: () => ({
    startNavigation: vi.fn(),
    endNavigation: vi.fn(),
    isLoading: { value: false },
  }),
}))

vi.mock('@/composables/useRoutePrefetch', () => ({
  useRoutePrefetch: () => ({
    triggerPrefetch: vi.fn(),
    cancelPendingPrefetch: vi.fn(),
    resetPrefetchState: vi.fn(),
  }),
}))

describe('supplier routes', () => {
  it('does not register the removed supplier apply page', async () => {
    const { default: router } = await import('@/router')

    expect(router.getRoutes().some((record) => record.path === '/supplier/apply')).toBe(false)
  })

  it('allows pending suppliers with a profile to land on dashboard and usage routes', async () => {
    const { default: router } = await import('@/router')
    const dashboard = router.getRoutes().find((record) => record.name === 'SupplierDashboard')
    const usage = router.getRoutes().find((record) => record.name === 'SupplierUsage')
    const accounts = router.getRoutes().find((record) => record.name === 'SupplierAccounts')

    expect(dashboard?.meta.requiresSupplierProfile).toBe(true)
    expect(dashboard?.meta.requiresSupplier).not.toBe(true)
    expect(usage?.meta.requiresSupplierProfile).toBe(true)
    expect(usage?.meta.requiresSupplier).not.toBe(true)
    expect(accounts?.meta.requiresSupplier).toBe(true)
  })
})
