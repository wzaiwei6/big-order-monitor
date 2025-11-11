<template>
  <section>
    <header class="header">
      <h3>最近成交记录</h3>
      <button class="btn" @click="$emit('export')">导出 CSV</button>
    </header>
    <table>
      <thead>
        <tr>
          <th>#</th>
          <th>方向</th>
          <th>价格</th>
          <th>数量</th>
          <th>成交时间</th>
          <th>存续</th>
        </tr>
      </thead>
      <tbody>
        <tr v-for="(item, index) in history" :key="item.id">
          <td>{{ index + 1 }}</td>
          <td>{{ renderSide(item.side) }}</td>
          <td>${{ formatNumber(item.price) }}</td>
          <td>{{ formatNumber(item.quantity) }}</td>
          <td>{{ formatTimestamp(item.filledTime) }}</td>
          <td>{{ formatDuration(item.durationSec) }}</td>
        </tr>
        <tr v-if="!history.length">
          <td colspan="6" class="placeholder">暂无成交数据</td>
        </tr>
      </tbody>
    </table>
  </section>
</template>

<script setup lang="ts">
import { computed } from "vue";

import { useOrderStore } from "@/stores/order";

const orderStore = useOrderStore();
const history = computed(() => orderStore.history);

function renderSide(side: "buy" | "sell") {
  return side === "buy" ? "🔴 买" : "🟢 卖";
}

function formatNumber(value: number) {
  return value.toLocaleString(undefined, { maximumFractionDigits: 2 });
}

function formatTimestamp(ts: number) {
  if (!ts) return "-";
  const date = new Date(ts * 1000);
  return date.toLocaleString();
}

function formatDuration(duration: number) {
  if (!Number.isFinite(duration)) return "-";
  if (duration <= 0) return "0s";
  if (duration < 60) return `${duration.toFixed(1)}s`;
  const minutes = duration / 60;
  if (minutes < 60) return `${minutes.toFixed(1)}m`;
  const hours = minutes / 60;
  return `${hours.toFixed(1)}h`;
}
</script>

<style scoped>
.header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 12px;
}

table {
  width: 100%;
  border-collapse: collapse;
}

th,
td {
  text-align: left;
  font-size: 12px;
  padding: 8px;
  border-bottom: 1px solid #1d2a44;
}

th {
  color: #9fb0cc;
  font-weight: 600;
}

.btn {
  padding: 6px 10px;
  background: #1d2a44;
  border: 1px solid #22314f;
  border-radius: 6px;
  color: #9bd1ff;
  cursor: pointer;
}

.placeholder {
  text-align: center;
  color: #9fb0cc;
}
</style>

