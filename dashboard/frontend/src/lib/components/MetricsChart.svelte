<script lang="ts">
  import { onMount, onDestroy } from 'svelte';
  import { Chart, registerables } from 'chart.js';
  
  Chart.register(...registerables);

  export let agentId: string;
  export let measurement: string;
  export let field: string;
  export let title: string;
  export let color: string = '#10b981';
  export let unit: string = '%';

  let canvas: HTMLCanvasElement;
  let chart: Chart | null = null;
  let interval: number;

  async function fetchData() {
    try {
      const response = await fetch(
        `http://host.docker.internal:8080/api/history/${agentId}/${measurement}?duration=1h`
      );
      const data = await response.json();

      if (!data || data.length === 0) return;

      // Filter for the specific field we want
      const fieldData = data.filter((record: any) => record._field === field);

      // Extract timestamps and values
      const labels = fieldData.map((record: any) => {
        const date = new Date(record.time);
        return date.toLocaleTimeString();
      });

      const values = fieldData.map((record: any) => record._value);

      updateChart(labels, values);
    } catch (error) {
      console.error('Failed to fetch chart data:', error);
    }
  }

  function updateChart(labels: string[], values: number[]) {
    if (!canvas) return;

    if (chart) {
      chart.data.labels = labels;
      chart.data.datasets[0].data = values;
      chart.update('none'); // Update without animation for performance
    } else {
      chart = new Chart(canvas, {
        type: 'line',
        data: {
          labels,
          datasets: [
            {
              label: title,
              data: values,
              borderColor: color,
              backgroundColor: `${color}20`,
              borderWidth: 2,
              fill: true,
              tension: 0.4,
              pointRadius: 0,
              pointHoverRadius: 4,
            },
          ],
        },
        options: {
          responsive: true,
          maintainAspectRatio: false,
          interaction: {
            intersect: false,
            mode: 'index',
          },
          plugins: {
            legend: {
              display: false,
            },
            tooltip: {
              backgroundColor: 'rgba(0, 0, 0, 0.8)',
              padding: 12,
              callbacks: {
                label: function(context) {
                  return `${context.parsed.y.toFixed(2)}${unit}`;
                }
              }
            },
          },
          scales: {
            x: {
              grid: {
                color: 'rgba(255, 255, 255, 0.1)',
              },
              ticks: {
                color: 'rgba(255, 255, 255, 0.6)',
                maxRotation: 0,
                autoSkipPadding: 20,
              },
            },
            y: {
              grid: {
                color: 'rgba(255, 255, 255, 0.1)',
              },
              ticks: {
                color: 'rgba(255, 255, 255, 0.6)',
                callback: function(value) {
                  return value + unit;
                }
              },
            },
          },
        },
      });
    }
  }

  onMount(() => {
    fetchData();
    // Refresh every 30 seconds
    interval = setInterval(fetchData, 30000);
  });

  onDestroy(() => {
    if (interval) clearInterval(interval);
    if (chart) chart.destroy();
  });
</script>

<div class="chart-container">
  <h4 class="text-sm font-semibold mb-2 text-gray-300">{title}</h4>
  <div class="chart-wrapper">
    <canvas bind:this={canvas}></canvas>
  </div>
</div>

<style>
  .chart-container {
    width: 100%;
  }

  .chart-wrapper {
    position: relative;
    height: 200px;
    width: 100%;
  }
</style>