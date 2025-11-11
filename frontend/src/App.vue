<template>
  <main class="wrapper">
    <header class="header">
      <h1>Binance 大额订单监控</h1>
      <span class="subtitle">V 2.1</span>
    </header>
    <CoinNav />
    <router-view />
  </main>
</template>

<script setup lang="ts">
import CoinNav from "./components/CoinNav.vue";
</script>

<style scoped>
.wrapper {
  min-height: 100vh;
  background: #0b1220;
  color: #e6e9ef;
  padding: 24px;
  font-family: -apple-system, BlinkMacSystemFont, "Segoe UI", sans-serif;
}

.header {
  display: flex;
  align-items: baseline;
  gap: 12px;
  margin-bottom: 24px;
}

.subtitle {
  font-size: 14px;
  color: #9fb0cc;
}
</style>

