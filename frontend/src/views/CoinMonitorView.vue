<template>
  <div class="layout">
    <div class="toolbar">
      <div class="coin-info">
        <h1>{{ coinConfig?.displayName }} 大额订单监控</h1>
        <span class="market-type">U 本位合约</span>
      </div>
      <div class="actions">
        <button class="toggle" @click="toggleConfig">{{ showConfig ? "关闭配置" : "显示配置" }}</button>
        <span class="status">{{ statusMessage }}</span>
      </div>
    </div>

    <WindowSummary />

    <transition name="drawer">
      <aside v-if="showConfig" class="drawer">
        <header class="drawer-header">
          <h2>连接配置</h2>
          <button class="close" @click="toggleConfig">×</button>
        </header>
        <ConfigPanel />
      </aside>
    </transition>
  </div>
</template>

<script setup lang="ts">
import { onMounted, ref, watch } from "vue";

import ConfigPanel from "@/components/ConfigPanel.vue";
import WindowSummary from "@/components/WindowSummary.vue";
import { getCoinConfig } from "@/config/coins";
import { useWebSocket } from "@/composables/useWebSocket";
import { useConfigStore } from "@/stores/config";

const props = defineProps<{
  coinId: string;
}>();

const showConfig = ref(false);
const configStore = useConfigStore();
const { connect, disconnect, statusMessage } = useWebSocket();

const coinConfig = ref(getCoinConfig(props.coinId));

function initCoinConfig() {
  const config = getCoinConfig(props.coinId);
  if (config) {
    coinConfig.value = config;
    configStore.symbol = config.symbol;
    configStore.marketType = config.marketType;
    configStore.threshold = config.defaultThreshold;
    configStore.depth = 20;
    configStore.thresholdOp = "gt";
  }
}

function toggleConfig() {
  showConfig.value = !showConfig.value;
}

// 监听 coinId 变化，重新初始化配置
watch(() => props.coinId, () => {
  disconnect();
  initCoinConfig();
  connect();
}, { immediate: false });

onMounted(() => {
  initCoinConfig();
  connect();
});
</script>

<style scoped>
.layout {
  position: relative;
  padding: 16px;
  max-width: 960px;
  margin: 0 auto;
  display: flex;
  flex-direction: column;
  gap: 24px;
}

.toolbar {
  display: flex;
  justify-content: space-between;
  align-items: center;
  flex-wrap: wrap;
  gap: 16px;
}

.coin-info {
  display: flex;
  align-items: center;
  gap: 12px;
}

.coin-info h1 {
  margin: 0;
  font-size: 24px;
  font-weight: 600;
  color: #e6e9ef;
}

.market-type {
  padding: 4px 10px;
  border-radius: 4px;
  background: #1b59b0;
  color: #ffffff;
  font-size: 12px;
  font-weight: 500;
}

.actions {
  display: flex;
  align-items: center;
  gap: 16px;
}

.toggle {
  padding: 10px 18px;
  border-radius: 999px;
  border: 1px solid #1d2a44;
  background: #0f1626;
  color: #9bd1ff;
  cursor: pointer;
  font-size: 13px;
  transition: all 0.2s ease;
}

.toggle:hover {
  background: #1b2338;
  border-color: #324565;
}

.status {
  font-size: 13px;
  color: #9fb0cc;
}

.drawer-enter-active,
.drawer-leave-active {
  transition: all 0.2s ease;
}

.drawer-enter-from,
.drawer-leave-to {
  opacity: 0;
  transform: translateY(-8px);
}

.drawer {
  position: fixed;
  top: 80px;
  right: 24px;
  width: 360px;
  background: #121a2a;
  border: 1px solid #1d2a44;
  border-radius: 12px;
  padding: 18px;
  box-shadow: 0 16px 32px rgba(0, 0, 0, 0.35);
  z-index: 20;
}

.drawer-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 12px;
}

.drawer-header h2 {
  margin: 0;
  font-size: 16px;
  color: #e6e9ef;
}

.close {
  background: none;
  border: none;
  color: #9fb0cc;
  font-size: 24px;
  cursor: pointer;
  padding: 0;
  width: 24px;
  height: 24px;
  display: flex;
  align-items: center;
  justify-content: center;
}

.close:hover {
  color: #e6e9ef;
}
</style>

