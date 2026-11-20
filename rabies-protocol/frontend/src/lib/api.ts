// ВАЖНО: путь содержит 'wails', так как это имя пакета в Go
import { 
    IsConfigured, 
    GetClinicConfig, 
    SaveClinicConfig, 
    ListVeterinarians, 
    CreateVeterinarian 
} from '../../wailsjs/go/wails/App';

// Гарантированно удаляет реактивные прокси Svelte, превращая объект в обычный JSON
function toPlainObject(obj: any) {
    return JSON.parse(JSON.stringify(obj));
}

export async function isConfigured(): Promise<boolean> {
    return await IsConfigured();
}

export async function getClinicConfig(): Promise<any> {
    return await GetClinicConfig();
}

export async function saveClinicConfig(cfg: any): Promise<void> {
    await SaveClinicConfig(toPlainObject(cfg));
}

export async function listVeterinarians(): Promise<any[]> {
    return await ListVeterinarians();
}

export async function createVeterinarian(v: any): Promise<void> {
    await CreateVeterinarian(toPlainObject(v));
}