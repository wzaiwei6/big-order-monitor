<template>
  <section class="window-panel">
    <header class="panel-top">
      <div class="monitor-row">
        <h2>{{ displayName }} 大额订单监控</h2>
        <span class="market-type">U 本位合约</span>
        <span class="threshold-chip">监控阈值 ≥ {{ formatThreshold(threshold) }} {{ thresholdUnit }}</span>
        <div v-if="lastUpdated" class="time-card">
          <span class="time-label">最近更新</span>
          <span class="timestamp">{{ formatTime(lastUpdated) }}</span>
        </div>
      </div>
    </header>

    <header class="header">
      <div>
        <h3>15 分钟买卖对比</h3>
        <p class="subhead">用短窗口观察短时主动成交偏向与放量方向。</p>
      </div>
      <span v-if="window15m" class="hint">阈值累积 {{ formatNumber(window15m.buyQty + window15m.sellQty) }}</span>
    </header>

    <div class="chart" v-if="window15m">
      <div class="bar">
        <div v-if="buyPercent > 0" class="segment buy" :style="{ width: buyPercent + '%' }">
          <span>买 {{ buyPercent.toFixed(1) }}%</span>
        </div>
        <div v-if="sellPercent > 0" class="segment sell" :style="{ width: sellPercent + '%' }">
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

const props = defineProps<{
  displayName: string;
  threshold: number;
  thresholdUnit: string;
  lastUpdated: number | null;
}>();

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

function formatThreshold(value: number) {
  return value.toLocaleString(undefined, { maximumFractionDigits: 2 });
}

function formatTime(timestamp: number) {
  return new Date(timestamp).toLocaleTimeString();
}
</script>

<style scoped>
.window-panel {
  min-height: 360px;
  padding: 22px;
  border-radius: var(--app-radius-lg);
  border: 1px solid var(--app-border-soft);
  background: var(--app-panel-bg);
  box-shadow: var(--app-shell-shadow), inset 0 1px 0 rgba(255, 255, 255, 0.04);
}

.panel-top {
  margin-bottom: 22px;
  padding-bottom: 20px;
  border-bottom: 1px solid var(--app-border-soft);
}

.monitor-row {
  display: flex;
  align-items: center;
  gap: 8px;
  flex-wrap: wrap;
  width: 100%;
}

.monitor-row h2 {
  margin: 0;
  font-family: Georgia, "Times New Roman", "Songti SC", serif;
  font-size: 26px;
  line-height: 1.1;
  font-weight: 500;
  letter-spacing: 0;
  color: var(--app-text);
}

.market-type {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  height: 26px;
  padding: 0 10px;
  border-radius: 999px;
  background: color-mix(in srgb, var(--app-accent) 14%, transparent);
  color: var(--app-accent-soft);
  border: 1px solid color-mix(in srgb, var(--app-accent) 22%, transparent);
  font-size: 11px;
  line-height: 1;
  font-weight: 600;
  letter-spacing: 0.03em;
}

.threshold-chip {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  height: 26px;
  padding: 0 10px;
  border-radius: 999px;
  background: color-mix(in srgb, var(--app-accent) 14%, transparent);
  border: 1px solid color-mix(in srgb, var(--app-accent) 22%, transparent);
  font-size: 11px;
  line-height: 1;
  font-weight: 600;
  letter-spacing: 0.03em;
  color: var(--app-accent-soft);
}

.time-card {
  margin-left: auto;
  display: flex;
  flex-direction: column;
  align-items: flex-end;
  gap: 2px;
  padding: 8px 10px;
  min-width: 0;
  border-radius: var(--app-radius-md);
  border: 1px solid var(--app-border);
  background: var(--app-chip-bg);
}

.time-label {
  font-size: 9px;
  letter-spacing: 0.08em;
  text-transform: uppercase;
  color: var(--app-text-soft);
}

.timestamp {
  font-size: 13px;
  line-height: 1;
  font-weight: 600;
  letter-spacing: -0.02em;
  color: var(--app-text);
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
  transition: width 420ms cubic-bezier(0.22, 1, 0.36, 1);
  will-change: width;
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
    min-height: 0;
    padding: 14px;
    border-radius: 14px;
  }

  .panel-top {
    margin-bottom: 14px;
    padding-bottom: 14px;
  }

  .monitor-row {
    align-items: flex-start;
    gap: 6px;
  }

  .monitor-row h2 {
    width: 100%;
    font-size: 20px;
  }

  .market-type,
  .threshold-chip {
    height: 24px;
    padding: 0 8px;
    font-size: 10px;
  }

  .time-card {
    margin-left: auto;
    padding: 6px 8px;
    border-radius: 12px;
    align-items: flex-end;
  }

  .time-label {
    font-size: 8px;
  }

  .timestamp {
    font-size: 12px;
  }

  .header {
    align-items: flex-start;
    flex-direction: column;
    gap: 8px;
    margin-bottom: 12px;
  }

  .header h3 {
    font-size: 19px;
  }

  .subhead {
    margin-top: 4px;
    font-size: 12px;
  }

  .hint {
    min-height: 28px;
    padding: 0 10px;
    font-size: 10px;
  }

  .chart {
    padding: 10px;
    margin-bottom: 12px;
    border-radius: 12px;
  }

  .bar {
    height: 34px;
    border-radius: 10px;
  }

  .segment {
    font-size: 12px;
  }

  .legend {
    flex-direction: column;
    gap: 6px;
    margin-top: 10px;
    font-size: 12px;
  }

  .others {
    grid-template-columns: 1fr;
    gap: 8px;
  }

  .card {
    padding: 10px 12px;
    border-radius: 12px;
  }

  .card header {
    margin-bottom: 8px;
    font-size: 17px;
  }

  .card p {
    font-size: 12px;
  }
}
</style>
