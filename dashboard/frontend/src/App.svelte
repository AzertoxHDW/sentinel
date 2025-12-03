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
  let addingAgentId: string | null = null
  let isLoadingMetrics = false;

  onMount(async () => {
    await loadAgents();
    
    const interval = setInterval(async () => {
      // Prevent overlapping requests
      if (isLoadingMetrics) {
        console.log('Skipping metrics refresh - previous request still in progress');
        return;
      }
      
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
    if (isLoadingMetrics) return;  // Prevent concurrent requests
    
    isLoadingMetrics = true;
    
    try {
      // Load all metrics in parallel with individual timeouts
      const promises = agents.map(async (agent) => {
        const controller = new AbortController();
        const timeoutId = setTimeout(() => controller.abort(), 3000);
        
        try {
          const response = await fetch(`/api/metrics/${agent.id}`, {
            signal: controller.signal
          });
          
          clearTimeout(timeoutId);
          
          if (response.ok) {
            const m = await response.json();
            agentMetrics.set(agent.id, m);
          } else {
            agentMetrics.delete(agent.id);
          }
        } catch (error) {
          clearTimeout(timeoutId);
          console.error(`Metrics fetch failed for ${agent.id}:`, error.name);
          agentMetrics.delete(agent.id);
        }
      });
      
      await Promise.allSettled(promises);
      agentMetrics = agentMetrics; // Trigger reactivity
    } finally {
      isLoadingMetrics = false;
    }
  }

  async function loadMetrics(agentId: string) {
    if (isLoadingMetrics) return;
    
    isLoadingMetrics = true;
    
    try {
      const controller = new AbortController();
      const timeoutId = setTimeout(() => controller.abort(), 3000);
      
      const response = await fetch(`/api/metrics/${agentId}`, {
        signal: controller.signal
      });
      
      clearTimeout(timeoutId);
      
      if (response.ok) {
        const newMetrics = await response.json();
        const now = Date.now();
        const timeDiff = (now - lastUpdateTime) / 1000;

        if (metrics && timeDiff > 0) {
          previousMetrics = metrics;
        }

        metrics = newMetrics;
        lastUpdateTime = now;
      }
    } catch (error) {
      console.error('Failed to load metrics:', error);
    } finally {
      isLoadingMetrics = false;
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
    const agentId = `${discovered.instance}:${discovered.port}`;  // Create unique ID
    addingAgentId = agentId;  // Set loading state
    
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
    } finally {
      addingAgentId = null;  // Clear loading state
    }
  }

function selectAgent(agent: Agent) {
  selectedAgent = agent;
  view = 'detail';
  metrics = null;  // Clear metrics immediately when switching
  previousMetrics = null;  // Clear previous metrics too
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
            <!-- svelte-ignore a11y_consider_explicit_label -->
            <button 
              on:click={backToOverview}
              class="p-2 hover:bg-gray-800 rounded-lg transition-colors"
            >
              <svg class="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M10 19l-7-7m0 0l7-7m-7 7h18" />
              </svg>
            </button>
          {/if}
          <div class="flex items-baseline gap-3">
  <h1 class="text-3xl font-mono tracking-tight">SENTINEL</h1>
  <span class="text-sm text-gray-500">v0.4-beta</span>
</div>
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
  <!-- svelte-ignore a11y_click_events_have_key_events -->
  <!-- svelte-ignore a11y_no_static_element_interactions -->
  <div class="fixed inset-0 bg-black/80 backdrop-blur-sm flex items-center justify-center z-50 p-4" on:click={closeModal}>
    <div class="bg-[#0d0d0d] rounded-2xl p-6 max-w-2xl w-full border border-gray-800" on:click|stopPropagation>
      <div class="flex items-center justify-between mb-6">
        <h3 class="text-lg font-medium">Discovered Agents</h3>
        <!-- svelte-ignore a11y_consider_explicit_label -->
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
            {@const agentId = `${discovered.instance}:${discovered.port}`}
            {@const isAdding = addingAgentId === agentId}
            
            <div class="flex items-center justify-between p-4 bg-gray-900/50 rounded-xl border border-gray-800">
              <div>
                <div class="font-medium">{discovered.instance}</div>
                <div class="text-sm text-gray-500 font-mono mt-0.5">
                  {discovered.ips[0]}:{discovered.port}
                </div>
              </div>
              <button
                on:click={() => addDiscoveredAgent(discovered)}
                disabled={isAdding}
                class="px-4 py-2 text-sm bg-emerald-500 hover:bg-emerald-600 text-black font-medium rounded-lg 
                       transition-colors disabled:opacity-50 disabled:cursor-not-allowed flex items-center gap-2"
              >
                {#if isAdding}
                  <span class="flex gap-1">
                    <span class="w-1.5 h-1.5 bg-black rounded-full animate-bounce" style="animation-delay: 0ms"></span>
                    <span class="w-1.5 h-1.5 bg-black rounded-full animate-bounce" style="animation-delay: 150ms"></span>
                    <span class="w-1.5 h-1.5 bg-black rounded-full animate-bounce" style="animation-delay: 300ms"></span>
                  </span>
                  Adding...
                {:else}
                  Add
                {/if}
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
    {#if view === 'detail' && selectedAgent}
  <div class="space-y-6">
    
    <!-- Header Card -->
    <div class="bg-[#0d0d0d] rounded-2xl p-6 border border-gray-800">
      <div class="flex items-start justify-between">
        <div>
          <h2 class="text-2xl font-medium mb-1">{selectedAgent.hostname}</h2>
          {#if metrics}
            <p class="text-sm text-gray-500">Uptime: {formatUptime(metrics.uptime)}</p>
          {:else}
            <p class="text-sm text-red-400">Offline - No metrics available</p>
          {/if}
        </div>
        <button
          on:click={() => deleteAgent(selectedAgent.id)}
          class="px-4 py-2 text-sm bg-red-500/10 hover:bg-red-500/20 text-red-400 rounded-lg transition-colors"
        >
          Remove
        </button>
      </div>
    </div>

    {#if metrics}

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
    {:else}
      <!-- Offline State -->
      <div class="bg-[#0d0d0d] rounded-2xl p-12 border border-gray-800 text-center">
        <svg class="w-16 h-16 mx-auto text-gray-600 mb-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
          <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M18.364 5.636a9 9 0 010 12.728m0 0l-2.829-2.829m2.829 2.829L21 21M15.536 8.464a5 5 0 010 7.072m0 0l-2.829-2.829m-4.243 2.829a4.978 4.978 0 01-1.414-2.83m-1.414 5.658a9 9 0 01-2.167-9.238m7.824 2.167a1 1 0 111.414 1.414m-1.414-1.414L3 3m8.293 8.293l1.414 1.414" />
        </svg>
        <h3 class="text-xl font-medium mb-2">Agent Offline</h3>
        <p class="text-gray-500 mb-4">Unable to connect to {selectedAgent.hostname}</p>
        <button 
          on:click={() => loadMetrics(selectedAgent.id)}
          class="px-4 py-2 bg-emerald-500 hover:bg-emerald-600 text-black font-medium rounded-lg transition-colors"
        >
          Retry Connection
        </button>
      </div>
    {/if}
    
  </div>
{/if}
  </div>
</main>