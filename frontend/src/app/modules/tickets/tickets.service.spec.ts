import { provideHttpClient } from '@angular/common/http';
import { HttpTestingController, provideHttpClientTesting } from '@angular/common/http/testing';
import { TestBed } from '@angular/core/testing';

import { TicketsService } from './tickets.service';

const TICKET = {
  number: 'CS-2026-0001',
  internal: false,
  subject: 'No puedo entrar',
  description: 'Desde esta mañana falla.',
  state: 'nuevo',
  createdAt: '2026-09-24T16:30:00Z',
  updatedAt: '2026-09-24T16:30:00Z',
};

/**
 * El servicio del módulo, que es **lo único que llama a `/api/tickets/**`**
 * (`docs/arquitectura.md`, sección 4): aquí se comprueba que cada cosa va a su ruta, y que lo que
 * necesita la pantalla —la lista de responsables— **no** se le pide a `/api/users`.
 */
describe('TicketsService', () => {
  let tickets: TicketsService;
  let http: HttpTestingController;

  beforeEach(() => {
    TestBed.configureTestingModule({
      providers: [provideHttpClient(), provideHttpClientTesting()],
    });

    tickets = TestBed.inject(TicketsService);
    http = TestBed.inject(HttpTestingController);
  });

  afterEach(() => {
    http.verify();
  });

  it('la bandeja lleva sus filtros, y no manda los que están vacíos', async () => {
    const peticion = tickets.listar({
      type: 'interno',
      state: '',
      q: '  sesión ',
      category: '',
      tag: '',
      page: 3,
    });

    const pidiendo = http.expectOne((peticion) => peticion.url === '/api/tickets');
    expect(pidiendo.request.method).toBe('GET');
    expect(pidiendo.request.params.get('page')).toBe('3');
    expect(pidiendo.request.params.get('type')).toBe('interno');
    expect(pidiendo.request.params.get('q')).toBe('sesión');
    // Un filtro vacío no viaja: el backend entiende «sin filtrar» por su ausencia.
    expect(pidiendo.request.params.has('state')).toBe(false);
    expect(pidiendo.request.params.has('category')).toBe(false);
    expect(pidiendo.request.params.has('tag')).toBe(false);

    pidiendo.flush({ tickets: [TICKET], total: 1, page: 3, perPage: 25 });
    await expect(peticion).resolves.toEqual({ tickets: [TICKET], total: 1, page: 3, perPage: 25 });
  });

  it('la categoría y la etiqueta viajan en la petición, y no se quedan en la pantalla', async () => {
    // Es el mismo fallo que tuvo `view`: si el filtro se queda en la pantalla y no en la URL, la lista
    // enseña lo de siempre y el filtro no hace nada (docs/interfaz-y-experiencia.md, sección 3.3).
    const peticion = tickets.listar({
      type: '',
      state: '',
      q: '',
      category: '4',
      tag: '  red-wifi ',
      page: 1,
    });

    const pidiendo = http.expectOne((peticion) => peticion.url === '/api/tickets');
    expect(pidiendo.request.params.get('category')).toBe('4');
    // La etiqueta va sin los espacios de fuera: es lo que se escribe al filtrar.
    expect(pidiendo.request.params.get('tag')).toBe('red-wifi');

    pidiendo.flush({ tickets: [], total: 0, page: 1, perPage: 20 });
    await expect(peticion).resolves.toEqual({ tickets: [], total: 0, page: 1, perPage: 20 });
  });

  it('«lo mío» viaja como `mine=1`, y sólo cuando se pide', async () => {
    // Es lo que hace que la bandeja sea la de cada uno (`docs/modules/tickets.md`, decisión 51): lo
    // asignado, lo abierto por quien mira y aquello donde ha comentado.
    const mia = tickets.listar({ type: '', state: '', q: '', page: 1, mine: true });
    const pidiendoMia = http.expectOne((peticion) => peticion.url === '/api/tickets');
    expect(pidiendoMia.request.params.get('mine')).toBe('1');
    pidiendoMia.flush({ tickets: [TICKET], total: 1, page: 1, perPage: 25 });
    await mia;

    const delTodo = tickets.listar({ type: 'principal', state: '', q: '', page: 1 });
    const pidiendoTodo = http.expectOne((peticion) => peticion.url === '/api/tickets');
    // Las listas del «todo» no mandan `mine`: piden todo lo que hay, y el tipo lo fija la pantalla.
    expect(pidiendoTodo.request.params.has('mine')).toBe(false);
    expect(pidiendoTodo.request.params.get('type')).toBe('principal');
    pidiendoTodo.flush({ tickets: [TICKET], total: 1, page: 1, perPage: 25 });
    await delTodo;
  });

  it('la vista viaja en la petición, y «lo mío» también', async () => {
    // El chip de vista cambia el filtro **y tiene que cambiar la petición**: cuando se quedó en la
    // pantalla y no en la URL, la lista de «Observo» enseñaba lo de siempre (lo cazó la prueba de
    // interfaz).
    const peticion = tickets.listar({
      type: '',
      state: '',
      q: '',
      page: 1,
      mine: true,
      view: 'watching',
    });

    const pidiendo = http.expectOne((peticion) => peticion.url === '/api/tickets');
    expect(pidiendo.request.params.get('mine')).toBe('1');
    expect(pidiendo.request.params.get('view')).toBe('watching');

    pidiendo.flush({ tickets: [], total: 0, page: 1, perPage: 20 });
    await expect(peticion).resolves.toEqual({ tickets: [], total: 0, page: 1, perPage: 20 });
  });

  it('volver a resumir va a la ruta de insights del propio ticket', async () => {
    // El botón «Volver a resumir» habla con `tickets`, no con `ai`: la pantalla sólo puede llamar a
    // su propia API (docs/modules/ai.md, sección 5).
    const peticion = tickets.regenerarResumenes('CS-2026-0001');

    const pidiendo = http.expectOne({
      url: '/api/tickets/CS-2026-0001/insights',
      method: 'POST',
    });

    pidiendo.flush({ ticket: TICKET });
    await expect(peticion).resolves.toEqual({ ticket: TICKET });
  });

  it('mejorar un borrador usa la ruta modular del ticket y manda sólo editor, texto y tono', async () => {
    const peticion = tickets.mejorarRedaccion(
      'CS-2026-0001', 'comment', 'Hola Ana, ya funciona.', 'friendly',
    );
    const pidiendo = http.expectOne({
      url: '/api/tickets/CS-2026-0001/writing/improve',
      method: 'POST',
    });
    expect(pidiendo.request.body).toEqual({
      editor: 'comment', draft: 'Hola Ana, ya funciona.', tone: 'friendly',
    });
    pidiendo.flush({ text: 'Hola, Ana. El servicio ya funciona.' });
    await expect(peticion).resolves.toEqual({ text: 'Hola, Ana. El servicio ya funciona.' });
  });

  it('cada acción va a su ruta, y el estado no va por el `PATCH`', async () => {
    void tickets.asignar('CS-2026-0001', 7);
    const asignar = http.expectOne({ url: '/api/tickets/CS-2026-0001/assign', method: 'POST' });
    expect(asignar.request.body).toEqual({ assigneeId: 7 });
    asignar.flush({ ticket: TICKET });

    void tickets.moverEstado('CS-2026-0001', { state: 'en progreso' });
    const estado = http.expectOne({ url: '/api/tickets/CS-2026-0001/state', method: 'POST' });
    expect(estado.request.body).toEqual({ state: 'en progreso' });
    estado.flush({ ticket: TICKET });

    // **Cerrar lleva su comentario** (decisión 81): es obligatorio, va en el cuerpo y es lo que el
    // backend guarda como comentario de la conversación.
    void tickets.moverEstado('CS-2026-0001', {
      state: 'cerrado',
      comment: 'Se resolvió por teléfono.',
    });
    const cerrar = http.expectOne({ url: '/api/tickets/CS-2026-0001/state', method: 'POST' });
    expect(cerrar.request.body).toEqual({
      state: 'cerrado',
      comment: 'Se resolvió por teléfono.',
    });
    cerrar.flush({ ticket: TICKET });

    void tickets.escalar('CS-2026-0001', 'Probado y descartado.');
    const escalar = http.expectOne({ url: '/api/tickets/CS-2026-0001/escalate', method: 'POST' });
    expect(escalar.request.body).toEqual({ reason: 'Probado y descartado.' });
    escalar.flush({ ticket: TICKET });

    void tickets.reabrir('CS-2026-0001');
    http
      .expectOne({ url: '/api/tickets/CS-2026-0001/reopen', method: 'POST' })
      .flush({ ticket: TICKET });

    // Editar es un `PATCH`, y sólo lleva el asunto o la descripción.
    void tickets.editar('CS-2026-0001', { subject: 'Otro asunto' });
    const editar = http.expectOne({ url: '/api/tickets/CS-2026-0001', method: 'PATCH' });
    expect(editar.request.body).toEqual({ subject: 'Otro asunto' });
    editar.flush({ ticket: TICKET });
  });

  it('los comentarios van a su ruta, con su identificador', async () => {
    void tickets.comentar('CS-2026-0001', 'Ya está arreglado.');
    const comentar = http.expectOne({ url: '/api/tickets/CS-2026-0001/comments', method: 'POST' });
    expect(comentar.request.body).toEqual({ body: 'Ya está arreglado.' });
    comentar.flush({
      comment: { id: 3, body: 'Ya está arreglado.', edited: false, deleted: false, createdAt: '' },
    });

    void tickets.editarComentario('CS-2026-0001', 3, 'Corregido.');
    const editar = http.expectOne({ url: '/api/tickets/CS-2026-0001/comments/3', method: 'PATCH' });
    expect(editar.request.body).toEqual({ body: 'Corregido.' });
    editar.flush({
      comment: { id: 3, body: 'Corregido.', edited: true, deleted: false, createdAt: '' },
    });

    void tickets.borrarComentario('CS-2026-0001', 3);
    http
      .expectOne({ url: '/api/tickets/CS-2026-0001/comments/3', method: 'DELETE' })
      .flush({ status: 'ok' });
  });

  it('el alta manda el correo del solicitante cuando Soporte crea en nombre de otro', async () => {
    const alta = tickets.crear({
      subject: 'Llamada',
      description: 'Lo pide Ana por teléfono.',
      categoryId: 4,
      tags: ['red-wifi', 'licencia-office'],
      requesterEmail: 'ana@ejemplo.com',
    });

    const peticion = http.expectOne({ url: '/api/tickets', method: 'POST' });
    expect(peticion.request.body).toEqual({
      subject: 'Llamada',
      description: 'Lo pide Ana por teléfono.',
      categoryId: 4,
      tags: ['red-wifi', 'licencia-office'],
      requesterEmail: 'ana@ejemplo.com',
    });

    peticion.flush({ ticket: TICKET });
    await expect(alta).resolves.toEqual({ ticket: TICKET });
  });

  it('editar manda la categoría y las etiquetas por el mismo `PATCH` de la edición', async () => {
    // Es lo que usa la ficha para cambiar la clasificación: el mismo `PATCH` del asunto y la
    // descripción, con `categoryId` y `tags` (docs/modules/tickets.md, sección 5).
    void tickets.editar('CS-2026-0001', { categoryId: 2, tags: ['vpn'] });

    const peticion = http.expectOne({ url: '/api/tickets/CS-2026-0001', method: 'PATCH' });
    expect(peticion.request.body).toEqual({ categoryId: 2, tags: ['vpn'] });

    peticion.flush({ ticket: TICKET });
  });

  it('el catálogo va a sus rutas: listar, crear, renombrar y el estado aparte', async () => {
    // **Retirar va por su propia ruta** y no por el `PATCH`: es cosa sólo del Administrador, y
    // mezclarlo con el nombre dejaría que Soporte lo hiciera de rebote (decisión 64).
    const catalogo = tickets.categorias();
    http.expectOne({ url: '/api/tickets/categories', method: 'GET' }).flush({
      categories: [{ id: 1, name: 'General', active: true, tickets: 3 }],
    });
    await expect(catalogo).resolves.toEqual({
      categories: [{ id: 1, name: 'General', active: true, tickets: 3 }],
    });

    void tickets.crearCategoria('Red');
    const crear = http.expectOne({ url: '/api/tickets/categories', method: 'POST' });
    expect(crear.request.body).toEqual({ name: 'Red' });
    crear.flush({ category: { id: 2, name: 'Red', active: true, tickets: 0 } });

    void tickets.renombrarCategoria(2, 'Redes');
    const renombrar = http.expectOne({ url: '/api/tickets/categories/2', method: 'PATCH' });
    expect(renombrar.request.body).toEqual({ name: 'Redes' });
    renombrar.flush({ category: { id: 2, name: 'Redes', active: true, tickets: 0 } });

    void tickets.cambiarEstadoDeCategoria(2, false);
    const estado = http.expectOne({ url: '/api/tickets/categories/2/state', method: 'POST' });
    expect(estado.request.body).toEqual({ active: false });
    estado.flush({ category: { id: 2, name: 'Redes', active: false, tickets: 0 } });
  });

  it('las etiquetas se piden con su búsqueda, y sin ella vienen las más usadas', async () => {
    const buscando = tickets.etiquetas('  re ');
    const pidiendo = http.expectOne((peticion) => peticion.url === '/api/tickets/tags');
    expect(pidiendo.request.method).toBe('GET');
    expect(pidiendo.request.params.get('q')).toBe('re');
    pidiendo.flush({ tags: [{ tag: 'red-wifi', tickets: 4 }] });
    await expect(buscando).resolves.toEqual({ tags: [{ tag: 'red-wifi', tickets: 4 }] });

    // Sin `q` no viaja el parámetro: el backend entiende «las más usadas».
    void tickets.etiquetas();
    const sinBusqueda = http.expectOne((peticion) => peticion.url === '/api/tickets/tags');
    expect(sinBusqueda.request.params.has('q')).toBe(false);
    sinBusqueda.flush({ tags: [] });
  });

  it('los responsables los da el módulo de tickets, no el de usuarios', async () => {
    // Un módulo del frontend sólo habla con su propia API: si esta llamada se fuera a `/api/users`,
    // la lista de responsables sería de otro módulo y la regla dura estaría rota.
    const peticion = tickets.responsables();

    const pidiendo = http.expectOne({ url: '/api/tickets/assignees', method: 'GET' });
    pidiendo.flush({ main: [], internal: [] });

    await expect(peticion).resolves.toEqual({ main: [], internal: [] });
  });

  it('un adjunto se sube con su comentario, y se descarga como datos', async () => {
    const archivo = new File(['un pdf'], 'prueba.pdf', { type: 'application/pdf' });

    void tickets.adjuntar('CS-2026-0001', archivo, 3);
    const subir = http.expectOne({ url: '/api/tickets/CS-2026-0001/attachments', method: 'POST' });
    expect(subir.request.body instanceof FormData).toBe(true);
    expect((subir.request.body as FormData).get('commentId')).toBe('3');
    subir.flush({
      attachment: {
        id: 9,
        filename: 'prueba.pdf',
        contentType: 'application/pdf',
        size: 6,
        createdAt: '',
      },
    });

    // La descarga pide el archivo como datos: la cabecera de la sesión es obligatoria, y un enlace no
    // la lleva.
    const descarga = tickets.descargar('CS-2026-0001', 9);
    const bajando = http.expectOne({
      url: '/api/tickets/CS-2026-0001/attachments/9',
      method: 'GET',
    });
    expect(bajando.request.responseType).toBe('blob');
    bajando.flush(new Blob(['un pdf']));

    await expect(descarga).resolves.toBeInstanceOf(Blob);
  });
});
