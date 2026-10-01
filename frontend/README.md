# Frontend (Angular 22)

Proyecto generado con Angular CLI 22.1.8. La forma de las carpetas y las convenciones están en
`../docs/arquitectura.md`, sección 6.

## Ejecutar

**No se ejecuta nada en la máquina**: Node 24 y el CLI viven en el contenedor. Desde la raíz del
repositorio:

```bash
docker compose -f dev.yml up -d
docker compose -f dev.yml exec frontend npm test -- --watch=false
docker compose -f dev.yml exec frontend npm run build
```

El servidor de desarrollo escucha en `http://127.0.0.1:11001` y recarga al guardar. En
desarrollo se entra por **https://dev.catalina-support.example.com** (nginx delante del
contenedor).

El servidor de desarrollo **rechaza con 403** cualquier `Host` que no conozca: por eso el dominio
de desarrollo está declarado en `serve.options.allowedHosts` de `angular.json`. Si se añade otro
dominio o puerto de acceso, hay que añadirlo ahí también.

El servidor de desarrollo **reenvía `/api` al backend** con `proxy.conf.json` (declarado en
`serve.options.proxyConfig` de `angular.json`), con destino **`http://backend:11002`** —el servicio
del backend dentro de la red del entorno, porque quien hace la petición es el contenedor del
frontend, no el navegador— y sin reescribir la ruta. Así **abrir `http://127.0.0.1:11001` funciona
sin nginx**. Es **sólo de desarrollo**: es una opción de `serve` y el `build` de producción no la ve.

## Estructura

```text
src/app
├── app.ts, app.html, app.css    # armazón: cabecera, contenido y pie
├── app.config.ts                # providers globales (router, http)
├── app.routes.ts                # rutas raíz; cada módulo aporta las suyas
├── core                         # armazón: layout, sesión, interceptores, guardas
├── shared                       # UI y utilidades reutilizables, sin lógica de negocio
└── modules                      # un módulo por carpeta (todavía vacío)
```

Nada llama a `HttpClient` desde un componente: cada módulo tiene su servicio, y ese servicio
sólo apunta a `/api/<su módulo>/**`, en relativo.

Para generar código nuevo, usar el CLI **dentro del contenedor**, para que la versión del CLI sea
la del proyecto:

```bash
docker compose -f dev.yml exec frontend npx ng generate component modules/<módulo>/components/<nombre>
```
