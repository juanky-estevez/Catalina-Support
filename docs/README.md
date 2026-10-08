# Documentación de Catalina-Support

> **Estado:** as-built
> **Última actualización:** 2026-10-08
>
> **Bloque 7 implementado y verificado el 2026-10-08.** `arquitectura.md` cierra la coherencia
> documental, automatiza las fronteras del frontend y deja decidido que no habrá un sistema de
> logging en el navegador. Las suites de frontend, build y backend pasaron; producción no se tocó.
>
> **Actualizado el 2026-10-08 (sexta vez):** la visibilidad de «Mejorar con IA» quedó implementada.
> El detalle devuelve `capabilities.aiWriting`; la interfaz sólo monta el botón y el modal cuando es
> verdadera. Configuración se distingue de salud y la IA continúa obligatoria. Pasaron backend, 225
> pruebas del frontend y Playwright en PC y móvil; los tres documentos vuelven a **as-built**, el
> entorno desechable quedó limpio y producción no se tocó.
>
> **Actualizado el 2026-10-08 (quinta vez):** `modules/ai.md`, `modules/tickets.md` e
> `interfaz-y-experiencia.md` pasan a **propuesta** para ocultar «Mejorar con IA» cuando no existe una
> configuración activa. El responsable eligió 1A–3A: configuración y no salud, capacidad dentro del
> detalle y actualización al recargar. No quedaron decisiones abiertas y la propuesta fue aprobada.
> Durante la implementación se confirmó que la IA obligatoria impide abrir tickets en una instalación
> sellada sin configurar. El responsable decidió mantener esa regla: el estado falso se probará en
> backend y componente, y Playwright conservará el recorrido configurado. La corrección fue aprobada
> y quedó implementada.
>
> **Actualizado el 2026-10-08 (cuarta vez):** el asistente de redacción aprobado quedó implementado.
> Soporte y Desarrollo pueden mejorar descripciones y comentarios con cinco tonos, revisar la
> propuesta y aplicarla al borrador sin publicar. El backend deriva el destinatario y aplica los
> permisos del ticket antes de delegar en IA. Pasaron backend, 224 pruebas del frontend y Playwright
> en PC y móvil; los cuatro documentos vuelven a **as-built** y producción no se tocó.
>
> **Actualizado el 2026-10-08 (tercera vez):** `modules/ai.md`, `modules/tickets.md`,
> `usuarios-y-permisos.md` e `interfaz-y-experiencia.md` pasan a **propuesta** para añadir el
> asistente de redacción con IA elegido en 1A–19A. Mejora borradores de Soporte y Desarrollo con
> contexto mínimo, revisión obligatoria y sin persistir prompts ni resultados. No quedan decisiones
> abiertas; el responsable aprobó explícitamente los cuatro documentos el 2026-10-08.
> Al preparar el código apareció una contradicción con la frontera modular. El responsable eligió
> 20A: la pantalla llamará a `/api/tickets/{number}/writing/improve` y `tickets` delegará en `ai`.
> El responsable aprobó explícitamente esta corrección; `modules/ai.md` y `modules/tickets.md`
> vuelven a **aprobado** para implementar.
> Durante la implementación se encontró que el plazo HTTP de 60 segundos cortaría un motor autorizado
> a tardar 240. El responsable eligió 21A: sólo la ruta de redacción dispondrá de 270 segundos. Los
> dos documentos vuelven a **aprobado** tras su aprobación explícita.
>
> **Actualizado el 2026-10-08 (segunda vez):** el bloque 6 cerró el repaso de formato de todas las
> pantallas en ambos idiomas, ocho temas y tres resoluciones. Corrigió en el componente compartido el
> único hallazgo real, un desbordamiento de `/users` en inglés y móvil. Pasaron 220 pruebas unitarias
> y Playwright completo (180 aprobadas y 38 omisiones previstas); quedaron cero hallazgos abiertos.
> Los cuatro documentos vuelven a **as-built** y producción permaneció fuera del recorrido.
>
> **Actualizado el 2026-10-08:** el bloque 5 verificó fuente y checksum de los tres modelos, corrigió
> el cambio 1.5B→3B y completó la validación real. La activación usa RSS recuperable, vuelve a medir,
> restaura el origen ante fallos y rechaza 7B sin interrumpir 3B; la interfaz muestra 2/4/6 GiB y el
> error con memoria requerida/disponible. Pasaron `ai-manager`, backend, 220 pruebas del frontend y
> el recorrido afectado en PC y móvil. La limpieza no dejó recursos `block5`. Los seis documentos
> del bloque vuelven a **as-built** y producción continúa aparcada hasta cerrar la versión 1.0.0.
>
> **Actualizado el 2026-10-07 (séptima vez):** se abre el bloque 5. `modules/ai.md`, `ambientes.md` y
> `prueba-local.md` pasan a **propuesta** para validar los GGUF reales 1.5B, 3B y 7B en un entorno
> desechable. El responsable eligió 1A–9A y cerró el repaso con 10A–14A: se miden recursos, tiempos,
> contrato y recuperación; se prueban `/setup` y `/settings` en PC y móvil; el 7B no se activa por
> falta de RAM; 3B y 7B se retiran primero, y la limpieza final destruye también el 1.5B junto con
> todos los recursos temporales. No quedan decisiones abiertas; los tres documentos fueron
> aprobados explícitamente el 2026-10-08 y la validación puede comenzar.
>
> **Actualizado el 2026-10-07 (sexta vez):** revisión de coherencia documental, sin cambios de
> comportamiento. Se alinean los resúmenes vigentes con los cinco pasos de `/setup`, el idioma
> global, las cinco tablas de configuración de una fila y las dos auxiliares de IA, y el motor local
> que sólo se levanta para esa modalidad. También se cierran referencias que aún describían como
> pendientes los cuatro bloques ya implementados y verificados. Todos sus documentos permanecen
> **as-built** y producción continúa aparcada hasta cerrar la versión 1.0.0.
>
> **Actualizado el 2026-10-07 (quinta vez):** el bloque 4 detecta diferencias entre la arquitectura
> aprobada y los clientes/administrador actuales de IA. `arquitectura.md`, `ambientes.md` y
> `modules/ai.md` pasaron a **propuesta** para cerrar redirecciones con credenciales, supervisión de
> `llama-server`, recursos efectivos y pruebas aisladas de los cuatro adaptadores. El responsable
> eligió 1A, 2A, 3A y 4A; el repaso cerró 5A, 6A, 7A y 8A sin asuntos abiertos. Falta únicamente la
> aprobación documental, concedida explícitamente el 2026-10-07. El bloque vuelve a **as-built** tras
> pasar backend, administrador local, 218 pruebas del frontend y 180 recorridos E2E ejecutados.
>
> **Actualizado el 2026-10-07 (cuarta vez):** al abrir el bloque 3 se encontró una contradicción en
> `primer-arranque.md`: la sección aprobada intercambiaba Correo e IA frente al resto del documento,
> el código y las pruebas. El responsable corrigió el orden: Correo permanece en el paso 4 e IA en
> el paso 5. El repaso final decidió confiar en la activación guardada al sellar, reunir privacidad
> y costo en una sola confirmación y reanudar descargas locales. El documento vuelve a
> **as-built** tras verificar servicios, 218 pruebas unitarias y 180 recorridos E2E ejecutados.
>
> **Actualizado el 2026-10-07 (tercera vez):** `flujos.md` vuelve a **as-built**. Los ocho avisos de
> tickets y los tres correos de cuenta consultan el idioma global vigente al enviarse, sin separar
> destinatarios por preferencias personales. Producción continúa aparcada hasta cerrar la 1.0.0.
>
> **Actualizado el 2026-10-07 (segunda vez):** `usuarios-y-permisos.md`, `modules/auth.md` y
> `modules/users.md` vuelven a **as-built**. La columna `users.language`, los campos de API y el
> estado de sesión desaparecen; los correos de cuenta consultan el idioma global. Producción sigue
> aparcada hasta el cierre de la versión 1.0.0.
>
> **Actualizado el 2026-10-07:** se corrige la propuesta coordinada del idioma. Sólo las once
> plantillas de correo se traducen; siempre pueden editarse manualmente y sólo un proveedor externo
> exige confirmar la estimación económica. Los resúmenes existentes se conservan visibles en su
> idioma original; los nuevos o recalculados usan el idioma global vigente. La corrección fue
> aprobada explícitamente y habilita implementar.
>
> **Actualizado el 2026-10-06 (tercera vez):** pasan a **propuesta** las enmiendas coordinadas de
> `modules/ai.md`, `modules/settings.md`, `primer-arranque.md`, `usuarios-y-permisos.md`,
> `modules/auth.md`, `modules/users.md`, `modules/mail.md`, `modules/tickets.md`, `flujos.md`,
> `interfaz-y-experiencia.md`, `arquitectura.md` y `ambientes.md`. Definen IA obligatoria —local,
> servidor propio o proveedor— administrada desde
> la interfaz, e idioma único para interfaz, correos e IA. El repaso quedó cerrado sin decisiones
> abiertas y el responsable aprobó explícitamente la propuesta. La implementación puede comenzar.
>
> **Actualizado el 2026-10-06 (segunda vez):** `prueba-local.md`, `ambientes.md` y
> `arquitectura.md` vuelven a **as-built**. La corrección conserva el código de pruebas de sólo
> lectura, arranca su backend una vez con `go run .` y retira `backend_tmp`. La ejecución canónica
> sin perfiles opcionales terminó con 177 casos aprobados y 39 omitidos, salida 0; `down -v` eliminó
> todos los recursos.
>
> **Actualizado el 2026-10-06:** `interfaz-y-experiencia.md`, `modules/settings.md` y
> `primer-arranque.md` vuelven a **as-built**. La enmienda sustituye los avisos generales situados al
> principio de `/settings` y `/setup` por un toast compartido, visible desde la posición actual. Los
> avisos permanentes permanecen en su sección. El repaso quedó cerrado, el responsable aprobó los
> tres documentos y la implementación pasó 220 pruebas unitarias, la compilación y los recorridos
> afectados de Playwright en PC y móvil.
>
> **Actualizado el 2026-10-05 (tercera vez)**: `modules/settings.md` e
> `interfaz-y-experiencia.md` vuelven a `as-built` tras mostrar explícitamente en `/settings` la zona
> seleccionada, con el mismo campo de sólo lectura ya usado en `/setup`. No cambia la persistencia,
> la API ni la base de datos. El repaso quedó cerrado, el responsable aprobó explícitamente ambas
> propuestas y las pruebas pasaron en PC y móvil.
>
> **Actualizado el 2026-10-05 (segunda vez)**: `interfaz-y-experiencia.md` vuelve a `as-built` tras
> encontrar que `/mail` coloca a la derecha su flecha de regreso aunque la regla aprobada y el propio
> comentario de la vista dicen que va a la izquierda. El inventario confirma que las otras tres
> cabeceras con este control ya cumplen. El alcance no incluye enlaces de flujo dentro del contenido;
> el repaso se cerró, el responsable aprobó explícitamente la enmienda y el cambio pasó las pruebas
> en PC y móvil.
>
> **Actualizado el 2026-10-05**: `primer-arranque.md` vuelve a `as-built`: el paso 3 de `/setup`
> muestra la región seleccionada en un campo de sólo lectura antes del buscador. La enmienda fue
> repasada, aprobada, implementada y verificada en PC y móvil. `interfaz-y-experiencia.md` recoge la
> misma regla visual.
>
> **Actualizado el 2026-10-04**: `prueba-local.md` vuelve a `as-built` tras implementar y verificar
> la migración automática antes del backend, `/setup` inicialmente en inglés y los valores
> editables de Mailpit en desarrollo.
>
> **Actualizado el 2026-10-03**: se añade `prueba-local.md`, implementado y verificado en Linux, para el
> recorrido local del arranque vacío, el esquema con un comando, los ejemplos que conservan
> configuración, Keycloak local y la suite aislada. Su enmienda §5.1.b también fue aprobada explícitamente.
>
> **Actualizado el 2026-10-02**: revisión del README para quien prueba el producto: instalación
> vacía primero y ejemplos opcionales, por corrección del responsable; `admin`/`admin` en local y
> `ADMIN_PASSWORD` en producción. Se corrige el pendiente desactualizado del asistente en
> `arquitectura.md` (7 enmiendas) y se reportan los límites reales del despliegue actual en
> `ambientes.md` (19 enmiendas). No se modifica código ni se abre producción.
>
> **Actualizado el 2026-10-01 (séptima vez)**: **los datos de ejemplo se siembran con un contenedor de
> un solo uso**, `docker compose -f dev.yml run --rm seed`, **igual en Linux, macOS y Windows** —deja
> fuera el problema del guion de bash, que en Windows no se ejecuta, y con él la copia de los
> adjuntos—. El servicio vive en `dev.yml` **con perfil propio**, así que **`up -d` no lo arranca**, y
> **los ejemplos son opcionales** (el esquema sigue siendo obligatorio). El contenedor corre **el guion
> de siempre, `scripts/dev-seed.sh`**, que también sigue valiendo en Linux y macOS fuera del
> contenedor. Quedan enmendados `ambientes.md` (**18 enmiendas**: secciones 3.1, 3.3, 5 y 9.3) y
> `arquitectura.md` (**6 enmiendas**: sección 10.3), y el `README.md` quita el aviso de Windows y el de
> los adjuntos. Comprobado el 2026-10-01: **dos pasadas seguidas** del comando («Listo: 11 cuentas, 25
> tickets y 5 adjuntos» las dos veces), 11 cuentas y 25 tickets en la base, **5 adjuntos en `_files/`**
> y `up -d` sin arrancar el servicio.
>
> **Actualizado el 2026-10-01 (sexta vez)**: **el servidor de desarrollo de Angular reenvía `/api` al
> backend**, así que **se entra en `http://127.0.0.1:11001` con `admin`/`admin` sin nginx** —el
> recorrido de quien se descarga el proyecto y sólo tiene Docker—. Se añade `frontend/proxy.conf.json`
> y se declara en `serve.options.proxyConfig` de `frontend/angular.json`, con destino
> **`http://backend:11002`** (el servicio dentro de la red del entorno) y **sin reescribir la ruta**.
> Es **sólo de desarrollo**: el `build` no la ve y **en producción sigue reenviando nginx**. El
> `README.md` quita el aviso de «pendiente de decisión». Enmendados: `arquitectura.md` (**5
> enmiendas**: la sección 9, la 10, la 11 y la fila de la 13, y nginx deja de ser necesario para
> entrar) y `ambientes.md` (**17 enmiendas**: sección 3.1 y 9.3). Comprobado el 2026-10-01 por la
> 11001: `GET /api/auth/methods` y `GET /api/health` → **200 JSON**, `POST /api/auth/login` con
> `admin`/`admin` → **200** con token, y Playwright contra `http://127.0.0.1:11001` → **14 casos en
> verde** (entrar y salir con la cuenta de fábrica).
>
> **Actualizado el 2026-10-01 (quinta vez)**: **el esquema y los ejemplos se pueden aplicar sin `bash`**.
> Los pasos de local del `README.md` quedan como un recorrido para quien llega de fuera —clonar,
> levantar, aplicar el esquema, entrar con `admin`/`admin`—, con el porqué de cada paso, y **los dos
> comandos de Docker que valen igual en PowerShell, CMD y bash**. `./scripts/dev-seed.sh` queda dicho
> como **lo que es, de Linux y macOS** (y como lo que es: **un extra de contenido, no un requisito para
> entrar**). `ambientes.md` pasa a **16 enmiendas** (su sección 3.2 estrena los dos comandos y las
> secciones 3.1, 3.3 y 5 separan las dos vías) y `arquitectura.md` a **4** (su sección 10.3, que
> empezaba por `up -d` sin el esquema, ya lo trae). Comprobado el 2026-10-01 sobre una base nueva
> (`catalina_support_prueba`, borrada después): el esquema deja 18 tablas, `GET /api/auth/methods` y
> `POST /api/auth/login` con `admin`/`admin` responden **200 sin datos de ejemplo**, y el mismo archivo
> aplicado dos veces seguidas da **0 errores**.
>
> **Actualizado el 2026-10-01 (cuarta vez)**: el `README.md` deja claro, en los pasos de local, que
> **en una base nueva `./scripts/dev-seed.sh` es lo que crea el esquema** —no sólo los ejemplos—, que
> **hasta que no se pasa la aplicación no tiene tablas** y que los síntomas son
> `relación "installation_settings" does not exist` y `relación "ai_insights" does not exist`; y
> **reordena los pasos** (entorno, después esquema y datos, y sólo entonces abrir). `ambientes.md`
> pasa a **15 enmiendas**: sus secciones 3.1, 3.3 y 5 dicen qué aplica el esquema en desarrollo y
> **cuándo** —antes de abrir la aplicación— y apuntan al `README.md` en vez de repetirlo.
>
> **Actualizado el 2026-10-01 (tercera vez)**: **el directorio de pruebas y Keycloak dejan `dev.yml`**
> y pasan a **`active-directory.yml`** y **`keycloak.yml`**, cada uno con **su propio comando**, y **el
> perfil `auth` desaparece**. La suite de los caminos de directorio pasa a **dos pasos** —levantar cada
> servicio con su archivo y correr la suite normal—, y los dos entran en la red
> **`catalina-support-dev`**, que **posee `dev.yml`** con un nombre fijo (sin `external`). Quedan
> enmendados `arquitectura.md` (**3 enmiendas**), `ambientes.md` (**14**) y `modules/auth.md` (**7**).
> Y el `README.md` dice ya **la contraseña de fábrica de desarrollo** —la de `config/env/dev.env`, que
> se versiona—, dejando claro que **la de producción es `ADMIN_PASSWORD` y no se escribe**.
>
> **Actualizado el 2026-10-01 (segunda vez)**: el `README.md` estrena **la tabla de las once cuentas
> de ejemplo de desarrollo** —nombre, correo, rol y contraseña—, y `ambientes.md` pasa a **13
> enmiendas** al **dejar de repetir esa lista** y apuntar a la tabla del `README.md` (sección 3.3),
> para que el dato viva en un solo sitio.
>
> **Actualizado el 2026-10-01**: los **tres elementos opcionales** —el motor de IA, el directorio de
> pruebas y Keycloak— quedan con **un comando de Docker cada uno** y dicho que son opcionales:
> `ambientes.md` pasa a **12 enmiendas** (sección 9.3: el directorio y Keycloak comparten el perfil
> `auth` pero **pueden levantarse por separado** nombrando el servicio, sin activar el perfil
> —comprobado—, y el comando de siempre `--profile auth up -d` sigue levantando los dos),
> `modules/ai.md` a **2 enmiendas** (se corrigen los recursos del motor: **1500m de memoria**, **sin
> tope de CPU** y **~1,44 GiB medidos**, donde decía «950 MB / 1,09 GB» y «2 CPU»; y se explica el
> **aviso esperado del volumen `ai_modelos`**) y `arquitectura.md` a **2 enmiendas** (la fila del motor
> medido, en la sección 13, con las mismas cifras: **1500 MiB de tope**, **~1,44 GiB —el 98% del
> tope—** y **sin tope de CPU**).
>
> **Actualizado el 2026-09-30 (tercera vez)**: los dos documentos de módulo afectados por las pruebas
> del asistente quedan al día: `modules/mail.md` pasa a **6 enmiendas** —estrena **`Probar`**, que
> conecta y autentica sin mandar ningún correo— y `modules/settings.md` a **8 enmiendas** —declara
> **`ProberDeCorreo`**, que el asistente reutiliza para el paso 4—.
>
> **Actualizado el 2026-09-30 (segunda vez)**: **el asistente de primer arranque prueba lo que pide**
> —sus dos endpoints públicos `POST /api/setup/entry/test` y `POST /api/setup/mail/test`, con el mismo
> candado del sello, y los botones de los pasos 2 y 4—, y la prueba del correo **conecta y autentica
> sin mandar ningún correo**. `primer-arranque.md` pasa a **1 enmienda**
> (`docs/primer-arranque.md`, sección 3.1).
>
> **Actualizado el 2026-09-30**: **el motor de IA se configura desde la pantalla de Configuración**
> —su dirección y su modelo, con su botón de «Probar la conexión»—, con el entorno como respaldo
> (`docs/modules/settings.md`, decisión 16, y `docs/modules/ai.md`, decisión 21). `modules/ai.md`
> pasa a **as-built**.
>
> **Actualizado el 2026-09-27**: nace `modules/ai.md`, el sexto módulo, con los dos campos que
> redacta el motor de IA en la lista de tickets.
>
> **Actualizado el 2026-09-26 (tercera vez)**: los adjuntos de texto y código —un `.sql` no se podía
> mandar— y **la imagen y el vídeo con tope de 480 × 360** que se abren en el visor al pulsarlos
> (`docs/modules/tickets.md`, decisiones 54 y 55).
>
> **Actualizado el 2026-09-26 (segunda vez)**: las listas de tickets pasan a ser **Mis tickets** (lo
> mío), **Tickets principales** y **Tickets internos**, y se quitan del menú «Nuevo ticket» y de la
> bandeja «Asignármelo» (decisiones del responsable; `docs/modules/tickets.md`, decisiones 50 a 53).
>
> **Actualizado el 2026-09-26**: se fija **cómo se cuenta una enmienda** (sección «El registro de
> enmiendas y cómo se cuenta») y se ponen los números de la tabla con esa regla, porque hasta hoy
> unos cuadraban y otros no; y `modules/mail.md` y `modules/users.md` pasan a **as-built**, que era
> lo que les tocaba desde que su código existe.

