<template>
  <div class="layout">
    <section v-if="error" class="error">
      <span>{{ error }}</span>
    </section>

    <WindowSummary
      :display-name="coinConfig?.displayName ?? props.coinId.toUpperCase()"
      :threshold="coinConfig?.defaultThreshold ?? 0"
      :threshold-unit="coinConfig?.displayName ?? props.coinId.toUpperCase()"
      :last-updated="lastUpdated"
    />
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
import { POLL_INTERVAL_MS } from "@/config/polling";
import { useOrderStore } from "@/stores/order";

const props = defineProps<{
  coinId: string;
}>();

const coinConfig = ref(getCoinConfig(props.coinId));
const orderStore = useOrderStore();
const error = ref("");
const lastUpdated = ref<number | null>(null);
let timer: number | null = null;
let abortController: AbortController | null = null;
let requestSeq = 0;
let loading = false;

function initCoinConfig() {
  const config = getCoinConfig(props.coinId);
  if (config) {
    coinConfig.value = config;
  }
}

async function loadMonitor() {
  if (loading) return;

  loading = true;
  const currentRequest = ++requestSeq;
  abortController = new AbortController();
  error.value = "";

  try {
    const response = await fetch(`${import.meta.env.VITE_API_BASE_URL || ""}/api/monitor/${props.coinId}`, {
      headers: { Accept: "application/json" },
      signal: abortController.signal
    });
    if (!response.ok) {
      throw new Error(`请求失败: ${response.status}`);
    }

    const payload = await response.json();
    if (currentRequest !== requestSeq) {
      return;
    }

    const nextStats = payload.stats ?? null;
    const nextAggregates = payload.windows ?? [];
    const nextHistory = (payload.orders ?? []).map((order: any) => ({
      id: `${order.side}-${order.price}-${order.filledTime}`,
      side: order.side,
      price: order.price,
      quantity: order.quantity,
      firstSeen: order.firstSeen,
      filledTime: order.filledTime,
      durationSec: order.durationSec
    }));
    const nextLastUpdated = payload.generatedAt ? payload.generatedAt * 1000 : Date.now();

    orderStore.setCoinSnapshot(props.coinId, {
      aggregates: nextAggregates,
      history: nextHistory,
      stats: nextStats,
      lastUpdated: nextLastUpdated
    });
    lastUpdated.value = nextLastUpdated;
    orderStore.setStatus("实时快照");
  } catch (err) {
    if (err instanceof DOMException && err.name === "AbortError") {
      return;
    }
    const message = err instanceof Error ? err.message : "未知错误";
    error.value = `加载失败：${message}`;
    orderStore.setStatus(error.value);
  } finally {
    if (currentRequest === requestSeq) {
      loading = false;
      abortController = null;
    }
  }
}

function startPolling() {
  if (timer) {
    window.clearInterval(timer);
  }

  const cachedSnapshot = orderStore.restoreCoinSnapshot(props.coinId);
  lastUpdated.value = cachedSnapshot?.lastUpdated ?? null;
  void loadMonitor();
  timer = window.setInterval(() => {
    void loadMonitor();
  }, POLL_INTERVAL_MS);
}

watch(() => props.coinId, () => {
  abortController?.abort();
  abortController = null;
  loading = false;
  initCoinConfig();
  startPolling();
}, { immediate: false });

onMounted(() => {
  initCoinConfig();
  startPolling();
});

onUnmounted(() => {
  abortController?.abort();
  abortController = null;
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

.error {
  padding: 14px 16px;
  background: rgba(181, 51, 51, 0.14);
  border: 1px solid rgba(181, 51, 51, 0.24);
  color: #f2c6c6;
  border-radius: var(--app-radius-md);
}

@media (max-width: 768px) {
}
</style>
