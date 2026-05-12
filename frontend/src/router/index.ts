import { createRouter, createWebHistory, RouteRecordRaw } from "vue-router";
import { COIN_LIST, getCoinConfig } from "@/config/coins";
import CoinMonitorView from "@/views/CoinMonitorView.vue";
import SummaryView from "@/views/SummaryView.vue";

const routes: RouteRecordRaw[] = [
  {
    path: "/",
    redirect: "/btc"
  },
  {
    path: "/summary",
    name: "summary",
    component: SummaryView
  },
  ...COIN_LIST.map((coin) => ({
    path: `/${coin.id}`,
    name: coin.id,
    component: CoinMonitorView,
    props: { coinId: coin.id }
  })),
  {
    path: "/:pathMatch(.*)*",
    redirect: (to) => {
      const coinId = String(to.params.pathMatch ?? "").split("/")[0].toLowerCase();
      return getCoinConfig(coinId) ? `/${coinId}` : "/btc";
    }
  }
];

const router = createRouter({
  history: createWebHistory(import.meta.env.BASE_URL),
  routes
});

export default router;
