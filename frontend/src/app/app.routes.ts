import { Routes } from '@angular/router';

import { Armazon } from './core/components/armazon';
import { instalacionComprobada, instalacionSinSellar } from './core/guards/instalacion.guard';
import { conCuentaPropia, conPapel, inicioSegunPapel, sesionActiva } from './core/guards/sesion.guard';

/**
 * Las rutas, en dos grupos.
 *
 * **Las pantallas de la sesión van a pantalla completa** —entrar, olvidar y establecer la
 * contraseña—: no hay armazón, porque todavía no hay nadie dentro. Todo lo demás va **dentro del
 * armazón**, con su menú lateral.
 *
 * **Las rutas van en inglés y en minúsculas**, aunque el idioma del producto sea español
 * (docs/modules/auth.md, decisión 21). Cada pantalla se carga con `loadComponent`, así que la que no
 * se visita no se descarga (docs/arquitectura.md, sección 6).
 *
 * Y por encima de todo está **el primer arranque**: mientras la instalación no esté sellada, no hay
 * puerta por la que entrar y lo único que tiene sentido es configurarla. Por eso todo lo demás vive
 * bajo la guarda `instalacionComprobada`, y `/setup` vive fuera, con la suya en la otra dirección
 * (`docs/primer-arranque.md`, secciones 2 y 6).
 */
export const routes: Routes = [
  // --- El asistente de primer arranque, sin armazón: no hay nada que navegar ---
  {
    path: 'setup',
    canActivate: [instalacionSinSellar],
    loadComponent: () => import('./core/pages/setup-page').then((m) => m.SetupPage),
  },

  // --- Todo lo demás: si la instalación no está sellada, esto lleva a /setup ---
  {
    path: '',
    canActivate: [instalacionComprobada],
    children: [
      // --- A pantalla completa: aquí todavía no hay sesión ---
      {
        path: 'login',
        loadComponent: () => import('./core/pages/login-page').then((m) => m.LoginPage),
      },
      {
        path: 'forgot-password',
        loadComponent: () =>
          import('./core/pages/forgot-password-page').then((m) => m.ForgotPasswordPage),
      },
      {
        // La misma pantalla sirve para el enlace del alta y el de la recuperación: en los dos casos se
        // llega sin saber la contraseña (docs/modules/auth.md, decisión 23).
        path: 'set-password',
        loadComponent: () =>
          import('./core/pages/set-password-page').then((m) => m.SetPasswordPage),
      },

      // --- Dentro del armazón: hace falta haber entrado ---
      {
        path: '',
        component: Armazon,
        canActivate: [sesionActiva],
        children: [
          {
            // La raíz **no es una pantalla**: reparte a cada papel a la suya. El inicio provisional
            // desapareció con las pantallas de producto (docs/interfaz-y-experiencia.md, sección 4.4).
            path: '',
            canActivate: [inicioSegunPapel],
            children: [],
          },
          {
            path: 'change-password',
            loadComponent: () =>
              import('./core/pages/change-password-page').then((m) => m.ChangePasswordPage),
          },
          {
            path: 'forbidden',
            loadComponent: () =>
              import('./core/pages/forbidden-page').then((m) => m.ForbiddenPage),
          },
          {
            // La bandeja: **«lo mío»** para el usuario, Soporte y Desarrollo —asignado, abierto por uno o
            // comentado—, y los dos tipos para el Administrador, de sólo lectura. Es la misma pantalla con
            // el nombre que le toca a cada uno (docs/interfaz-y-experiencia.md, secciones 3.2, 3.3 y 3.7).
            path: 'tickets',
            data: { listado: 'mios' },
            loadComponent: () => import('./modules/tickets/bandeja-page').then((m) => m.BandejaPage),
          },
          {
            // Las dos listas del «todo», sólo para Soporte y Desarrollo, que son los dos papeles que
            // pueden verlo todo (decisión 52 de docs/modules/tickets.md). **Van antes que `tickets/:numero`**:
            // si no, `main` se leería como el número de un ticket.
            path: 'tickets/main',
            data: { listado: 'principal' },
            canActivate: [conPapel('soporte', 'desarrollo')],
            loadComponent: () => import('./modules/tickets/bandeja-page').then((m) => m.BandejaPage),
          },
          {
            path: 'tickets/internal',
            data: { listado: 'interno' },
            canActivate: [conPapel('soporte', 'desarrollo')],
            loadComponent: () => import('./modules/tickets/bandeja-page').then((m) => m.BandejaPage),
          },
          {
            // **El catálogo de categorías y etiquetas**: una pantalla del módulo `tickets` y no una sección
            // de Configuración (decisión 64). La ven Soporte, que lo mantiene, y el Administrador, que
            // además retira. **Va antes que `tickets/:numero`**: si no, `categories` se leería como el número
            // de un ticket y el detalle contestaría «ese ticket no existe».
            path: 'tickets/categories',
            canActivate: [conPapel('soporte', 'administrador')],
            loadComponent: () =>
              import('./modules/tickets/categorias-page').then((m) => m.CategoriasPage),
          },
          {
            // El alta: usuario, Soporte y Desarrollo crean tickets (Soporte también en nombre de otro).
            // **No tiene entrada en el menú** desde el 2026-09-26: se entra por el botón de la bandeja.
            path: 'tickets/new',
            canActivate: [conPapel('usuario', 'soporte', 'desarrollo')],
            loadComponent: () =>
              import('./modules/tickets/nuevo-ticket-page').then((m) => m.NuevoTicketPage),
          },
          {
            // El detalle, y **el de un interno es el mismo**: se lee por su número, con el conmutador de la
            // vista doble cuando el ticket tiene interno (docs/interfaz-y-experiencia.md, secciones 3.4 y 5).
            path: 'tickets/:numero',
            loadComponent: () => import('./modules/tickets/ticket-page').then((m) => m.TicketPage),
          },
          {
            // El perfil propio: lo cambia cualquiera, salvo la cuenta de fábrica, que no tiene perfil
            // (docs/modules/users.md, sección 8).
            path: 'profile',
            canActivate: [conCuentaPropia],
            loadComponent: () => import('./modules/users/profile-page').then((m) => m.ProfilePage),
          },
          {
            // El módulo de usuarios: la lista la ven Soporte y Administrador.
            path: 'users',
            canActivate: [conPapel('soporte', 'administrador')],
            loadComponent: () => import('./modules/users/users-page').then((m) => m.UsersPage),
          },
          {
            // La ficha de una cuenta: **sólo un Administrador** (docs/modules/users.md, decisión 2).
            path: 'users/:id',
            canActivate: [conPapel('administrador')],
            loadComponent: () => import('./modules/users/user-page').then((m) => m.UserPage),
          },
          {
            // El editor de los correos: **una pantalla del módulo `mail`**, a la que Configuración enlaza.
            // Llamar a `/api/mail/**` desde `settings` rompería la regla dura de modularidad
            // (docs/modules/mail.md, decisión 15).
            path: 'mail',
            canActivate: [conPapel('administrador')],
            loadComponent: () => import('./modules/mail/correos-page').then((m) => m.CorreosPage),
          },
          {
            // La configuración de la instalación: **sólo un Administrador**
            // (docs/modules/settings.md, sección 6).
            path: 'settings',
            canActivate: [conPapel('administrador')],
            loadComponent: () => import('./core/pages/settings-page').then((m) => m.SettingsPage),
          },
        ],
      },
    ],
  },

  // Cualquier dirección que no exista lleva al inicio, que ya decide si hay que entrar.
  { path: '**', redirectTo: '' },
];
