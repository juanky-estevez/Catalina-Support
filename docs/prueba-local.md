# Prueba local: instalación vacía, ejemplos y verificación aislada

> **Estado:** as-built
> **Última actualización:** 2026-10-03
>
> **Propuesta del 2026-10-03, aprobada explícitamente por el responsable.** Reúne las decisiones del
> responsable de las dos tandas de preguntas. La primera eligió local primero, esquema con un
> comando de Docker, pruebas locales y Keycloak opcional en localhost. La segunda eligió separar
> direcciones pública e interna de Keycloak, aislar las pruebas y conservar la configuración al
> cargar ejemplos. **Corrección previa del responsable:** la instalación vacía es el recorrido
> principal; el agente había sugerido comenzar con ejemplos. Esta propuesta mantiene esa corrección.
> Aprobación recibida después del repaso: «Lo apruebo». Implementado y verificado;
> la enmienda §5.1.b también recibió aprobación explícita.

## 1. Problema y resultado esperado

Quien descarga el proyecto debe poder seguir el README sin el dominio, nginx, certificados ni datos
previos de esta máquina. Primero levanta una instalación vacía, aplica el esquema con un comando,
completa el asistente y entra con la cuenta de fábrica. Puede crear cuentas y tickets propios,
cargar ejemplos si los quiere y probar AD o Keycloak como servicios opcionales.

Las pruebas automáticas se ejecutan contra una instalación separada. No cambian las cuentas,
tickets, configuración, correo ni adjuntos de la instalación que está usando esa persona.

Esta propuesta cubre **desarrollo y evaluación local**. No habilita desplegar producción.

## 2. Hallazgos que motivan la propuesta

| Hallazgo | Evidencia actual | Cambio propuesto |
| --- | --- | --- |
| Aplicar sólo el esquema exige copiar SQL y llamar a `psql` manualmente | README y `dev.yml` | Servicio de un solo uso para el esquema |
| Playwright apunta al dominio del proyecto | `dev.yml`, `playwright.config.ts` y preparación | Compose propio y direcciones de su red |
| Angular rechaza el nombre `frontend` | `frontend/angular.json` y comprobación del 2026-10-02 | Admitir explícitamente ese host de desarrollo |
| Los ejemplos sustituyen los ajustes del usuario | Últimos bloques de `v1.0.0_dev.sql` | Separar datos de ejemplo y configuración |
| Keycloak anuncia el dominio original | `keycloak.yml` y reino importado | Dirección pública local y conexión interna independiente |
| El cliente OIDC no lee ni comprueba `issuer` en descubrimiento | `backend/modules/auth/services/keycloak.go` | Comprobar el emisor público antes de utilizar los endpoints |

Los límites del script de producción ya reportados permanecen fuera de alcance. No se ejecuta
`prod-build.sh`, no se abre el dominio ni se crea la etiqueta de versión.

## 3. Recorrido principal: sin ejemplos

Comandos vigentes, desde la raíz del repositorio, iguales en PowerShell, CMD y bash:

```text
docker compose -f dev.yml up -d --build
docker compose -f dev.yml run --rm migrate
```

`migrate` es el nombre aprobado para el servicio que sólo aplica el esquema vigente:

- Tiene perfil propio y no arranca con `up -d`.
- Espera a PostgreSQL saludable, monta la migración en sólo lectura y usa el cliente de PostgreSQL
  dentro del contenedor. No necesita Docker, Go, Node ni `psql` en la máquina.
- Usa `ON_ERROR_STOP=1`, conserva la transacción e idempotencia del SQL y falla con código distinto
  de cero si la migración falla. El mensaje de éxito sólo aparece tras completarla.
- No carga ejemplos, no borra datos, no configura SMTP y no sella una base nueva.
- Repetirlo no reabre una instalación sellada ni altera sus ajustes.
- Mientras la 1.0.0 siga abierta aplica `v1.0.0.sql`; no introduce un sistema general de migraciones.

