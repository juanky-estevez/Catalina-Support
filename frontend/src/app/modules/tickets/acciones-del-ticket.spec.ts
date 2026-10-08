import { provideHttpClient } from '@angular/common/http';
import { HttpTestingController, provideHttpClientTesting } from '@angular/common/http/testing';
import { type ComponentFixture, TestBed } from '@angular/core/testing';
import { type WritableSignal, signal } from '@angular/core';

import { TranslationService } from '../../core/i18n/translation.service';
import { SessionService, type Usuario } from '../../core/services/session.service';
import { AccionesDelTicket } from './acciones-del-ticket';
import { type Ticket } from './tickets.service';

/**
 * **El desplegable de estados** (corrección del responsable del 2026-09-29, decisiones 80 y 81): lo
 * que se prueba aquí es lo que el documento fija.
 *
 * · Que **enseñe estados y no acciones**, con el actual como **primera opción, marcado y sin poder
 *   elegirlo**, y **sólo los alcanzables** para ese papel y ese estado.
 * · Que elegir un estado que necesita un texto **abra su ventana y lo pida**, y que `en progreso`
 *   **sólo confirme**.
 * · Que **cerrar no deje confirmar sin comentario**, con el aviso de la clave que traduce el `422`.
 * · Que el cambio se mande al backend como `{ to, comment }`.
 * · Que **vuelva a su sitio** después de elegir: es un control de mando, no un campo con valor.
 * · Que sin nada que hacer **no se enseñe**.
 *
 * El aspecto y las etiquetas accesibles, en un navegador, los cubre la capa de Playwright.
 */

const ANA = {
  id: 9,
  name: 'Ana',
  lastName: 'Ruiz',
  email: 'ana@demo.com',
  role: 'usuario',
};

/** Un principal recién abierto: se puede empezar, escalar o cerrar. */
const NUEVO: Ticket = {
  number: 'CS-2026-0001',
  internal: false,
  subject: 'No va la red',
  description: '',
  state: 'nuevo',
  requester: ANA,
  createdAt: '2026-09-28T08:00:00Z',
  updatedAt: '2026-09-28T08:00:00Z',
};

/** Ya en marcha: se puede preguntar, escalar, resolver o cerrar. */
const EN_PROGRESO: Ticket = { ...NUEVO, state: 'en progreso' };

/** Un principal cerrado: lo único que queda es reabrirlo. */
const CERRADO: Ticket = { ...NUEVO, state: 'cerrado' };

/** Un interno en marcha, que es el ticket con el que trabaja Desarrollo. */
const INTERNO: Ticket = {
  ...NUEVO,
  number: 'INT-CS-2026-0001',
  internal: true,
  state: 'en progreso',
  parent: 'CS-2026-0001',
};