Índice de la documentación del proyecto. La **Regla 0** de `AGENTS.md` es obligatoria: no se
escribe código sin un documento aprobado antes.

## Documentos

| Documento | Estado | Cubre |
| --- | --- | --- |
| `README.md` (este archivo) | as-built | Índice, convenciones y estados |
| `prueba-local.md` | **as-built** | Recorrido visual reproducible completado sobre datos desechables y limpiado |
| `arquitectura.md` | **as-built** | Bloque 7: coherencia documental y fronteras automáticas del frontend; sin logging de aplicación |
| `propósito-y-alcance.md` | **as-built** (3 enmiendas) | Qué problema resuelve la mesa de ayuda, los dos equipos y los cuatro papeles, el modelo de tickets (principal e interno, numeración y estados) y qué queda fuera. Sin decisiones abiertas. Las tres enmiendas son del 2026-09-22, antes de aprobarlo: el detalle del acceso, el repaso (el sexto aviso, el límite de 25 MB, las marcas de editado y eliminado) y la **regla 5** que cambió al escribir `flujos.md` |
| `usuarios-y-permisos.md` | **as-built** | Soporte y Desarrollo tienen el permiso acotado de mejorar sus borradores con IA |
| `flujos.md` | **as-built** | Los recorridos conservan destinatarios y datos; todos sus correos usan el idioma global vigente |
| `modules/settings.md` | **as-built** | Transporta memoria requerida/disponible en fallos de activación local |
| `modules/mail.md` | **as-built** | Usa el idioma global y genera borradores revisables o manuales de las once plantillas al cambiarlo |
| `modules/auth.md` | **as-built** | Identidad y sesión no llevan idioma personal; los correos de cuenta usan el idioma global |
| `modules/users.md` | **as-built** | Las cuentas no guardan ni exponen idioma; alta, ficha y perfil usan el contrato global |
| `interfaz-y-experiencia.md` | **as-built** | Oculta la ayuda de redacción cuando el detalle no declara una IA configurada |
| `primer-arranque.md` | **as-built** (6 enmiendas) | Cinco pasos: idioma y nombre, entrada, ubicación, correo e IA obligatoria; verificado en PC y móvil |
| `ambientes.md` | **as-built** | Matriz visual aislada verificada y limpiada; producción no se tocó |
| `modules/ai.md` | **as-built** | Expone internamente sólo si existe una configuración activa, sin comprobar salud |
| `modules/tickets.md` | **as-built** | Lleva `capabilities.aiWriting` en el detalle y la recalcula en cada lectura |

