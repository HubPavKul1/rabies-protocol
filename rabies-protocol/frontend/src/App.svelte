<script lang="ts">
  import { onMount } from 'svelte';
  import { isConfigured } from './lib/api';
  import Setup from './routes/Setup.svelte';
  import Dashboard from './routes/Dashboard.svelte';

  let configured = false;
  let loading = true;

  onMount(async () => {
    try {
      configured = await isConfigured();
    } catch (e) {
      console.error("Ошибка проверки конфигурации:", e);
    }
    loading = false;
  });
</script>

{#if loading}
  <div class="loading">Загрузка приложения...</div>
{:else if !configured}
  <Setup on:configured={() => configured = true} />
{:else}
  <Dashboard />
{/if}

<style>
  :global(body) {
    margin: 0;
    font-family: 'Segoe UI', Tahoma, sans-serif;
    background: #f5f7fa;
    color: #333;
  }
  .loading {
    display: flex;
    align-items: center;
    justify-content: center;
    height: 100vh;
    font-size: 18px;
    color: #2c5282;
  }
</style>