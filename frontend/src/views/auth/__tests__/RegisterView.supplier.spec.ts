import { flushPromises, mount } from '@vue/test-utils'
import { beforeEach, describe, expect, it, vi } from 'vitest'
import RegisterView from '@/views/auth/RegisterView.vue'

const {
  pushMock,
  showErrorMock,
  showSuccessMock,
  registerMock,
  getPublicSettingsMock,
  validatePromoCodeMock,
  validateInvitationCodeMock,
} = vi.hoisted(() => ({
  pushMock: vi.fn(),
  showErrorMock: vi.fn(),
  showSuccessMock: vi.fn(),
  registerMock: vi.fn(),
  getPublicSettingsMock: vi.fn(),
  validatePromoCodeMock: vi.fn(),
  validateInvitationCodeMock: vi.fn(),
}))

let authState = {
  isAdmin: false,
  isSupplier: false,
  hasSupplierProfile: false,
}

vi.mock('vue-router', () => ({
  useRouter: () => ({
    push: pushMock,
  }),
  useRoute: () => ({
    query: {},
  }),
}))

vi.mock('vue-i18n', () => ({
  createI18n: () => ({
    global: {
      locale: { value: 'zh' },
      setLocaleMessage: vi.fn(),
    },
  }),
  useI18n: () => ({
    t: (key: string, params?: Record<string, string | number>) => {
      if (key === 'auth.accountCreatedSuccess') {
        return `Account created for ${params?.siteName ?? 'Sub2API'}`
      }
      return key
    },
    locale: { value: 'zh-CN' },
  }),
}))

vi.mock('@/stores', () => ({
  useAuthStore: () => ({
    get isAdmin() {
      return authState.isAdmin
    },
    get isSupplier() {
      return authState.isSupplier
    },
    get hasSupplierProfile() {
      return authState.hasSupplierProfile
    },
    register: (...args: any[]) => registerMock(...args),
  }),
  useAppStore: () => ({
    showError: (...args: any[]) => showErrorMock(...args),
    showSuccess: (...args: any[]) => showSuccessMock(...args),
  }),
}))

vi.mock('@/api/auth', async () => {
  const actual = await vi.importActual<typeof import('@/api/auth')>('@/api/auth')
  return {
    ...actual,
    getPublicSettings: (...args: any[]) => getPublicSettingsMock(...args),
    validatePromoCode: (...args: any[]) => validatePromoCodeMock(...args),
    validateInvitationCode: (...args: any[]) => validateInvitationCodeMock(...args),
    isWeChatWebOAuthEnabled: () => false,
  }
})

vi.mock('@/utils/oauthAffiliate', async () => {
  const actual = await vi.importActual<typeof import('@/utils/oauthAffiliate')>('@/utils/oauthAffiliate')
  return {
    ...actual,
    loadAffiliateReferralCode: () => '',
    resolveAffiliateReferralCode: () => '',
    clearAffiliateReferralCode: vi.fn(),
  }
})

function mountRegisterView() {
  return mount(RegisterView, {
    global: {
      stubs: {
        AuthLayout: { template: '<div><slot /><slot name="footer" /></div>' },
        EmailOAuthButtons: true,
        LinuxDoOAuthSection: true,
        OidcOAuthSection: true,
        WechatOAuthSection: true,
        LoginAgreementPrompt: true,
        TurnstileWidget: true,
        Icon: true,
        RouterLink: true,
        transition: false,
      },
    },
  })
}

describe('RegisterView supplier signup', () => {
  beforeEach(() => {
    pushMock.mockReset()
    showErrorMock.mockReset()
    showSuccessMock.mockReset()
    registerMock.mockReset()
    getPublicSettingsMock.mockReset()
    validatePromoCodeMock.mockReset()
    validateInvitationCodeMock.mockReset()
    sessionStorage.clear()
    localStorage.clear()
    authState = {
      isAdmin: false,
      isSupplier: false,
      hasSupplierProfile: false,
    }

    getPublicSettingsMock.mockResolvedValue({
      registration_enabled: true,
      email_verify_enabled: false,
      promo_code_enabled: false,
      invitation_code_enabled: false,
      turnstile_enabled: false,
      turnstile_site_key: '',
      site_name: 'Sub2API',
      linuxdo_oauth_enabled: false,
      wechat_oauth_enabled: false,
      oidc_oauth_enabled: false,
      github_oauth_enabled: false,
      google_oauth_enabled: false,
      registration_email_suffix_whitelist: [],
      login_agreement_enabled: false,
      login_agreement_documents: [],
    })
  })

  it('requires supplier company name before submitting', async () => {
    const wrapper = mountRegisterView()
    await flushPromises()

    await wrapper.findAll('button[type="button"]').find((button) => button.text().includes('供应商用户'))!.trigger('click')
    await wrapper.get('#email').setValue('supplier@example.com')
    await wrapper.get('#password').setValue('secret-123')
    await wrapper.get('form').trigger('submit.prevent')

    expect(registerMock).not.toHaveBeenCalled()
    expect(showErrorMock).toHaveBeenCalledWith('请输入供应商主体名称')
  })

  it('submits supplier profile and redirects pending supplier to dashboard', async () => {
    registerMock.mockImplementation(async () => {
      authState.hasSupplierProfile = true
      return {}
    })
    const wrapper = mountRegisterView()
    await flushPromises()

    await wrapper.findAll('button[type="button"]').find((button) => button.text().includes('供应商用户'))!.trigger('click')
    await wrapper.get('#email').setValue(' supplier@example.com ')
    await wrapper.get('#password').setValue('secret-123')
    await wrapper.get('[data-testid="register-supplier-company-name"]').setValue('  Example Supplier  ')
    await wrapper.get('[data-testid="register-supplier-contact-name"]').setValue('  Alice  ')
    await wrapper.get('[data-testid="register-supplier-contact-email"]').setValue(' contact@example.com ')
    await wrapper.get('[data-testid="register-supplier-contact-phone"]').setValue(' 13800000000 ')
    await wrapper.get('[data-testid="register-supplier-notes"]').setValue('  OpenAI resources  ')
    await wrapper.get('form').trigger('submit.prevent')
    await flushPromises()

    expect(registerMock).toHaveBeenCalledWith(
      expect.objectContaining({
        email: 'supplier@example.com',
        password: 'secret-123',
        account_type: 'supplier',
        supplier_profile: {
          company_name: 'Example Supplier',
          contact_name: 'Alice',
          contact_email: 'contact@example.com',
          contact_phone: '13800000000',
          notes: 'OpenAI resources',
        },
      })
    )
    expect(pushMock).toHaveBeenCalledWith('/supplier/dashboard')
  })
})
