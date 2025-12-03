const API_BASE = '/api';

export interface Agent {
  id: string;
  hostname: string;
  ip_address: string;
  port: number;
  added_at: string;
  last_seen: string;
  status: 'online' | 'offline' | 'unknown';
}

export interface DiscoveredAgent {
  hostname: string;
  instance: string;
  port: number;
  ips: string[];
}

export interface SystemMetrics {
  timestamp: string;
  hostname: string;
  uptime: number;
  cpu: {
    usage_percent: number;
    core_count: number;
    load_avg?: number[];
    model: string;
  };
  memory: {
    total: number;
    available: number;
    used: number;
    used_percent: number;
  };
  disk: Array<{
    device: string;
    mount_point: string;
    fs_type: string;
    total: number;
    used: number;
    free: number;
    used_percent: number;
  }>;
  network: Array<{
    interface: string;
    bytes_sent: number;
    bytes_recv: number;
    packets_sent: number;
    packets_recv: number;
  }>;
}

async function fetchWithTimeout(url: string, options: RequestInit = {}, timeout = 5000): Promise<Response> {
  const controller = new AbortController();
  const timeoutId = setTimeout(() => controller.abort(), timeout);
  
  try {
    const response = await fetch(url, {
      ...options,
      signal: controller.signal
    });
    clearTimeout(timeoutId);
    return response;
  } catch (error) {
    clearTimeout(timeoutId);
    throw error;
  }
}

export const api = {
  async getAgents(): Promise<Agent[]> {
    const response = await fetchWithTimeout(`${API_BASE}/agents`);
    return response.json();
  },

  async addAgent(agent: { ip_address: string; port: number; hostname: string }): Promise<void> {
    await fetchWithTimeout(`${API_BASE}/agents`, {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify(agent),
    });
  },

  async removeAgent(id: string): Promise<void> {
    await fetchWithTimeout(`${API_BASE}/agents/${id}`, {
      method: 'DELETE',
    });
  },

  async discoverAgents(): Promise<DiscoveredAgent[]> {
    const response = await fetchWithTimeout(`${API_BASE}/agents/discover`);
    return response.json();
  },

  async getMetrics(agentId: string): Promise<SystemMetrics> {
    const response = await fetchWithTimeout(`${API_BASE}/metrics/${agentId}`, {}, 3000);
    return response.json();
  },

  async checkHealth(): Promise<{ status: string }> {
    const response = await fetchWithTimeout(`${API_BASE}/health`);
    return response.json();
  },
};