import { provideHttpClient } from '@angular/common/http';
import { HttpTestingController, provideHttpClientTesting } from '@angular/common/http/testing';
import { type ComponentFixture, TestBed } from '@angular/core/testing';
import { type WritableSignal, signal } from '@angular/core';
import { provideRouter } from '@angular/router';
import { vi } from 'vitest';

import { TranslationService } from '../../core/i18n/translation.service';
import { SessionService, type Usuario } from '../../core/services/session.service';
import { TicketVista, type Responsables } from './ticket-vista';
import { type Categoria, type DetalleDeTicket, type Persona } from './tickets.service';

/**
 * **Todo se edita donde se muestra** (decisión 74). Aquí se prueba lo que la vista manda cuando alguien
 * usa cada control —el lápiz del asunto, el botón de etiquetas, el «Añadir» de los observadores— y,
 * sobre todo, **que a quien no puede editar no se le ofrezca el control**: la interfaz no ofrece lo que
 * el backend rechazaría. El aspecto y las etiquetas accesibles, en un navegador, los cubre la capa de
 * Playwright.
 */

const SOPORTE: Persona = {
  id: 1,
  name: 'María',
  lastName: 'Pérez',
  email: 'maria@demo.com',
  role: 'soporte',
};

const DESARROLLO: Persona = {
  id: 2,
  name: 'Juan',
  lastName: 'Gómez',
  email: 'juan@demo.com',
  role: 'desarrollo',
};

const CATEGORIAS: readonly Categoria[] = [
  { id: 1, name: 'Red', active: true, tickets: 3 },
  { id: 2, name: 'Software', active: true, tickets: 1 },
];

const RESPONSABLES: Responsables = { main: [SOPORTE, DESARROLLO], internal: [] };

/** Un principal abierto, con categoría, dos etiquetas y un observador ya puesto. */
const DETALLE: DetalleDeTicket = {
  ticket: {
    number: 'CS-2026-0001',
    internal: false,
    subject: 'No va la red',
    description: '<p>Sin conexión desde esta mañana.</p>',
    state: 'en progreso',
    category: { id: 1, name: 'Red', active: true },
    tags: ['red-wifi', 'licencia-office'],
    requester: SOPORTE,
    createdAt: '2026-09-28T08:00:00Z',
    updatedAt: '2026-09-28T09:00:00Z',
  },
  comments: [],
  attachments: [],
  history: [],
  observers: [{ id: 10, account: SOPORTE }],
};

/**
 * Un interno abierto del mismo caso: **hereda el asunto, la categoría y las etiquetas del principal**
 * (decisión 67), que es lo que la ficha enseña cuando se mira un interno. Es el ticket desde el que
 * Desarrollo etiqueta (decisión 85).
 */
const INTERNO_DETALLE: DetalleDeTicket = {
  ticket: {
    number: 'INT-CS-2026-0001',
    internal: true,
    subject: 'No va la red',
    description: '<p>Sin conexión desde esta mañana.</p>',
    state: 'en progreso',
    parent: 'CS-2026-0001',
    category: { id: 1, name: 'Red', active: true },
    tags: ['red-wifi', 'licencia-office'],
    requester: SOPORTE,
    createdAt: '2026-09-28T08:00:00Z',
    updatedAt: '2026-09-28T09:00:00Z',
  },
  comments: [],
  attachments: [],
  history: [],
  observers: [],
};

