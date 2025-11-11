<template>
  <div class="summary-layout">
    <header class="summary-header">
      <div class="title-group">
        <h1>15 分钟多币种汇总</h1>
        <p class="subtitle">查看 7 个监控币种最近 15 分钟的买卖成交概要</p>
      </div>
      <div class="actions">
        <button class="refresh" :disabled="loading" @click="loadSummary">
          {{ loading ? "加载中..." : "刷新" }}
        </button>
        <span class="timestamp" v-if="lastUpdated">
          更新于 {{ formatTime(lastUpdated) }}
        </span>
      </div>
    </header>

    <section v-if="error" class="error">
      <span>⚠️ {{ error }}</span>
    </section>

    <section v-if="enrichedEntries.length" class="summary-list">
      <article v-for="item in enrichedEntries" :key="item.coinId" class="summary-row">
        <header class="row-header">
          <div class="coin-info">
            <h2>{{ item.displayName }}</h2>
            <span class="symbol">{{ item.symbol.toUpperCase() }}</span>
          </div>
          <div class="totals">
            <span>总买：{{ formatNumber(item.buyQty) }} · {{ item.buyCnt }} 笔</span>
            <span>总卖：{{ formatNumber(item.sellQty) }} · {{ item.sellCnt }} 笔</span>
          </div>
        </header>
        <div class="bar">
          <div class="segment buy" :style="{ width: item.buyPercent + '%' }">
            <span>{{ item.buyPercent.toFixed(1) }}% 买</span>
          </div>
          <div class="segment sell" :style="{ width: item.sellPercent + '%' }">
            <span>{{ item.sellPercent.toFixed(1) }}% 卖</span>
          </div>
        </div>
      </article>
    </section>

    <section v-else-if="!loading" class="empty">
      <p>暂无数据</p>
    </section>
  </div>
</template>

<script setup lang="ts">
import { computed, onMounted, onUnmounted, ref } from "vue";

interface SummaryEntry {
  coinId: string;
  symbol: string;
  displayName: string;
  buyQty: number;
  sellQty: number;
  buyCnt: number;
  sellCnt: number;
}

interface SummaryResponse {
  generatedAt: number;
  window: string;
  coins: SummaryEntry[];
}

const entries = ref<SummaryEntry[]>([]);
const lastUpdated = ref<number | null>(null);
const loading = ref(false);
const error = ref("");
let timer: number | null = null;

const enrichedEntries = computed(() =>
  entries.value.map((item) => {
    const total = item.buyQty + item.sellQty;
    const buyPercent = total > 0 ? (item.buyQty / total) * 100 : 50;
    const sellPercent = 100 - buyPercent;
    return {
      ...item,
      buyPercent,
      sellPercent
    };
  })
);

const API_BASE = import.meta.env.VITE_API_BASE_URL || "";

async function loadSummary() {
  if (loading.value) return;

  loading.value = true;
  error.value = "";

  try {
    const resp = await fetch(`${API_BASE}/api/summary/15m`, {
      headers: { Accept: "application/json" }
    });
    if (!resp.ok) {
      throw new Error(`请求失败: ${resp.status}`);
    }
    const data = (await resp.json()) as SummaryResponse;
    entries.value = data.coins ?? [];
    lastUpdated.value = data.generatedAt * 1000;
  } catch (err) {
    console.error("加载汇总失败", err);
    error.value = err instanceof Error ? err.message : "未知错误";
  } finally {
    loading.value = false;
  }
}

function formatNumber(value: number) {
  return value.toLocaleString(undefined, { maximumFractionDigits: 2 });
}

function formatTime(timestamp: number) {
  return new Date(timestamp).toLocaleTimeString();
}

onMounted(() => {
  loadSummary();
  timer = window.setInterval(loadSummary, 30000);
});

onUnmounted(() => {
  if (timer) {
    window.clearInterval(timer);
    timer = null;
  }
});
</script>

<style scoped>
.summary-layout {
  max-width: 960px;
  margin: 0 auto;
  padding: 16px;
  display: flex;
  flex-direction: column;
  gap: 24px;
}

.summary-header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 16px;
  flex-wrap: wrap;
}

.title-group h1 {
  margin: 0;
  font-size: 26px;
  color: #e6e9ef;
}

.subtitle {
  margin: 4px 0 0;
  color: #9fb0cc;
  font-size: 14px;
}

.actions {
  display: flex;
  align-items: center;
  gap: 12px;
}

.refresh {
  padding: 10px 18px;
  border-radius: 999px;
  border: 1px solid #1d2a44;
  background: #1b59b0;
  color: #ffffff;
  cursor: pointer;
  font-size: 13px;
  transition: opacity 0.2s ease;
}

.refresh:disabled {
  opacity: 0.7;
  cursor: not-allowed;
}

.timestamp {
  font-size: 13px;
  color: #9fb0cc;
}

.error {
  padding: 14px 16px;
  background: rgba(255, 143, 143, 0.1);
  border: 1px solid #ff8f8f;
  color: #ffbcbc;
  border-radius: 10px;
}

.summary-list {
  display: flex;
  flex-direction: column;
  gap: 18px;
}

.summary-row {
  background: #0f1626;
  border: 1px solid #1d2a44;
  border-radius: 12px;
  padding: 18px;
  display: flex;
  flex-direction: column;
  gap: 14px;
}

.row-header {
  display: flex;
  justify-content: space-between;
  flex-wrap: wrap;
  gap: 12px;
  color: #e6e9ef;
}

.coin-info {
  display: flex;
  align-items: baseline;
  gap: 10px;
}

.coin-info h2 {
  margin: 0;
  font-size: 20px;
}

.symbol {
  font-size: 12px;
  color: #9fb0cc;
  letter-spacing: 1px;
}

.totals {
  display: flex;
  gap: 16px;
  font-size: 13px;
  color: #9fb0cc;
}

.bar {
  display: flex;
  height: 44px;
  border-radius: 10px;
  overflow: hidden;
  background: #1b2338;
}

.segment {
  display: flex;
  align-items: center;
  justify-content: center;
  font-weight: 600;
  font-size: 14px;
  color: #ffffff;
}

.segment.buy {
  background: linear-gradient(90deg, #7a2434, #ff9aa5);
}

.segment.sell {
  background: linear-gradient(90deg, #1e6a38, #76e0a6);
}

.empty {
  padding: 40px 0;
  text-align: center;
  color: #9fb0cc;
  font-size: 14px;
}
</style>
