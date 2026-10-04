# Primer arranque: la instalación desde cero

> **Estado:** as-built
> **Última actualización:** 2026-10-03
>
> **Enmendado el 2026-10-03**, conforme a `docs/prueba-local.md` aprobado: El paso de Keycloak
> admite `internalIssuer` opcional. La preparación de Playwright sobre una base nueva recorre los
> cuatro pasos en PC y móvil antes de sellar, comprueba la conexión de Mailpit y termina en móvil.
> Esta verificación automatizada sustituye la limitación histórica del punto 3 de las desviaciones.
>
> **Enmendado el 2026-09-30**: **el asistente prueba lo que pide**, como prometía la §3 y la §6. Se
> añaden dos endpoints públicos más —`POST /api/setup/entry/test` y `POST /api/setup/mail/test`— con
> el mismo candado del sello (**409** `setup.alreadyInstalled`), el paso 2 y el paso 4 estrenan su
> botón de «Probar la conexión» y la prueba del correo **comprueba la conexión y la autenticación sin
> mandar ningún correo**, porque en el asistente no hay destinatario. Los datos que llegan se validan
> igual que al guardar, con las mismas claves. Con esto **quedan corregidas las desviaciones 1 y 1.b**
> del bloque de abajo. Cuenta como **1 enmienda** (`docs/README.md`).
>
> **Escrito a petición del responsable** (2026-09-30) y **aprobado e implementado el mismo día**: quería
> poder probar **la vista de primer arranque en otra máquina**, y que esa vista **pida toda la
> información necesaria para arrancar** —dominio, dirección, puerto, correo, método de entrada— sin
> tener que editar archivos en el servidor.
>
> **Lo que se desvió de esta propuesta, contado aquí** (Regla 0):
> 1. **Ni la prueba de conexión del directorio o Keycloak (paso 2) ni la del correo (paso 4) estaban en
>    el asistente** al implementarlo. La API de `/api/setup/**` no tenía endpoints de prueba y
>    **todavía no hay sesión** con la que probar: los botones de «Probar la conexión» de Configuración
>    piden ser Administrador, y aquí no hay ninguno. Los datos se guardaban y se validaban —el método
>    tiene que estar configurado para poder terminar—, y **la prueba se hacía después**. **Corregido el
>    2026-09-30**: los dos endpoints existen y son públicos, con el sello como único candado, y el
>    paso 2 y el paso 4 traen su botón (§3.1).
>
> 1.b. **La prueba del correo saliente desde el asistente no se hacía.** El paso 4 guardaba y validaba,
>    pero no probaba nada: para mandar un correo de prueba hay que tener ya el correo configurado y un
>    destinatario, y en el asistente no hay ninguno. **Corregido el 2026-09-30**: el paso 4 tiene su
>    botón y prueba **la conexión y la autenticación**, no el envío, que es lo que se puede comprobar
>    sin destinatario (§3.1).
> 2. **El asistente no tiene botón propio de «probar el motor de IA»**: el resumen **dice** si responde
>    —el backend lo pregunta al arrancar el paso y lo devuelve en `aiAvailable`—, que es lo que hace
>    falta saber antes de terminar.
> 3. **Limitación histórica, corregida el 2026-10-03 en `tests.yml`:** el camino completo de instalación no estaba en la suite de interfaz. El contenedor de las pruebas
>    no habla con la base y no se puede quitar el sello desde él, así que el recorrido de los cuatro
>    pasos se verifica **por la API, a mano** (los cuatro pasos, los 422 de cada paso, las dos pruebas
>    de conexión y el 409 al volver a configurar), y lo que sí está en la suite es **el candado**: en
>    una instalación ya configurada, `/setup` lleva a la entrada y la API lo rechaza.
> 4. **Los caminos de la API son** `POST /api/setup/installation`, `/entry`, `/entry/test`, `/location`,
>    `/mail`, `/mail/test` y `/finish`.

## 1. Qué problema resuelve

Hoy una instalación nueva **arranca ya configurada** con valores de fábrica, y la única puerta es **la
cuenta de fábrica**, cuya contraseña está en `config/env/*.env`. Para dejarla en serio hay que:

1. editar el archivo de entorno en el servidor (correo, contraseñas, dirección);
2. entrar con la cuenta de fábrica;
3. recorrer **Configuración** —que tiene ocho tarjetas— para poner el nombre, el idioma, el método de
   entrada, la región y la dirección.

**Funciona**, y está documentado en `docs/ambientes.md`. Pero **no se puede probar sin editar un
archivo**, y quien llega de fuera no sabe por dónde empezar. Esta propuesta añade **una vista de
instalación** que se enseña **sólo la primera vez** y pide lo necesario, en orden, con sus pruebas.

