<template>
  <main class="wrapper">
    <div class="backdrop backdrop-a"></div>
    <div class="backdrop backdrop-b"></div>
    <div class="shell">
      <header class="header">
        <div class="brand">
          <span class="eyebrow">Binance Futures Radar</span>
          <div class="title-row">
            <h1>大额订单监控</h1>
            <span class="subtitle">V 2.2</span>
          </div>
          <p class="description">聚焦盘口大额挂单、短周期买卖失衡与成交节奏。</p>
        </div>
        <div class="header-actions">
          <button type="button" class="theme-toggle" @click="toggleTheme" :aria-label="themeAriaLabel">
            <span class="theme-pill" :class="{ active: themeMode === 'day' }" aria-hidden="true">
              <svg viewBox="0 0 24 24" fill="none">
                <circle cx="12" cy="12" r="4.2" stroke="currentColor" stroke-width="1.8" />
                <path d="M12 2.8v2.4M12 18.8v2.4M21.2 12h-2.4M5.2 12H2.8M18.5 5.5l-1.7 1.7M7.2 16.8l-1.7 1.7M18.5 18.5l-1.7-1.7M7.2 7.2 5.5 5.5" stroke="currentColor" stroke-width="1.8" stroke-linecap="round" />
              </svg>
            </span>
            <span class="theme-pill" :class="{ active: themeMode === 'night' }" aria-hidden="true">
              <svg viewBox="0 0 24 24" fill="none">
                <path d="M15.4 3.8a7.9 7.9 0 1 0 4.8 14.4A8.9 8.9 0 1 1 15.4 3.8Z" stroke="currentColor" stroke-width="1.8" stroke-linejoin="round" />
              </svg>
            </span>
          </button>
        </div>
      </header>
      <CoinNav />
      <router-view />
    </div>
  </main>
</template>

<script setup lang="ts">
import { computed, onMounted, ref } from "vue";

import CoinNav from "./components/CoinNav.vue";

type ThemeMode = "day" | "night";

const THEME_KEY = "order-monitor-theme";
const themeMode = ref<ThemeMode>("day");

const themeAriaLabel = computed(() =>
  themeMode.value === "day" ? "切换到黑夜模式" : "切换到白天模式"
);

function applyTheme(nextTheme: ThemeMode) {
  themeMode.value = nextTheme;
  document.documentElement.setAttribute("data-theme", nextTheme);
  document.documentElement.style.colorScheme = nextTheme === "night" ? "dark" : "light";
  window.localStorage.setItem(THEME_KEY, nextTheme);
}

function toggleTheme() {
  applyTheme(themeMode.value === "day" ? "night" : "day");
}

onMounted(() => {
  const savedTheme = window.localStorage.getItem(THEME_KEY);
  if (savedTheme === "day" || savedTheme === "night") {
    applyTheme(savedTheme);
    return;
  }

  applyTheme("day");
});
</script>

<style scoped>
:global(:root) {
  --app-bg: #f4ede4;
  --app-surface: #fffaf4;
  --app-surface-soft: #f7efe4;
  --app-surface-strong: #efe5d7;
  --app-border: #d8c9b8;
  --app-border-soft: #e4d6c8;
  --app-text: #272e36;
  --app-text-muted: #6d665f;
  --app-text-soft: #91887e;
  --app-accent: #b96d45;
  --app-accent-soft: #cf855d;
  --app-accent-contrast: #ffffff;
  --app-buy: #c75b69;
  --app-buy-strong: #a64554;
  --app-buy-soft: #df7885;
  --app-sell: #38a06f;
  --app-sell-strong: #2c835a;
  --app-sell-soft: #73c997;
  --app-chip-bg: rgba(255, 255, 255, 0.7);
  --app-hover-bg: rgba(185, 109, 69, 0.08);
  --app-table-head-bg: #eadfce;
  --app-table-row-border: rgba(171, 151, 129, 0.28);
  --app-bar-track: #e8ddd0;
  --app-header-bg: rgba(255, 250, 244, 0.88);
  --app-nav-bg: rgba(255, 252, 247, 0.82);
  --app-panel-bg: linear-gradient(180deg, rgba(255, 250, 244, 0.96), rgba(244, 236, 225, 0.96));
  --app-card-bg: linear-gradient(180deg, rgba(255, 253, 249, 0.96), rgba(246, 238, 228, 0.96));
  --app-shell-shadow: 0 24px 60px rgba(145, 115, 82, 0.12);
  --app-glow-a: rgba(190, 118, 70, 0.2);
  --app-glow-b: rgba(86, 161, 122, 0.14);
  --app-page-top: #f7f1e8;
  --app-page-bottom: #eee4d7;
  --app-radius-sm: 12px;
  --app-radius-md: 16px;
  --app-radius-lg: 24px;
}