La cadena de producto **está completa**: siete documentos que cubren qué se construye, cómo se
comporta, cómo se ve y cómo se despliega. Además, **cada módulo tiene su documento**, que se escribe
justo antes de implementarlo, en este orden:

1. `modules/mail.md` — **entero**: backend y el editor.
2. `modules/settings.md` — **la pantalla de Configuración está hecha entera**: la marca, el color, el
   prefijo, el reparto y el idioma de la instalación.
3. `modules/auth.md` — **entero**: los tres caminos de entrada, hechos y verificados.
4. `modules/users.md` — **entero**: backend, las tres pantallas y las tres acciones del directorio.
5. `modules/tickets.md` — **entero**: backend y pantallas, con los adjuntos de los comentarios.

## Convenciones

### Cabecera obligatoria

Todo documento de `docs/` empieza así:

```markdown
# Título

> **Estado:** propuesta | aprobado | as-built
> **Última actualización:** AAAA-MM-DD
```

- `propuesta`: escrito, pendiente de aprobación. No se toca código de esa área.
- `aprobado`: aprobado por el responsable del proyecto. Habilita implementar.
- `as-built`: el código existe y el documento describe lo que hace hoy.

### El registro de enmiendas y cómo se cuenta

Debajo de la cabecera, cada documento lleva **su registro de enmiendas**, de la más reciente a la más
antigua: **una entrada por cambio**, con su fecha y qué cambió y por qué.

