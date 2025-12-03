<script lang="ts">
  import type { Agent, SystemMetrics } from '../api';
  import { formatBytes, formatBytesRounded } from '../utils';
  import Gauge from './Gauge.svelte';

  export let agent: Agent;
  export let metrics: SystemMetrics | null;
  export let onClick: () => void;
  export let onDelete: () => void;

  $: isOnline = metrics !== null && metrics !== undefined;
  $: cpuWarning = metrics && metrics.cpu.usage_percent > 85;
  $: memWarning = metrics && metrics.memory.used_percent > 85;
  $: diskWarning = metrics && metrics.disk.some(d => d.used_percent > 90);
  
  $: hasWarnings = cpuWarning || memWarning || diskWarning;

  // Get total disk capacity
  $: totalDiskCapacity = metrics ? metrics.disk.reduce((sum, d) => sum + d.total, 0) : 0;

  function handleDelete(event: Event) {
    event.stopPropagation();
    if (confirm(`Remove ${agent.hostname}?`)) {
      onDelete();
    }
  }

  function getMetricColor(value: number, warningThreshold: number = 85): string {
    if (value >= warningThreshold) return 'text-amber-400';
    if (value >= 70) return 'text-yellow-500';
    return 'text-emerald-400';
  }

  // Shorten CPU model name
  function shortenCPUModel(model: string): string {
    return model
      .replace(/\(R\)/g, '')
      .replace(/\(TM\)/g, '')
      .replace(/CPU/g, '')
      .replace(/Processor/g, '')
      .replace(/\s+/g, ' ')
      .trim();
  }
</script>

<div class="group relative">
  <button
    on:click={onClick}
    class="relative w-full text-left p-6 rounded-2xl border transition-all duration-300 {
      !isOnline
        ? 'border-red-950 bg-red-950/20 hover:bg-red-950/30' 
        : hasWarnings
        ? 'border-amber-900/30 bg-amber-950/10 hover:border-amber-800/50 hover:bg-amber-950/20'
        : 'border-gray-800 bg-[#0d0d0d] hover:border-gray-700 hover:bg-[#111]'
    }"
  >
    <!-- Header -->
    <div class="flex items-start justify-between mb-6">
      <div>
        <h3 class="text-lg font-medium tracking-tight uppercase">{agent.hostname}</h3>
        <p class="text-sm text-gray-500 font-mono mt-0.5">{agent.ip_address}</p>
      </div>
      
      <!-- Status -->
      <div class="flex items-center gap-2">
        {#if isOnline}
          <div class="w-2 h-2 rounded-full bg-emerald-400 animate-pulse"></div>
          <span class="text-xs text-emerald-400">online</span>
        {:else}
          <div class="w-2 h-2 rounded-full bg-red-500"></div>
          <span class="text-xs text-red-400">offline</span>
        {/if}
      </div>
    </div>

    {#if !isOnline}
      <!-- Offline State -->
      <div class="py-8 text-center">
        <svg class="w-12 h-12 mx-auto text-gray-700 mb-3" fill="none" stroke="currentColor" viewBox="0 0 24 24">
          <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M18.364 5.636a9 9 0 010 12.728m0 0l-2.829-2.829m2.829 2.829L21 21M15.536 8.464a5 5 0 010 7.072m0 0l-2.829-2.829m-4.243 2.829a4.978 4.978 0 01-1.414-2.83m-1.414 5.658a9 9 0 01-2.167-9.238m7.824 2.167a1 1 0 111.414 1.414m-1.414-1.414L3 3m8.293 8.293l1.414 1.414" />
        </svg>
        <p class="text-sm text-red-400/80">Agent offline</p>
        <p class="text-xs text-gray-600 mt-1">Click to view details</p>
      </div>
    {:else if metrics}
      <!-- Metrics -->
      <div class="space-y-6">
        <!-- CPU & Memory Gauges -->
        <div class="grid grid-cols-2 gap-6">
          <Gauge value={metrics.cpu.usage_percent} label="CPU" />
          <Gauge value={metrics.memory.used_percent} label="Memory" />
        </div>

        <!-- Hardware Specs -->
        <div class="pt-4 border-t border-gray-800 space-y-2 text-xs">
          <!-- CPU Model -->
          <div class="flex items-start justify-between gap-2">
            <span class="text-gray-500">CPU</span>
            <span class="text-gray-400 text-right font-mono leading-tight">
              {shortenCPUModel(metrics.cpu.model)}
            </span>
          </div>
          
          <!-- RAM -->
          <div class="flex items-center justify-between">
            <span class="text-gray-500">RAM</span>
            <span class="text-gray-400 font-mono">
              {formatBytesRounded(metrics.memory.total)}
            </span>
          </div>

          <!-- Storage -->
          {#if metrics.disk.length > 0}
            <div class="flex items-center justify-between">
              <span class="text-gray-500">Storage</span>
              <span class="text-gray-400 font-mono">
                {formatBytesRounded(totalDiskCapacity)}
              </span>
            </div>
          {/if}
        </div>

        <!-- Disk Usage -->
        {#if metrics.disk.length > 0}
          <div class="pt-4 border-t border-gray-800">
            <div class="text-xs text-gray-500 uppercase tracking-wider mb-3">Root Usage</div>
            <div class="space-y-2">
              {#each metrics.disk as disk}
                <div class="flex items-center justify-between">
                  <span class="text-xs font-mono text-gray-400">{disk.mount_point}</span>
                  <span class="text-xs font-mono {getMetricColor(disk.used_percent, 90)}">
                    {disk.used_percent.toFixed(0)}%
                  </span>
                </div>
              {/each}
            </div>
          </div>
        {/if}
      </div>

      <!-- Footer -->
      <div class="mt-6 pt-4 border-t border-gray-800 flex items-center justify-between">
        <span class="text-xs text-gray-500">{metrics.cpu.core_count} cores</span>
        <span class="text-xs text-gray-500">View details →</span>
      </div>
    {/if}
  </button>

  <!-- Delete Button (hover reveal) -->
  <button
    on:click={handleDelete}
    class="absolute top-4 right-4 p-2 bg-red-500/10 hover:bg-red-500/20 rounded-lg 
           transition-all duration-200 opacity-0 group-hover:opacity-100"
    title="Remove agent"
  >
    <svg class="w-4 h-4 text-red-400" fill="none" stroke="currentColor" viewBox="0 0 24 24">
      <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M19 7l-.867 12.142A2 2 0 0116.138 21H7.862a2 2 0 01-1.995-1.858L5 7m5 4v6m4-6v6m1-10V4a1 1 0 00-1-1h-4a1 1 0 00-1 1v3M4 7h16" />
    </svg>
  </button>
</div>