:global(:root[data-theme="day"]) {
  --app-bg: #f4ede4;
  --app-surface: #fffaf4;
  --app-surface-soft: #f7efe4;
  --app-surface-strong: #efe5d7;
  --app-border: #d8c9b8;
  --app-border-soft: #e4d6c8;
  --app-text: #272e36;
  --app-text-muted: #6d665f;
  --app-text-soft: #91887e;
  --app-accent: #b96d45;
  --app-accent-soft: #cf855d;
  --app-accent-contrast: #ffffff;
  --app-buy: #c75b69;
  --app-buy-strong: #a64554;
  --app-buy-soft: #df7885;
  --app-sell: #38a06f;
  --app-sell-strong: #2c835a;
  --app-sell-soft: #73c997;
  --app-chip-bg: rgba(255, 255, 255, 0.7);
  --app-hover-bg: rgba(185, 109, 69, 0.08);
  --app-table-head-bg: #eadfce;
  --app-table-row-border: rgba(171, 151, 129, 0.28);
  --app-bar-track: #e8ddd0;
  --app-header-bg: rgba(255, 250, 244, 0.88);
  --app-nav-bg: rgba(255, 252, 247, 0.82);
  --app-panel-bg: linear-gradient(180deg, rgba(255, 250, 244, 0.96), rgba(244, 236, 225, 0.96));
  --app-card-bg: linear-gradient(180deg, rgba(255, 253, 249, 0.96), rgba(246, 238, 228, 0.96));
  --app-shell-shadow: 0 24px 60px rgba(145, 115, 82, 0.12);
  --app-glow-a: rgba(190, 118, 70, 0.2);
  --app-glow-b: rgba(86, 161, 122, 0.14);
  --app-page-top: #f7f1e8;
  --app-page-bottom: #eee4d7;
}

:global(:root[data-theme="night"]) {
  --app-bg: #141413;
  --app-surface: #1b1b1a;
  --app-surface-soft: #232321;
  --app-surface-strong: #2a2a27;
  --app-border: #353531;
  --app-border-soft: #2c2c29;
  --app-text: #f3efe7;
  --app-text-muted: #b8b1a7;
  --app-text-soft: #8f887f;
  --app-accent: #d18957;
  --app-accent-soft: #e7a475;
  --app-accent-contrast: #ffffff;
  --app-buy: #d15d6e;
  --app-buy-strong: #ab4756;
  --app-buy-soft: #e27d8b;
  --app-sell: #43ab78;
  --app-sell-strong: #31835b;
  --app-sell-soft: #7ad0a0;
  --app-chip-bg: rgba(255, 255, 255, 0.04);
  --app-hover-bg: rgba(255, 255, 255, 0.04);
  --app-table-head-bg: #242420;
  --app-table-row-border: rgba(86, 84, 78, 0.42);
  --app-bar-track: #242420;
  --app-header-bg: rgba(26, 26, 25, 0.9);
  --app-nav-bg: rgba(23, 23, 22, 0.86);
  --app-panel-bg: linear-gradient(180deg, rgba(30, 30, 29, 0.96), rgba(24, 24, 23, 0.96));
  --app-card-bg: linear-gradient(180deg, rgba(36, 36, 34, 0.96), rgba(29, 29, 27, 0.96));
  --app-shell-shadow: 0 24px 60px rgba(0, 0, 0, 0.24);
  --app-glow-a: rgba(209, 137, 87, 0.08);
  --app-glow-b: rgba(67, 171, 120, 0.07);
  --app-page-top: #121211;
  --app-page-bottom: #171715;
}