## 2. Cómo sabe la aplicación que no está instalada

**Un sello en la propia fila de la configuración**: `installation_settings.installed_at`
(`timestamptz`, **nulo** mientras nadie haya terminado el asistente). Va **en `v1.0.0.sql`** —la
versión sigue abierta— con **relleno para las instalaciones que ya existen**: una base que ya estaba
configurada **queda sellada** con la fecha de su última actualización, así que el asistente **no
aparece** en producción por el simple hecho de actualizar.

**Una instalación está «sin instalar» sólo si el sello es nulo.** Eso hace que el asistente:

- **se enseñe** en una instalación recién creada, y
- **no se pueda volver a recorrer** en una ya instalada, ni por la pantalla ni por la API
  (**409** `setup.alreadyInstalled`), que es lo que impide que alguien reescriba la configuración de
  una instalación en marcha.

## 3. Qué pregunta, y en qué orden

**Cuatro pasos**, con su línea de avance y un resumen final. Cada paso guarda al avanzar, así que
cerrar el navegador a medias no pierde lo hecho: al volver, el asistente sigue donde estaba.

| Paso | Qué pide | Por qué |
| --- | --- | --- |
| **1 · La instalación** | **El nombre** y **el idioma** | Es lo que se lee en la entrada, en el menú y en los correos |
| **2 · Cómo se entra** | **El método** —local, Active Directory o Keycloak— y **sus datos**, **con su prueba**: el botón comprueba lo que corresponda al método elegido. **La cuenta de fábrica se explica siempre**, sea cual sea el método: cómo se llama y **de dónde sale su contraseña** (`ADMIN_PASSWORD`, del entorno) | Es la puerta: sin esto no entra nadie, y la de fábrica entra con **cualquiera** de los tres métodos. **La contraseña no se pide aquí**: sigue en el entorno, y esa regla —que la puerta de fábrica **no dependa de la base**— no se toca |
| **3 · Dónde está** | **La región horaria** —de la lista con buscador— y **la dirección pública** —esquema, host y puerto—, con **el aviso si no es https** | Deciden cómo se leen las fechas y a dónde apuntan los enlaces de los correos |
| **4 · El correo** | **El servidor saliente**: host, puerto, TLS, usuario, contraseña y remitente, **con su prueba** | Sin él no sale ningún correo: ni un alta, ni un restablecer, ni un aviso |

**Y el resumen**, antes de terminar: lo que se ha configurado, **con la prueba del motor de IA**
(«responde» o «no está, la mesa de ayuda funciona igual») y un botón para **terminar la instalación**,
que es lo que pone el sello.

**Los tres métodos se pueden elegir desde el principio**, sin esperar a que estén configurados: en
Configuración un método sin sus datos se ofrece apagado, y aquí **no puede ser**, porque si no habría
forma de rellenar los campos del que se quiere usar. Elegirlo abre sus campos, y al avanzar se
comprueba que estén completos.

### 3.1 Las pruebas de conexión del asistente

Los dos pasos que configuran algo que hay que comprobar traen su botón de **«Probar la conexión»**,
como Configuración. **La prueba no guarda nada**: manda lo que hay escrito en pantalla y contesta si
eso funciona, así que se puede corregir antes de avanzar.

| Paso | Qué prueba | Qué **no** hace |
| --- | --- | --- |
| **2 · Cómo se entra** | **El directorio (AD) o Keycloak, según el método elegido**: con AD abre la conexión y valida la cuenta de servicio; con Keycloak lee el documento del reino. Con el método **local** no hay nada que conectar y el botón no se enseña. | No guarda nada, no valida la contraseña de nadie —eso sólo se sabe cuando alguien entra— y no toca la base. |
| **4 · El correo** | **La conexión y la autenticación del correo saliente**: host, puerto, TLS, usuario y contraseña. Es la misma conexión y la misma autenticación que usa el envío. | **No manda ningún correo**: no hay MAIL FROM, ni RCPT TO, ni DATA. En el asistente no hay destinatario, y por eso no se puede —ni se debe— probar el envío desde aquí. |

**Las contesta quien ya sabía hacerlo**, sin duplicar nada: el directorio y Keycloak los prueba el
módulo `auth` —el mismo que contesta a los botones de Configuración— y el correo lo prueba el módulo
`mail`, con la conexión que usa para enviar. El módulo `settings` declara lo que necesita
(`Prober`, `ProberDeCorreo`) y `main.go` se lo da, como manda la regla de modularidad
(`docs/arquitectura.md`, sección 4).