Al abrir `http://127.0.0.1:11001`, una base nueva enseña `/setup`. El README indica los cuatro pasos,
el método local y los valores de Mailpit (`mail`, `1025`, sin TLS ni credenciales), con su prueba.
Después de terminar se entra con `admin` / `admin`, se crean cuentas y se abren sus enlaces en
`http://127.0.0.1:11004`. La contraseña de producción sigue siendo `ADMIN_PASSWORD` del entorno.

No se cambia el flujo del asistente ni se añade un acceso que lo salte.

## 4. Ejemplos opcionales que conservan la configuración

El comando sigue siendo `docker compose -f dev.yml run --rm seed`.

En una instalación **sellada**, no modifica `installation_settings`, `directory_settings`,
`keycloak_settings` ni `ticket_settings`: conserva nombre, idioma, marca, región, dirección,
SMTP, método de entrada, directorios, IA, prefijo y reparto. Conserva el sello.

En una base **sin sellar**, completa la preparación local de los ejemplos con valores de demo:
entrada local, dirección `http://127.0.0.1:11001`, SMTP de Mailpit y ajustes necesarios para las
cuentas y tickets de ejemplo; deja la demo sellada. El README advierte que, en este caso, el seeder
sustituye el recorrido del asistente y puede reemplazar sus pasos todavía incompletos.

Se mantiene la semántica destructiva de los datos de ejemplo: se recrean tickets, categorías,
etiquetas, contadores, cuentas `@demo.com` y archivos adjuntos de desarrollo; se preserva la marca.
Conservar configuración **no** significa conservar esos datos. El aviso debe aparecer antes del
comando, tanto en el README como en su salida.

Las once cuentas mantienen sus credenciales de demo. Si la instalación conservada usa AD o
Keycloak, las cuentas locales de ejemplo no entran hasta volver a Local con la cuenta de fábrica.
No se cambia el método para hacerlas entrar automáticamente.

## 5. Keycloak: identidad pública y conexión interna

### 5.1 Configuración

Se añade una dirección interna **opcional** al bloque de Keycloak, tanto en Configuración como en
el paso de entrada del asistente. Vive en `keycloak_settings`, con migración idempotente dentro de
la versión abierta, y viaja en los DTO de settings y en `config.OIDC`.

Nombre de contrato aprobado: `internalIssuer` en JSON, `internal_issuer` en PostgreSQL. Es una URL
completa del mismo reino, no otro emisor ni un secreto. Vacía significa usar las direcciones
públicas, conservando las instalaciones que ya funcionan. No se añaden variables de entorno para
sustituir una configuración que vive en la base.

| Dato | Local en el navegador de la máquina |
| --- | --- |
| Aplicación y dirección pública | `http://127.0.0.1:11001` |
| Emisor público | `http://127.0.0.1:11006/sso/realms/catalina-support` |
| Dirección interna del mismo reino | `http://keycloak:8080/sso/realms/catalina-support` |
| Vuelta | `http://127.0.0.1:11001/api/auth/keycloak/callback` |
| Cliente y secreto de pruebas | Los del reino versionado, documentados como credenciales locales |

Keycloak sigue en su compose propio y en la red del entorno. El navegador llega directamente al
puerto 11006; no se necesita añadir `/sso` al proxy de Angular ni instalar nginx.
El reino permite la vuelta local exacta. No se añaden comodines de redirección.
La dirección anunciada por el contenedor debe poder configurarse para el entorno aislado de pruebas;
no se obliga a editar el reino ni a tener un archivo de hosts en la máquina.

### 5.1.b Enmienda aprobada: configurar un método todavía vacío

**Estado de esta enmienda: aprobada explícitamente por el responsable el 2026-10-03 («Lo apruebo»).** Hallazgo al verificar
Keycloak desde una instalación vacía: Configuración deshabilita AD y Keycloak si sus campos están
vacíos, pero los campos sólo se muestran al elegir ese método. No hay forma de rellenarlos por la
interfaz. El asistente ya permite esa elección provisional.

Se permite elegir AD o Keycloak en el formulario de Configuración aunque sus datos estén
vacíos. Elegir sólo muestra los campos; **el método en funcionamiento no cambia hasta guardar**.
El backend conserva su validación: guardar un método incompleto devuelve 422. Se conserva la cuenta
de fábrica. No se añaden pantallas, endpoints ni permisos. Se verificó que una instalación local
vacía permite rellenar y probar los dos métodos sin cargar ejemplos y que un guardado incompleto
no cambia el método persistido. La aprobación habilita el cambio del selector y su prueba de regresión.

