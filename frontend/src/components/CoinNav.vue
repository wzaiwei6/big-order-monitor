<template>
  <nav class="coin-nav">
    <router-link
      v-for="coin in coins"
      :key="coin.id"
      :to="`/${coin.id}`"
      class="coin-btn"
      :class="{ active: currentCoin === coin.id }"
    >
      {{ coin.displayName }}
    </router-link>
  </nav>
</template>

<script setup lang="ts">
import { computed } from "vue";
import { useRoute } from "vue-router";
import { COIN_LIST } from "@/config/coins";

const route = useRoute();
const coins = COIN_LIST;

const currentCoin = computed(() => {
  const path = route.path;
  return path.substring(1) || "btc";
});
</script>

<style scoped>
.coin-nav {
  display: flex;
  gap: 8px;
  flex-wrap: wrap;
  padding: 16px 0;
  border-bottom: 1px solid #1d2a44;
  margin-bottom: 24px;
}

.coin-btn {
  padding: 8px 16px;
  border-radius: 6px;
  border: 1px solid #1d2a44;
  background: #0f1626;
  color: #9fb0cc;
  text-decoration: none;
  font-size: 14px;
  font-weight: 500;
  transition: all 0.2s ease;
  cursor: pointer;
}

.coin-btn:hover {
  background: #1b2338;
  border-color: #324565;
  color: #e6e9ef;
}

.coin-btn.active {
  background: #1b59b0;
  border-color: #1b59b0;
  color: #ffffff;
}
</style>

