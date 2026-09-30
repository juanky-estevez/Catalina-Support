import { provideHttpClient } from '@angular/common/http';
import { HttpTestingController, provideHttpClientTesting } from '@angular/common/http/testing';
import { type ComponentFixture, TestBed } from '@angular/core/testing';
import { type WritableSignal, signal } from '@angular/core';
import { provideRouter } from '@angular/router';

import { TranslationService } from '../../core/i18n/translation.service';
import { SessionService, type Usuario } from '../../core/services/session.service';
import { BandejaPage } from './bandeja-page';
import { type PaginaDeTickets, type Ticket } from './tickets.service';

/**
 * **La tabla de resumen de las bandejas** (decisión 86) y **el refresco de los dos resúmenes del motor**
 * (enmienda del 2026-09-29): el número va solo, la categoría y las etiquetas viven juntas en
 * «Clasificación», el estado sigue en la suya, y la lista **se vuelve a leer sola** mientras algo siga
 * pendiente, con un tope de intentos.
 *
 * El aspecto en un navegador lo cubre la capa de Playwright; aquí se prueba lo que la pantalla enseña y
 * lo que pide.
 */

const ANA = {
  id: 9,
  name: 'Ana',
  lastName: 'Ruiz',
  email: 'ana@demo.com',
  role: 'usuario',
};

/** Un ticket ya resumido: el motor no está pendiente, así que no se programa ninguna relectura. */
const CON_CLASIFICACION: Ticket = {
  number: 'CS-2026-0001',
  internal: false,
  subject: 'No va la red',
  description: '',
  state: 'nuevo',
  category: { id: 1, name: 'Red', active: true },
  tags: ['red-wifi', 'urgente'],
  requester: ANA,
  insights: {
    motivo: { state: 'listo', es: 'No va la red', en: 'Network is down' },
    ultimaAccion: { state: 'listo', es: 'Se avisó', en: 'Notified' },
  },
  createdAt: '2026-09-28T08:00:00Z',
  updatedAt: '2026-09-28T08:00:00Z',
};

const PAGINA: PaginaDeTickets = { tickets: [CON_CLASIFICACION], total: 1, page: 1, perPage: 20 };

/** El mismo ticket, pero con el motor todavía trabajando: es lo que dispara la relectura. */
const PAGINA_PENDIENTE: PaginaDeTickets = {
  ...PAGINA,
  tickets: [
    {
      ...CON_CLASIFICACION,
      insights: {
        motivo: { state: 'pendiente' },
        ultimaAccion: { state: 'pendiente' },
      },
    },
  ],
};

const SIN_CATEGORIAS = { categories: [] };

