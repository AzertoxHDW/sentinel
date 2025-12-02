<script lang="ts">
  import { onMount, onDestroy } from 'svelte';
  import { Chart, registerables } from 'chart.js';
  
  Chart.register(...registerables);

  export let agentId: string;
  export let interface_name: string;

  let canvas: HTMLCanvasElement;
  let chart: Chart | null = null;
  let interval: number;

  function formatSpeed(bytesPerSecond: number): string {
    if (bytesPerSecond < 1024) return `${bytesPerSecond.toFixed(0)} B/s`;
    if (bytesPerSecond < 1024 * 1024) return `${(bytesPerSecond / 1024).toFixed(1)} KB/s`;
    if (bytesPerSecond < 1024 * 1024 * 1024) return `${(bytesPerSecond / (1024 * 1024)).toFixed(1)} MB/s`;
    return `${(bytesPerSecond / (1024 * 1024 * 1024)).toFixed(2)} GB/s`;
  }

  async function fetchData() {
    try {
      const response = await fetch(
        `http://host.docker.internal:8080/api/history/${agentId}/network?duration=1h`
      );
      const data = await response.json();

      if (!data || data.length === 0) return;

      // Filter for this specific interface
      const interfaceData = data.filter((record: any) => 
        record.interface === interface_name
      );

      // Separate sent and received data
      const sentData = interfaceData.filter((r: any) => r._field === 'bytes_sent')
        .sort((a: any, b: any) => new Date(a.time).getTime() - new Date(b.time).getTime());
      const recvData = interfaceData.filter((r: any) => r._field === 'bytes_recv')
        .sort((a: any, b: any) => new Date(a.time).getTime() - new Date(b.time).getTime());

      if (sentData.length < 2 || recvData.length < 2) return;

      // Calculate speeds (bytes per second between data points)
      const labels: string[] = [];
      const downloadSpeeds: number[] = [];
      const uploadSpeeds: number[] = [];

      for (let i = 1; i < recvData.length; i++) {
        const timeDiff = (new Date(recvData[i].time).getTime() - new Date(recvData[i-1].time).getTime()) / 1000;
        
        if (timeDiff > 0) {
          const date = new Date(recvData[i].time);
          labels.push(date.toLocaleTimeString());
          
          const downloadSpeed = Math.max(0, (recvData[i]._value - recvData[i-1]._value) / timeDiff);
          const uploadSpeed = Math.max(0, (sentData[i]._value - sentData[i-1]._value) / timeDiff);
          
          downloadSpeeds.push(downloadSpeed);
          uploadSpeeds.push(uploadSpeed);
        }
      }

      updateChart(labels, downloadSpeeds, uploadSpeeds);
    } catch (error) {
      console.error('Failed to fetch network data:', error);
    }
  }

  function updateChart(labels: string[], downloadSpeeds: number[], uploadSpeeds: number[]) {
    if (!canvas) return;

    if (chart) {
      chart.data.labels = labels;
      chart.data.datasets[0].data = downloadSpeeds;
      chart.data.datasets[1].data = uploadSpeeds;
      chart.update('none');
    } else {
      chart = new Chart(canvas, {
        type: 'line',
        data: {
          labels,
          datasets: [
            {
              label: 'Download',
              data: downloadSpeeds,
              borderColor: '#3b82f6',
              backgroundColor: '#3b82f620',
              borderWidth: 2,
              fill: true,
              tension: 0.4,
              pointRadius: 0,
              pointHoverRadius: 4,
            },
            {
              label: 'Upload',
              data: uploadSpeeds,
              borderColor: '#10b981',
              backgroundColor: '#10b98120',
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
              display: true,
              position: 'top',
              labels: {
                color: 'rgba(255, 255, 255, 0.6)',
                usePointStyle: true,
                padding: 15,
              }
            },
            tooltip: {
              backgroundColor: 'rgba(0, 0, 0, 0.8)',
              padding: 12,
              callbacks: {
                label: function(context) {
                  return `${context.dataset.label}: ${formatSpeed(context.parsed.y)}`;
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
                  return formatSpeed(Number(value));
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
    interval = setInterval(fetchData, 30000);
  });

  onDestroy(() => {
    if (interval) clearInterval(interval);
    if (chart) chart.destroy();
  });
</script>

<div class="chart-container">
  <h4 class="text-sm font-semibold mb-2 text-gray-300">{interface_name}</h4>
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