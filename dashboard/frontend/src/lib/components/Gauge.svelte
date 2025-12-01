<script lang="ts">
  export let value: number; // 0-100
  export let label: string;
  export let warningThreshold: number = 85;

  $: color = value >= warningThreshold ? '#fbbf24' : '#10b981'; // amber or emerald
  $: dashArray = (value / 100) * 157; // Half circle circumference (π * 50)
</script>

<div class="flex flex-col items-center">
  <!-- Label at top -->
  <div class="text-xs text-gray-500 uppercase tracking-wider mb-2">{label}</div>

  <!-- Gauge -->
  <div class="relative w-full">
    <!-- Gauge SVG -->
    <svg class="w-full h-auto" viewBox="0 0 120 70" style="display: block;">
      <!-- Background arc (bottom half circle) -->
      <path
        d="M 20 60 A 40 40 0 0 1 100 60"
        fill="none"
        stroke="#1a1a1a"
        stroke-width="10"
        stroke-linecap="round"
      />
      
      <!-- Value arc -->
      <path
        d="M 20 60 A 40 40 0 0 1 100 60"
        fill="none"
        stroke={color}
        stroke-width="10"
        stroke-linecap="round"
        stroke-dasharray="{dashArray} 157"
        class="transition-all duration-500 ease-out"
      />
    </svg>

    <!-- Center value -->
    <div class="absolute inset-x-0 bottom-2 flex flex-col items-center">
      <div class="text-3xl font-mono font-medium" style="color: {color}">
        {value.toFixed(0)}
      </div>
      <div class="text-xs text-gray-500 mt-0.5">%</div>
    </div>
  </div>
</div>