const DEFAULT_POLL_INTERVAL_MS = 500;
const MIN_POLL_INTERVAL_MS = 250;

function parsePollInterval() {
  const rawValue = Number(import.meta.env.VITE_POLL_INTERVAL_MS);
  if (!Number.isFinite(rawValue) || rawValue <= 0) {
    return DEFAULT_POLL_INTERVAL_MS;
  }

  return Math.max(MIN_POLL_INTERVAL_MS, Math.floor(rawValue));
}

export const POLL_INTERVAL_MS = parsePollInterval();

export const POLL_INTERVAL_LABEL =
  POLL_INTERVAL_MS >= 1000
    ? `${Number((POLL_INTERVAL_MS / 1000).toFixed(2))}s`
    : `${POLL_INTERVAL_MS}ms`;
