<template>
  <div class="layout">
    <div class="toolbar">
      <button class="toggle" @click="toggleConfig">{{ showConfig ? "关闭配置" : "显示配置" }}</button>
      <span class="status">{{ statusMessage }}</span>
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
import { onMounted, ref } from "vue";

import ConfigPanel from "@/components/ConfigPanel.vue";
import WindowSummary from "@/components/WindowSummary.vue";
import { useWebSocket } from "@/composables/useWebSocket";

const showConfig = ref(false);
const { connect, statusMessage } = useWebSocket();

onMounted(() => {
  connect();
});

function toggleConfig() {
  showConfig.value = !showConfig.value;
}
</script>

<style scoped>
.layout {
  position: relative;
  padding: 32px 16px;
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
}

.toggle {
  padding: 10px 18px;
  border-radius: 999px;
  border: 1px solid #1d2a44;
  background: #0f1626;
  color: #9bd1ff;
  cursor: pointer;
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
}

.close {
  border: none;
  background: transparent;
  font-size: 20px;
  color: #9fb0cc;
  cursor: pointer;
}

@media (max-width: 768px) {
  .drawer {
    right: 12px;
    left: 12px;
    width: auto;
  }
}
</style>

