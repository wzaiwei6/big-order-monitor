<template>
  <div class="layout">
    <section class="hero">
      <div class="coin-info">
        <div class="hero-title">
          <h1>{{ coinConfig?.displayName }} 大额订单监控</h1>
          <span class="market-type">U 本位合约</span>
        </div>
      </div>
      <div class="actions">
        <div class="time-card" v-if="lastUpdated">
          <span class="time-label">最近更新</span>
          <span class="timestamp">{{ formatTime(lastUpdated) }}</span>
        </div>
      </div>
    </section>

    <section v-if="error" class="error">
      <span>{{ error }}</span>
    </section>

    <WindowSummary />
    <RealtimeStats />
    <OrderHistory />
  </div>
</template>

<script setup lang="ts">
import { onMounted, onUnmounted, ref, watch } from "vue";

import OrderHistory from "@/components/OrderHistory.vue";
import RealtimeStats from "@/components/RealtimeStats.vue";
import WindowSummary from "@/components/WindowSummary.vue";
import { getCoinConfig } from "@/config/coins";
import { useOrderStore } from "@/stores/order";

const props = defineProps<{
  coinId: string;
}>();

const coinConfig = ref(getCoinConfig(props.coinId));
const orderStore = useOrderStore();
const error = ref("");
const lastUpdated = ref<number | null>(null);
let timer: number | null = null;

function initCoinConfig() {
  const config = getCoinConfig(props.coinId);
  if (config) {
    coinConfig.value = config;
  }
}

async function loadMonitor() {
  error.value = "";

  try {
    const response = await fetch(`${import.meta.env.VITE_API_BASE_URL || ""}/api/monitor/${props.coinId}`, {
      headers: { Accept: "application/json" }
    });
    if (!response.ok) {
      throw new Error(`请求失败: ${response.status}`);
    }

    const payload = await response.json();
    orderStore.setStats(payload.stats ?? null);
    orderStore.setAggregates(payload.windows ?? []);
    orderStore.setHistory(
      (payload.orders ?? []).map((order: any) => ({
        id: `${order.side}-${order.price}-${order.filledTime}`,
        side: order.side,
        price: order.price,
        quantity: order.quantity,
        firstSeen: order.firstSeen,
        filledTime: order.filledTime,
        durationSec: order.durationSec
      }))
    );

    lastUpdated.value = payload.generatedAt ? payload.generatedAt * 1000 : Date.now();
    orderStore.setStatus("实时快照");
  } catch (err) {
    const message = err instanceof Error ? err.message : "未知错误";
    error.value = `加载失败：${message}`;
    orderStore.setStatus(error.value);
  }
}

function startPolling() {
  if (timer) {
    window.clearInterval(timer);
  }

  void loadMonitor();
  timer = window.setInterval(() => {
    void loadMonitor();
  }, 1000);
}

function formatTime(timestamp: number) {
  return new Date(timestamp).toLocaleTimeString();
}

watch(() => props.coinId, () => {
  initCoinConfig();
  orderStore.setStats(null);
  orderStore.setAggregates([]);
  orderStore.setHistory([]);
  startPolling();
}, { immediate: false });

onMounted(() => {
  initCoinConfig();
  startPolling();
});

onUnmounted(() => {
  if (timer) {
    window.clearInterval(timer);
    timer = null;
  }
});
</script>

<style scoped>
.layout {
  position: relative;
  width: 100%;
  margin: 0 auto;
  display: flex;
  flex-direction: column;
  gap: 20px;
}

.hero {
  display: flex;
  justify-content: space-between;
  align-items: flex-start;
  flex-wrap: wrap;
  gap: 18px;
  padding: 16px 18px;
  border-radius: var(--app-radius-lg);
  border: 1px solid var(--app-border-soft);
  background: var(--app-header-bg);
  box-shadow: var(--app-shell-shadow), inset 0 1px 0 rgba(255, 255, 255, 0.04);
}

.coin-info {
  display: flex;
  flex-direction: column;
  gap: 0;
  max-width: 620px;
}

.hero-title {
  display: flex;
  align-items: center;
  gap: 10px;
  flex-wrap: wrap;
}

.hero-title h1 {
  margin: 0;
  font-family: Georgia, "Times New Roman", "Songti SC", serif;
  font-size: clamp(18px, 2.2vw, 24px);
  line-height: 1.08;
  font-weight: 500;
  letter-spacing: -0.03em;
  color: var(--app-text);
}

.market-type {
  padding: 4px 8px;
  border-radius: 999px;
  background: color-mix(in srgb, var(--app-accent) 14%, transparent);
  color: var(--app-accent-soft);
  border: 1px solid color-mix(in srgb, var(--app-accent) 22%, transparent);
  font-size: 10px;
  font-weight: 600;
  letter-spacing: 0.03em;
}

.actions {
  display: flex;
  justify-content: flex-end;
  flex: 1;
  min-width: 132px;
}

.time-card {
  display: flex;
  flex-direction: column;
  align-items: flex-end;
  gap: 4px;
  padding: 8px 10px;
  min-width: 118px;
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

.error {
  padding: 14px 16px;
  background: rgba(181, 51, 51, 0.14);
  border: 1px solid rgba(181, 51, 51, 0.24);
  color: #f2c6c6;
  border-radius: var(--app-radius-md);
}

@media (max-width: 768px) {
  .hero {
    padding: 14px 12px;
    border-radius: 18px;
  }

  .actions {
    width: 100%;
    justify-content: flex-start;
  }

  .time-card {
    align-items: flex-start;
  }

  .timestamp {
    font-size: 12px;
  }
}
</style>