### 5.2 Reglas de conexión y validación

- El descubrimiento se obtiene desde `internalIssuer`, si está configurado; en caso contrario,
  desde el emisor público.
- El `issuer` anunciado debe coincidir con el emisor público configurado. Ausente o distinto:
  la prueba y el acceso fallan con una clave traducida, antes de canjear credenciales.
- La URL de autorización enviada al navegador permanece pública.
- Sólo para peticiones del backend se sustituye el prefijo público del reino por el interno:
  descubrimiento, canje del código y lectura de UserInfo. Se conserva el resto de la ruta y consulta.
- La sustitución debe comparar URLs analizadas y límites de ruta, no reemplazar texto arbitrario.
  Los endpoints deben pertenecer al mismo origen y reino esperado; no se transforma un host o
  prefijo parecido ni se sigue una redirección a otro origen con secretos o tokens.
- Se admiten HTTP y HTTPS, puertos y rutas de reino; se rechazan credenciales incrustadas,
  fragmentos y consultas en las dos direcciones de configuración. No se desactiva la validación TLS.
- Vaciar la dirección interna restaura la conexión pública. Cambiar los ajustes invalida el
  descubrimiento recordado, como corresponde a una configuración leída en cada intento.
- Se mantienen el cliente confidencial, el `state` firmado y la sesión en el fragmento de vuelta.
  No se añade PKCE ni una nueva validación de ID tokens en este trabajo.

Ambos campos se explican con texto sencillo y ejemplos, en español e inglés. No se muestra la
URL interna al usuario que entra por Keycloak; el administrador es quien la configura.

## 6. Pruebas en una instalación aislada

Se implementa `tests.yml` en la raíz, con proyecto distinto del de desarrollo y producción:

- Contenedores, redes y volúmenes propios, sin nombres globales que coincidan con los actuales.
- Fuente montada sólo para construir o ejecutar; adjuntos, logs, dependencias y base en volúmenes
  propios. No monta `_files/`, `_logs/` ni los volúmenes de la instalación local.
- Frontend, backend, PostgreSQL y Mailpit sin puertos publicados en la máquina.
- LDAP y Keycloak opcionales mediante un perfil de pruebas de directorio, dentro de esta red.
  No reutilizan los servicios ni el reino de la instalación local.
- No entra en la red compartida del motor de IA. La suite debe funcionar con IA ausente.
- El navegador usa `http://frontend.localhost:11001`, permitido explícitamente en Angular. Los enlaces de
  correo usan esa misma base en la configuración exclusiva de pruebas.
- El Keycloak aislado anuncia `http://keycloak:8080/sso`; cliente, emisor, URL interna y vuelta
  coinciden con las direcciones accesibles al navegador del contenedor.
- Chromium resuelve `frontend.localhost` al servicio `frontend` sólo dentro del navegador de pruebas
  aisladas; el sufijo localhost proporciona un contexto seguro, para ejercer el portapapeles real como en localhost. No se simula la API ni se cambia
  la seguridad de la aplicación.
- Todas las utilidades de Playwright comparten el mismo origen; no quedan respaldos divergentes
  entre configuración, preparación y limpieza.

El README muestra un recorrido de comandos Docker: preparar una instancia de pruebas nueva,
arrancar sus servicios, ejecutar Playwright y retirar **sólo** sus recursos. Las dependencias se
instalan en contenedores; el fallo del runner conserva su código de salida. El comando predeterminado
no utiliza una dirección externa ni convierte `BASE_URL` en una forma implícita de tocar la instalación
local. Un destino externo es una opción avanzada explícita y no forma parte de este recorrido.

La preparación parte del esquema y configura por la API sólo la instalación aislada. Incluye un
caso de navegador de los cuatro pasos desde vacío **antes de sellarla**, seguido de la suite usual.
Las pruebas crean sus cuentas y fixtures, sin exigir el seeder opcional del usuario. Los endpoints
requieren la misma protección que en el producto, sin puertas exclusivas para pruebas.

