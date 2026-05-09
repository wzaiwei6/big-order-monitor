<template>
  <div class="summary-layout">
    <header class="summary-header">
      <div class="title-group">
        <h1>15 分钟多币种汇总</h1>
        <p class="subtitle">查看 7 个监控币种最近 15 分钟的买卖成交概要</p>
      </div>
      <div class="actions">
        <button class="refresh" @click="loadSummary">
          手动刷新
        </button>
        <span class="status" :class="{ error: statusType === 'error' }">
          <span v-if="statusLabel">{{ statusLabel }}</span>
        </span>
        <span class="timestamp" v-if="lastUpdated">更新于 {{ formatTime(lastUpdated) }}</span>
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
            <span class="threshold-chip">监控阈值 ≥ {{ formatNumber(item.threshold) }} {{ item.displayName }}</span>
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
import { getCoinConfig } from "@/config/coins";
import { POLL_INTERVAL_MS } from "@/config/polling";

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
const statusMessage = ref("后端采集器运行中...");
const statusType = ref<"normal" | "error">("normal");
let timer: number | null = null;

const statusLabel = computed(() => {
  if (statusType.value === "error") {
    return statusMessage.value;
  }
  return "";
});

const enrichedEntries = computed(() =>
  entries.value.map((item) => {
    const total = item.buyQty + item.sellQty;
    const buyPercent = total > 0 ? (item.buyQty / total) * 100 : 50;
    const sellPercent = 100 - buyPercent;
    return {
      ...item,
      threshold: getCoinConfig(item.coinId)?.defaultThreshold ?? 0,
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
    statusMessage.value = "";
    statusType.value = "normal";
  } catch (err) {
    console.error("加载汇总失败", err);
    error.value = err instanceof Error ? err.message : "未知错误";
    statusMessage.value = `拉取失败，稍后重试 (${error.value})`;
    statusType.value = "error";
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

onMounted(async () => {
  await loadSummary();
  timer = window.setInterval(loadSummary, POLL_INTERVAL_MS);
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
  width: 100%;
  margin: 0 auto;
  padding: 0;
  display: flex;
  flex-direction: column;
  gap: 16px;
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
  font-family: Georgia, "Times New Roman", "Songti SC", serif;
  font-size: 24px;
  font-weight: 500;
  letter-spacing: 0;
  color: var(--app-text);
}

.subtitle {
  margin: 4px 0 0;
  color: var(--app-text-muted);
  font-size: 14px;
}

.actions {
  display: flex;
  align-items: center;
  gap: 12px;
  flex-wrap: wrap;
}

.refresh {
  padding: 9px 16px;
  border-radius: 999px;
  border: 1px solid color-mix(in srgb, var(--app-accent) 28%, transparent);
  background: var(--app-accent);
  color: var(--app-accent-contrast);
  cursor: pointer;
  font-size: 13px;
  font-weight: 600;
  transition: background 0.2s ease, border-color 0.2s ease;
}

.refresh:hover {
  background: var(--app-accent-soft);
  border-color: color-mix(in srgb, var(--app-accent-soft) 34%, transparent);
}

.timestamp {
  font-size: 13px;
  color: var(--app-text-muted);
}

.status {
  display: inline-flex;
  align-items: center;
  gap: 8px;
  min-width: 7px;
  font-size: 13px;
  color: var(--app-text-muted);
}

.status::before {
  content: "";
  width: 7px;
  height: 7px;
  border-radius: 50%;
  background: var(--app-sell);
}

.status.error {
  color: #ffb0b0;
}

.status.error::before {
  background: #d96a74;
}

.error {
  padding: 14px 16px;
  background: rgba(181, 51, 51, 0.14);
  border: 1px solid rgba(181, 51, 51, 0.24);
  color: #f2c6c6;
  border-radius: var(--app-radius-md);
}

.summary-list {
  display: flex;
  flex-direction: column;
  gap: 12px;
}

.summary-row {
  background: var(--app-panel-bg);
  border: 1px solid var(--app-border);
  border-radius: var(--app-radius-lg);
  padding: 14px 18px;
  display: flex;
  flex-direction: column;
  gap: 10px;
  box-shadow: var(--app-shell-shadow), inset 0 1px 0 rgba(255, 255, 255, 0.04);
}

.row-header {
  display: grid;
  grid-template-columns: minmax(0, 1fr) auto;
  align-items: start;
  gap: 8px 18px;
  min-height: 28px;
  color: var(--app-text);
}

.coin-info {
  display: flex;
  align-items: center;
  gap: 10px;
  min-height: 28px;
  flex-wrap: wrap;
}

.coin-info h2 {
  margin: 0;
  font-size: 17px;
  letter-spacing: 0;
}

.symbol {
  font-size: 12px;
  color: var(--app-text-soft);
  letter-spacing: 1px;
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
  color: var(--app-accent-soft);
  font-size: 11px;
  line-height: 1;
  font-weight: 600;
  letter-spacing: 0.03em;
  white-space: nowrap;
}

.totals {
  display: grid;
  grid-auto-flow: column;
  gap: 16px;
  align-items: start;
  justify-content: end;
  min-height: 28px;
  font-size: 12px;
  color: var(--app-text-muted);
  text-align: right;
}

.totals span {
  white-space: nowrap;
}

.bar {
  display: flex;
  height: 38px;
  border-radius: 12px;
  overflow: hidden;
  background: var(--app-bar-track);
}

.segment {
  display: flex;
  align-items: center;
  justify-content: center;
  font-weight: 600;
  font-size: 13px;
  color: #ffffff;
  transition: width 420ms cubic-bezier(0.22, 1, 0.36, 1);
  will-change: width;
}

.segment.buy {
  background: linear-gradient(90deg, var(--app-buy-strong), var(--app-buy) 58%, var(--app-buy-soft));
}

.segment.sell {
  background: linear-gradient(90deg, var(--app-sell-strong), var(--app-sell) 58%, var(--app-sell-soft));
}

.empty {
  padding: 40px 0;
  text-align: center;
  color: var(--app-text-muted);
  font-size: 14px;
}

@media (max-width: 768px) {
  .summary-layout {
    gap: 12px;
  }

  .summary-header {
    align-items: flex-start;
    gap: 10px;
  }

  .title-group h1 {
    font-size: 20px;
  }

  .subtitle {
    font-size: 12px;
  }

  .actions {
    width: 100%;
    gap: 8px;
  }

  .refresh {
    min-height: 32px;
    padding: 0 12px;
    font-size: 12px;
  }

  .status,
  .timestamp {
    font-size: 12px;
  }

  .summary-list {
    gap: 10px;
  }

  .summary-row {
    padding: 12px;
    gap: 8px;
    border-radius: 14px;
  }

  .row-header {
    grid-template-columns: 1fr;
    gap: 8px;
  }

  .coin-info {
    gap: 6px;
    min-height: 26px;
  }

  .coin-info h2 {
    font-size: 16px;
  }

  .symbol {
    font-size: 10px;
  }

  .threshold-chip {
    height: 24px;
    padding: 0 8px;
    font-size: 10px;
  }

  .totals {
    gap: 8px;
    grid-auto-flow: row;
    justify-content: start;
    min-height: 0;
    font-size: 11px;
    text-align: left;
  }

  .bar {
    height: 32px;
    border-radius: 10px;
  }

  .segment {
    font-size: 12px;
  }
}
</style>
