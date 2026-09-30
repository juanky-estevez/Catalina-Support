import { provideHttpClient } from '@angular/common/http';
import { HttpTestingController, provideHttpClientTesting } from '@angular/common/http/testing';
import { type ComponentFixture, TestBed } from '@angular/core/testing';
import { type WritableSignal, signal } from '@angular/core';

import { TranslationService } from '../../core/i18n/translation.service';
import { SessionService } from '../../core/services/session.service';
import { CategoriasPage } from './categorias-page';

/**
 * **Categorías y etiquetas** (decisión 72): las dos mitades con el mismo peso. Aquí se prueba lo que
 * la pantalla manda y lo que enseña —que Soporte no vea lo que sólo el Administrador puede hacer, que
 * crear una etiqueta la normalice y que retirarla pregunte antes—. El aspecto, en un navegador, lo
 * cubre la capa de Playwright.
 */
const CATEGORIAS = {
  categories: [
    { id: 1, name: 'Red', active: true, tickets: 3 },
    { id: 2, name: 'Software', active: false, tickets: 0 },
  ],
};

const ETIQUETAS = {
  tags: [
    { tag: 'red-wifi', tickets: 2 },
    { tag: 'licencia-office', tickets: 0 },
  ],
};

/**
 * jsdom no trae el `<dialog>` nativo que usa `app-dialogo`. Aquí sólo hace falta que abrir y cerrar no
 * reviente: lo que se prueba es lo que la pantalla manda, no cómo lo pinta el navegador.
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

describe('CategoriasPage', () => {
  let http: HttpTestingController;
  let esAdministrador: WritableSignal<boolean>;

  beforeEach(async () => {
    localStorage.clear();
    esAdministrador = signal(false);

    await TestBed.configureTestingModule({
      imports: [CategoriasPage],
      providers: [
        provideHttpClient(),
        provideHttpClientTesting(),
        // La sesión se sustituye por lo único que mira la pantalla: el papel de quien entra.
        { provide: SessionService, useValue: { esAdministrador } },
      ],
    }).compileComponents();

    // La pantalla se prueba en español, que es el idioma con el que se lee en el documento.
    TestBed.inject(TranslationService).cambiar('es');
    http = TestBed.inject(HttpTestingController);
  });

  /** Deja correr las promesas del servicio: el backend de prueba contesta en el acto. */
  function esperar(): Promise<void> {
    return new Promise((resolve) => setTimeout(resolve));
  }

  afterEach(() => {
    http.verify();
    localStorage.clear();
  });

  /** Monta la pantalla y responde lo primero que pide: el catálogo y las etiquetas. */
  async function montar(): Promise<ComponentFixture<CategoriasPage>> {
    const fixture = TestBed.createComponent(CategoriasPage);
    fixture.detectChanges();

    http.expectOne('/api/tickets/categories').flush(CATEGORIAS);
    // **El catálogo entero** (`all=1`): aquí se mantienen, y una etiqueta recién creada que no lleve
    // ningún ticket tiene que verse (decisión 68-quinquies). Sin el listado completo, el tope de diez la
    // dejaba fuera.
    http.expectOne('/api/tickets/tags?all=1').flush(ETIQUETAS);
    await esperar();
    fixture.detectChanges();

    return fixture;
  }

  /** Los botones que hay dentro de las filas de las dos listas. */
  function botonesDeLasFilas(host: HTMLElement, texto: string): HTMLButtonElement[] {
    return [...host.querySelectorAll('li button')].filter(
      (boton) => boton.textContent?.trim() === texto,
    ) as HTMLButtonElement[];
  }

  /** La fila de la lista donde se lee ese texto (la categoría o la etiqueta). */
  function filaCon(host: HTMLElement, texto: string): HTMLLIElement {
    const fila = [...host.querySelectorAll('li')].find((candidata) =>
      candidata.textContent?.includes(texto),
    );
    if (!fila) {
      throw new Error(`No hay ninguna fila con «${texto}»`);
    }
    return fila;
  }

  /** Pulsa un botón de esa fila por su texto. */
  function pulsarEnLaFila(host: HTMLElement, fila: string, texto: string): void {
    const boton = [...filaCon(host, fila).querySelectorAll('button')].find(
      (candidato) => candidato.textContent?.trim() === texto,
    );
    if (!boton) {
      throw new Error(`No hay ningún botón «${texto}» en la fila «${fila}»`);
    }
    boton.click();
  }

  /** Pulsa un botón por su texto, en toda la pantalla. */
  function pulsar(host: HTMLElement, texto: string): void {
    const boton = [...host.querySelectorAll('button')].find(
      (candidato) => candidato.textContent?.trim() === texto,
    ) as HTMLButtonElement | undefined;
    if (!boton) {
      throw new Error(`No hay ningún botón «${texto}»`);
    }
    boton.click();
  }

  /** El diálogo donde vive un campo, para no confundirlo con el de la otra mitad. */
  function dialogoDelCampo(host: HTMLElement, identificador: string): HTMLDialogElement {
    return host.querySelector(`#${identificador}`)!.closest('dialog') as HTMLDialogElement;
  }

  it('enseña las dos mitades con su descripción y su cuenta de tickets', async () => {
    const fixture = await montar();
    const texto = (fixture.nativeElement as HTMLElement).textContent ?? '';

    expect(texto).toContain('Categorías');
    expect(texto).toContain('Etiquetas');
    expect(texto).toContain('Con esto se clasifica cada ticket');
    expect(texto).toContain('Con esto se matiza lo que una categoría no distingue');
    expect(texto).toContain('3 tickets');
    expect(texto).toContain('2 tickets');
    // Una etiqueta del catálogo que no la lleva nadie lo dice, en vez de enseñar un cero mudo.
    expect(texto).toContain('Todavía no la lleva ningún ticket.');
  });

  it('a Soporte no se le enseña retirar; al Administrador sí, en las dos mitades', async () => {
    const fixture = await montar();
    const host = fixture.nativeElement as HTMLElement;

    // La interfaz no ofrece lo que no se puede hacer: Soporte crea y renombra, pero no retira.
    expect(botonesDeLasFilas(host, 'Retirar').length).toBe(0);

    esAdministrador.set(true);
    fixture.detectChanges();

    // Una por la categoría activa y una por cada etiqueta: las dos etiquetas se pueden retirar.
    expect(botonesDeLasFilas(host, 'Retirar').length).toBe(3);
  });

  it('crear una etiqueta la normaliza y la manda al backend', async () => {
    // **Mantener el catálogo es del Administrador** (decisión 83): sin esto, la pantalla no
    // enseña los botones de crear ni de renombrar.
    esAdministrador.set(true);

    const fixture = await montar();
    const host = fixture.nativeElement as HTMLElement;

    pulsar(host, 'Nueva etiqueta');
    fixture.detectChanges();

    const campo = host.querySelector<HTMLInputElement>('#nueva-etiqueta')!;
    campo.value = 'Premios 2026';
    campo.dispatchEvent(new Event('input'));
    fixture.detectChanges();
    // Al teclear se normaliza: el campo enseña lo que va a quedar (decisión 68).
    expect(campo.value).toBe('premios-2026');

    dialogoDelCampo(host, 'nueva-etiqueta')
      .querySelector('form')!
      .dispatchEvent(new Event('submit', { cancelable: true }));

    // **Sin `all`**: crear es un POST al catálogo; el `all=1` es del listado de mantenimiento.
    const peticion = http.expectOne('/api/tickets/tags');
    expect(peticion.request.method).toBe('POST');
    expect(peticion.request.body).toEqual({ tag: 'premios-2026' });
    peticion.flush({ tag: { tag: 'premios-2026', tickets: 0 } });

    await esperar();
    http.expectOne('/api/tickets/categories').flush(CATEGORIAS);
    http.expectOne('/api/tickets/tags?all=1').flush(ETIQUETAS);
    await esperar();
    fixture.detectChanges();

    expect(host.textContent).toContain('Etiqueta creada.');
  });

  it('un nombre que queda vacío al normalizar avisa en vez de mandarlo', async () => {
    // **Mantener el catálogo es del Administrador** (decisión 83): sin esto, la pantalla no
    // enseña los botones de crear ni de renombrar.
    esAdministrador.set(true);

    const fixture = await montar();
    const host = fixture.nativeElement as HTMLElement;

    pulsar(host, 'Nueva etiqueta');
    fixture.detectChanges();

    const campo = host.querySelector<HTMLInputElement>('#nueva-etiqueta')!;
    campo.value = '¡¡!!';
    campo.dispatchEvent(new Event('input'));
    fixture.detectChanges();

    dialogoDelCampo(host, 'nueva-etiqueta')
      .querySelector('form')!
      .dispatchEvent(new Event('submit', { cancelable: true }));
    fixture.detectChanges();

    // No sale ninguna petición: el aviso es la clave `tickets.tag.required`, con su texto.
    expect(host.textContent).toContain('no puede quedar vacía');
  });

  it('retirar una etiqueta pregunta antes, diciendo a cuántos tickets afecta', async () => {
    esAdministrador.set(true);
    const fixture = await montar();
    const host = fixture.nativeElement as HTMLElement;

    pulsarEnLaFila(host, 'red-wifi', 'Retirar');
    fixture.detectChanges();

    // El aviso lleva la etiqueta y su cuenta: retirarla la quita de todos esos tickets.
    expect(host.textContent).toContain('«red-wifi»');
    expect(host.textContent).toContain('de 2 tickets');

    // El botón de confirmar vive en el diálogo, no en una fila. Se confirma y se manda el `DELETE`.
    const dialogo = [...host.querySelectorAll('dialog')].at(-1) as HTMLDialogElement;
    const confirmar = [...dialogo.querySelectorAll('button')].find(
      (boton) => boton.textContent?.trim() === 'Retirar',
    ) as HTMLButtonElement;
    confirmar.click();

    const peticion = http.expectOne('/api/tickets/tags/red-wifi');
    expect(peticion.request.method).toBe('DELETE');
    peticion.flush({ status: 'ok' });

    await esperar();
    http.expectOne('/api/tickets/categories').flush(CATEGORIAS);
    http.expectOne('/api/tickets/tags?all=1').flush(ETIQUETAS);
    await esperar();
    fixture.detectChanges();

    expect(host.textContent).toContain('Etiqueta retirada y quitada de sus tickets.');
  });

  it('renombrar una etiqueta manda el nombre que se está viendo y el nuevo', async () => {
    // **Mantener el catálogo es del Administrador** (decisión 83): sin esto, la pantalla no
    // enseña los botones de crear ni de renombrar.
    esAdministrador.set(true);

    const fixture = await montar();
    const host = fixture.nativeElement as HTMLElement;

    pulsarEnLaFila(host, 'red-wifi', 'Renombrar');
    fixture.detectChanges();

    const campo = host.querySelector<HTMLInputElement>('#etiqueta-red-wifi')!;
    campo.value = 'wifi-red';
    campo.dispatchEvent(new Event('input'));
    fixture.detectChanges();

    // La fila está en modo edición: el «Guardar» es el de al lado del propio campo.
    const enLaFila = campo.closest('li')!;
    const guardar = [...enLaFila.querySelectorAll('button')].find(
      (boton) => boton.textContent?.trim() === 'Guardar',
    ) as HTMLButtonElement;
    guardar.click();

    const peticion = http.expectOne('/api/tickets/tags/red-wifi');
    expect(peticion.request.method).toBe('PATCH');
    expect(peticion.request.body).toEqual({ tag: 'wifi-red' });
    peticion.flush({ tag: { tag: 'wifi-red', tickets: 2 } });

    await esperar();
    http.expectOne('/api/tickets/categories').flush(CATEGORIAS);
    http.expectOne('/api/tickets/tags?all=1').flush(ETIQUETAS);
    await esperar();
  });
});