**Una enmienda es una entrada del registro que empieza por `Enmendado el …`.** Eso es lo que dice el
número que aparece en la tabla de arriba y en `AGENTS.md`, y se puede comprobar contando esas
entradas en la cabecera del documento. **`Pasa a as-built el …` no cuenta**: no cambia nada del
documento, cambia su estado.

Dos cosas se escriben aparte, y no son enmiendas:

- **La corrección del responsable**, cuando cambió lo propuesto, se anota **dentro de la entrada**
  («es una corrección del responsable», «yo proponía… y decidió…»).
- **Lo que se descubrió al implementar** también va dentro de su entrada, diciendo qué se encontró.

El registro es memoria de **por qué** un documento está como está. No se tachan ni se borran
entradas: si algo se corrige, se añade la entrada nueva y se dice qué queda corregido.

### Nombres y carpetas

- **Los documentos de producto** viven en `docs/`, y se llaman en español, en minúsculas, sin acentos
  ni espacios: `usuarios-y-permisos.md`.
- **Los documentos de módulo** viven en `docs/modules/` y se llaman **como el módulo**:
  `modules/auth.md`, `modules/users.md`. El nombre del módulo manda ahí, aunque sea en inglés, porque
  es el nombre que tiene su carpeta en el código.
- Un documento de módulo cuenta **cómo se construye** ese módulo. Lo que es decisión de producto —qué
  puede cada papel, cómo se comportan los tickets— vive en los documentos de `docs/`, y el de módulo
  no lo repite: lo da por escrito y apunta a él.

### Correspondencia con el código

La tabla documento ↔ código vive en `AGENTS.md`. Cuando un documento pase a `as-built`, hay que
añadir o actualizar su fila allí en el mismo cambio.

### Reglas de escritura

- Español, frases cortas, sin relleno y sin adornos de marketing.
- Cada documento dice **qué está decidido**, **qué está pendiente** y **qué queda fuera de alcance**.
- Lo que no esté decidido se marca como pendiente; no se rellena con suposiciones.
- Cuando un documento describa algo ya implementado, describe lo que el código hace de verdad,
  no lo que se pretendía.
