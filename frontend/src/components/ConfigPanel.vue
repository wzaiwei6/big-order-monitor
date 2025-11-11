<template>
  <div class="config">
    <form class="form" @submit.prevent="handleSubmit">
      <div class="actions">
        <button type="submit" class="btn primary">连接</button>
        <button type="button" class="btn" @click="disconnect">断开</button>
      </div>
      <label>
        <span>交易对</span>
        <input v-model="config.symbol" placeholder="btcusdt" readonly />
      </label>

      <div class="info-row">
        <span class="label">市场类型</span>
        <span class="value">U 本位合约 (USDTM)</span>
      </div>

      <label>
        <span>深度档位</span>
        <select v-model.number="config.depth">
          <option :value="5">5</option>
          <option :value="10">10</option>
          <option :value="20">20</option>
          <option :value="100">100</option>
        </select>
      </label>

      <label>
        <span>阈值比较符</span>
        <select v-model="config.thresholdOp">
          <option value="gt">大于</option>
          <option value="lt">小于</option>
        </select>
      </label>

      <label>
        <span>阈值</span>
        <input v-model.number="config.threshold" type="number" min="0" step="0.1" />
      </label>

      <label>
        <span>自定义网关</span>
        <input v-model="config.gateway" placeholder="ws://127.0.0.1:8765/ws" />
      </label>

      <label>
        <span>网关代理</span>
        <input v-model="config.gatewayProxy" placeholder="http://127.0.0.1:7890" />
      </label>

      <label class="checkbox">
        <input v-model="config.autoReconnect" type="checkbox" />
        <span>自动重连</span>
      </label>



      <p class="status">{{ statusMessage }}</p>
    </form>
  </div>
</template>

<script setup lang="ts">
import { computed } from "vue";

import { useConfigStore } from "@/stores/config";
import { useWebSocket } from "@/composables/useWebSocket";

const configStore = useConfigStore();
const config = computed({
  get: () => configStore.$state,
  set: (value) => configStore.$patch(value)
});

const { connect, disconnect, statusMessage } = useWebSocket();

function handleSubmit() {
  connect();
}
</script>

<style scoped>
.config {
  display: flex;
  flex-direction: column;
  gap: 16px;
}

.form {
  display: flex;
  flex-direction: column;
  gap: 14px;
}

label {
  display: flex;
  flex-direction: column;
  gap: 6px;
  font-size: 13px;
  color: #9fb0cc;
}

input,
select {
  background: #0f1626;
  color: #e6e9ef;
  border: 1px solid #22314f;
  border-radius: 6px;
  padding: 10px 12px;
}

.checkbox {
  flex-direction: row;
  align-items: center;
  gap: 10px;
}

.actions {
  display: flex;
  gap: 10px;
}

.btn {
  flex: 1;
  padding: 10px 16px;
  border-radius: 6px;
  border: 1px solid #22314f;
  background: #1d2a44;
  color: #e6e9ef;
  cursor: pointer;
  transition: background 0.2s ease;
}

.btn.primary {
  background: #1b59b0;
  border-color: #1b59b0;
}

.btn:hover {
  filter: brightness(1.1);
}

.status {
  font-size: 12px;
  color: #9fb0cc;
}

.info-row {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 10px 12px;
  background: #0f1626;
  border: 1px solid #22314f;
  border-radius: 6px;
  font-size: 13px;
}

.info-row .label {
  color: #9fb0cc;
}

.info-row .value {
  color: #e6e9ef;
  font-weight: 500;
}

input[readonly] {
  opacity: 0.7;
  cursor: not-allowed;
}
</style>

