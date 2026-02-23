ALTER TABLE empresa_sync_states
  ADD COLUMN IF NOT EXISTS download_blocked_until TIMESTAMPTZ;