.wrapper {
  min-height: 100vh;
  position: relative;
  overflow: hidden;
  background:
    radial-gradient(circle at top center, var(--app-glow-a), transparent 30%),
    linear-gradient(180deg, var(--app-page-top) 0%, var(--app-page-bottom) 100%);
  color: var(--app-text);
  padding: 28px 24px 56px;
  font-family: Inter, "Avenir Next", "PingFang SC", "Microsoft YaHei", sans-serif;
  transition: background 0.25s ease, color 0.25s ease;
}

.shell {
  position: relative;
  z-index: 1;
  max-width: 1320px;
  margin: 0 auto;
}

.backdrop {
  position: absolute;
  inset: auto;
  pointer-events: none;
  filter: blur(28px);
  opacity: 0.34;
}

.backdrop-a {
  top: 92px;
  left: -40px;
  width: 220px;
  height: 220px;
  border-radius: 50%;
  background: var(--app-glow-a);
}

.backdrop-b {
  top: 240px;
  right: -40px;
  width: 240px;
  height: 240px;
  border-radius: 50%;
  background: var(--app-glow-b);
}

.header {
  position: relative;
  display: flex;
  justify-content: space-between;
  align-items: flex-start;
  gap: 18px;
  flex-wrap: wrap;
  padding: 22px 24px 24px;
  margin-bottom: 18px;
  border: 1px solid var(--app-border-soft);
  border-radius: var(--app-radius-lg);
  background: var(--app-header-bg);
  box-shadow: var(--app-shell-shadow), inset 0 1px 0 rgba(255, 255, 255, 0.04);
}

.brand {
  display: flex;
  flex-direction: column;
  gap: 8px;
}

.eyebrow {
  font-size: 11px;
  letter-spacing: 0.18em;
  text-transform: uppercase;
  color: var(--app-accent-soft);
}

.title-row {
  display: flex;
  align-items: baseline;
  gap: 12px;
  flex-wrap: wrap;
}

.title-row h1 {
  margin: 0;
  font-family: Georgia, "Times New Roman", "Songti SC", serif;
  font-size: clamp(28px, 4vw, 42px);
  line-height: 1.08;
  letter-spacing: -0.03em;
  font-weight: 500;
}

.description {
  max-width: 720px;
  margin: 0;
  font-size: 15px;
  line-height: 1.6;
  color: var(--app-text-muted);
}

.subtitle {
  display: inline-flex;
  align-items: center;
  min-height: 30px;
  padding: 0 12px;
  border-radius: 999px;
  background: color-mix(in srgb, var(--app-accent) 14%, transparent);
  border: 1px solid color-mix(in srgb, var(--app-accent) 24%, transparent);
  font-size: 12px;
  font-weight: 600;
  letter-spacing: 0.08em;
  color: var(--app-accent-soft);
}

.header-actions {
  display: flex;
  justify-content: flex-end;
  flex: 1;
  min-width: 168px;
}

.theme-toggle {
  display: inline-flex;
  align-items: center;
  gap: 4px;
  padding: 4px;
  border: 1px solid var(--app-border);
  border-radius: 999px;
  background: var(--app-chip-bg);
  cursor: pointer;
  transition: background 0.2s ease, border-color 0.2s ease, transform 0.2s ease;
}

.theme-toggle:hover {
  background: var(--app-hover-bg);
  border-color: var(--app-border-soft);
  transform: translateY(-1px);
}

.theme-pill {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  width: 42px;
  min-height: 34px;
  padding: 0;
  border-radius: 999px;
  color: var(--app-text-soft);
  transition: background 0.2s ease, color 0.2s ease, box-shadow 0.2s ease;
}

.theme-pill svg {
  width: 18px;
  height: 18px;
}

.theme-pill.active {
  background: var(--app-accent);
  color: var(--app-accent-contrast);
  box-shadow: 0 8px 20px color-mix(in srgb, var(--app-accent) 28%, transparent);
}

@media (max-width: 768px) {
  .wrapper {
    padding: 16px 14px 28px;
  }

  .header {
    padding: 18px 16px 20px;
    border-radius: 18px;
  }

  .header-actions {
    width: 100%;
    justify-content: flex-start;
  }

  .description {
    font-size: 13px;
  }
}
</style>
