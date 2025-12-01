<script lang="ts">
  import { onMount } from 'svelte';
  import { api, type Agent, type SystemMetrics, type DiscoveredAgent } from './lib/api';
  import { formatBytes, formatUptime, formatPercent } from './lib/utils';
  import MetricsChart from './lib/components/MetricsChart.svelte';

  let agents: Agent[] = [];
  let selectedAgent: Agent | null = null;
  let metrics: SystemMetrics | null = null;
  let loading = false;
  let discoveredAgents: DiscoveredAgent[] = [];
  let showDiscoveryModal = false;

  onMount(async () => {
    await loadAgents();
    const interval = setInterval(async () => {
      if (selectedAgent) {
        await loadMetrics(selectedAgent.id);
      }
    }, 5000);

    return () => clearInterval(interval);
  });

  async function loadAgents() {
    try {
      agents = await api.getAgents();
      if (agents.length > 0 && !selectedAgent) {
        selectedAgent = agents[0];
        await loadMetrics(selectedAgent.id);
      }
    } catch (error) {
      console.error('Failed to load agents:', error);
    }
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
      showDiscoveryModal = true;
      console.log('Discovered agents:', discoveredAgents);
    } catch (error) {
      console.error('Discovery failed:', error);
    } finally {
      loading = false;
    }
  }

  async function addDiscoveredAgent(discovered: DiscoveredAgent) {
    try {
      // Use first IP address
      const ipAddress = discovered.ips[0];
      await api.addAgent({
        ip_address: ipAddress,
        port: discovered.port,
        hostname: discovered.instance,
      });
      
      // Reload agents list
      await loadAgents();
      
      // Remove from discovered list
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
    loadMetrics(agent.id);
  }

  function closeModal() {
    showDiscoveryModal = false;
  }
</script>

<main class="min-h-screen bg-gradient-to-br from-gray-900 via-black to-gray-900 p-8">
  <div class="max-w-7xl mx-auto">
    <!-- Header -->
    <div class="mb-8">
      <h1 class="text-4xl font-bold bg-gradient-to-r from-green-400 to-blue-500 bg-clip-text text-transparent">
        Sentinel
      </h1>
      <p class="text-gray-400 mt-2">System Monitoring Dashboard</p>
    </div>

    <!-- Controls -->
    <div class="mb-6 flex gap-4">
      <button 
        on:click={discoverAgents}
        disabled={loading}
        class="px-4 py-2 bg-green-600 hover:bg-green-700 rounded-lg transition-colors disabled:opacity-50"
      >
        {loading ? 'Scanning...' : 'Discover Agents'}
      </button>
      <button 
        on:click={loadAgents}
        class="px-4 py-2 bg-blue-600 hover:bg-blue-700 rounded-lg transition-colors"
      >
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

    <!-- Agent List -->
    <div class="mb-8 bg-gray-800/50 backdrop-blur-sm rounded-xl p-6 border border-gray-700">
      <h2 class="text-xl font-semibold mb-4">Agents</h2>
      {#if agents.length === 0}
        <p class="text-gray-400">No agents registered. Click "Discover Agents" to find them.</p>
      {:else}
        <div class="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-4">
          {#each agents as agent}
            <button
              on:click={() => selectAgent(agent)}
              class="text-left p-4 rounded-lg border-2 transition-all {selectedAgent?.id === agent.id 
                ? 'border-green-500 bg-green-500/10' 
                : 'border-gray-700 bg-gray-800/30 hover:border-gray-600'}"
            >
              <div class="font-semibold">{agent.hostname}</div>
              <div class="text-sm text-gray-400">{agent.ip_address}:{agent.port}</div>
              <div class="mt-2">
                <span class="inline-block px-2 py-1 text-xs rounded {
                  agent.status === 'online' ? 'bg-green-500/20 text-green-400' :
                  agent.status === 'offline' ? 'bg-red-500/20 text-red-400' :
                  'bg-gray-500/20 text-gray-400'
                }">
                  {agent.status}
                </span>
              </div>
            </button>
          {/each}
        </div>
      {/if}
    </div>

    <!-- Rest of the metrics display stays the same -->
    {#if metrics && selectedAgent}
      <div class="space-y-6">
        <!-- System Info -->
        <div class="bg-gray-800/50 backdrop-blur-sm rounded-xl p-6 border border-gray-700">
          <h2 class="text-2xl font-bold mb-4">{metrics.hostname}</h2>
          <div class="text-gray-400">
            Uptime: {formatUptime(metrics.uptime)}
          </div>
        </div>

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