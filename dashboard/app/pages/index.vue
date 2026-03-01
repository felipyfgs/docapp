<script setup lang="ts">
import { sub, format } from 'date-fns'
import type { Period, Range, DashboardResponse } from '~/types'

const range = shallowRef<Range>({
  start: sub(new Date(), { days: 30 }),
  end: new Date()
})
const period = ref<Period>('daily')

const queryParams = computed(() => ({
  from: format(range.value.start, 'yyyy-MM-dd'),
  to: format(range.value.end, 'yyyy-MM-dd'),
  group_by: period.value
}))

const { data, status, refresh } = await useFetch<DashboardResponse>('/api/dashboard', {
  lazy: true,
  query: queryParams
})
</script>

<template>
  <UDashboardPanel id="home">
    <template #header>
      <UDashboardNavbar title="Dashboard">
        <template #leading>
          <UDashboardSidebarCollapse />
        </template>
      </UDashboardNavbar>

      <UDashboardToolbar>
        <template #left>
          <HomeDateRangePicker v-model="range" class="-ms-1" />
          <HomePeriodSelect v-model="period" :range="range" />
        </template>
      </UDashboardToolbar>
    </template>

    <template #body>
      <div v-if="status === 'pending'" class="flex items-center justify-center py-20">
        <UIcon name="i-lucide-loader-2" class="size-6 animate-spin text-muted" />
      </div>
      <div v-else-if="status === 'error'" class="flex flex-col items-center justify-center gap-3 py-20">
        <UIcon name="i-lucide-circle-alert" class="size-8 text-muted" />
        <p class="text-sm text-muted">
          Erro ao carregar dados do dashboard.
        </p>
        <UButton label="Tentar novamente" variant="outline" color="neutral" icon="i-lucide-refresh-cw" @click="refresh()" />
      </div>
      <template v-else>
        <HomeStats :stats="data?.stats ?? null" />
        <HomeChart
          :data="data?.chart ?? null"
          :period="period"
        />
        <HomeSales :data="data?.recentes ?? null" />
      </template>
    </template>
  </UDashboardPanel>
</template>
