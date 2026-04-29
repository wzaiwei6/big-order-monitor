import { computed, onUnmounted, ref } from "vue";
import { storeToRefs } from "pinia";

import { useConfigStore } from "@/stores/config";
import { useOrderStore } from "@/stores/order";

interface WsMessage<T = any> {
  type: string;
  payload?: T;
}

interface ConnectionStatusPayload {
  status: string;
  message: string;
  symbol: string;
  marketType: string;
  threshold: number;
  thresholdOp: string;
}

interface HeartbeatPayload {
  type: "heartbeat";
  timestamp: number;
}

interface StatsPayload {
  buyWallQty: number;
  sellWallQty: number;
  buyWallCount: number;
  sellWallCount: number;
  buyValue: number;
  sellValue: number;
}

interface AggregatePayload {
  windows: Array<{
    window: string;
    buyQty: number;
    sellQty: number;
    buyCnt: number;
    sellCnt: number;
  }>;
}

interface OrdersPayload {
  orders: Array<{
    side: "buy" | "sell";
    price: number;
    quantity: number;
    firstSeen: number;
    filledTime: number;
    durationSec: number;
  }>;
}

export function useWebSocket() {
  const configStore = useConfigStore();
  const orderStore = useOrderStore();
  const { statusMessage } = storeToRefs(orderStore);

  const socket = ref<WebSocket | null>(null);
  const isConnected = computed(() => socket.value?.readyState === WebSocket.OPEN);

  function buildUrl() {
    const base = import.meta.env.VITE_WS_URL || "ws://localhost:8081";
    const params = new URLSearchParams({
      symbol: configStore.symbol,
      marketType: configStore.marketType,
      depth: String(configStore.depth),
      threshold: String(configStore.threshold),
      thresholdOp: configStore.thresholdOp
    });

    if (configStore.gateway) params.set("gateway", configStore.gateway);
    if (configStore.gatewayProxy) params.set("gatewayProxy", configStore.gatewayProxy);

    return `${base}/ws?${params.toString()}`;
  }

  function connect() {
    const url = buildUrl();
    
    // 清空旧数据，准备接收新数据
    orderStore.setHistory([]);
    orderStore.setAggregates([]);
    orderStore.setStats(null);
    orderStore.setStatus("连接中...");

    const ws = new WebSocket(url);
    socket.value = ws;

    ws.onopen = () => {
      orderStore.setStatus("✅ 已连接，等待数据...");
    };

    ws.onmessage = (event: MessageEvent<string>) => {
      try {
        const message: WsMessage = JSON.parse(event.data);
        handleMessage(message);
      } catch (error) {
        console.warn("无法解析 WebSocket 消息", error, event.data);
      }
    };

    ws.onerror = () => {
      orderStore.setStatus("❌ 连接错误");
    };

    ws.onclose = () => {
      orderStore.setStatus("🔌 已断开");
      socket.value = null;
      if (configStore.autoReconnect) {
        setTimeout(connect, 2000);
      }
    };
  }

  function handleMessage(message: WsMessage) {
    switch (message.type) {
      case "connection_status":
        handleStatus(message.payload as ConnectionStatusPayload);
        break;
      case "heartbeat":
        handleHeartbeat(message.payload as HeartbeatPayload);
        break;
      case "stats_update":
        handleStats(message.payload as StatsPayload);
        break;
      case "aggregate_update":
        handleAggregates(message.payload as AggregatePayload);
        break;
      case "order_filled":
        handleOrders(message.payload as OrdersPayload);
        break;
      default:
        console.debug("未知消息", message);
    }
  }

  function handleStatus(payload?: ConnectionStatusPayload) {
    if (!payload) return;
    const { status, message } = payload;
    const prefix = status === "connected" ? "✅" : status === "error" ? "❌" : "ℹ️";
    orderStore.setStatus(`${prefix} ${message}`);
  }

  function handleHeartbeat(payload?: HeartbeatPayload) {
    if (!payload) return;
    orderStore.setStatus(`心跳 ${new Date(payload.timestamp * 1000).toLocaleTimeString()}`);
  }

  function handleStats(payload?: StatsPayload) {
    if (!payload) return;
    orderStore.setStats({
      buyWallQty: payload.buyWallQty,
      sellWallQty: payload.sellWallQty,
      buyWallCount: payload.buyWallCount,
      sellWallCount: payload.sellWallCount,
      buyValue: payload.buyValue,
      sellValue: payload.sellValue
    });
  }

  function handleAggregates(payload?: AggregatePayload) {
    if (!payload) return;
    orderStore.setAggregates(payload.windows ?? []);
  }

  function handleOrders(payload?: OrdersPayload) {
    if (!payload || !payload.orders) return;
    
    // 如果是批量数据（历史数据），直接替换
    if (payload.orders.length > 10) {
      const items = payload.orders.map((order) => ({
        id: `${order.side}-${order.price}-${order.filledTime}`,
        side: order.side,
        price: order.price,
        quantity: order.quantity,
        firstSeen: order.firstSeen,
        filledTime: order.filledTime,
        durationSec: order.durationSec
      }));
      orderStore.setHistory(items);
    } else {
      // 单条或少量数据，逐条添加
      payload.orders.forEach((order) => {
        orderStore.pushHistory({
          id: `${order.side}-${order.price}-${order.filledTime}`,
          side: order.side,
          price: order.price,
          quantity: order.quantity,
          firstSeen: order.firstSeen,
          filledTime: order.filledTime,
          durationSec: order.durationSec
        });
      });
    }
  }

  function disconnect() {
    configStore.autoReconnect = false;
    socket.value?.close();
  }

  onUnmounted(() => {
    socket.value?.close();
  });

  return {
    connect,
    disconnect,
    isConnected,
    statusMessage
  };
}
