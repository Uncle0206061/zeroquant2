<template>
  <router-view v-if="isAuthPage" />
  <template v-else>
    <WsStatusBar />
    <router-view />
  </template>
</template>

<script setup lang="ts">
import { computed } from 'vue'
import { useRoute } from 'vue-router'
import WsStatusBar from '@/components/WsStatusBar.vue'

const route = useRoute()
const isAuthPage = computed(() => route.path === '/login' || route.path === '/register')
</script>