Los informes, trazas y capturas se conservan en una carpeta de resultados identificada como de
pruebas; retirar la base desechable no borra esas evidencias. No se promete ejecución simultánea de
varias suites en este alcance. Una pasada nueva comienza con base y archivos de prueba nuevos,
también después de una pasada interrumpida; la limpieza nunca busca recursos por un prefijo amplio.

## 7. Documentos y archivos afectados

Este documento coordina el cambio; las normas de cada área siguen en su documento actual.
Las enmiendas específicas quedan registradas en los documentos afectados y describen la implementación.

| Documento actualizado | Qué decide o refleja | Archivos afectados |
| --- | --- | --- |
| `docs/ambientes.md` | Esquema con un comando, seeder conservando configuración y suite desechable | `dev.yml`, `tests.yml`, `scripts/dev-seed.sh`, migraciones y `tests/e2e/**` |
| `docs/arquitectura.md` | Red y volúmenes aislados, hosts admitidos y separación OIDC | Compose, `frontend/angular.json`, cableado del backend |
| `docs/modules/auth.md` | Descubrimiento, emisor público y conexión interna de Keycloak | `backend/modules/auth/**`, `backend/shared/config/**`, `keycloak.yml`, `config/keycloak/**` |
| `docs/modules/settings.md` | Nuevo campo interno, lectura, escritura y prueba de conexión | `backend/modules/settings/**`, migración de `keycloak_settings`, pantalla de Configuración |
| `docs/primer-arranque.md` | Campo interno opcional en el paso de entrada y cobertura desde cero | DTO y servicios de setup, `setup-page.*`, caso Playwright |
| `docs/interfaz-y-experiencia.md` | Campo y ayudas de Keycloak en las dos pantallas, textos en ambos idiomas | Configuración, asistente y diccionarios de i18n |
| `README.md` | Recorrido vacío, ejemplos opcionales y pruebas locales reproducibles | Instrucciones y ejemplos de comandos |
| `docs/README.md` y `AGENTS.md` | Estado as-built y correspondencia documental | Índice y tabla de cobertura |

Los documentos afectados se actualizaron en este trabajo, con la enmienda fechada en su cabecera.

## 8. Criterios de aceptación

1. Desde base nueva: `up` y `run --rm migrate` dejan cero cuentas y cero tickets; el navegador
   muestra el asistente y al terminar se puede entrar con `admin` / `admin`.
2. El comando de esquema se puede repetir: conserva datos, ajustes y sello. SQL inválido hace
   fallar el comando y no se anuncia éxito.
3. Alta local: el correo llega a Mailpit, su enlace abre esta instalación y permite establecer
   contraseña e iniciar sesión con la cuenta nueva.
4. Seeder sobre instalación sellada: compara antes y después las cuatro tablas de configuración
   y la marca; se conservan mientras se recrean los ejemplos previstos y sus adjuntos.
5. Seeder sobre base nueva o asistente incompleto: deja una demo local utilizable y sellada, sin
   dominio externo; repetirlo no vuelve a cambiar la configuración de la demo ya sellada.
6. Keycloak local: contenedor opcional, prueba de conexión, entrada con una persona del reino,
   vuelta exacta y fragmento borrado. Sin contenedor, Local y la cuenta de fábrica siguen funcionando.
7. OIDC: emisor incorrecto, reino equivocado, endpoints fuera del prefijo y redirecciones externas
   se rechazan; se demuestra que el backend usa la dirección interna y el navegador la pública.
8. Suite aislada: asistente desde vacío, cuentas locales y correos en PC y móvil; LDAP y Keycloak
   cuando se solicita el perfil. Servicios opcionales ausentes se indican como casos omitidos.
9. Una pasada fallida o interrumpida no cambia la instalación del usuario y no deja datos de
   pruebas que contaminen la siguiente. Los informes se conservan y los fallos no se convierten en éxito.
10. Comandos y montajes se prueban en Linux. Si no hay máquinas Windows o macOS disponibles,
    se declara esa limitación y no se afirma haber verificado esos sistemas.

