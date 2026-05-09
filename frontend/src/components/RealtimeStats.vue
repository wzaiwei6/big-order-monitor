<template>
  <section class="stats-panel">
    <header class="header">
      <h3>实时统计</h3>
      <span :class="badgeClass">{{ badgeText }}</span>
    </header>
    <div v-if="stats" class="grid">
      <div class="item">
        <p class="label">买单墙数量</p>
        <p class="value">{{ stats.buyWallCount }} <span>/ {{ formatNumber(stats.buyWallQty) }}</span></p>
        <p class="caption">观察当前大额买单堆积规模</p>
      </div>
      <div class="item">
        <p class="label">卖单墙数量</p>
        <p class="value">{{ stats.sellWallCount }} <span>/ {{ formatNumber(stats.sellWallQty) }}</span></p>
        <p class="caption">观察当前大额卖单堆积规模</p>
      </div>
      <div class="item">
        <p class="label">买单金额</p>
        <p class="value">${{ formatNumber(stats.buyValue) }}</p>
        <p class="caption">按照盘口价格估算的买墙名义金额</p>
      </div>
      <div class="item">
        <p class="label">卖单金额</p>
        <p class="value">${{ formatNumber(stats.sellValue) }}</p>
        <p class="caption">按照盘口价格估算的卖墙名义金额</p>
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
  if (!stats.value) return "等待中";
  if (stats.value.buyWallQty > stats.value.sellWallQty * 1.5) return "买盘占优";
  if (stats.value.sellWallQty > stats.value.buyWallQty * 1.5) return "卖盘占优";
  return "基本均衡";
});

function formatNumber(value: number) {
  return value.toLocaleString(undefined, { maximumFractionDigits: 2 });
}
</script>

<style scoped>
.stats-panel {
  padding: 22px;
  border-radius: var(--app-radius-lg);
  border: 1px solid var(--app-border-soft);
  background: var(--app-panel-bg);
  box-shadow: var(--app-shell-shadow), inset 0 1px 0 rgba(255, 255, 255, 0.04);
}

.header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  gap: 12px;
  margin-bottom: 18px;
}

.header h3 {
  margin: 0;
  font-family: Georgia, "Times New Roman", "Songti SC", serif;
  font-size: 26px;
  font-weight: 500;
  letter-spacing: -0.02em;
}

.badge {
  display: inline-flex;
  align-items: center;
  gap: 8px;
  padding: 8px 12px;
  border-radius: 999px;
  background: var(--app-chip-bg);
  border: 1px solid var(--app-border);
  font-size: 13px;
  font-weight: 600;
  color: var(--app-text-muted);
}

.badge::before {
  content: "";
  width: 7px;
  height: 7px;
  border-radius: 50%;
  background: #8d877d;
}

.badge.buy {
  background: color-mix(in srgb, var(--app-buy) 14%, transparent);
  border-color: color-mix(in srgb, var(--app-buy) 30%, transparent);
  color: color-mix(in srgb, var(--app-buy) 40%, var(--app-text) 60%);
}

.badge.buy::before {
  background: var(--app-buy);
}

.badge.sell {
  background: color-mix(in srgb, var(--app-sell) 14%, transparent);
  border-color: color-mix(in srgb, var(--app-sell) 28%, transparent);
  color: color-mix(in srgb, var(--app-sell) 46%, var(--app-text) 54%);
}

.badge.sell::before {
  background: var(--app-sell);
}

.grid {
  display: grid;
  grid-template-columns: repeat(2, minmax(0, 1fr));
  gap: 14px;
}

.item {
  min-height: 134px;
  padding: 18px;
  border-radius: var(--app-radius-md);
  border: 1px solid var(--app-border);
  background: var(--app-card-bg);
}

.label {
  margin: 0 0 12px;
  font-size: 12px;
  letter-spacing: 0.08em;
  text-transform: uppercase;
  color: var(--app-text-soft);
}

.value {
  margin: 0;
  font-size: 28px;
  line-height: 1.15;
  font-weight: 600;
  letter-spacing: 0;
  color: var(--app-text);
}

.value span {
  font-size: 0.56em;
  color: var(--app-text-muted);
}

.caption {
  margin: 18px 0 0;
  max-width: 280px;
  font-size: 13px;
  line-height: 1.55;
  color: var(--app-text-muted);
}

.placeholder {
  margin: 0;
  font-size: 14px;
  color: var(--app-text-muted);
}

@media (max-width: 768px) {
  .stats-panel {
    padding: 14px;
    border-radius: 14px;
  }

  .header {
    align-items: center;
    margin-bottom: 12px;
  }

  .header h3 {
    font-size: 20px;
  }

  .badge {
    padding: 6px 10px;
    font-size: 11px;
  }

  .grid {
    grid-template-columns: repeat(2, minmax(0, 1fr));
    gap: 8px;
  }

  .item {
    min-height: 92px;
    padding: 12px;
    border-radius: 12px;
  }

  .label {
    margin-bottom: 8px;
    font-size: 10px;
    letter-spacing: 0.04em;
  }

  .value {
    font-size: 18px;
    line-height: 1.25;
  }

  .caption {
    display: none;
  }
}
</style>
