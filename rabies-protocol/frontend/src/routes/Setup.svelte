<script lang="ts">
  import { createEventDispatcher } from 'svelte';
  import { saveClinicConfig, createVeterinarian } from '../lib/api';

  const dispatch = createEventDispatcher();

  let clinic = { name: '', address: '', phone: '', email: '', license: '' };
  let vetName = '';
  let vetPosition = 'Ветеринарный врач';
  let saving = false;

  async function handleSave() {
    saving = true;
    try {
      await saveClinicConfig(clinic);
      await createVeterinarian({ full_name: vetName, position: vetPosition });
      dispatch('configured');
    } catch (e) {
      alert("Ошибка сохранения: " + e);
    } finally {
      saving = false;
    }
  }
</script>

<div class="setup">
  <h1>🏥 Первичная настройка RabiesProtocol</h1>
  <p>Пожалуйста, введите данные вашей клиники и ветеринарного врача.</p>

  <section>
    <h2>Сведения о клинике</h2>
    <label>Название клиники *
      <input bind:value={clinic.name} required placeholder="ООО «ВетКлиника»" />
    </label>
    <label>Юридический адрес *
      <input bind:value={clinic.address} required />
    </label>
    <label>Контактный телефон
      <input bind:value={clinic.phone} placeholder="+7 (XXX) XXX-XX-XX" />
    </label>
  </section>

  <section>
    <h2>Ветеринарный врач</h2>
    <label>ФИО полностью *
      <input bind:value={vetName} required placeholder="Иванов Иван Иванович" />
    </label>
    <label>Должность
      <input bind:value={vetPosition} />
    </label>
  </section>

  <button on:click={handleSave} disabled={!clinic.name || !clinic.address || !vetName || saving}>
    {saving ? 'Сохранение...' : 'Сохранить и начать работу'}
  </button>
</div>

<style>
  .setup {
    max-width: 600px;
    margin: 40px auto;
    padding: 30px;
    background: white;
    border-radius: 8px;
    box-shadow: 0 4px 12px rgba(0,0,0,0.08);
  }
  h1 { margin-top: 0; color: #2c5282; font-size: 24px; }
  p { color: #666; margin-bottom: 24px; }
  section { margin: 24px 0; padding-top: 16px; border-top: 1px solid #eee; }
  h2 { font-size: 18px; color: #2d3748; margin-bottom: 16px; }
  label {
    display: block;
    margin: 12px 0;
    font-size: 14px;
    font-weight: 500;
  }
  input {
    display: block;
    width: 100%;
    margin-top: 6px;
    padding: 10px;
    border: 1px solid #cbd5e0;
    border-radius: 6px;
    font-size: 14px;
    box-sizing: border-box;
    transition: border-color 0.2s;
  }
  input:focus {
    outline: none;
    border-color: #4299e1;
    box-shadow: 0 0 0 3px rgba(66, 153, 225, 0.1);
  }
  button {
    width: 100%;
    padding: 12px;
    background: #2c5282;
    color: white;
    border: none;
    border-radius: 6px;
    font-size: 16px;
    font-weight: 600;
    cursor: pointer;
    margin-top: 24px;
    transition: background 0.2s;
  }
  button:hover:not(:disabled) { background: #2a4365; }
  button:disabled { background: #a0aec0; cursor: not-allowed; }
</style>