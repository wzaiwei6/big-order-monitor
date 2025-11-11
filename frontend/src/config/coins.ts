export interface CoinConfig {
  id: string;
  symbol: string;
  displayName: string;
  defaultThreshold: number;
  marketType: "usdtm"; // 固定为 U 本位合约
}

export const COINS: Record<string, CoinConfig> = {
  btc: {
    id: "btc",
    symbol: "btcusdt",
    displayName: "BTC",
    defaultThreshold: 3,
    marketType: "usdtm"
  },
  eth: {
    id: "eth",
    symbol: "ethusdt",
    displayName: "ETH",
    defaultThreshold: 300,
    marketType: "usdtm"
  },
  sol: {
    id: "sol",
    symbol: "solusdt",
    displayName: "SOL",
    defaultThreshold: 800,
    marketType: "usdtm"
  },
  wld: {
    id: "wld",
    symbol: "wldusdt",
    displayName: "WLD",
    defaultThreshold: 10000,
    marketType: "usdtm"
  },
  doge: {
    id: "doge",
    symbol: "dogeusdt",
    displayName: "DOGE",
    defaultThreshold: 200000,
    marketType: "usdtm"
  },
  fil: {
    id: "fil",
    symbol: "filusdt",
    displayName: "FIL",
    defaultThreshold: 10000,
    marketType: "usdtm"
  },
  bnb: {
    id: "bnb",
    symbol: "bnbusdt",
    displayName: "BNB",
    defaultThreshold: 50,
    marketType: "usdtm"
  }
};

export const COIN_LIST = Object.values(COINS);

export function getCoinConfig(coinId: string): CoinConfig | undefined {
  return COINS[coinId.toLowerCase()];
}

