<script lang="ts">
  import { onMount } from 'svelte';
  import Setup from './routes/Setup.svelte';
  import Dashboard from './routes/Dashboard.svelte';
  import { isConfigured } from './lib/api';

  let configured = false;
  let loading = true;
  let errorMsg = '';

  onMount(async () => {
    try {
      // Пытаемся проверить БД. Если ошибка - ловим её и всё равно показываем Setup
      configured = await isConfigured().catch(e => {
        errorMsg = "Не удалось связаться с Go-бэкендом: " + e;
        return false; 
      });
    } catch (e) {
      errorMsg = "Критическая ошибка инициализации: " + e;
    }
    loading = false;
  });
</script>

{#if errorMsg}
  <div class="error-screen">
    <h2>⚠️ ОШИБКА ПРИЛОЖЕНИЯ</h2>
    <p>{errorMsg}</p>
    <p><strong>Что делать:</strong></p>
    <ol>
      <li>Закройте приложение.</li>
      <li>Удалите папку <code>data</code> в папке с программой.</li>
      <li>Запустите <code>run.bat</code> заново.</li>
    </ol>
    <p style="font-size: 12px; color: #666;">Нажмите F12, чтобы увидеть подробности в консоли.</p>
  </div>
{:else if loading}
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
  .error-screen {
    padding: 40px;
    max-width: 600px;
    margin: 40px auto;
    background: #fff5f5;
    border: 1px solid #fc8181;
    border-radius: 8px;
    color: #c53030;
  }
  .error-screen h2 { margin-top: 0; }
  .error-screen code { background: #fed7d7; padding: 2px 4px; border-radius: 4px; }
</style>