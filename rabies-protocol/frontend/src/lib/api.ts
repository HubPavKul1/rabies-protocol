// Обертка над Wails bindings для удобства вызова из Svelte
export async function isConfigured(): Promise<boolean> {
  return (window as any).go.main.App.IsConfigured();
}

export async function getClinicConfig(): Promise<any> {
  return (window as any).go.main.App.GetClinicConfig();
}

export async function saveClinicConfig(cfg: any): Promise<void> {
  return (window as any).go.main.App.SaveClinicConfig(cfg);
}

export async function listVeterinarians(): Promise<any[]> {
  return (window as any).go.main.App.ListVeterinarians();
}

export async function createVeterinarian(v: any): Promise<void> {
  return (window as any).go.main.App.CreateVeterinarian(v);
}