**Los datos que llegan se validan igual que al guardar**, con las mismas claves: el método tiene que
ser uno de los tres y estar configurado, el directorio y Keycloak tienen que estar completos, y el
correo necesita servidor, puerto y remitente con arroba. La contraseña o el secreto vacíos usan **lo
que ya esté guardado**, que es lo que permite probar sin volver a escribir un secreto que no sale
nunca por la API. Cuando la prueba no llega, la pantalla lee `settings.directory.unreachable`,
`settings.keycloak.unreachable` o `setup.mail.unreachable`, según el caso, y el detalle técnico queda
en el log.

## 4. Lo que el asistente **no** pregunta, y por qué

- **El modelo de IA.** Vive **en los contenedores** (`ai.yml`), y el asistente sólo **prueba** el
  motor. El modelo ocupa ~1,1 GB y se sirve desde un volumen: es un asunto de la máquina que lo
  levanta, no de la configuración del producto. El README dice cuál se baja por defecto, cómo
  cambiarlo por otro `.gguf` y **qué recursos consume**.
- **La contraseña de la cuenta de fábrica.** Sigue siendo `ADMIN_PASSWORD` del entorno
  (`docs/usuarios-y-permisos.md`, sección 8). El asistente **cuenta de dónde sale** en lugar de
  pedirla, porque si la contraseña viviera en la base, la puerta de fábrica **dejaría de ser
  independiente** del resto de la configuración.
- **El certificado y el dominio de nginx.** Son del servidor: van en el README, no en la aplicación.

## 5. El correo saliente se muda de sitio

Hoy el SMTP se lee del **entorno**; el paso 4 lo pide en la pantalla, así que pasa a **vivir en la
base**, como el directorio y Keycloak (`docs/modules/settings.md`). Se hace con el patrón que ya está
en el repositorio: **el módulo `mail` declara lo que necesita** (`Instalacion`, una interfaz con el
SMTP resuelto) y **`main.go` se lo da** desde el módulo de configuración. **Y las variables `SMTP_*` se retiran del entorno**, por la misma razón que se retiraron las del
directorio y Keycloak (corrección del responsable, 2026-09-30): **un dato que vive en la base no se
configura en dos sitios**. En desarrollo el archivo de ejemplos deja puesto el buzón de pruebas, y
**una instalación que tuviera el correo en el entorno lo pone una vez** en la vista de instalación o
en Configuración.

## 6. Dónde vive, y con qué se prueba

- **La pantalla**: una ruta propia, `/setup`, **sin el armazón del menú** —no hay nada que navegar
  todavía— y con la marca de fábrica.
- **La API**: `/api/setup/**`, en el módulo `settings`, **disponible sólo mientras el sello sea
  nulo**. Guardar dos veces la misma instalación se rechaza con **409** y su clave. Además de los cinco
  caminos que guardan y sellan, están **las dos pruebas** del paso 2 y del paso 4 —`POST
  /api/setup/entry/test` y `POST /api/setup/mail/test`—, que reciben los datos escritos, **no tocan la
  base**, contestan `{"status":"ok"}` o la clave del fallo, y **con la instalación sellada contestan
  también 409** `setup.alreadyInstalled` (§3.1).
- **Las pruebas**: las unitarias de cada paso y **las dos de los endpoints de prueba** —con dobles, sin
  salir a la red: el 409 con el sello puesto, la validación de los datos malos y que con datos buenos
  se llama a la prueba del módulo—, la del remitente del correo —que conecta y autentica sin mandar
  nada— y **la que más importa en la suite de interfaz**: que **una instalación ya configurada no
  enseña el asistente y su API lo rechaza**. El recorrido de los cuatro pasos desde cero no está en la
  suite —el contenedor de las pruebas no toca la base y no puede quitar el sello—: se verifica por la
  API, a mano (desviación 3).

## 7. Lo que queda fuera, y hay que decirlo

- **Elegir el modelo de IA desde la aplicación** no entra: queda dicho arriba por qué.
- **Reinstalar**: una vez sellada, la configuración se cambia en **Configuración**. Para volver a
  empezar de cero hay que **borrar la base** y aplicar el esquema otra vez, y eso es del servidor.
- **Mientras el asistente esté sin terminar, la instalación no se expone**: cualquiera que llegue a
  una instalación sin sellar puede configurarla. El README lo dirá en su paso a paso: **primero se
  configura, después se abre el dominio**.

## 8. Qué habilita este documento

Aprobar esta propuesta habilita: la columna del sello en `v1.0.0.sql` con su relleno, la ruta
`/setup` con sus cuatro pasos y su resumen, los endpoints `/api/setup/**` con su candado, el correo
saliente en la base —y fuera del entorno—, y las pruebas de las tres capas. **Nada de eso se
escribe hasta que el responsable lo apruebe.**