## 9. Fuera de alcance

Producción y su script, publicación del dominio, etiqueta de versión, cambios de roles o del flujo
de tickets, migraciones de versiones futuras, sustitución del motor de IA, PKCE, gestión de cuentas
de servicio de Keycloak y ejecución concurrente de suites. Tampoco se cambia la contraseña de fábrica
ni se obliga a instalar Go o Node en la máquina.

## 10. Registro del repaso

El responsable respondió **1A, 2A, 3A** al repaso del documento:

1. **Seeder:** completa y sella la demo cuando la base es nueva o el asistente está incompleto.
   Se aceptó para poder probar inmediatamente, con el aviso de que sustituye los pasos incompletos.
   En una instalación sellada conserva la configuración, como fija la sección 4.
2. **Keycloak:** la dirección interna es opcional y visible tanto en Configuración como en el
   asistente. Se aceptó poder configurarla sin editar archivos y rechazar el descubrimiento si
   el `issuer` falta o no coincide, como fija la sección 5.
3. **Suite:** retira exclusivamente los datos y recursos de su instancia desechable, conserva
   los informes y permite incluir LDAP y Keycloak mediante un perfil opcional. Se eligió que
   cada pasada comience limpia, como fija la sección 6.

No quedan decisiones de diseño abiertas. Después del repaso, el responsable aprobó explícitamente
el documento con «Lo apruebo». La implementación y su verificación quedaron completas, y el documento se cierra como `as-built`.

## 11. Verificación de la implementación

Comprobaciones realizadas en copias aisladas de Linux, sin modificar datos de desarrollo o producción:

- Esquema desde cero: 0 cuentas, 0 tickets y sello nulo. Repetir la migración conserva las
  11 cuentas, los 25 tickets y el hash de las cuatro tablas de configuración. SQL inválido
  devuelve código 3 y no imprime el mensaje de éxito.
- Asistente, SMTP de Mailpit, admin/admin, alta local, enlace de contraseña y entrada real
  desde localhost verificados en navegador. La API del asistente sellado devuelve 409.
- Seeder sobre instalación sellada: los cuatro ajustes conservan exactamente el mismo hash
  antes y después; se recrean 11 cuentas, 25 tickets y 5 adjuntos. Sobre base nueva y sello nulo
  deja método local, URL localhost y sello puesto.
- Keycloak real: autorización pública en localhost, tráfico interno del backend, prueba y
  guardado desde Configuración, acceso de una persona del reino y fragmento eliminado al volver.
- Selector vacío, campos visibles y rechazo 422 sin cambiar el método local verificados en
  PC y móvil. Las 216 pruebas de unidad del frontend pasan. Go: `go test ./...` y `go vet ./...` pasan.
- Suite con directorios: 214 casos, 193 aprobados, 20 omitidos y uno interrumpido por una
  recarga de Angular durante un ajuste de formato. Se conserva su informe, con salida 1.
  En una instalación nueva y estable, el caso interrumpido pasó en PC y móvil: 2 aprobados
  y 24 de directorio omitidos correctamente al no levantar los servicios. Entre ambas
  comprobaciones quedan 194 casos distintos verificados y 20 omisiones previstas en la suite completa.
  No se oculta ni se convierte en éxito la salida fallida de la primera pasada.
- LDAP presente: el mismo endpoint de detección devuelve 200; sin servicio los casos se omiten.
- Las pasadas nuevas comienzan con `down -v` exclusivo de `tests.yml`; la retirada de los
  contenedores y volúmenes conserva los informes y capturas. Las fuentes no reciben escrituras
  del runner y no se usan volúmenes, adjuntos ni correos del entorno del usuario.

No se han ejecutado estas comprobaciones en Windows o macOS. Producción permanece fuera de alcance.

Hallazgo durante la verificación sin servicios opcionales: `directorio.spec.ts` declaraba
`respondeElDirectorio`, pero no lo utilizaba antes de ejecutar sus casos. La documentación
prometía omitirlos cuando LDAP estuviera ausente. Se aplica la detección en `beforeAll`,
conforme al comportamiento ya aprobado en las secciones 6 y 8; no cambia el producto.
