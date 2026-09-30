import { HttpClient, HttpParams } from '@angular/common/http';
import { Injectable, inject } from '@angular/core';
import { firstValueFrom } from 'rxjs';

/** Una cuenta, tal y como la cuenta el backend (`docs/modules/users.md`, sección 2). */
export interface Cuenta {
  readonly id: number;
  readonly name: string;
  readonly lastName: string;
  readonly email: string;
  readonly role: string;
  readonly origin: string;
  readonly language: string;
  readonly isActive: boolean;
  /** Si tiene contraseña propia. En las cuentas de directorio es falsa: la suya no es nuestra. */
  readonly hasPassword: boolean;
  readonly externalId?: string;
  /** Cuándo entró por última vez, en ISO. Vacío quiere decir que no ha entrado nunca. */
  readonly lastLoginAt?: string;
}

/** Una página de cuentas, con lo que necesita la paginación. */
export interface PaginaDeCuentas {
  readonly users: readonly Cuenta[];
  readonly total: number;
  readonly page: number;
  readonly perPage: number;
}

/** Lo que contestan el alta y las acciones: la cuenta, y lo que conviene saber de lo que pasó. */
export interface RespuestaDeCuenta {
  readonly user: Cuenta;
  /** Si se le ha lanzado el correo del enlace. */
  readonly invited?: boolean;
  /** Avisos sobre algo que **sí se ha hecho**, como claves que la pantalla traduce. */
  readonly warnings?: readonly string[];
}

/** Los filtros de la lista. Vacío quiere decir «sin filtrar». */
export interface FiltrosDeCuentas {
  readonly role: string;
  readonly origin: string;
  /** `''` todas, `'true'` activas, `'false'` desactivadas. */
  readonly active: string;
  readonly q: string;
  readonly page: number;
}

/** Los filtros de partida: la lista entera, por la primera página. */
export const FILTROS_VACIOS: FiltrosDeCuentas = { role: '', origin: '', active: '', q: '', page: 1 };

/** Lo que hace falta para dar de alta una cuenta. */
export interface AltaDeCuenta {
  readonly name: string;
  readonly lastName: string;
  readonly email: string;
  readonly role: string;
  readonly origin: string;
  /** Vacío quiere decir **el idioma de la instalación**, que es el que pone el backend. */
  readonly language: string;
}

/** Lo que un Administrador puede cambiar de una cuenta. */
export interface CambiosDeCuenta {
  readonly name?: string;
  readonly lastName?: string;
  readonly email?: string;
  readonly role?: string;
  readonly language?: string;
}

/** Lo que cualquiera cambia de su propio perfil: su nombre, sus apellidos y su idioma. */
export interface CambiosDePerfil {
  readonly name?: string;
  readonly lastName?: string;
  readonly language?: string;
}

/**
 * El único servicio del módulo `users` del frontend.
 *
 * **Todas las llamadas del módulo pasan por aquí**, y todas van a `/api/users/**`: es la regla dura de
 * modularidad (`docs/arquitectura.md`, sección 4). Los componentes no usan `HttpClient`
 * directamente, así que la forma de la API se cambia en un solo sitio.
 *
 * Los errores **no se traducen aquí**: viajan como claves y la pantalla las traduce
 * (`docs/interfaz-y-experiencia.md`, sección 8).
 */
@Injectable({ providedIn: 'root' })
export class UsersService {
  private readonly http = inject(HttpClient);

  /** La lista, con sus filtros y su paginación. */
  listar(filtros: FiltrosDeCuentas): Promise<PaginaDeCuentas> {
    let parametros = new HttpParams().set('page', String(filtros.page));

    if (filtros.role) {
      parametros = parametros.set('role', filtros.role);
    }
    if (filtros.origin) {
      parametros = parametros.set('origin', filtros.origin);
    }
    if (filtros.active) {
      parametros = parametros.set('active', filtros.active);
    }
    if (filtros.q.trim()) {
      parametros = parametros.set('q', filtros.q.trim());
    }

    return firstValueFrom(this.http.get<PaginaDeCuentas>('/api/users', { params: parametros }));
  }

  /** La ficha de una cuenta. Sólo la puede pedir un Administrador. */
  ficha(id: number): Promise<RespuestaDeCuenta> {
    return firstValueFrom(this.http.get<RespuestaDeCuenta>(`/api/users/${id}`));
  }

  /** Da de alta una cuenta. Si es local, el enlace de alta sale en el mismo momento. */
  crear(alta: AltaDeCuenta): Promise<RespuestaDeCuenta> {
    return firstValueFrom(this.http.post<RespuestaDeCuenta>('/api/users', alta));
  }

  /** Cambia una cuenta: es de Administrador, salvo el nombre y los apellidos, que Soporte también. */
  cambiar(id: number, cambios: CambiosDeCuenta): Promise<RespuestaDeCuenta> {
    return firstValueFrom(this.http.patch<RespuestaDeCuenta>(`/api/users/${id}`, cambios));
  }

  /** Cambia el perfil propio: el nombre, los apellidos y el idioma, y nada más. */
  cambiarPerfil(cambios: CambiosDePerfil): Promise<RespuestaDeCuenta> {
    return firstValueFrom(this.http.patch<RespuestaDeCuenta>('/api/users/me', cambios));
  }

  /** Desactiva una cuenta. Devuelve, si lo hay, el aviso de lo que conviene saber. */
  desactivar(id: number): Promise<RespuestaDeCuenta> {
    return firstValueFrom(this.http.post<RespuestaDeCuenta>(`/api/users/${id}/deactivate`, {}));
  }

  /** Reactiva una cuenta. */
  activar(id: number): Promise<RespuestaDeCuenta> {
    return firstValueFrom(this.http.post<RespuestaDeCuenta>(`/api/users/${id}/activate`, {}));
  }

  /** Cambia el origen de una cuenta: es una de las dos acciones peligrosas del módulo. */
  cambiarOrigen(id: number, origin: string, externalId = ''): Promise<RespuestaDeCuenta> {
    return firstValueFrom(
      this.http.post<RespuestaDeCuenta>(`/api/users/${id}/origin`, { origin, externalId }),
    );
  }

  /** Lanza otra vez el enlace de contraseña: es la salida para quien no recibió el primero. */
  mandarEnlace(id: number): Promise<RespuestaDeCuenta> {
    return firstValueFrom(this.http.post<RespuestaDeCuenta>(`/api/users/${id}/reset-password`, {}));
  }
}
