<template>
  <section>
    <header class="header">
      <h3>15 分钟买卖对比</h3>
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
.header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 12px;
}

.hint {
  font-size: 12px;
  color: #9fb0cc;
}

.chart {
  background: #0f1626;
  border: 1px solid #1d2a44;
  border-radius: 10px;
  padding: 16px;
  margin-bottom: 16px;
}

.bar {
  display: flex;
  height: 42px;
  border-radius: 8px;
  overflow: hidden;
  background: #1b2338;
}

.segment {
  display: flex;
  align-items: center;
  justify-content: center;
  font-weight: 600;
  font-size: 14px;
  color: #fff;
}

.segment.buy {
  background: linear-gradient(90deg, #7a2434, #ff9aa5);
}

.segment.sell {
  background: linear-gradient(90deg, #1e6a38, #76e0a6);
}

.legend {
  display: flex;
  justify-content: space-between;
  margin-top: 12px;
  font-size: 13px;
  color: #9fb0cc;
}

.placeholder {
  font-size: 13px;
  color: #9fb0cc;
  margin-bottom: 16px;
}

.others {
  display: grid;
  gap: 12px;
  grid-template-columns: repeat(auto-fit, minmax(180px, 1fr));
}

.card {
  background: #0f1626;
  border: 1px solid #1d2a44;
  border-radius: 8px;
  padding: 12px;
  color: #e6e9ef;
}

.card header {
  font-weight: 600;
  margin-bottom: 6px;
}

.card p {
  margin: 0;
  font-size: 13px;
  color: #9fb0cc;
}
</style>

