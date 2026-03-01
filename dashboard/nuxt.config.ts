// https://nuxt.com/docs/api/configuration/nuxt-config
const runtimeEnv = globalThis as { process?: { env?: Record<string, string | undefined> } }

export default defineNuxtConfig({
  modules: [
    '@nuxt/eslint',
    '@nuxt/ui',
    '@vueuse/nuxt'
  ],

  devtools: {
    enabled: true
  },

  css: ['~/assets/css/main.css'],

  runtimeConfig: {
    coreApiUrl: runtimeEnv.process?.env?.CORE_API_URL || 'http://localhost:8080'
  },

  routeRules: {
    '/api/**': {
      cors: true
    }
  },

  compatibilityDate: '2024-07-11',

  vite: {
    optimizeDeps: {
      include: [
        'zod',
        '@tanstack/table-core',
        'date-fns',
        '@vue/devtools-core',
        '@vue/devtools-kit'
      ]
    }
  },

  eslint: {
    config: {
      stylistic: {
        commaDangle: 'never',
        braceStyle: '1tbs'
      }
    }
  }
})
