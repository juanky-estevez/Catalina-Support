import { provideHttpClient } from '@angular/common/http';
import { HttpTestingController, provideHttpClientTesting } from '@angular/common/http/testing';
import { type ComponentFixture, TestBed } from '@angular/core/testing';
import { type WritableSignal, signal } from '@angular/core';
import { Router, provideRouter } from '@angular/router';
import { vi } from 'vitest';

import { TranslationService } from '../../core/i18n/translation.service';
import { SessionService, type Usuario } from '../../core/services/session.service';
import { TicketPage } from './ticket-page';
import { type DetalleDeTicket } from './tickets.service';

/**
 * **Dónde vive el desplegable de acciones** (decisión 78), que es lo que se prueba aquí: con un solo
 * ticket a la vista hay **un control** en la fila de arriba; con los dos, **uno por columna**, porque
 * cada ticket tiene su estado y desde un solo control no se sabría a cuál se le cambia.
 *
 * El resto de la pantalla —lo que manda cada control, la vista doble— se prueba en las otras suites.
 */

const ANA = {
  id: 9,
  name: 'Ana',
  lastName: 'Ruiz',
  email: 'ana@demo.com',
  role: 'usuario',
};

const PRINCIPAL: DetalleDeTicket = {
  ticket: {
    number: 'CS-2026-0001',
    internal: false,
    subject: 'No va la red',
    description: '',
    state: 'nuevo',
    child: 'INT-CS-2026-0001',
    requester: ANA,
    createdAt: '2026-09-28T08:00:00Z',
    updatedAt: '2026-09-28T08:00:00Z',
  },
  comments: [],
  attachments: [],
  history: [],
};

const INTERNO: DetalleDeTicket = {
  ticket: {
    number: 'INT-CS-2026-0001',
    internal: true,
    subject: 'No va la red',
    description: '',
    state: 'en progreso',
    parent: 'CS-2026-0001',
    requester: ANA,
    createdAt: '2026-09-28T08:00:00Z',
    updatedAt: '2026-09-28T08:00:00Z',
  },
  comments: [],
  attachments: [],
  history: [],
};

/**
 * jsdom no trae el `<dialog>` nativo que usa `app-dialogo`. La pantalla monta dos —el de borrar un
 * comentario y el visor—, así que basta con que abrir y cerrar no revienten.
 */
beforeAll(() => {
  HTMLDialogElement.prototype.showModal = function (this: HTMLDialogElement) {
    this.open = true;
  };
  HTMLDialogElement.prototype.close = function (this: HTMLDialogElement) {
    this.open = false;
    this.dispatchEvent(new Event('close'));
  };
});

describe('TicketPage', () => {
  let http: HttpTestingController;
  let usuario: WritableSignal<Usuario | null>;

  beforeEach(async () => {
    localStorage.clear();
    usuario = signal<Usuario | null>(null);

    await TestBed.configureTestingModule({
      imports: [TicketPage],
      providers: [
        provideHttpClient(),
        provideHttpClientTesting(),
        provideRouter([]),
        // La sesión se sustituye por lo único que mira la pantalla: quién ha entrado y su papel.
        {
          provide: SessionService,
          useValue: { usuario, esAdministrador: signal(false) },
        },
      ],
    }).compileComponents();

    TestBed.inject(TranslationService).cambiar('es');
    http = TestBed.inject(HttpTestingController);
  });

  afterEach(() => {
    http.verify();
    localStorage.clear();
  });

  /** Deja correr las promesas: el backend de prueba contesta en el acto. */
  function esperar(): Promise<void> {
    return new Promise((resolve) => setTimeout(resolve));
  }

  /** Monta la pantalla con un principal que tiene interno, que es cuando existe la vista doble. */
  async function montar(): Promise<ComponentFixture<TicketPage>> {
    usuario.set({
      id: 1,
      name: 'María',
      lastName: 'Pérez',
      email: 'maria@demo.com',
      role: 'soporte',
      origin: 'local',
      factory: false,
    });

    const fixture = TestBed.createComponent(TicketPage);
    fixture.componentRef.setInput('numero', 'CS-2026-0001');
    // `ngOnInit` arranca la carga: el principal, su interno, los responsables y las categorías.
    fixture.detectChanges();

    http.expectOne('/api/tickets/CS-2026-0001').flush(PRINCIPAL);
    await esperar();
    fixture.detectChanges();

    http.expectOne('/api/tickets/INT-CS-2026-0001').flush(INTERNO);
    await esperar();
    fixture.detectChanges();

    http.expectOne('/api/tickets/assignees').flush({ main: [], internal: [] });
    await esperar();
    fixture.detectChanges();

    http.expectOne('/api/tickets/categories').flush({ categories: [] });
    await esperar();
    fixture.detectChanges();

    return fixture;
  }

  /** Pulsa el conmutador de la vista por su texto. */
  function verLosDos(host: HTMLElement, fixture: ComponentFixture<TicketPage>): void {
    const boton = [...host.querySelectorAll<HTMLButtonElement>('button')].find(
      (candidato) => candidato.textContent?.trim() === 'Los dos',
    );
    expect(boton).toBeTruthy();
    boton!.click();
    fixture.detectChanges();
  }

  it('con un solo ticket hay un desplegable, en la fila de arriba', async () => {
    const fixture = await montar();
    const host = fixture.nativeElement as HTMLElement;

    // Uno solo, y **fuera** de las vistas de ticket: vive en la cabecera de la página.
    expect(host.querySelectorAll('app-acciones-del-ticket').length).toBe(1);
    expect(host.querySelector('app-ticket-vista app-acciones-del-ticket')).toBeNull();
  });

  it('con los dos tickets hay uno por columna, en la cabecera de cada una', async () => {
    const fixture = await montar();
    const host = fixture.nativeElement as HTMLElement;

    verLosDos(host, fixture);

    const columnas = [...host.querySelectorAll('app-ticket-vista')];
    expect(columnas.length).toBe(2);

    // **Uno en cada columna**: no hay un control único de arriba del que no se sabría a quién cambia.
    for (const columna of columnas) {
      expect(columna.querySelectorAll('header app-acciones-del-ticket').length).toBe(1);
    }

    expect(host.querySelectorAll('app-acciones-del-ticket').length).toBe(2);
  });

  it('el botón de volver es un icono a la izquierda del número y sigue llevando a la bandeja', async () => {
    const fixture = await montar();
    const host = fixture.nativeElement as HTMLElement;
    const router = TestBed.inject(Router);
    const navegar = vi.spyOn(router, 'navigate').mockResolvedValue(true);

    const volver = host.querySelector<HTMLButtonElement>('button[aria-label="Volver"]');
    expect(volver).toBeTruthy();

    // **A la izquierda del número**: el botón va antes que el `h1` en la cabecera.
    const cabecera = host.querySelector('main > div');
    const primero = cabecera?.firstElementChild;
    expect(primero?.querySelector('button[aria-label="Volver"]')).toBeTruthy();
    expect(primero?.querySelector('h1')).toBeTruthy();

    volver!.click();
    fixture.detectChanges();

    expect(navegar).toHaveBeenCalledWith(['/tickets']);
  });
});