/**
 * jsdom no trae el `<dialog>` nativo que usa `app-dialogo`. En estas pruebas no se abre ninguna ventana,
 * así que basta con que la vista se monte sin reventar.
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

describe('TicketVista', () => {
  let http: HttpTestingController;
  let usuario: WritableSignal<Usuario | null>;
  let esAdministrador: WritableSignal<boolean>;

  beforeEach(async () => {
    localStorage.clear();
    usuario = signal<Usuario | null>(null);
    esAdministrador = signal(false);

    await TestBed.configureTestingModule({
      imports: [TicketVista],
      providers: [
        provideHttpClient(),
        provideHttpClientTesting(),
        // La cabecera de un interno enlaza con su principal: el `RouterLink` necesita un Router.
        provideRouter([]),
        // La sesión se sustituye por lo único que mira la vista: quién ha entrado y su papel.
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

  /** Monta la vista con ese papel, que es lo único que cambia lo que se puede hacer dentro. */
  function montar(rol: string, detalle: DetalleDeTicket = DETALLE): ComponentFixture<TicketVista> {
    usuario.set({
      id: 1,
      name: 'María',
      lastName: 'Pérez',
      email: 'maria@demo.com',
      role: rol,
      origin: 'local',
      language: 'es',
      factory: false,
    });

    const fixture = TestBed.createComponent(TicketVista);
    fixture.componentRef.setInput('detalle', detalle);
    fixture.componentRef.setInput('responsables', RESPONSABLES);
    fixture.componentRef.setInput('categorias', CATEGORIAS);
    fixture.detectChanges();

    return fixture;
  }

  /** Pulsa un botón por su nombre accesible. */
  function pulsarAria(host: HTMLElement, etiqueta: string): void {
    const boton = host.querySelector<HTMLButtonElement>(`button[aria-label="${etiqueta}"]`);
    expect(boton).toBeTruthy();
    boton!.click();
  }

  /** Y uno por su texto, que es el nombre de los botones que no son sólo un icono. */
  function pulsarTexto(host: HTMLElement, texto: string): void {
    const boton = [...host.querySelectorAll<HTMLButtonElement>('button')].find(
      (candidato) => candidato.textContent?.trim() === texto,
    );
    expect(boton).toBeTruthy();
    boton!.click();
  }

  /** Una casilla del modal por el nombre de su etiqueta. */
  function casilla(contenedor: HTMLElement, tag: string): HTMLInputElement {
    const input = contenedor.querySelector<HTMLInputElement>(`input[aria-label="${tag}"]`);
    expect(input).toBeTruthy();
    return input!;
  }

  /** Marca o desmarca una etiqueta como lo haría una persona. */
  function marcar(contenedor: HTMLElement, tag: string, marcada: boolean): void {
    const input = casilla(contenedor, tag);
    input.checked = marcada;
    input.dispatchEvent(new Event('change'));
  }

  /**
   * Abre el modal de etiquetas (decisión 82) y comprueba que **pide el catálogo entero**: `all=1` y sin
   * `q`, porque no es la lista de sugerencias sino el catálogo.
   */
  async function abrirElModal(
    fixture: ComponentFixture<TicketVista>,
    host: HTMLElement,
  ): Promise<HTMLElement> {
    // El botón se llama **«Editar las etiquetas»** desde que existe el modal (decisión 82); el nombre
    // viejo era el de la propia línea, y por eso se busca por este.
    pulsarTexto(host, 'Editar las etiquetas');
    fixture.detectChanges();

    const peticion = http.expectOne((req) => req.url === '/api/tickets/tags');
    expect(peticion.request.method).toBe('GET');
    expect(peticion.request.params.get('all')).toBe('1');
    expect(peticion.request.params.has('q')).toBe(false);
    peticion.flush({
      tags: [
        { tag: 'red-wifi', tickets: 5 },
        { tag: 'licencia-office', tickets: 2 },
        { tag: 'vpn', tickets: 0 },
      ],
    });

    await esperar();
    fixture.detectChanges();

    const modal = host.querySelector('app-modal-etiquetas');
    expect(modal).toBeTruthy();
    return modal as HTMLElement;
  }

  it('el lápiz del asunto lo convierte en un campo y lo guarda con el `PATCH` del ticket', async () => {
    const fixture = montar('soporte');
    const host = fixture.nativeElement as HTMLElement;

    // Sin lápiz no hay campo: el asunto se lee, no se edita.
    expect(host.querySelector('#asunto-CS-2026-0001')).toBeNull();

    pulsarAria(host, 'Editar el asunto');
    fixture.detectChanges();

    const campo = host.querySelector<HTMLInputElement>('#asunto-CS-2026-0001');
    expect(campo).toBeTruthy();
    expect(campo!.value).toBe('No va la red');

    campo!.value = 'No va la red del tercero';
    campo!.dispatchEvent(new Event('input'));
    fixture.detectChanges();

    pulsarTexto(host, 'Guardar');

    const peticion = http.expectOne('/api/tickets/CS-2026-0001');
    expect(peticion.request.method).toBe('PATCH');
    // **Sólo el asunto**: no se manda la descripción ni la clasificación, que no se han tocado.
    expect(peticion.request.body).toEqual({ subject: 'No va la red del tercero' });
    peticion.flush({ ticket: DETALLE.ticket });

    await esperar();
    fixture.detectChanges();

    expect(host.textContent).toContain('Ticket actualizado.');
  });

  it('el modal pide el catálogo entero y el buscador filtra por nombre', async () => {
    const fixture = montar('soporte');
    const host = fixture.nativeElement as HTMLElement;
    const modal = await abrirElModal(fixture, host);

    // Todas las del catálogo, con la cuenta de cada una.
    expect(casilla(modal, 'red-wifi')).toBeTruthy();
    expect(casilla(modal, 'licencia-office')).toBeTruthy();
    expect(casilla(modal, 'vpn')).toBeTruthy();
    expect(modal.textContent).toContain('5');
    expect(modal.textContent).toContain('Todavía no la lleva ningún ticket.');

    const buscador = modal.querySelector<HTMLInputElement>('#buscar-etiqueta-CS-2026-0001');
    expect(buscador).toBeTruthy();
    expect(modal.querySelector('label[for="buscar-etiqueta-CS-2026-0001"]')?.textContent).toContain(
      'Buscar por etiqueta',
    );

    buscador!.value = 'lic';
    buscador!.dispatchEvent(new Event('input'));
    fixture.detectChanges();

    // Sólo queda la que coincide por nombre.
    expect(casilla(modal, 'licencia-office')).toBeTruthy();
    expect(modal.querySelector('input[aria-label="red-wifi"]')).toBeNull();
    expect(modal.querySelector('input[aria-label="vpn"]')).toBeNull();
  });

  it('marcar y desmarcar varias y guardar manda la lista entera en un solo `PATCH`', async () => {
    const fixture = montar('soporte');
    const host = fixture.nativeElement as HTMLElement;
    const modal = await abrirElModal(fixture, host);

    // **Las que el ticket ya lleva salen marcadas** y las demás no.
    expect(casilla(modal, 'red-wifi').checked).toBe(true);
    expect(casilla(modal, 'licencia-office').checked).toBe(true);
    expect(casilla(modal, 'vpn').checked).toBe(false);

    // Se desmarca una y se marca otra: varias a la vez, y nada se manda hasta guardar.
    marcar(modal, 'red-wifi', false);
    marcar(modal, 'vpn', true);
    fixture.detectChanges();
    expect(http.match((req) => req.url === '/api/tickets/CS-2026-0001').length).toBe(0);

    pulsarTexto(modal, 'Guardar');

    // **Un solo `PATCH`** con la lista entera: el backend reemplaza las etiquetas del ticket.
    const peticion = http.expectOne('/api/tickets/CS-2026-0001');
    expect(peticion.request.method).toBe('PATCH');
    expect(peticion.request.body).toEqual({ tags: ['licencia-office', 'vpn'] });
    peticion.flush({ ticket: DETALLE.ticket });

    await esperar();
    fixture.detectChanges();

    expect(host.textContent).toContain('Etiquetas guardadas.');
  });

  it('cancelar en el modal no manda nada', async () => {
    const fixture = montar('soporte');
    const host = fixture.nativeElement as HTMLElement;
    const modal = await abrirElModal(fixture, host);

    marcar(modal, 'vpn', true);
    pulsarTexto(modal, 'Cancelar');
    fixture.detectChanges();

    expect(http.match((req) => req.url === '/api/tickets/CS-2026-0001').length).toBe(0);
    // La ventana se cierra sin guardar: lo que se había marcado no sale de aquí.
    expect((modal.querySelector('dialog') as HTMLDialogElement).open).toBe(false);
  });

  it('cambiar la categoría en su línea la guarda al elegir', async () => {
    const fixture = montar('soporte');
    const host = fixture.nativeElement as HTMLElement;

    const selector = host.querySelector<HTMLSelectElement>('#categoria-CS-2026-0001');
    expect(selector).toBeTruthy();

    selector!.value = '2';
    selector!.dispatchEvent(new Event('change'));

    const peticion = http.expectOne('/api/tickets/CS-2026-0001');
    expect(peticion.request.method).toBe('PATCH');
    // Sólo la categoría: las etiquetas no se tocan al clasificar.
    expect(peticion.request.body).toEqual({ categoryId: 2 });
    peticion.flush({ ticket: DETALLE.ticket });

    await esperar();
    fixture.detectChanges();

    expect(host.textContent).toContain('Categoría cambiada.');
  });

  it('el botón del responsable abre el modal y asignar llama al servicio', async () => {
    const fixture = montar('soporte');
    const host = fixture.nativeElement as HTMLElement;

    // **El responsable ya no se cambia con un campo suelto**: junto a su nombre hay un botón que abre
    // el modal (quinta enmienda del 2026-09-29, punto 2).
    pulsarAria(host, 'Asignar a');
    fixture.detectChanges();

    const selector = host.querySelector<HTMLSelectElement>('#asignar-CS-2026-0001');
    expect(selector).toBeTruthy();
    expect(selector!.closest('dialog')?.open).toBe(true);

    selector!.value = '2';
    selector!.dispatchEvent(new Event('change'));
    fixture.detectChanges();

    pulsarTexto(host, 'Asignar');

    const peticion = http.expectOne('/api/tickets/CS-2026-0001/assign');
    expect(peticion.request.method).toBe('POST');
    expect(peticion.request.body).toEqual({ assigneeId: 2 });
    peticion.flush({ ticket: DETALLE.ticket });

    await esperar();
    fixture.detectChanges();

    expect(host.textContent).toContain('Ticket asignado.');
  });

  it('a quien no puede reasignar no se le enseña el botón del responsable', () => {
    const fixture = montar('usuario');
    const host = fixture.nativeElement as HTMLElement;

    expect(host.querySelector('button[aria-label="Asignar a"]')).toBeNull();
    expect(host.querySelector<HTMLSelectElement>('#asignar-CS-2026-0001')).toBeNull();
  });

  it('con un resumen pendiente la ficha se relee sola y para cuando los dos están listos', () => {
    vi.useFakeTimers();
    try {
      const conPendiente: DetalleDeTicket = {
        ...DETALLE,
        ticket: {
          ...DETALLE.ticket,
          insights: {
            motivo: { state: 'pendiente' },
            ultimaAccion: { state: 'pendiente' },
          },
        },
      };

      const fixture = montar('soporte', conPendiente);
      let recargas = 0;
      fixture.componentInstance.cambio.subscribe(() => (recargas += 1));

      // **El motor tarda, así que la ficha se relee sola** mientras haya algo pendiente.
      vi.advanceTimersByTime(8000);
      expect(recargas).toBe(1);

      // El padre recarga y sigue pendiente: se programa la siguiente relectura.
      fixture.componentRef.setInput('detalle', { ...conPendiente });
      fixture.detectChanges();
      vi.advanceTimersByTime(8000);
      expect(recargas).toBe(2);

      // **Los dos listos: se para** y no se vuelve a sondear.
      fixture.componentRef.setInput('detalle', { ...DETALLE });
      fixture.detectChanges();
      vi.advanceTimersByTime(80000);
      expect(recargas).toBe(2);
    } finally {
      vi.useRealTimers();
    }
  });

  it('añadir un observador llama al servicio con su cuenta y lo guarda al momento', async () => {
    const fixture = montar('soporte');
    const host = fixture.nativeElement as HTMLElement;

    // Al abrir el buscador salen los que pueden observarlo y todavía no lo hacen: el que ya observa
    // —María— no se ofrece.
    pulsarTexto(host, 'Añadir un observador');
    fixture.detectChanges();

    const nombresOfrecidos = [...host.querySelectorAll<HTMLButtonElement>('button')].map((boton) =>
      boton.textContent?.trim(),
    );
    expect(nombresOfrecidos).not.toContain('María Pérez');
    expect(nombresOfrecidos).toContain('Juan Gómez');

    pulsarTexto(host, 'Juan Gómez');

    const peticion = http.expectOne('/api/tickets/CS-2026-0001/observers');
    expect(peticion.request.method).toBe('POST');
    expect(peticion.request.body).toEqual({ accountId: 2 });
    peticion.flush({ ticket: DETALLE.ticket });

    await esperar();
    fixture.detectChanges();

    expect(host.textContent).toContain('Observador añadido.');
  });

  it('a Desarrollo no se le enseña ningún lápiz ni el control de la clasificación del principal', () => {
    const fixture = montar('desarrollo');
    const host = fixture.nativeElement as HTMLElement;

    // La interfaz no ofrece lo que el backend rechazaría: Desarrollo no escribe en el principal.
    expect(host.querySelector('button[aria-label="Editar el asunto"]')).toBeNull();
    expect(host.querySelector('button[aria-label="Editar la descripción"]')).toBeNull();
    expect(host.querySelector('#asunto-CS-2026-0001')).toBeNull();
    expect(host.querySelector('#categoria-CS-2026-0001')).toBeNull();

    // **Y en el principal tampoco etiqueta** (decisión 85): las etiquetas las pone desde su interno, no
    // desde el principal. Se leen, pero sin botón ni modal.
    const botones = [...host.querySelectorAll<HTMLButtonElement>('button')].map((boton) =>
      boton.textContent?.trim(),
    );
    expect(botones).not.toContain('Editar las etiquetas');
    expect(host.querySelector('app-modal-etiquetas')).toBeNull();

    // Y aun así sigue leyendo la categoría y las etiquetas como fichas de sólo lectura.
    expect(host.textContent).toContain('Red');
    expect(host.textContent).toContain('red-wifi');
  });

  it('Desarrollo ve el modal en el interno y guardar manda el `PATCH` con la lista entera', async () => {
    const fixture = montar('desarrollo', INTERNO_DETALLE);
    const host = fixture.nativeElement as HTMLElement;

    // En el interno Desarrollo **sí** etiqueta (decisión 85), pero sigue sin lápices ni categoría.
    expect(host.querySelector('button[aria-label="Editar el asunto"]')).toBeNull();
    expect(host.querySelector('#categoria-INT-CS-2026-0001')).toBeNull();

    const modal = await abrirElModal(fixture, host);

    // Las del ticket —las del principal, que el interno hereda— salen marcadas.
    expect(casilla(modal, 'red-wifi').checked).toBe(true);
    expect(casilla(modal, 'licencia-office').checked).toBe(true);
    expect(casilla(modal, 'vpn').checked).toBe(false);

    marcar(modal, 'red-wifi', false);
    marcar(modal, 'vpn', true);
    fixture.detectChanges();

    pulsarTexto(modal, 'Guardar');

    // **El `PATCH` va al número del ticket que se está viendo** —el interno—: el backend lo aplica al
    // principal, que es donde viven las etiquetas (decisiones 67 y 85).
    const peticion = http.expectOne('/api/tickets/INT-CS-2026-0001');
    expect(peticion.request.method).toBe('PATCH');
    expect(peticion.request.body).toEqual({ tags: ['licencia-office', 'vpn'] });
    peticion.flush({ ticket: INTERNO_DETALLE.ticket });

    await esperar();
    fixture.detectChanges();

    expect(host.textContent).toContain('Etiquetas guardadas.');
  });

  it('Soporte etiqueta el principal, no el interno', () => {
    const fixture = montar('soporte', INTERNO_DETALLE);
    const host = fixture.nativeElement as HTMLElement;

    // Decisión 85: cada equipo etiqueta desde el ticket en el que trabaja; el interno es de Desarrollo.
    const botones = [...host.querySelectorAll<HTMLButtonElement>('button')].map((boton) =>
      boton.textContent?.trim(),
    );
    expect(botones).not.toContain('Editar las etiquetas');
    expect(host.querySelector('app-modal-etiquetas')).toBeNull();
    // Las etiquetas del principal se siguen leyendo.
    expect(host.textContent).toContain('red-wifi');
  });

  it('al usuario no se le enseña ni la línea de etiquetas ni el botón', () => {
    const fixture = montar('usuario');
    const host = fixture.nativeElement as HTMLElement;

    // Decisión 82: el usuario no pone etiquetas **ni las ve**; la línea entera es de Soporte y Desarrollo.
    expect(host.textContent).not.toContain('Etiquetas');
    expect(host.textContent).not.toContain('red-wifi');
    expect(host.querySelector('app-modal-etiquetas')).toBeNull();
  });
});
