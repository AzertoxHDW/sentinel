<script lang="ts">
  import { onMount } from 'svelte';
  import { api, type Agent, type SystemMetrics, type DiscoveredAgent } from './lib/api';
  import { formatBytes, formatUptime, formatPercent } from './lib/utils';
  import MetricsChart from './lib/components/MetricsChart.svelte';
  import NetworkChart from './lib/components/NetworkChart.svelte';
  import AgentCard from './lib/components/AgentCard.svelte';

  let agents: Agent[] = [];
  let selectedAgent: Agent | null = null;
  let metrics: SystemMetrics | null = null;
  let previousMetrics: SystemMetrics | null = null;
  let agentMetrics: Map<string, SystemMetrics> = new Map();
  let loading = false;
  let discoveredAgents: DiscoveredAgent[] = [];
  let showDiscoveryModal = false;
  let view: 'overview' | 'detail' = 'overview';
  let lastUpdateTime: number = Date.now();

  onMount(async () => {
    await loadAgents();
    
    const interval = setInterval(async () => {
      if (view === 'overview') {
        await loadAllAgentMetrics();
      } else if (selectedAgent) {
        await loadMetrics(selectedAgent.id);
      }
    }, 5000);

    return () => clearInterval(interval);
  });

  async function loadAgents() {
    try {
      agents = await api.getAgents();
      if (agents.length > 0) {
        await loadAllAgentMetrics();
      }
    } catch (error) {
      console.error('Failed to load agents:', error);
    }
  }

  async function loadAllAgentMetrics() {
    for (const agent of agents) {
      try {
        const m = await api.getMetrics(agent.id);
        agentMetrics.set(agent.id, m);
      } catch (error) {
        console.error(`Failed to load metrics for ${agent.id}:`, error);
      }
    }
    agentMetrics = agentMetrics;
  }

  async function loadMetrics(agentId: string) {
    try {
      const newMetrics = await api.getMetrics(agentId);
      const now = Date.now();
      const timeDiff = (now - lastUpdateTime) / 1000; // seconds

      if (metrics && timeDiff > 0) {
        previousMetrics = metrics;
      }

      metrics = newMetrics;
      lastUpdateTime = now;
    } catch (error) {
      console.error('Failed to load metrics:', error);
    }
  }

  function calculateNetworkSpeed(currentBytes: number, previousBytes: number, timeDiff: number): number {
    if (!previousBytes || timeDiff <= 0) return 0;
    return Math.max(0, (currentBytes - previousBytes) / timeDiff);
  }

  function formatSpeed(bytesPerSecond: number): string {
    if (bytesPerSecond < 1024) return `${bytesPerSecond.toFixed(0)} B/s`;
    if (bytesPerSecond < 1024 * 1024) return `${(bytesPerSecond / 1024).toFixed(1)} KB/s`;
    if (bytesPerSecond < 1024 * 1024 * 1024) return `${(bytesPerSecond / (1024 * 1024)).toFixed(1)} MB/s`;
    return `${(bytesPerSecond / (1024 * 1024 * 1024)).toFixed(2)} GB/s`;
  }

  async function discoverAgents() {
    loading = true;
    try {
      discoveredAgents = await api.discoverAgents();
      if (discoveredAgents.length > 0) {
        showDiscoveryModal = true;
      }
    } catch (error) {
      console.error('Discovery failed:', error);
    } finally {
      loading = false;
    }
  }

  async function addDiscoveredAgent(discovered: DiscoveredAgent) {
    try {
      const ipAddress = discovered.ips[0];
      await api.addAgent({
        ip_address: ipAddress,
        port: discovered.port,
        hostname: discovered.instance,
      });
      
      await loadAgents();
      discoveredAgents = discoveredAgents.filter(a => a.instance !== discovered.instance);
      
      if (discoveredAgents.length === 0) {
        showDiscoveryModal = false;
      }
    } catch (error) {
      console.error('Failed to add agent:', error);
    }
  }

  function selectAgent(agent: Agent) {
    selectedAgent = agent;
    view = 'detail';
    loadMetrics(agent.id);
  }

  function backToOverview() {
    view = 'overview';
    selectedAgent = null;
  }

  async function deleteAgent(agentId: string) {
    try {
      await api.removeAgent(agentId);
      if (selectedAgent?.id === agentId) {
        backToOverview();
      }
      await loadAgents();
    } catch (error) {
      console.error('Failed to delete agent:', error);
    }
  }

  function closeModal() {
    showDiscoveryModal = false;
  }
