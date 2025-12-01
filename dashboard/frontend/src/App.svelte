<script lang="ts">
  import { onMount } from 'svelte';
  import { api, type Agent, type SystemMetrics, type DiscoveredAgent } from './lib/api';
  import { formatBytes, formatUptime, formatPercent } from './lib/utils';
  import MetricsChart from './lib/components/MetricsChart.svelte';
  import AgentCard from './lib/components/AgentCard.svelte';

  let agents: Agent[] = [];
  let selectedAgent: Agent | null = null;
  let metrics: SystemMetrics | null = null;
  let agentMetrics: Map<string, SystemMetrics> = new Map();
  let loading = false;
  let discoveredAgents: DiscoveredAgent[] = [];
  let showDiscoveryModal = false;
  let view: 'overview' | 'detail' = 'overview';

  onMount(async () => {
    await loadAgents();
    
    // Auto-refresh all agent metrics every 5 seconds
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
    agentMetrics = agentMetrics; // Trigger reactivity
  }

  async function loadMetrics(agentId: string) {
    try {
      metrics = await api.getMetrics(agentId);
    } catch (error) {
      console.error('Failed to load metrics:', error);
    }
  }

  async function discoverAgents() {
    loading = true;
    try {
      discoveredAgents = await api.discoverAgents();
      if (discoveredAgents.length > 0) {
        showDiscoveryModal = true;
      } else {
        alert('No new agents discovered');
      }
    } catch (error) {
      console.error('Discovery failed:', error);
      alert('Discovery failed: ' + error);
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

  async function deleteAgent(agentId: string) {
  try {
    await api.removeAgent(agentId);
    
    // If we're viewing this agent, go back to overview
    if (selectedAgent?.id === agentId) {
      backToOverview();
    }
    
    // Reload agents
    await loadAgents();
  } catch (error) {
    console.error('Failed to delete agent:', error);
    alert('Failed to delete agent: ' + error);
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

  function closeModal() {
    showDiscoveryModal = false;
  }
</script>

<main class="min-h-screen bg-gradient-to-br from-gray-900 via-black to-gray-900 p-8">
  <div class="max-w-7xl mx-auto">
    <!-- Header -->
    <div class="mb-8 flex items-center justify-between">
      <div>
        <h1 class="text-4xl font-bold bg-gradient-to-r from-green-400 to-blue-500 bg-clip-text text-transparent">
          Sentinel
        </h1>
        <p class="text-gray-400 mt-2">System Monitoring Dashboard</p>
      </div>
      
      {#if view === 'detail'}
        <button 
          on:click={backToOverview}
          class="px-4 py-2 bg-gray-700 hover:bg-gray-600 rounded-lg transition-colors flex items-center gap-2"
        >
          <svg class="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
            <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M10 19l-7-7m0 0l7-7m-7 7h18" />
          </svg>
          Back to Overview
        </button>
      {/if}
    </div>

    <!-- Controls -->
    <div class="mb-6 flex gap-4">
      <button 
        on:click={discoverAgents}
        disabled={loading}
        class="px-4 py-2 bg-green-600 hover:bg-green-700 rounded-lg transition-colors disabled:opacity-50 flex items-center gap-2"
      >
        <svg class="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
          <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M21 21l-6-6m2-5a7 7 0 11-14 0 7 7 0 0114 0z" />
        </svg>
        {loading ? 'Scanning...' : 'Discover Agents'}
      </button>
      <button 
        on:click={loadAgents}
        class="px-4 py-2 bg-blue-600 hover:bg-blue-700 rounded-lg transition-colors flex items-center gap-2"
      >
        <svg class="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
          <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M4 4v5h.582m15.356 2A8.001 8.001 0 004.582 9m0 0H9m11 11v-5h-.581m0 0a8.003 8.003 0 01-15.357-2m15.357 2H15" />
        </svg>
        Refresh
      </button>
    </div>

    <!-- Discovery Modal -->
    {#if showDiscoveryModal}
      <div class="fixed inset-0 bg-black/50 backdrop-blur-sm flex items-center justify-center z-50" on:click={closeModal}>
        <div class="bg-gray-800 rounded-xl p-6 max-w-2xl w-full mx-4 border border-gray-700" on:click|stopPropagation>
          <div class="flex justify-between items-center mb-4">
            <h3 class="text-xl font-semibold">Discovered Agents</h3>
            <button on:click={closeModal} class="text-gray-400 hover:text-white">
              <svg class="w-6 h-6" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M6 18L18 6M6 6l12 12" />
              </svg>
            </button>
          </div>

          {#if discoveredAgents.length === 0}
            <p class="text-gray-400">No new agents found on the network.</p>
          {:else}
            <div class="space-y-3">
              {#each discoveredAgents as discovered}
                <div class="flex items-center justify-between p-4 bg-gray-900/50 rounded-lg border border-gray-700">
                  <div>
                    <div class="font-semibold">{discovered.instance}</div>
                    <div class="text-sm text-gray-400">
                      {discovered.ips.join(', ')} : {discovered.port}
                    </div>
                  </div>
                  <button
                    on:click={() => addDiscoveredAgent(discovered)}
                    class="px-4 py-2 bg-green-600 hover:bg-green-700 rounded-lg transition-colors"
                  >
                    Add Agent
                  </button>
                </div>
              {/each}
            </div>
          {/if}
        </div>
      </div>
    {/if}

    <!-- OVERVIEW VIEW -->
    {#if view === 'overview'}
      {#if agents.length === 0}
        <div class="bg-gray-800/50 backdrop-blur-sm rounded-xl p-12 border border-gray-700 text-center">
          <svg class="w-16 h-16 mx-auto text-gray-600 mb-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
            <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M5 12h14M5 12a2 2 0 01-2-2V6a2 2 0 012-2h14a2 2 0 012 2v4a2 2 0 01-2 2M5 12a2 2 0 00-2 2v4a2 2 0 002 2h14a2 2 0 002-2v-4a2 2 0 00-2-2m-2-4h.01M17 16h.01" />
          </svg>
          <h3 class="text-xl font-semibold mb-2">No Agents Registered</h3>
          <p class="text-gray-400 mb-6">Click "Discover Agents" to find and add monitoring agents on your network.</p>
          <button 
            on:click={discoverAgents}
            class="px-6 py-3 bg-green-600 hover:bg-green-700 rounded-lg transition-colors"
          >
            Discover Agents
          </button>
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
    <!-- System Info with Delete Button -->
    <div class="bg-gray-800/50 backdrop-blur-sm rounded-xl p-6 border border-gray-700">
      <div class="flex items-center justify-between mb-4">
        <h2 class="text-2xl font-bold">{metrics.hostname}</h2>
        <button
          on:click={() => deleteAgent(selectedAgent.id)}
          class="px-4 py-2 bg-red-600 hover:bg-red-700 rounded-lg transition-colors flex items-center gap-2"
        >
          <svg class="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
            <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M19 7l-.867 12.142A2 2 0 0116.138 21H7.862a2 2 0 01-1.995-1.858L5 7m5 4v6m4-6v6m1-10V4a1 1 0 00-1-1h-4a1 1 0 00-1 1v3M4 7h16" />
          </svg>
          Remove Agent
        </button>
      </div>
      <div class="text-gray-400">
        Uptime: {formatUptime(metrics.uptime)}
      </div>
    </div>
    <!-- Rest of detail view stays the same -->

        <!-- Historical Charts -->
        <div class="bg-gray-800/50 backdrop-blur-sm rounded-xl p-6 border border-gray-700">
          <h3 class="text-xl font-semibold mb-6">Historical Metrics (Last Hour)</h3>
          <div class="grid grid-cols-1 lg:grid-cols-2 gap-6">
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

        <!-- Current CPU -->
        <div class="bg-gray-800/50 backdrop-blur-sm rounded-xl p-6 border border-gray-700">
          <h3 class="text-xl font-semibold mb-4">CPU (Current)</h3>
          <div class="space-y-2">
            <div class="flex justify-between">
              <span>Usage:</span>
              <span class="font-mono">{formatPercent(metrics.cpu.usage_percent)}</span>
            </div>
            <div class="w-full bg-gray-700 rounded-full h-2">
              <div 
                class="bg-gradient-to-r from-green-400 to-blue-500 h-2 rounded-full transition-all"
                style="width: {metrics.cpu.usage_percent}%"
              ></div>
            </div>
            <div class="text-sm text-gray-400">
              {metrics.cpu.core_count} cores
            </div>
          </div>
        </div>

        <!-- Current Memory -->
        <div class="bg-gray-800/50 backdrop-blur-sm rounded-xl p-6 border border-gray-700">
          <h3 class="text-xl font-semibold mb-4">Memory (Current)</h3>
          <div class="space-y-2">
            <div class="flex justify-between">
              <span>Usage:</span>
              <span class="font-mono">{formatPercent(metrics.memory.used_percent)}</span>
            </div>
            <div class="w-full bg-gray-700 rounded-full h-2">
              <div 
                class="bg-gradient-to-r from-purple-400 to-pink-500 h-2 rounded-full transition-all"
                style="width: {metrics.memory.used_percent}%"
              ></div>
            </div>
            <div class="text-sm text-gray-400">
              {formatBytes(metrics.memory.used)} / {formatBytes(metrics.memory.total)}
            </div>
          </div>
        </div>

        <!-- Disks -->
        <div class="bg-gray-800/50 backdrop-blur-sm rounded-xl p-6 border border-gray-700">
          <h3 class="text-xl font-semibold mb-4">Disks</h3>
          <div class="space-y-4">
            {#each metrics.disk as disk}
              <div class="space-y-2">
                <div class="flex justify-between text-sm">
                  <span class="font-mono">{disk.mount_point}</span>
                  <span>{formatPercent(disk.used_percent)}</span>
                </div>
                <div class="w-full bg-gray-700 rounded-full h-2">
                  <div 
                    class="bg-gradient-to-r from-yellow-400 to-orange-500 h-2 rounded-full transition-all"
                    style="width: {disk.used_percent}%"
                  ></div>
                </div>
                <div class="text-xs text-gray-400">
                  {formatBytes(disk.used)} / {formatBytes(disk.total)} free
                </div>
              </div>
            {/each}
          </div>
        </div>

        <!-- Network -->
        <div class="bg-gray-800/50 backdrop-blur-sm rounded-xl p-6 border border-gray-700">
          <h3 class="text-xl font-semibold mb-4">Network</h3>
          <div class="grid grid-cols-1 md:grid-cols-2 gap-4">
            {#each metrics.network as net}
              <div class="p-4 bg-gray-800/50 rounded-lg">
                <div class="font-mono text-sm mb-2">{net.interface}</div>
                <div class="space-y-1 text-sm">
                  <div class="flex justify-between">
                    <span class="text-gray-400">↓ Received:</span>
                    <span>{formatBytes(net.bytes_recv)}</span>
                  </div>
                  <div class="flex justify-between">
                    <span class="text-gray-400">↑ Sent:</span>
                    <span>{formatBytes(net.bytes_sent)}</span>
                  </div>
                </div>
              </div>
            {/each}
          </div>
        </div>
      </div>
    {/if}
  </div>
</main>