/**
 * jsdom no trae el `<dialog>` nativo que usa `app-dialogo`. Basta con que abrir y cerrar no reviente:
 * lo que se prueba es el comportamiento del desplegable, no cómo pinta el navegador la ventana.
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

describe('AccionesDelTicket', () => {
  let http: HttpTestingController;
  let usuario: WritableSignal<Usuario | null>;
  let esAdministrador: WritableSignal<boolean>;

  beforeEach(async () => {
    localStorage.clear();
    usuario = signal<Usuario | null>(null);
    esAdministrador = signal(false);

    await TestBed.configureTestingModule({
      imports: [AccionesDelTicket],
      providers: [
        provideHttpClient(),
        provideHttpClientTesting(),
        // La sesión se sustituye por lo único que mira el componente: quién ha entrado y su papel.
        { provide: SessionService, useValue: { usuario, esAdministrador } },
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

  /** Monta el desplegable con ese ticket y ese papel. */
  function montar(ticket: Ticket, rol: string): ComponentFixture<AccionesDelTicket> {
    usuario.set({
      id: 1,
      name: 'María',
      lastName: 'Pérez',
      email: 'maria@demo.com',
      role: rol,
      origin: 'local',
      factory: false,
    });

    const fixture = TestBed.createComponent(AccionesDelTicket);
    fixture.componentRef.setInput('ticket', ticket);
    fixture.detectChanges();

    return fixture;
  }

  /** Los estados que ofrece el desplegable, sin la opción del estado actual (que no es un destino). */
  function estados(host: HTMLElement): string[] {
    return [...host.querySelectorAll<HTMLOptionElement>('option')]
      .map((opcion) => opcion.value)
      .filter((valor) => !valor.startsWith('actual:'));
  }

  /** Las opciones tal y como se leen, con la del estado actual delante. */
  function opciones(host: HTMLElement): HTMLOptionElement[] {
    return [...host.querySelectorAll<HTMLOptionElement>('option')];
  }

  /** El `<select>` del desplegable. */
  function selector(host: HTMLElement): HTMLSelectElement {
    const select = host.querySelector<HTMLSelectElement>('select');
    expect(select).toBeTruthy();
    return select!;
  }

  /** Elige un estado como lo haría el navegador. */
  function elegir(fixture: ComponentFixture<AccionesDelTicket>, estado: string): HTMLSelectElement {
    const host = fixture.nativeElement as HTMLElement;
    const select = selector(host);
    select.value = estado;
    select.dispatchEvent(new Event('change'));
    fixture.detectChanges();

    return select;
  }

  /** Escribe en el área de texto de la ventana. */
  function escribir(fixture: ComponentFixture<AccionesDelTicket>, texto: string): void {
    const host = fixture.nativeElement as HTMLElement;
    const campo = host.querySelector<HTMLTextAreaElement>('textarea');
    expect(campo).toBeTruthy();
    campo!.value = texto;
    campo!.dispatchEvent(new Event('input'));
    fixture.detectChanges();
  }

  /** Pulsa un botón por su texto, que es el nombre de los botones que no son sólo un icono. */
  function pulsarTexto(host: HTMLElement, texto: string): void {
    const boton = [...host.querySelectorAll<HTMLButtonElement>('button')].find(
      (candidato) => candidato.textContent?.trim() === texto,
    );
    expect(boton).toBeTruthy();
    boton!.click();
  }

  it('con un principal nuevo, Soporte ve estados: en progreso, escalado y cerrado', () => {
    const host = montar(NUEVO, 'soporte').nativeElement as HTMLElement;

    // **Estados, no acciones**: los valores son los del backend. Y **el actual va primero**, marcado y
    // sin poder elegirlo: no es un destino, es dónde está el ticket.
    expect(estados(host)).toEqual(['en progreso', 'escalado', 'cerrado']);

    const lista = opciones(host);
    expect(lista[0].value.startsWith('actual:')).toBe(true);
    expect(lista[0].textContent?.trim()).toBe('Nuevo');
    expect(lista[0].disabled).toBe(true);
    expect(lista[0].selected).toBe(true);
    expect(selector(host).selectedIndex).toBe(0);

    // Y sus nombres son los de los estados, no los de las acciones.
    expect(lista.map((opcion) => opcion.textContent?.trim())).toEqual([
      'Nuevo',
      'En progreso',
      'Escalado',
      'Cerrado',
    ]);
  });

  it('en cuanto el principal está en marcha, aparecen en espera, escalado, resuelto y cerrado', () => {
    const host = montar(EN_PROGRESO, 'soporte').nativeElement as HTMLElement;

    expect(estados(host)).toEqual(['en espera', 'escalado', 'resuelto', 'cerrado']);
    // Empezar ya no: no se empieza dos veces lo que ya está en marcha.
    expect(estados(host)).not.toContain('en progreso');
  });

  it('Desarrollo, en un interno en marcha, ve los estados de su trabajo', () => {
    const host = montar(INTERNO, 'desarrollo').nativeElement as HTMLElement;

    // `en espera` (pedirle algo a Soporte), `resuelto` y `cerrado` (que sin resolver devuelve el caso).
    expect(estados(host)).toEqual(['en espera', 'resuelto', 'cerrado']);
  });

  it('el usuario lee los estados con el lenguaje de su pantalla, no con la jerga interna', () => {
    // Un principal del propio usuario —por eso el solicitante lleva su id—: lo único que puede hacer
    // es cerrarlo.
    const host = montar({ ...NUEVO, requester: { ...ANA, id: 1 } }, 'usuario')
      .nativeElement as HTMLElement;

    expect(estados(host)).toEqual(['cerrado']);
    // «Recibido», que es como su pantalla llama a `nuevo` (`estadosDelUsuario`).
    expect(opciones(host)[0].textContent?.trim()).toBe('Recibido');
    expect(opciones(host)[1].textContent?.trim()).toBe('Cerrado');
  });

  it('con un ticket cerrado, Soporte sólo puede volver a en progreso (reabrirlo)', () => {
    const host = montar(CERRADO, 'soporte').nativeElement as HTMLElement;

    expect(estados(host)).toEqual(['en progreso']);
    expect(opciones(host)[0].textContent?.trim()).toBe('Cerrado');
  });

  it('Desarrollo no ve el desplegable en un principal: no es cosa suya', () => {
    const host = montar(NUEVO, 'desarrollo').nativeElement as HTMLElement;

    expect(host.querySelector('select')).toBeNull();
    expect(host.querySelector('dialog')).toBeNull();
  });

  it('el Administrador no ve el desplegable: no mueve estados', () => {
    esAdministrador.set(true);
    const host = montar(NUEVO, 'administrador').nativeElement as HTMLElement;

    // Ni control ni ventana: no hay nada que hacer, así que no se enseña nada.
    expect(host.querySelector('select')).toBeNull();
    expect(host.querySelector('dialog')).toBeNull();
  });

  it('elegir «En espera» abre su ventana, no pide texto y vuelve a su sitio', () => {
    const fixture = montar(EN_PROGRESO, 'soporte');
    const host = fixture.nativeElement as HTMLElement;

    const select = elegir(fixture, 'en espera');

    // **Vuelve a su sitio**: no se queda enseñando el último estado elegido. La opción del estado
    // actual es siempre la primera.
    expect(select.selectedIndex).toBe(0);
    expect(select.value).not.toBe('en espera');

    // Y sale **una sola ventana**, con su título y **sin campo ninguno**: «En espera» sólo se confirma
    // (decisión 88). El único estado que pide texto es «Cerrado», con el comentario del cierre.
    const ventana = host.querySelector<HTMLDialogElement>('dialog');
    expect(ventana?.open).toBe(true);
    expect(host.textContent).toContain('¿Pasar el ticket a «En espera»?');
    expect(ventana?.querySelector('input, textarea')).toBeNull();
    expect(ventana?.textContent).toContain('El ticket queda esperando la respuesta del usuario');
  });

  it('elegir escalado pide el motivo en su ventana y lo manda al escalar', async () => {
    const fixture = montar(NUEVO, 'soporte');
    const host = fixture.nativeElement as HTMLElement;

    elegir(fixture, 'escalado');

    // El título dice a cuál se pasa, y la ventana pide el motivo.
    expect(host.querySelector<HTMLDialogElement>('dialog')?.open).toBe(true);
    expect(host.textContent).toContain('¿Pasar el ticket a «Escalado»?');
    expect(host.textContent).toContain('Motivo');

    escribir(fixture, 'Probado y descartado.');
    pulsarTexto(host, 'Confirmar');

    const peticion = http.expectOne('/api/tickets/CS-2026-0001/escalate');
    expect(peticion.request.method).toBe('POST');
    expect(peticion.request.body).toEqual({ reason: 'Probado y descartado.' });
    peticion.flush({ ticket: { ...NUEVO, state: 'escalado' } });

    await esperar();
    fixture.detectChanges();

    expect(host.textContent).toContain('Ticket escalado. Se ha avisado a Desarrollo.');
  });

  it('cerrado no deja confirmar sin comentario y lo avisa', async () => {
    const fixture = montar(EN_PROGRESO, 'soporte');
    const host = fixture.nativeElement as HTMLElement;

    elegir(fixture, 'cerrado');
    expect(host.textContent).toContain('Comentario del cierre');

    pulsarTexto(host, 'Confirmar');
    fixture.detectChanges();

    // **No sale ninguna petición** y la ventana sigue abierta, con el aviso de por qué falta.
    http.expectNone('/api/tickets/CS-2026-0001/state');
    expect(host.querySelector<HTMLDialogElement>('dialog')?.open).toBe(true);
    expect(host.textContent).toContain('Para cerrar el ticket hay que decir por qué.');

    // Con el comentario, el cierre sí sale.
    escribir(fixture, 'Se resolvió por teléfono.');
    pulsarTexto(host, 'Confirmar');

    const peticion = http.expectOne('/api/tickets/CS-2026-0001/state');
    expect(peticion.request.method).toBe('POST');
    expect(peticion.request.body).toEqual({ to: 'cerrado', comment: 'Se resolvió por teléfono.' });
    peticion.flush({ ticket: CERRADO });

    await esperar();
    fixture.detectChanges();

    expect(host.textContent).toContain('Ticket cerrado.');
  });

  it('en progreso sólo confirma: la ventana no pide ningún texto', async () => {
    const fixture = montar(NUEVO, 'soporte');
    const host = fixture.nativeElement as HTMLElement;

    elegir(fixture, 'en progreso');

    expect(host.querySelector<HTMLDialogElement>('dialog')?.open).toBe(true);
    expect(host.querySelector('textarea')).toBeNull();

    pulsarTexto(host, 'Confirmar');

    // Sin texto que mandar, el cuerpo es sólo el destino.
    const peticion = http.expectOne('/api/tickets/CS-2026-0001/state');
    expect(peticion.request.body).toEqual({ to: 'en progreso' });
    peticion.flush({ ticket: EN_PROGRESO });

    await esperar();
    fixture.detectChanges();

    expect(host.textContent).toContain('El ticket ha cambiado de estado.');
  });

  it('reabrir un cerrado explica lo que va a pasar y usa la reapertura', async () => {
    const fixture = montar(CERRADO, 'soporte');
    const host = fixture.nativeElement as HTMLElement;

    elegir(fixture, 'en progreso');

    expect(host.textContent).toContain(
      'Este ticket está cerrado: para seguir hablando hay que reabrirlo.',
    );
    pulsarTexto(host, 'Confirmar');

    const peticion = http.expectOne('/api/tickets/CS-2026-0001/reopen');
    peticion.flush({ ticket: EN_PROGRESO });

    await esperar();
    fixture.detectChanges();

    expect(host.textContent).toContain('Ticket reabierto.');
  });

  it('elegir la opción del estado actual no hace nada', () => {
    const fixture = montar(EN_PROGRESO, 'soporte');
    const host = fixture.nativeElement as HTMLElement;

    const select = selector(host);
    select.value = select.options[0].value;
    select.dispatchEvent(new Event('change'));
    fixture.detectChanges();

    expect(host.querySelector<HTMLDialogElement>('dialog')?.open).toBe(false);
    http.expectNone('/api/tickets/CS-2026-0001/state');
  });
});