</script>

<main class="min-h-screen bg-[#0a0a0a] text-gray-100">
  <div class="max-w-[1400px] mx-auto px-6 py-12">
    
    <!-- Header -->
    <header class="mb-12 flex items-center justify-between">
      <div>
        <div class="flex items-center gap-3 mb-2">
          {#if view === 'detail'}
            <button 
              on:click={backToOverview}
              class="p-2 hover:bg-gray-800 rounded-lg transition-colors"
            >
              <svg class="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M10 19l-7-7m0 0l7-7m-7 7h18" />
              </svg>
            </button>
          {/if}
          <h1 class="text-3xl font-semibold tracking-tight">Sentinel - Beta2</h1>
        </div>
        <p class="text-sm text-gray-500">Infrastructure monitoring</p>
      </div>
      
      <div class="flex items-center gap-3">
        <button 
          on:click={discoverAgents}
          disabled={loading}
          class="px-4 py-2 text-sm bg-emerald-500 hover:bg-emerald-600 text-black font-medium rounded-lg 
                 transition-all disabled:opacity-50 disabled:cursor-not-allowed"
        >
          {loading ? 'Scanning...' : 'Discover'}
        </button>
        <button 
          on:click={loadAgents}
          class="p-2 hover:bg-gray-800 rounded-lg transition-colors"
          title="Refresh"
        >
          <svg class="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
            <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M4 4v5h.582m15.356 2A8.001 8.001 0 004.582 9m0 0H9m11 11v-5h-.581m0 0a8.003 8.003 0 01-15.357-2m15.357 2H15" />
          </svg>
        </button>
      </div>
    </header>

    <!-- Discovery Modal -->
    {#if showDiscoveryModal}
      <div class="fixed inset-0 bg-black/80 backdrop-blur-sm flex items-center justify-center z-50 p-4" on:click={closeModal}>
        <div class="bg-[#0d0d0d] rounded-2xl p-6 max-w-2xl w-full border border-gray-800" on:click|stopPropagation>
          <div class="flex items-center justify-between mb-6">
            <h3 class="text-lg font-medium">Discovered Agents</h3>
            <button on:click={closeModal} class="p-2 hover:bg-gray-800 rounded-lg transition-colors">
              <svg class="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M6 18L18 6M6 6l12 12" />
              </svg>
            </button>
          </div>

          {#if discoveredAgents.length === 0}
            <p class="text-sm text-gray-500">No new agents found</p>
          {:else}
            <div class="space-y-3">
              {#each discoveredAgents as discovered}
                <div class="flex items-center justify-between p-4 bg-gray-900/50 rounded-xl border border-gray-800">
                  <div>
                    <div class="font-medium">{discovered.instance}</div>
                    <div class="text-sm text-gray-500 font-mono mt-0.5">
                      {discovered.ips[0]}:{discovered.port}
                    </div>
                  </div>
                  <button
                    on:click={() => addDiscoveredAgent(discovered)}
                    class="px-4 py-2 text-sm bg-emerald-500 hover:bg-emerald-600 text-black font-medium rounded-lg transition-colors"
                  >
                    Add
                  </button>
                </div>
              {/each}
            </div>
          {/if}
        </div>
      </div>
    {/if}

    <!-- OVERVIEW -->
    {#if view === 'overview'}
      {#if agents.length === 0}
        <div class="flex items-center justify-center min-h-[60vh]">
          <div class="text-center max-w-md">
            <div class="w-16 h-16 mx-auto mb-6 rounded-2xl bg-gray-900 flex items-center justify-center">
              <svg class="w-8 h-8 text-gray-600" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M5 12h14M5 12a2 2 0 01-2-2V6a2 2 0 012-2h14a2 2 0 012 2v4a2 2 0 01-2 2M5 12a2 2 0 00-2 2v4a2 2 0 002 2h14a2 2 0 002-2v-4a2 2 0 00-2-2m-2-4h.01M17 16h.01" />
              </svg>
            </div>
            <h3 class="text-xl font-medium mb-2">No agents yet</h3>
            <p class="text-sm text-gray-500 mb-6">Discover and add agents to start monitoring your infrastructure</p>
            <button 
              on:click={discoverAgents}
              class="px-6 py-2.5 bg-emerald-500 hover:bg-emerald-600 text-black font-medium rounded-lg transition-colors"
            >
              Discover Agents
            </button>
          </div>
        </div>
      {:else}
        <div class="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-6">
          {#each agents as agent}
            <AgentCard 
              {agent} 
              metrics={agentMetrics.get(agent.id) || null}
              onClick={() => selectAgent(agent)}
              onDelete={() => deleteAgent(agent.id)}
            />
          {/each}
        </div>
      {/if}
    {/if}

    <!-- DETAIL VIEW -->
    {#if view === 'detail' && metrics && selectedAgent}
      <div class="space-y-6">
        
        <!-- Header Card -->
        <div class="bg-[#0d0d0d] rounded-2xl p-6 border border-gray-800">
          <div class="flex items-start justify-between">
            <div>
              <h2 class="text-2xl font-medium mb-1">{metrics.hostname}</h2>
              <p class="text-sm text-gray-500">Uptime: {formatUptime(metrics.uptime)}</p>
            </div>
            <button
              on:click={() => deleteAgent(selectedAgent.id)}
              class="px-4 py-2 text-sm bg-red-500/10 hover:bg-red-500/20 text-red-400 rounded-lg transition-colors"
            >
              Remove
            </button>
          </div>
        </div>

        <!-- Charts CPU & RAM -->
        <div class="bg-[#0d0d0d] rounded-2xl p-6 border border-gray-800">
          <h3 class="text-sm font-medium text-gray-400 uppercase tracking-wider mb-6">Historical Metrics</h3>
          <div class="grid grid-cols-1 lg:grid-cols-2 gap-8">
            <MetricsChart 
              agentId={selectedAgent.id}
              measurement="cpu"
              field="usage_percent"
              title="CPU Usage"
              color="#10b981"
              unit="%"
            />
            <MetricsChart 
              agentId={selectedAgent.id}
              measurement="memory"
              field="used_percent"
              title="Memory Usage"
              color="#8b5cf6"
              unit="%"
            />
          </div>
        </div>

        <!-- Current Metrics Grid -->
        <div class="grid grid-cols-1 lg:grid-cols-2 gap-6">
          
          <!-- CPU -->
<div class="bg-[#0d0d0d] rounded-2xl p-6 border border-gray-800">
  <div class="flex items-center justify-between mb-4">
    <div class="text-xs text-gray-500 uppercase tracking-wider">CPU</div>
    <div class="text-xs text-gray-400 font-mono">{metrics.cpu.core_count} cores</div>
  </div>
  <div class="flex items-baseline gap-2 mb-2">
    <span class="text-4xl font-mono font-medium">{metrics.cpu.usage_percent.toFixed(1)}</span>
    <span class="text-lg text-gray-500">%</span>
  </div>
  <div class="h-2 bg-gray-900 rounded-full overflow-hidden mb-4">
    <div 
      class="h-full bg-emerald-400 rounded-full transition-all duration-500"
      style="width: {metrics.cpu.usage_percent}%"
    ></div>
  </div>
  <p class="text-sm text-gray-500">{metrics.cpu.model}</p>
</div>

          <!-- Memory -->
          <div class="bg-[#0d0d0d] rounded-2xl p-6 border border-gray-800">
            <div class="text-xs text-gray-500 uppercase tracking-wider mb-4">Memory</div>
            <div class="flex items-baseline gap-2 mb-4">
              <span class="text-4xl font-mono font-medium">{metrics.memory.used_percent.toFixed(1)}</span>
              <span class="text-lg text-gray-500">%</span>
            </div>
            <div class="h-2 bg-gray-900 rounded-full overflow-hidden mb-4">
              <div 
                class="h-full bg-purple-400 rounded-full transition-all duration-500"
                style="width: {metrics.memory.used_percent}%"
              ></div>
            </div>
            <p class="text-sm text-gray-500">{formatBytes(metrics.memory.used)} / {formatBytes(metrics.memory.total)}</p>
          </div>
        </div>

        <!-- Disks -->
        <div class="bg-[#0d0d0d] rounded-2xl p-6 border border-gray-800">
          <div class="text-xs text-gray-500 uppercase tracking-wider mb-6">Storage</div>
          <div class="space-y-6">
            {#each metrics.disk as disk}
              <div>
                <div class="flex items-center justify-between mb-2">
                  <span class="text-sm font-mono text-gray-400">{disk.mount_point}</span>
                  <span class="text-sm font-mono">{disk.used_percent.toFixed(1)}%</span>
                </div>
                <div class="h-1.5 bg-gray-900 rounded-full overflow-hidden mb-2">
                  <div 
                    class="h-full bg-yellow-400 rounded-full transition-all duration-500"
                    style="width: {disk.used_percent}%"
                  ></div>
                </div>
                <p class="text-xs text-gray-500">
                  {formatBytes(disk.used)} / {formatBytes(disk.total)}
                </p>
              </div>
            {/each}
          </div>
        </div>

        <!-- Network Stats (Current Speed) -->
{#if metrics.network.length > 0}
  <div class="bg-[#0d0d0d] rounded-2xl p-6 border border-gray-800">
    <div class="text-xs text-gray-500 uppercase tracking-wider mb-6">Network Speed</div>
    <div class="grid grid-cols-1 md:grid-cols-2 gap-4">
      {#each metrics.network as net}
        {@const prevNet = previousMetrics?.network.find(n => n.interface === net.interface)}
        {@const timeDiff = (Date.now() - lastUpdateTime) / 1000 + 5} <!-- approximate 5s refresh -->
        {@const downloadSpeed = prevNet ? calculateNetworkSpeed(net.bytes_recv, prevNet.bytes_recv, timeDiff) : 0}
        {@const uploadSpeed = prevNet ? calculateNetworkSpeed(net.bytes_sent, prevNet.bytes_sent, timeDiff) : 0}
        
        <div class="p-4 bg-gray-900/30 rounded-xl">
          <div class="font-mono text-sm mb-3 text-gray-400">{net.interface}</div>
          <div class="space-y-2">
            <div class="flex items-center justify-between text-sm">
              <span class="text-gray-500 flex items-center gap-2">
                <svg class="w-4 h-4 text-blue-400" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                  <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M19 14l-7 7m0 0l-7-7m7 7V3" />
                </svg>
                Download
              </span>
              <span class="font-mono text-blue-400">{formatSpeed(downloadSpeed)}</span>
            </div>
            <div class="flex items-center justify-between text-sm">
              <span class="text-gray-500 flex items-center gap-2">
                <svg class="w-4 h-4 text-emerald-400" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                  <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M5 10l7-7m0 0l7 7m-7-7v18" />
                </svg>
                Upload
              </span>
              <span class="font-mono text-emerald-400">{formatSpeed(uploadSpeed)}</span>
            </div>
          </div>
        </div>
      {/each}
    </div>
  </div>
{/if}

        <!-- Network Activity Charts -->
    {#if metrics.network.length > 0}
      <div class="bg-[#0d0d0d] rounded-2xl p-6 border border-gray-800">
        <div class="grid grid-cols-1 lg:grid-cols-2 gap-8">
          {#each metrics.network as net}
            <NetworkChart 
              agentId={selectedAgent.id}
              interface_name={net.interface}
            />
          {/each}
        </div>
      </div>
    {/if}
      </div>
    {/if}
  </div>
</main>