<script lang="ts">
  let currentView = 'welcome';
</script>

<div class="app">
  <nav>
    <div class="logo">🐾 RabiesProtocol</div>
    <div class="menu">
      <button class:active={currentView === 'welcome'} on:click={() => currentView = 'welcome'}>
        Главная
      </button>
      <button class:active={currentView === 'protocols'} on:click={() => currentView = 'protocols'}>
        Протоколы
      </button>
      <button class:active={currentView === 'settings'} on:click={() => currentView = 'settings'}>
        Настройки
      </button>
    </div>
  </nav>

  <main>
    {#if currentView === 'welcome'}
      <div class="welcome-card">
        <h2>Добро пожаловать в RabiesProtocol!</h2>
        <p>Система для оформления протоколов осмотра и обоснованных отказов в исследовании на бешенство.</p>
        <button class="primary-btn" on:click={() => currentView = 'protocols'}>
          Создать новый протокол
        </button>
      </div>
    {:else if currentView === 'protocols'}
      <div class="placeholder">
        <h2>Список протоколов</h2>
        <p>Здесь будет мастер создания протокола и таблица сохраненных записей.</p>
        <p><em>(Этот функционал мы добавим следующим шагом)</em></p>
      </div>
    {:else if currentView === 'settings'}
      <div class="placeholder">
        <h2>Настройки</h2>
        <p>Здесь можно будет изменить данные клиники или добавить новых врачей.</p>
      </div>
    {/if}
  </main>
</div>

<style>
  .app { display: flex; flex-direction: column; height: 100vh; }
  nav {
    display: flex;
    align-items: center;
    justify-content: space-between;
    background: #2c5282;
    color: white;
    padding: 0 24px;
    height: 60px;
  }
  .logo { font-size: 18px; font-weight: bold; }
  .menu { display: flex; gap: 8px; }
  .menu button {
    background: none;
    color: rgba(255,255,255,0.8);
    border: none;
    padding: 8px 16px;
    cursor: pointer;
    font-size: 14px;
    border-radius: 4px;
    transition: all 0.2s;
  }
  .menu button:hover { background: rgba(255,255,255,0.1); color: white; }
  .menu button.active { background: rgba(255,255,255,0.2); color: white; font-weight: 600; }
  
  main { flex: 1; padding: 32px; overflow: auto; }
  .welcome-card {
    background: white;
    padding: 40px;
    border-radius: 8px;
    box-shadow: 0 2px 8px rgba(0,0,0,0.05);
    text-align: center;
    max-width: 600px;
    margin: 0 auto;
  }
  .welcome-card h2 { color: #2c5282; margin-top: 0; }
  .primary-btn {
    margin-top: 24px;
    padding: 12px 24px;
    background: #38a169;
    color: white;
    border: none;
    border-radius: 6px;
    font-size: 16px;
    cursor: pointer;
    font-weight: 600;
  }
  .primary-btn:hover { background: #2f855a; }
  .placeholder {
    background: white;
    padding: 40px;
    border-radius: 8px;
    text-align: center;
    color: #718096;
  }
</style>