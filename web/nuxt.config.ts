// https://nuxt.com/docs/api/configuration/nuxt-config
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
    coreApiUrl: process.env.CORE_API_URL || 'http://localhost:8080'
  },

  routeRules: {
    '/api/**': {
      cors: true
    }
  },

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

  compatibilityDate: '2024-07-11',

  eslint: {
    config: {
      stylistic: {
        commaDangle: 'never',
        braceStyle: '1tbs'
      }
    }
  }
})
