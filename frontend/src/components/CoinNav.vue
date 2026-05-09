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
    <router-link
      to="/summary"
      class="coin-btn"
      :class="{ active: currentCoin === 'summary' }"
    >
      汇总
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
  const segment = route.path.replace(/^\//, "");
  return segment || "btc";
});
</script>

<style scoped>
.coin-nav {
  display: flex;
  align-items: center;
  gap: 8px;
  flex-wrap: wrap;
  width: 100%;
  box-sizing: border-box;
  padding: 8px;
  margin-bottom: 28px;
  border: 1px solid var(--app-border-soft);
  border-radius: var(--app-radius-md);
  background: var(--app-nav-bg);
  box-shadow: inset 0 1px 0 rgba(255, 255, 255, 0.04);
}

.coin-btn {
  position: relative;
  display: flex;
  align-items: center;
  justify-content: center;
  min-width: 80px;
  min-height: 40px;
  padding: 0 16px;
  border-radius: 12px;
  border: 1px solid transparent;
  background: transparent;
  color: var(--app-text-muted);
  text-decoration: none;
  font-size: 14px;
  font-weight: 600;
  letter-spacing: 0.03em;
  transition: border-color 0.2s ease, color 0.2s ease, background 0.2s ease;
  cursor: pointer;
}

.coin-btn:hover {
  border-color: var(--app-border);
  color: var(--app-text);
  background: var(--app-hover-bg);
}

.coin-btn.active {
  background: var(--app-accent);
  border-color: color-mix(in srgb, var(--app-accent) 32%, transparent);
  color: var(--app-accent-contrast);
}

@media (max-width: 768px) {
  .coin-nav {
    flex-wrap: nowrap;
    gap: 6px;
    padding: 6px;
    margin-bottom: 14px;
    border-radius: 14px;
    overflow-x: auto;
    scrollbar-width: none;
    -webkit-overflow-scrolling: touch;
  }

  .coin-nav::-webkit-scrollbar {
    display: none;
  }

  .coin-btn {
    flex: 0 0 auto;
    min-width: 64px;
    min-height: 34px;
    padding: 0 12px;
    border-radius: 10px;
    font-size: 13px;
  }
}
</style>
