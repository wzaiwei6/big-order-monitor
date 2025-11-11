<template>
  <section>
    <header class="header">
      <h3>实时统计</h3>
      <span :class="badgeClass">{{ badgeText }}</span>
    </header>
    <div v-if="stats" class="grid">
      <div class="item">
        <p class="label">买单墙数量</p>
        <p class="value">{{ stats.buyWallCount }} / {{ formatNumber(stats.buyWallQty) }}</p>
      </div>
      <div class="item">
        <p class="label">卖单墙数量</p>
        <p class="value">{{ stats.sellWallCount }} / {{ formatNumber(stats.sellWallQty) }}</p>
      </div>
      <div class="item">
        <p class="label">买单金额</p>
        <p class="value">${{ formatNumber(stats.buyValue) }}</p>
      </div>
      <div class="item">
        <p class="label">卖单金额</p>
        <p class="value">${{ formatNumber(stats.sellValue) }}</p>
      </div>
    </div>
    <p v-else class="placeholder">等待数据...</p>
  </section>
</template>

<script setup lang="ts">
import { computed } from "vue";

import { useOrderStore } from "@/stores/order";

const orderStore = useOrderStore();
const stats = computed(() => orderStore.stats);

const badgeClass = computed(() => {
  if (!stats.value) return "badge";
  if (stats.value.buyWallQty > stats.value.sellWallQty * 1.5) return "badge buy";
  if (stats.value.sellWallQty > stats.value.buyWallQty * 1.5) return "badge sell";
  return "badge";
});

const badgeText = computed(() => {
  if (!stats.value) return "⚖️ 等待中";
  if (stats.value.buyWallQty > stats.value.sellWallQty * 1.5) return "🔴 买盘占优";
  if (stats.value.sellWallQty > stats.value.buyWallQty * 1.5) return "🟢 卖盘占优";
  return "⚖️ 基本均衡";
});

function formatNumber(value: number) {
  return value.toLocaleString(undefined, { maximumFractionDigits: 2 });
}
</script>

<style scoped>
.header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 16px;
}

.badge {
  display: inline-flex;
  align-items: center;
  gap: 6px;
  padding: 4px 10px;
  border-radius: 999px;
  background: #202a3b;
  border: 1px solid #324565;
  font-size: 12px;
}

.badge.buy {
  background: #3b1720;
  border-color: #7a2434;
  color: #ff9aa5;
}

.badge.sell {
  background: #14351c;
  border-color: #1e6a38;
  color: #76e0a6;
}

.grid {
  display: grid;
  grid-template-columns: repeat(2, minmax(0, 1fr));
  gap: 16px;
}

.label {
  font-size: 12px;
  color: #9fb0cc;
}

.value {
  font-size: 20px;
  font-weight: 600;
}

.placeholder {
  font-size: 13px;
  color: #9fb0cc;
}
</style>

