import { ApplicationConfig, provideBrowserGlobalErrorListeners } from '@angular/core';
import { provideHttpClient, withInterceptors } from '@angular/common/http';
import { TitleStrategy, provideRouter, withComponentInputBinding } from '@angular/router';

import { routes } from './app.routes';
import { apiInterceptor } from './core/interceptors/api.interceptor';
import { TituloDeLaInstalacion } from './core/services/title.strategy';

export const appConfig: ApplicationConfig = {
  providers: [
    provideBrowserGlobalErrorListeners(),
    // El frontend llama SIEMPRE a rutas relativas (/api/**): nginx es quien enruta al
    // backend, así que no hay ninguna URL que configurar ni CORS que resolver
    // (docs/arquitectura.md, secciones 4 y 9).
    //
    // El interceptor es el único sitio donde se pone la cabecera de la sesión y donde se atiende
    // el 401: las pantallas no saben que existe un token (docs/modules/auth.md, sección 9).
    provideHttpClient(withInterceptors([apiInterceptor])),
    provideRouter(routes, withComponentInputBinding()),
    // La pestaña del navegador se llama como la instalación, no como el producto de fábrica
    // (docs/modules/settings.md).
    { provide: TitleStrategy, useClass: TituloDeLaInstalacion },
  ],
};
