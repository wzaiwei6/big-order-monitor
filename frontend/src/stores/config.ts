import { defineStore } from "pinia";
import type { CoinConfig } from "@/config/coins";

type MarketType = "spot" | "usdtm" | "coinm";
type ThresholdOp = "gt" | "lt";

interface ConfigState {
  symbol: string;
  marketType: MarketType;
  depth: number;
  threshold: number;
  thresholdOp: ThresholdOp;
  gateway: string;
  gatewayProxy: string;
  autoReconnect: boolean;
}

export const useConfigStore = defineStore("config", {
  state: (): ConfigState => ({
    symbol: "btcusdt",
    marketType: "usdtm",
    depth: 20,
    threshold: 3,
    thresholdOp: "gt",
    gateway: "",
    gatewayProxy: "",
    autoReconnect: true
  }),
  actions: {
    initFromCoinConfig(config: CoinConfig) {
      this.symbol = config.symbol;
      this.marketType = config.marketType;
      this.threshold = config.defaultThreshold;
    }
  }
});

