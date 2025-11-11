import { createRouter, createWebHistory, RouteRecordRaw } from "vue-router";
import CoinMonitorView from "@/views/CoinMonitorView.vue";

const routes: RouteRecordRaw[] = [
  {
    path: "/",
    redirect: "/btc"
  },
  {
    path: "/btc",
    name: "btc",
    component: CoinMonitorView,
    props: { coinId: "btc" }
  },
  {
    path: "/eth",
    name: "eth",
    component: CoinMonitorView,
    props: { coinId: "eth" }
  },
  {
    path: "/sol",
    name: "sol",
    component: CoinMonitorView,
    props: { coinId: "sol" }
  },
  {
    path: "/wld",
    name: "wld",
    component: CoinMonitorView,
    props: { coinId: "wld" }
  },
  {
    path: "/doge",
    name: "doge",
    component: CoinMonitorView,
    props: { coinId: "doge" }
  },
  {
    path: "/fil",
    name: "fil",
    component: CoinMonitorView,
    props: { coinId: "fil" }
  },
  {
    path: "/bnb",
    name: "bnb",
    component: CoinMonitorView,
    props: { coinId: "bnb" }
  }
];

const router = createRouter({
  history: createWebHistory(import.meta.env.BASE_URL),
  routes
});

export default router;

