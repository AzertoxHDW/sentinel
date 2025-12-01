<script lang="ts">
  import type { Agent, SystemMetrics } from '../api';
  import { formatPercent } from '../utils';

  export let agent: Agent;
  export let metrics: SystemMetrics | null;
  export let onClick: () => void;
  export let onDelete: () => void;

  // Determine status color and warnings
  $: cpuWarning = metrics && metrics.cpu.usage_percent > 80;
  $: memWarning = metrics && metrics.memory.used_percent > 80;
  $: diskWarning = metrics && metrics.disk.some(d => d.used_percent > 90);
  
  $: statusColor = agent.status === 'online' ? 'bg-green-500' : 
                   agent.status === 'offline' ? 'bg-red-500' : 'bg-gray-500';
  
  $: hasWarnings = cpuWarning || memWarning || diskWarning;

  function handleDelete(event: Event) {
    event.stopPropagation(); // Prevent card click
    if (confirm(`Remove agent "${agent.hostname}"?`)) {
      onDelete();
    }
  }
</script>

<div class="relative">
  <button
    on:click={onClick}
    class="relative w-full text-left p-6 rounded-xl border-2 transition-all duration-200 {
      agent.status === 'offline' ? 'border-red-500/30 bg-red-500/5' :
      hasWarnings ? 'border-yellow-500/50 bg-yellow-500/10 hover:border-yellow-400' :
      'border-gray-700 bg-gray-800/30 hover:border-gray-600 hover:bg-gray-800/50'
    }"
  >
    <!-- Status indicator -->
    <div class="absolute top-4 right-4">
      <div class="w-3 h-3 rounded-full {statusColor} {agent.status === 'online' ? 'animate-pulse' : ''}"></div>
    </div>

    <!-- Hostname -->
    <div class="mb-4">
      <h3 class="text-xl font-bold">{agent.hostname}</h3>
      <p class="text-sm text-gray-400">{agent.ip_address}:{agent.port}</p>
    </div>

    {#if agent.status === 'offline'}
      <div class="text-red-400 text-sm">Offline - No data available</div>
    {:else if metrics}
      <!-- Metrics Grid -->
      <div class="grid grid-cols-2 gap-4 mb-4">
        <!-- CPU -->
        <div>
          <div class="text-xs text-gray-400 mb-1">CPU</div>
          <div class="flex items-center gap-2">
            <span class="text-2xl font-bold {cpuWarning ? 'text-yellow-400' : 'text-white'}">
              {metrics.cpu.usage_percent.toFixed(0)}%
            </span>
            {#if cpuWarning}
              <span class="text-yellow-400">⚠️</span>
            {/if}
          </div>
          <div class="w-full bg-gray-700 rounded-full h-1.5 mt-2">
            <div 
              class="h-1.5 rounded-full transition-all {cpuWarning ? 'bg-yellow-400' : 'bg-gradient-to-r from-green-400 to-blue-500'}"
              style="width: {metrics.cpu.usage_percent}%"
            ></div>
          </div>
        </div>

        <!-- Memory -->
        <div>
          <div class="text-xs text-gray-400 mb-1">Memory</div>
          <div class="flex items-center gap-2">
            <span class="text-2xl font-bold {memWarning ? 'text-yellow-400' : 'text-white'}">
              {metrics.memory.used_percent.toFixed(0)}%
            </span>
            {#if memWarning}
              <span class="text-yellow-400">⚠️</span>
            {/if}
          </div>
          <div class="w-full bg-gray-700 rounded-full h-1.5 mt-2">
            <div 
              class="h-1.5 rounded-full transition-all {memWarning ? 'bg-yellow-400' : 'bg-gradient-to-r from-purple-400 to-pink-500'}"
              style="width: {metrics.memory.used_percent}%"
            ></div>
          </div>
        </div>
      </div>

      <!-- Disk Usage -->
      <div class="mb-4">
        <div class="text-xs text-gray-400 mb-2">Disks</div>
        <div class="space-y-2">
          {#each metrics.disk.slice(0, 2) as disk}
            <div class="flex items-center justify-between text-sm">
              <span class="font-mono text-gray-400">{disk.mount_point}</span>
              <span class="{disk.used_percent > 90 ? 'text-yellow-400 font-semibold' : 'text-gray-300'}">
                {disk.used_percent.toFixed(0)}%
              </span>
            </div>
          {/each}
          {#if metrics.disk.length > 2}
            <div class="text-xs text-gray-500">+{metrics.disk.length - 2} more</div>
          {/if}
        </div>
      </div>

      <!-- Footer info -->
      <div class="text-xs text-gray-500 border-t border-gray-700 pt-3">
        {metrics.cpu.core_count} cores • Click for details
      </div>
    {:else}
      <div class="text-gray-400 text-sm">Loading metrics...</div>
    {/if}
  </button>

  <!-- Delete Button (outside the card button to prevent propagation issues) -->
  <button
    on:click={handleDelete}
    class="absolute bottom-4 right-4 p-2 bg-red-600/80 hover:bg-red-600 rounded-lg transition-colors z-10"
    title="Remove agent"
  >
    <svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
      <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M19 7l-.867 12.142A2 2 0 0116.138 21H7.862a2 2 0 01-1.995-1.858L5 7m5 4v6m4-6v6m1-10V4a1 1 0 00-1-1h-4a1 1 0 00-1 1v3M4 7h16" />
    </svg>
  </button>
</div>