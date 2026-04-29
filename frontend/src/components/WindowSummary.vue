<template>
  <section class="window-panel">
    <header class="header">
      <div>
        <h3>15 分钟买卖对比</h3>
        <p class="subhead">用短窗口观察短时主动成交偏向与放量方向。</p>
      </div>
      <span v-if="window15m" class="hint">阈值累积 {{ formatNumber(window15m.buyQty + window15m.sellQty) }}</span>
    </header>

    <div class="chart" v-if="window15m">
      <div class="bar">
        <div class="segment buy" :style="{ width: buyPercent + '%' }">
          <span>买 {{ buyPercent.toFixed(1) }}%</span>
        </div>
        <div class="segment sell" :style="{ width: sellPercent + '%' }">
          <span>卖 {{ sellPercent.toFixed(1) }}%</span>
        </div>
      </div>
      <div class="legend">
        <span>买单量 {{ formatNumber(window15m.buyQty) }} · {{ window15m.buyCnt }} 笔</span>
        <span>卖单量 {{ formatNumber(window15m.sellQty) }} · {{ window15m.sellCnt }} 笔</span>
      </div>
    </div>
    <p v-else class="placeholder">暂无 15 分钟数据</p>

    <div class="others" v-if="otherWindows.length">
      <article v-for="item in otherWindows" :key="item.window" class="card">
        <header>{{ item.window }}</header>
        <p>买：{{ formatNumber(item.buyQty) }} · {{ item.buyCnt }} 笔</p>
        <p>卖：{{ formatNumber(item.sellQty) }} · {{ item.sellCnt }} 笔</p>
      </article>
    </div>
  </section>
</template>

<script setup lang="ts">
import { computed } from "vue";
import { useOrderStore } from "@/stores/order";

const orderStore = useOrderStore();
const aggregates = computed(() => orderStore.aggregates);

const window15m = computed(() => aggregates.value.find((item) => item.window === "15m"));

const orderedWindows = ["30m", "1h", "4h"];
const otherWindows = computed(() =>
  orderedWindows
    .map((label) => aggregates.value.find((item) => item.window === label))
    .filter((item): item is typeof aggregates.value[number] => Boolean(item))
);

const buyPercent = computed(() => {
  if (!window15m.value) return 50;
  const total = window15m.value.buyQty + window15m.value.sellQty;
  if (total <= 0) return 50;
  return (window15m.value.buyQty / total) * 100;
});

const sellPercent = computed(() => 100 - buyPercent.value);

function formatNumber(value: number) {
  return value.toLocaleString(undefined, { maximumFractionDigits: 2 });
}
</script>

<style scoped>
.window-panel {
  padding: 22px;
  border-radius: var(--app-radius-lg);
  border: 1px solid var(--app-border-soft);
  background: var(--app-panel-bg);
  box-shadow: var(--app-shell-shadow), inset 0 1px 0 rgba(255, 255, 255, 0.04);
}

.header {
  display: flex;
  justify-content: space-between;
  align-items: flex-end;
  gap: 16px;
  margin-bottom: 18px;
}

.header h3 {
  margin: 0;
  font-family: Georgia, "Times New Roman", "Songti SC", serif;
  font-size: 22px;
  font-weight: 500;
  letter-spacing: -0.02em;
}

.subhead {
  margin: 6px 0 0;
  font-size: 13px;
  color: var(--app-text-muted);
}

.hint {
  display: inline-flex;
  align-items: center;
  min-height: 34px;
  padding: 0 12px;
  border-radius: 999px;
  background: var(--app-chip-bg);
  border: 1px solid var(--app-border);
  font-size: 11px;
  color: var(--app-text-muted);
}

.chart {
  background: var(--app-card-bg);
  border: 1px solid var(--app-border);
  border-radius: var(--app-radius-md);
  padding: 16px;
  margin-bottom: 18px;
}

.bar {
  display: flex;
  height: 42px;
  border-radius: 12px;
  overflow: hidden;
  background: var(--app-bar-track);
}

.segment {
  display: flex;
  align-items: center;
  justify-content: center;
  font-weight: 600;
  font-size: 14px;
  letter-spacing: -0.02em;
  color: #fff;
}

.segment.buy {
  background: linear-gradient(90deg, var(--app-buy-strong), var(--app-buy) 58%, var(--app-buy-soft));
}

.segment.sell {
  background: linear-gradient(90deg, var(--app-sell-strong), var(--app-sell) 58%, var(--app-sell-soft));
}

.legend {
  display: flex;
  justify-content: space-between;
  gap: 16px;
  margin-top: 14px;
  font-size: 13px;
  color: var(--app-text-muted);
}

.placeholder {
  font-size: 14px;
  color: var(--app-text-muted);
  margin-bottom: 16px;
}

.others {
  display: grid;
  gap: 12px;
  grid-template-columns: repeat(auto-fit, minmax(180px, 1fr));
}

.card {
  background: var(--app-card-bg);
  border: 1px solid var(--app-border);
  border-radius: var(--app-radius-md);
  padding: 14px;
  color: var(--app-text);
}

.card header {
  margin-bottom: 10px;
  font-family: Georgia, "Times New Roman", "Songti SC", serif;
  font-size: 20px;
  line-height: 1;
  font-weight: 500;
  letter-spacing: -0.02em;
}

.card p {
  margin: 6px 0 0;
  font-size: 13px;
  color: var(--app-text-muted);
}

@media (max-width: 768px) {
  .window-panel {
    padding: 18px;
    border-radius: 18px;
  }

  .header {
    align-items: flex-start;
    flex-direction: column;
  }

  .header h3 {
    font-size: 20px;
  }

  .subhead {
    font-size: 13px;
  }

  .bar {
    height: 40px;
  }

  .segment {
    font-size: 13px;
  }

  .legend {
    flex-direction: column;
    font-size: 14px;
  }
}
</style>
