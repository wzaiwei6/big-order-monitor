<template>
  <section class="history-panel">
    <header class="header">
      <div>
        <h3>最近成交记录</h3>
        <p class="subhead">保留最近 100 笔达到阈值的大额挂单成交事件。</p>
      </div>
      <span class="count">{{ history.length }} 笔</span>
    </header>
    <div class="table-wrap">
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
            <td><span :class="['side-pill', item.side]">{{ renderSide(item.side) }}</span></td>
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
    </div>
  </section>
</template>

<script setup lang="ts">
import { computed } from "vue";

import { useOrderStore } from "@/stores/order";

const orderStore = useOrderStore();
const history = computed(() => orderStore.history);

function renderSide(side: "buy" | "sell") {
  return side === "buy" ? "买" : "卖";
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
.history-panel {
  padding: 22px;
  border-radius: var(--app-radius-lg);
  border: 1px solid var(--app-border-soft);
  background: var(--app-surface);
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
  font-size: 26px;
  font-weight: 500;
  letter-spacing: -0.02em;
}

.subhead {
  margin: 6px 0 0;
  font-size: 14px;
  color: var(--app-text-muted);
}

.count {
  display: inline-flex;
  align-items: center;
  min-height: 34px;
  padding: 0 12px;
  border-radius: 999px;
  background: var(--app-chip-bg);
  border: 1px solid var(--app-border);
  font-size: 12px;
  color: var(--app-text-muted);
}

.table-wrap {
  overflow: auto;
  border-radius: var(--app-radius-md);
  border: 1px solid var(--app-border);
  background: var(--app-surface-strong);
}

table {
  width: 100%;
  border-collapse: collapse;
}

th {
  position: sticky;
  top: 0;
  z-index: 1;
  background: var(--app-table-head-bg);
  color: var(--app-text-soft);
  font-weight: 600;
  letter-spacing: 0.08em;
  text-transform: uppercase;
  font-size: 11px;
}

th,
td {
  text-align: left;
  padding: 14px 16px;
  border-bottom: 1px solid var(--app-table-row-border);
  white-space: nowrap;
}

td {
  font-size: 13px;
  color: var(--app-text);
}

tbody tr {
  transition: background 0.2s ease;
}

tbody tr:hover {
  background: var(--app-hover-bg);
}

.side-pill {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  min-width: 44px;
  min-height: 26px;
  padding: 0 12px;
  border-radius: 999px;
  font-size: 11px;
  font-weight: 700;
  letter-spacing: 0.08em;
}

.side-pill.buy {
  background: color-mix(in srgb, var(--app-buy) 14%, transparent);
  border: 1px solid color-mix(in srgb, var(--app-buy) 28%, transparent);
  color: color-mix(in srgb, var(--app-buy) 42%, var(--app-text) 58%);
}

.side-pill.sell {
  background: color-mix(in srgb, var(--app-sell) 14%, transparent);
  border: 1px solid color-mix(in srgb, var(--app-sell) 28%, transparent);
  color: color-mix(in srgb, var(--app-sell) 46%, var(--app-text) 54%);
}

.placeholder {
  text-align: center;
  color: var(--app-text-muted);
  padding: 28px 16px;
}

@media (max-width: 768px) {
  .history-panel {
    padding: 14px;
    border-radius: 14px;
  }

  .header {
    align-items: flex-start;
    flex-direction: column;
    gap: 8px;
    margin-bottom: 12px;
  }

  .header h3 {
    font-size: 20px;
  }

  .subhead {
    font-size: 12px;
  }

  .count {
    min-height: 28px;
    padding: 0 10px;
    font-size: 11px;
  }

  .table-wrap {
    border-radius: 12px;
    -webkit-overflow-scrolling: touch;
  }

  th,
  td {
    padding: 10px 12px;
  }

  td {
    font-size: 12px;
  }
}
</style>
