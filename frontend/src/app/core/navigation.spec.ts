import { ENTRADAS, claveDeEntrada, entradasPara } from './navigation';

describe('entradasPara', () => {
  it('el menú enseña lo que existe, y sólo lo que existe', () => {
    // La lista ya es la del documento de interfaz (sección 3.2). **«Nuevo ticket» salió el
    // 2026-09-26** —se confundía con un módulo— y en su lugar están las dos listas del «todo».
    const rutas = ENTRADAS.map((entrada) => entrada.ruta);

    expect(rutas).toEqual([
      '/tickets',
      '/tickets/main',
      '/tickets/internal',
      '/tickets/categories',
      '/users',
      '/settings',
    ]);
  });

  it('cada papel tiene sus entradas, y ninguna de más', () => {
    expect(entradasPara('usuario').map((entrada) => entrada.ruta)).toEqual(['/tickets']);

    expect(entradasPara('soporte').map((entrada) => entrada.ruta)).toEqual([
      '/tickets',
      '/tickets/main',
      '/tickets/internal',
      '/tickets/categories',
      '/users',
    ]);

    expect(entradasPara('desarrollo').map((entrada) => entrada.ruta)).toEqual([
      '/tickets',
      '/tickets/main',
      '/tickets/internal',
    ]);

    expect(entradasPara('administrador').map((entrada) => entrada.ruta)).toEqual([
      '/tickets',
      '/tickets/categories',
      '/users',
      '/settings',
    ]);
  });

  it('«Categorías y etiquetas» la ven Soporte y el Administrador', () => {
    // El catálogo es de `tickets` y **no una sección de Configuración** (decisión 64): Soporte lo
    // mantiene y el Administrador, además, retira. Desarrollo y el usuario no lo ven.
    const entrada = ENTRADAS.find((candidata) => candidata.ruta === '/tickets/categories');

    expect(entrada?.clave).toBe('categorias');
    expect(entrada?.papeles).toEqual(['soporte', 'administrador']);
    for (const papel of ['usuario', 'desarrollo']) {
      expect(entradasPara(papel).map((candidata) => candidata.ruta)).not.toContain(
        '/tickets/categories',
      );
    }
  });

  it('la bandeja se llama distinto según quien mire', () => {
    // Una sola pantalla con dos nombres —«Mis tickets» para los tres que tienen tickets propios, y
    // «Bandeja» para el Administrador, que no tiene—, y el menú no se inventa dos entradas que llevan
    // al mismo sitio.
    const bandeja = ENTRADAS[0];

    expect(claveDeEntrada(bandeja, 'usuario')).toBe('misTickets');
    expect(claveDeEntrada(bandeja, 'desarrollo')).toBe('misTickets');
    expect(claveDeEntrada(bandeja, 'soporte')).toBe('misTickets');
    expect(claveDeEntrada(bandeja, 'administrador')).toBe('bandeja');
  });

  it('las dos listas del «todo» son de Soporte y Desarrollo', () => {
    // Son los dos papeles que pueden verlo todo (`docs/modules/tickets.md`, decisión 52).
    for (const ruta of ['/tickets/main', '/tickets/internal']) {
      const entrada = ENTRADAS.find((candidata) => candidata.ruta === ruta);

      expect(entrada?.papeles).toEqual(['soporte', 'desarrollo']);
      expect(entradasPara('usuario').map((candidata) => candidata.ruta)).not.toContain(ruta);
      expect(entradasPara('administrador').map((candidata) => candidata.ruta)).not.toContain(ruta);
    }
  });

  it('Configuración sólo la ve un administrador', () => {
    expect(entradasPara('administrador').map((entrada) => entrada.ruta)).toContain('/settings');

    for (const papel of ['usuario', 'soporte', 'desarrollo']) {
      expect(entradasPara(papel).map((entrada) => entrada.ruta)).not.toContain('/settings');
    }
  });

  it('crear un ticket no está en el menú, para nadie', () => {
    // Es una acción de la bandeja, no un módulo (decisión 53): la ruta sigue existiendo y se entra
    // por el botón que hay dentro.
    expect(ENTRADAS.map((entrada) => entrada.ruta)).not.toContain('/tickets/new');
  });

  it('un papel que no conocemos no ve nada de nadie', () => {
    // Si el backend mandara un papel nuevo, el menú no se le adelanta enseñando entradas de más: el
    // backend lo rechazaría igual, pero el menú no promete lo que no hay.
    expect(entradasPara('superusuario')).toEqual([]);
    expect(entradasPara('')).toEqual([]);
  });

  it('cada entrada tiene su ruta y su clave de texto', () => {
    for (const entrada of ENTRADAS) {
      expect(entrada.ruta.startsWith('/')).toBe(true);
      expect(entrada.clave.length).toBeGreaterThan(0);
    }
  });
});
