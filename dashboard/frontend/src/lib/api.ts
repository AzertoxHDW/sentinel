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

export const api = {
  async getAgents(): Promise<Agent[]> {
    const res = await fetch(`${API_BASE}/agents`);
    return res.json();
  },

  async addAgent(agent: Omit<Agent, 'id' | 'added_at' | 'last_seen' | 'status'>): Promise<Agent> {
    const res = await fetch(`${API_BASE}/agents`, {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify(agent),
    });
    return res.json();
  },

  async removeAgent(id: string): Promise<void> {
    await fetch(`${API_BASE}/agents/${id}`, { method: 'DELETE' });
  },

  async discoverAgents(): Promise<DiscoveredAgent[]> {
    const res = await fetch(`${API_BASE}/agents/discover`);
    return res.json();
  },

  async getMetrics(agentId: string): Promise<SystemMetrics> {
    const res = await fetch(`${API_BASE}/metrics/${agentId}`);
    return res.json();
  },

  async checkHealth(): Promise<{ status: string }> {
    const res = await fetch(`${API_BASE}/health`);
    return res.json();
  },
};