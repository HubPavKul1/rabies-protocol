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
      // Функция toPlainObject внутри api.ts сама очистит данные от прокси Svelte
      await saveClinicConfig(clinic);
      await createVeterinarian({ full_name: vetName, position: vetPosition });
      
      dispatch('configured');
    } catch (e) {
      console.error("Ошибка при сохранении:", e);
      alert("Ошибка сохранения: " + e);
    } finally {
      saving = false;
    }
  }
</script>