describe('BandejaPage', () => {
  let http: HttpTestingController;
  let usuario: WritableSignal<Usuario | null>;

  beforeEach(async () => {
    localStorage.clear();
    usuario = signal<Usuario | null>(null);

    await TestBed.configureTestingModule({
      imports: [BandejaPage],
      providers: [
        provideHttpClient(),
        provideHttpClientTesting(),
        provideRouter([]),
        // La sesión se sustituye por lo único que mira la pantalla: quién ha entrado y su papel.
        { provide: SessionService, useValue: { usuario, esAdministrador: signal(false) } },
      ],
    }).compileComponents();

    TestBed.inject(TranslationService).cambiar('es');
    http = TestBed.inject(HttpTestingController);
  });

  afterEach(() => {
    http.verify();
    localStorage.clear();
  });

  /** Deja correr las promesas del servicio: el backend de prueba contesta en el acto. */
  function esperar(): Promise<void> {
    return new Promise((resolve) => setTimeout(resolve));
  }

  /** Monta la pantalla con un usuario normal y responde la lista y el catálogo. */
  async function montar(pagina: PaginaDeTickets = PAGINA): Promise<ComponentFixture<BandejaPage>> {
    usuario.set({
      id: 1,
      name: 'María',
      lastName: 'Pérez',
      email: 'maria@demo.com',
      role: 'usuario',
      origin: 'local',
      language: 'es',
      factory: false,
    });

    const fixture = TestBed.createComponent(BandejaPage);
    fixture.detectChanges();

    http.expectOne((peticion) => peticion.url === '/api/tickets').flush(pagina);
    http.expectOne('/api/tickets/categories').flush(SIN_CATEGORIAS);
    await esperar();
    fixture.detectChanges();

    return fixture;
  }

  /** Las cabeceras de la tabla, en orden. */
  function encabezados(host: HTMLElement): (string | undefined)[] {
    return [...host.querySelectorAll('thead th')].map((celda) => celda.textContent?.trim());
  }

  /** La fila de la tabla. */
  function celdasDeLaFila(host: HTMLElement): HTMLElement[] {
    return [...host.querySelectorAll('tbody tr:first-child td')] as HTMLElement[];
  }

  it('la columna del número no lleva chips; la clasificación los lleva, y el estado sigue en la suya', async () => {
    const fixture = await montar();
    const host = fixture.nativeElement as HTMLElement;

    // Las columnas, en el orden del documento: Ticket · Asunto · Motivo · Última acción ·
    // Clasificación · Estado · …
    expect(encabezados(host).slice(0, 6)).toEqual([
      'Ticket',
      'Asunto',
      'Motivo',
      'Última acción',
      'Clasificación',
      'Estado',
    ]);

    const celdas = celdasDeLaFila(host);

    // **El número solo**: ni la categoría ni las etiquetas van debajo.
    expect(celdas[0].textContent?.trim()).toBe('CS-2026-0001');
    expect(celdas[0].querySelector('span')).toBeNull();
    expect(celdas[0].textContent).not.toContain('Red');
    expect(celdas[0].textContent).not.toContain('red-wifi');

    // **La clasificación los agrupa**: el chip de la categoría y el de cada etiqueta.
    const clasificacion = celdas[4];
    expect(clasificacion.textContent).toContain('Red');
    expect(clasificacion.textContent).toContain('red-wifi');
    expect(clasificacion.textContent).toContain('urgente');

    // **El estado sigue en su columna**, con su chip.
    const estado = celdas[5];
    expect(estado.querySelector('app-estado')).not.toBeNull();
    // Y no se ha colado en la clasificación.
    expect(clasificacion.querySelector('app-estado')).toBeNull();
  });

  it('vuelve a leer sola mientras algo esté pendiente, y para al llegar al tope', async () => {
    vi.useFakeTimers();
    try {
      usuario.set({
        id: 1,
        name: 'María',
        lastName: 'Pérez',
        email: 'maria@demo.com',
        role: 'usuario',
        origin: 'local',
        language: 'es',
        factory: false,
      });

      const fixture = TestBed.createComponent(BandejaPage);
      fixture.detectChanges();

      http.expectOne((peticion) => peticion.url === '/api/tickets').flush(PAGINA_PENDIENTE);
      http.expectOne('/api/tickets/categories').flush(SIN_CATEGORIAS);
      await vi.advanceTimersByTimeAsync(0);

      // Recién cargada, todavía no ha releído: la espera del motor aún no ha pasado.
      http.expectNone((peticion) => peticion.url === '/api/tickets');

      // **Ocho relecturas** (el tope), cada una con el motor todavía pendiente.
      for (let intento = 0; intento < 8; intento += 1) {
        await vi.advanceTimersByTimeAsync(8000);
        http.expectOne((peticion) => peticion.url === '/api/tickets').flush(PAGINA_PENDIENTE);
        await vi.advanceTimersByTimeAsync(0);
      }

      // Al llegar al tope **se para**: por mucho que se espere, no vuelve a pedir la lista.
      await vi.advanceTimersByTimeAsync(80_000);
      http.expectNone((peticion) => peticion.url === '/api/tickets');

      fixture.destroy();
    } finally {
      vi.useRealTimers();
    }
  });

  it('deja de releer en cuanto el motor termina', async () => {
    vi.useFakeTimers();
    try {
      usuario.set({
        id: 1,
        name: 'María',
        lastName: 'Pérez',
        email: 'maria@demo.com',
        role: 'usuario',
        origin: 'local',
        language: 'es',
        factory: false,
      });

      const fixture = TestBed.createComponent(BandejaPage);
      fixture.detectChanges();

      http.expectOne((peticion) => peticion.url === '/api/tickets').flush(PAGINA_PENDIENTE);
      http.expectOne('/api/tickets/categories').flush(SIN_CATEGORIAS);
      await vi.advanceTimersByTimeAsync(0);

      // La primera relectura ya encuentra los resúmenes listos.
      await vi.advanceTimersByTimeAsync(8000);
      http.expectOne((peticion) => peticion.url === '/api/tickets').flush(PAGINA);
      await vi.advanceTimersByTimeAsync(0);

      // Y no hay una segunda.
      await vi.advanceTimersByTimeAsync(80_000);
      http.expectNone((peticion) => peticion.url === '/api/tickets');

      fixture.destroy();
    } finally {
      vi.useRealTimers();
    }
  });
});
