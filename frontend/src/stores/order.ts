import { defineStore } from "pinia";

export interface AggregatedWindow {
  window: string;
  buyQty: number;
  sellQty: number;
  buyCnt: number;
  sellCnt: number;
}

export interface OrderHistoryItem {
  id: string;
  side: "buy" | "sell";
  price: number;
  quantity: number;
  firstSeen: number;
  filledTime: number;
  durationSec: number;
}

export interface RealtimeStat {
  buyWallQty: number;
  sellWallQty: number;
  buyWallCount: number;
  sellWallCount: number;
  buyValue: number;
  sellValue: number;
}

interface OrderState {
  aggregates: AggregatedWindow[];
  history: OrderHistoryItem[];
  stats: RealtimeStat | null;
  statusMessage: string;
  snapshotsByCoin: Record<string, {
    aggregates: AggregatedWindow[];
    history: OrderHistoryItem[];
    stats: RealtimeStat | null;
    lastUpdated: number | null;
  }>;
}

export const useOrderStore = defineStore("order", {
  state: (): OrderState => ({
    aggregates: [],
    history: [],
    stats: null,
    statusMessage: "未连接",
    snapshotsByCoin: {}
  }),
  actions: {
    setAggregates(payload: AggregatedWindow[]) {
      this.aggregates = payload;
    },
    setHistory(payload: OrderHistoryItem[]) {
      this.history = payload;
    },
    pushHistory(item: OrderHistoryItem) {
      this.history = [item, ...this.history].slice(0, 200);
    },
    setStats(payload: RealtimeStat | null) {
      this.stats = payload;
    },
    setStatus(message: string) {
      this.statusMessage = message;
    },
    setCoinSnapshot(coinId: string, payload: {
      aggregates: AggregatedWindow[];
      history: OrderHistoryItem[];
      stats: RealtimeStat | null;
      lastUpdated: number | null;
    }) {
      this.snapshotsByCoin[coinId] = payload;
      this.aggregates = payload.aggregates;
      this.history = payload.history;
      this.stats = payload.stats;
    },
    restoreCoinSnapshot(coinId: string) {
      const snapshot = this.snapshotsByCoin[coinId];
      if (!snapshot) {
        this.aggregates = [];
        this.history = [];
        this.stats = null;
        return null;
      }

      this.aggregates = snapshot.aggregates;
      this.history = snapshot.history;
      this.stats = snapshot.stats;
      return snapshot;
    }
  }
});
