import { ApplicationConfig, provideBrowserGlobalErrorListeners } from '@angular/core';
import { provideHttpClient } from '@angular/common/http';
import { provideRouter } from '@angular/router';
import { routes } from './app.routes';

export const appConfig: ApplicationConfig = {
  providers: [
    provideBrowserGlobalErrorListeners(),
    // El frontend llama SIEMPRE a rutas relativas (/api/**): nginx es quien enruta al
    // backend, así que no hay ninguna URL que configurar ni CORS que resolver
    // (docs/arquitectura.md, secciones 4 y 9).
    provideHttpClient(),
    provideRouter(routes),
  ],
};
