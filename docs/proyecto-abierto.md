# Proyecto abierto, contribuciones y marca

> **Estado:** aprobado
> **Última actualización:** 2026-10-10
>
> **Segunda enmienda aprobada e implementada el 2026-10-10.** La dependencia directa
> `github.com/juanky-estevez/go-logs v1.0.0` no publicaba una licencia. El responsable eligió 27A,
> aprobó la enmienda y autorizó la publicación concreta: `go-logs v1.0.1` contiene MIT y Catalina
> Support ya resuelve esa versión. La sección 11 conserva la evidencia.
>
> **Tercera enmienda aprobada e implementada el 2026-10-10.** El frontend generaba avisos de
> terceros al compilar, pero el repositorio no conservaba un inventario equivalente y comprobable
> para el backend. El responsable cerró 28A–37A; la sección 12 registra el archivo versionado, los
> textos completos, su control automático y la evidencia.
>
> **Primera enmienda aprobada e implementada el 2026-10-10.** La auditoría encontró seis avisos npm:
> dos críticos y cuatro altos. El responsable eligió y aprobó 26A; Angular quedó actualizado de
> forma coordinada a 22.2.2, el lockfile se regeneró y la sección 10 conserva la evidencia.

## 1. Objetivo

Catalina Support se publicará como proyecto de código abierto para que cualquier persona pueda
examinarlo, instalarlo, probarlo, modificarlo y contribuir. La edición comunitaria seguirá siendo
el producto completo: no tendrá límites artificiales de usuarios, tickets ni funciones.

La publicación persigue primero que el proyecto se pueda evaluar como producto y como trabajo de
ingeniería. También conserva caminos futuros de sostenimiento: servicio alojado, instalación y
soporte administrados, consultoría y licencias comerciales para organizaciones que no puedan usar
AGPL. Ninguna de esas ofertas forma parte de este trabajo.

El repositorio continuará inicialmente en la cuenta personal
`juanky-estevez/Catalina-Support`. Podrá transferirse más adelante sin cambiar estas reglas.

## 2. Licencia del código

El código que se publique a partir de este cambio usará **GNU Affero General Public License,
versión 3 únicamente (`AGPL-3.0-only`)**. El archivo raíz `LICENSE`, la declaración del README y los
metadatos del repositorio deben coincidir. Los archivos nuevos que necesiten un identificador SPDX
usarán `AGPL-3.0-only`.

El cambio no revoca permisos ya concedidos: quien haya recibido una copia bajo MIT podrá seguir
usando esa copia bajo MIT. La nueva licencia se aplica a la versión publicada después del cambio y
a sus contribuciones posteriores.

El titular inicial será **Juan C. Estevez**. La AGPL permite uso comercial, modificación y
redistribución; obliga a conservar sus avisos y a ofrecer el código fuente correspondiente, también
cuando una versión modificada se usa para prestar el programa mediante una red. El README explicará
esta consecuencia en lenguaje directo y enlazará el texto completo de la licencia, sin sustituirlo.

## 3. Contribuciones y CLA

El repositorio incorporará un perfil comunitario completo:

- `CONTRIBUTING.md`, con el ciclo documentación antes que código, preparación del entorno, pruebas,
  formato de commits y proceso de revisión;
- `CODE_OF_CONDUCT.md`, basado en Contributor Covenant y con un canal privado de contacto;
- `SECURITY.md`, con versiones atendidas y el uso de GitHub Private Vulnerability Reporting;
- `SUPPORT.md`, que separa ayuda de uso, propuestas y reportes de seguridad;
- plantillas de incidencias, solicitud de función y pull request;
- `CLA.md` y el registro mínimo de aceptaciones;
- `TRADEMARKS.md`, con la política descrita en la sección 4.

Las discusiones generales vivirán en GitHub Discussions; los defectos y trabajos concretos, en
Issues. Las vulnerabilidades no se publicarán como issue: se enviarán por Private Vulnerability
Reporting.

El CLA seguirá el modelo de licencia de Harmony: cada contribuyente conserva sus derechos y concede
a Juan C. Estevez permisos suficientes para distribuir la contribución bajo AGPL y también bajo
otras licencias, incluidas licencias comerciales o propietarias. No es una cesión de titularidad.
Se exigirá para código, recursos gráficos y documentación sustancial; no para abrir incidencias ni
para correcciones triviales de erratas.

Cada pull request sujeto al CLA tendrá una comprobación automática. Una GitHub Action fijada a un
hash de commit, no a una etiqueta mutable, registrará en el propio repositorio sólo el usuario de
GitHub, la versión del CLA, la fecha y el pull request. El registro será público. La contribución no
se integrará mientras falte la aceptación. Antes de adoptar una licencia comercial se recomienda
revisión jurídica independiente del texto definitivo y de la cadena de derechos.

## 4. Marca

`TRADEMARKS.md` separará los permisos de la licencia del código de los permisos sobre el nombre y
los logotipos de Catalina Support.

Una compilación oficial sin cambios puede conservar el nombre y los logotipos. También puede
conservar la atribución a Catalina Support una instalación oficial que cambie desde Configuración
su nombre, logotipo o color: esa personalización prevista por el producto no la convierte en un
fork.

Un fork o una compilación con cambios puede explicar de forma nominativa que está «basado en
Catalina Support», pero debe usar un nombre y una identidad visual distintos y no puede dar a
entender que es oficial, patrocinado o aprobado. Cualquier uso adicional requiere autorización
escrita. La política permitirá referencias veraces en artículos, comparativas y enlaces.

## 5. Atribución visible y fuente correspondiente

La pantalla de entrada y la zona inferior del menú mostrarán **«Código fuente y licencia»**. Será un
enlace discreto y accesible, separado de las acciones principales. Llevará al repositorio y
explicará `AGPL-3.0-only`.

La marca pública que ya entrega el backend añadirá `license` y `sourceUrl`, de modo que ambas vistas
usen la misma fuente. Para una versión publicada, `sourceUrl` apuntará a la etiqueta que corresponda
exactamente a `version`. Para una compilación de desarrollo apuntará al repositorio e identificará
la compilación como desarrollo. El frontend no duplicará esos valores en constantes.

La URL no se podrá cambiar desde Configuración: identifica el código distribuido, no la instalación.
Una compilación derivada debe modificar sus metadatos de construcción para enlazar su propia fuente
correspondiente. Si los datos faltan por compatibilidad con un backend anterior, la interfaz no
mostrará un enlace inventado.

## 6. Auditoría previa a publicar

El repositorio seguirá privado mientras se prepara y revisa un resultado concreto. La auditoría
abarcará el árbol actual y todo el historial de Git, con búsqueda de secretos, credenciales, datos
personales, archivos grandes y referencias a infraestructura real. También comprobará dependencias,
avisos de terceros, archivos generados, permisos de contenedores y que una persona externa pueda
seguir el README desde una instalación vacía.

Ya se conocen dos hallazgos:

1. dieciocho commits contienen `<correo personal retirado>` como correo del autor;
2. documentos y configuraciones versionados contienen dominios y rutas del despliegue real.

Antes de publicar se reescribirán los autores afectados con el correo `noreply` de GitHub. Esta
operación cambiará todos los hashes descendientes; se hará una sola vez, después de terminar los
cambios y guardar una referencia privada recuperable. Se verificará el historial resultante antes
de retirar esa referencia.

Los dominios, rutas y datos propios de la infraestructura se sustituirán por ejemplos y variables.
Los valores reales permanecerán sólo en configuración no versionada. La documentación seguirá
siendo reproducible y distinguirá con claridad ejemplos de valores que el operador debe elegir.

**Evidencia local del 2026-10-10.** Gitleaks revisó por separado el conjunto exacto de archivos que
se publicará y los 19 commits anteriores. Los únicos cuatro avisos del árbol eran dos claves
declaradas como desarrollo, la misma clave en el entorno aislado de pruebas y un JWT sin firma que
prueba el rechazo de `alg=none`; quedaron permitidos por ruta y contenido exactos. Dos huellas
históricas de esos mismos valores se registran por commit, archivo, regla y línea. Los secretos de
producción y tokens encontrados en el directorio de trabajo viven exclusivamente en
`config/env/prod.env` y `tmp/`, ambos ignorados y ausentes del índice. Con esas excepciones acotadas,
el conjunto publicable y el historial terminaron sin hallazgos. El mayor blob histórico es un logo
de 603 822 bytes; no hay archivos grandes inesperados.

Se creó y verificó el respaldo privado
`/tmp/catalina-support-before-public-history.bundle`, SHA-256
`ca61d18cac089c71faf576040df4e39ec0533fde9629b61888ef3404156a9`. Después se reescribieron una
sola vez `main` y `dev`: todos los autores y committers personales usan ahora el correo público
`noreply`, los dominios reales pasaron a `example.com` y las rutas reales a
`/srv/catalina-support`. Las ramas locales resultantes son `main` en `b9df02f` y `dev` en
`1260932` antes de incorporar la actualización de huellas de Gitleaks. La publicación remota de
estas ramas y la retirada de las referencias locales antiguas quedan para después de la
verificación final del historial reescrito.

Si aparece un secreto o un dato sensible adicional, la publicación se detendrá y el hallazgo se
presentará antes de limpiar o reescribir el historial. No se asumirá que borrar el archivo actual
lo elimina de Git.

## 7. Preparación de GitHub y publicación

Antes del cambio de visibilidad se prepararán y verificarán:

1. licencia, avisos, documentos comunitarios y plantillas;
2. CLA automático y protección de la rama para exigir sus comprobaciones junto con las pruebas;
3. Issues, Discussions y Private Vulnerability Reporting;
4. descripción, temas y enlace del repositorio;
5. auditoría limpia del árbol y del historial reescrito;
6. instalación local desde cero siguiendo únicamente el README;
7. enlaces de fuente y licencia en la pantalla de entrada y el menú;
8. una comparación final entre el repositorio local y lo que quedará público.

La publicación será de una versión **en desarrollo y previa a 1.0**. No se crea ahora una etiqueta de
candidata ni `v1.0.0`. El README lo indicará sin presentar estabilidad que aún no existe.

Cambiar la visibilidad de GitHub será el último paso y requerirá una aprobación explícita sobre el
resultado ya preparado. Este documento no autoriza abrir el repositorio, desplegar producción,
abrir el dominio, publicar una versión, aceptar donaciones ni ofrecer un plan comercial.

## 8. Verificación y cierre

Para pasar este documento a `as-built` deben quedar registrados:

- la revisión del texto de licencia, CLA, marca y archivos comunitarios;
- el resultado del escaneo del árbol y de todas las revisiones del historial;
- la comprobación de que no permanece el correo personal ni infraestructura real;
- la prueba del flujo de CLA en un pull request de ensayo;
- las pruebas de backend y frontend afectadas por `license` y `sourceUrl`;
- el recorrido en PC y móvil del enlace en entrada y menú;
- la instalación limpia ejecutada desde el README;
- `git diff --check` y el estado final de los documentos.

La visibilidad permanecerá privada durante estas verificaciones. Una vez preparado el resultado,
se presentará al responsable el inventario de cambios, hallazgos y evidencia para que decida si se
hace público.

## 9. Decisiones y repaso

El responsable eligió 1A–5A: AGPL-3.0-only, política de marca separada, CLA ligero, auditoría antes
de publicar y edición comunitaria completa. Eligió 6A–10A: titularidad personal inicial, CLA
verificable por pull request, alcance de contribuciones sustanciales, reporte privado de seguridad y
parada ante hallazgos sensibles. Eligió 11A–15A: enlace visible, uso nominativo de marca, repositorio
personal, herramientas comunitarias completas y aprobación final antes de cambiar la visibilidad.

En el repaso eligió 16A–20A: CLA sin cesión y con permiso de relicencia, registro local fijado por
hash, fuente correspondiente a la versión, conjunto comunitario completo y publicación previa a
1.0. Finalmente eligió 21A–25A: reescribir el correo personal, sustituir infraestructura real,
registro público mínimo del CLA, metadatos servidos por el backend y atribución conservada en las
personalizaciones oficiales. No quedan decisiones abiertas.

## 10. Enmienda implementada: dependencias Angular vulnerables

La auditoría del árbol encontró seis avisos en el lockfile vigente: dos críticos y cuatro altos. El
aviso de `@angular/router` afecta a renderizado en servidor, que Catalina Support no usa, pero la
versión permanece dentro del rango vulnerable. Los avisos críticos llegan por la herramienta de
compilación y `piscina`; los demás afectan a herramientas transitivas del CLI y a mapas de fuente.
Todos tienen corrección disponible.

El frontend actualizará de forma coordinada los paquetes Angular a **22.2.2** y regenerará
`package-lock.json` con el gestor declarado por el proyecto. No se aplicará una actualización mayor,
no se desactivarán avisos y no se añadirán excepciones de auditoría. Se revisará el diff del lockfile
para comprobar que sólo resuelve la nueva familia y sus transitivas.

El criterio de cierre será:

1. `npm audit` sin vulnerabilidades altas o críticas;
2. las fronteras, las pruebas unitarias y la compilación completa del frontend aprobadas;
3. el recorrido Playwright afectado en PC y móvil aprobado;
4. ausencia de cambios funcionales fuera de compatibilidad con Angular 22.2.2;
5. registro del resultado real en este documento antes de volver a `as-built`.

El responsable eligió 26A. La actualización se integra en la auditoría previa a publicar; no cambia
la licencia, la política de marca, el CLA ni la decisión separada sobre la visibilidad del
repositorio. No quedan decisiones abiertas.

**Evidencia del 2026-10-10.** Todos los paquetes Angular directos quedaron fijados en 22.2.2 y el
lockfile se regeneró con el npm declarado por el proyecto. `npm audit --audit-level=high` terminó
con cero vulnerabilidades; las 227 pruebas unitarias del frontend, la compilación completa y los 12
casos afectados de Playwright en PC y móvil pasaron. El entorno desechable se retiró al terminar.

## 11. Enmienda implementada: licencia de `go-logs`

La auditoría de dependencias encontró que `github.com/juanky-estevez/go-logs v1.0.0`, usado por el
backend para sus cuatro niveles de registro, no contiene archivo ni declaración de licencia. Su
árbol publicado incluye código, pruebas y configuración genérica; no contiene credenciales. Al no
existir permiso explícito de redistribución, Catalina Support no se publicará dependiendo de esa
versión.

El repositorio `go-logs`, también titularidad de Juan C. Estevez, incorporará una licencia **MIT** y
un aviso de copyright de 2026. No cambiará su API ni su comportamiento. Después de verificar sus
pruebas se creará y publicará la etiqueta **`v1.0.1`**, cuyo contenido incluirá la licencia. No se
retaggea `v1.0.0` ni se reescribe su historial.

Catalina Support actualizará `go.mod` y `go.sum` a `github.com/juanky-estevez/go-logs v1.0.1`,
descargará el módulo desde su origen y comprobará que el artefacto resuelto contiene `LICENSE`. La
suite completa del backend y `go vet ./...` deben pasar sin cambios de comportamiento ni de formato
de registros.

La etiqueta y el push de `go-logs` son publicaciones externas e irreversibles en la práctica. Se
presentará primero el commit concreto del repositorio externo y se pedirá una aprobación final para
publicarlos. Si el repositorio tiene sus propias instrucciones de contribución, se aplicarán antes
de preparar ese commit.

El responsable eligió 27A. Esta enmienda corrige únicamente la cadena de licencias; no incorpora el
código del módulo a Catalina Support ni cambia su política AGPL. No quedan decisiones abiertas.

**Evidencia del 2026-10-10.** El commit publicado de `go-logs` es
`6c13674c9f9c5794302fcae002409fc30e51836d`, con autor y committer en el correo público `noreply`;
`main` y la etiqueta anotada `v1.0.1` apuntan a ese cambio. El módulo pasó `go test ./...` y
`go vet ./...` en su contenedor. Catalina Support actualizó `go.mod` y `go.sum` a `v1.0.1`, comprobó
el `LICENSE` MIT dentro del artefacto descargado y volvió a pasar toda la suite y `go vet ./...` del
backend.

## 12. Enmienda implementada: avisos de terceros

La auditoría confirmó que la compilación de Angular produce `3rdpartylicenses.txt`, mientras que el
backend no genera ni conserva un inventario equivalente. Las dependencias directas revisadas tienen
licencias compatibles con la publicación prevista; `go-logs`, que era la excepción sin permiso
explícito, quedó corregida en la sección 11. El hueco restante es de trazabilidad: una persona que
examina el repositorio no dispone de una relación única y reproducible de los componentes de
terceros que se distribuyen.

El repositorio incorporará **`THIRD_PARTY_NOTICES.md`**, generado de forma determinista y
versionado. Incluirá las dependencias de producción y las herramientas cuyos componentes se
distribuyan con la aplicación; excluirá las dependencias usadas únicamente para desarrollo o
pruebas. El conjunto se calculará con las versiones bloqueadas y el grafo que alcanza la
compilación de producción, no sólo con las dependencias declaradas directamente. Cada entrada del
índice identificará el componente, la versión resuelta, la licencia y un enlace estable a su origen.
Los avisos y textos completos que las licencias obliguen a conservar vivirán, también versionados,
en **`third_party_licenses/`**; el índice compacto enlazará cada componente con su texto local.

Un guion del repositorio regenerará el archivo desde las fuentes de dependencias bloqueadas, sin
resolver versiones nuevas. Regenerará tanto el índice como `third_party_licenses/`. La CI ejecutará
el mismo generador en modo de comprobación y fallará si cualquiera de los archivos versionados
difiere, si una dependencia incluida no declara licencia o si no se puede conservar un aviso
obligatorio. Así, un cambio en `go.mod`, `go.sum`, `package.json` o `package-lock.json` no podrá
integrarse dejando el inventario atrás.

La compilación del frontend seguirá produciendo su archivo propio dentro del artefacto. El inventario
raíz no lo sustituye: ofrece revisión en GitHub y cubre además el backend. Las imágenes externas y
las imágenes base quedan fuera de este archivo: el proyecto hoy las referencia o construye
localmente, pero no publica imágenes propias. Si más adelante distribuye imágenes, ese trabajo debe
auditar y entregar los avisos de sus capas.

Una licencia ausente, desconocida o incompatible detendrá tanto la generación como la CI. No habrá
una lista de excepciones que convierta ese error en advertencia. La implementación no cambia
licencias de terceros, no incorpora su código al repositorio y no altera el comportamiento de la
aplicación. `scripts/prod-build.sh` copiará el índice y la carpeta de licencias junto a los
artefactos publicados del backend y del frontend; un despliegue preparado por el guion conservará
los avisos aunque se separe del clon de Git.

El responsable eligió 28A–30A: archivo versionado, alcance de producción y comprobación automática
en CI. En el primer repaso eligió 31A, 32B, 33B y 34A: grafo real sobre versiones bloqueadas,
imágenes fuera del inventario actual, índice compacto con enlaces y bloqueo de licencias
desconocidas. La elección 33B corrigió la presentación propuesta por el agente. En el repaso final
eligió 35A–37A: los textos exigidos quedan en una carpeta separada y versionada, generada y
comprobada junto al índice, y ambos se copian a los artefactos. De ese modo el índice sigue compacto
sin depender de enlaces externos para cumplir los avisos. No quedan decisiones abiertas.

**Evidencia del 2026-10-10.** `scripts/third-party-notices.go` recorrió el grafo normal de
compilación del backend y las entradas no marcadas como desarrollo en `package-lock.json`: generó
29 componentes —19 de Go y 10 de npm— con versión, licencia, fuente y texto local. Una segunda
ejecución en modo `--check` no produjo diferencias. La prueba negativa añadió contenido a un texto
generado y el control falló nombrando exactamente ese archivo; después el guion documentado lo
regeneró y volvió a pasar. El generador pasó `gofmt` y `go vet`; los guiones pasaron `bash -n` o
`sh -n`, y `git diff --check` pasó. `npm ci` resolvió 311 paquetes desde el lockfile con cero
vulnerabilidades.

La CI ejecuta el control en un trabajo propio con Go y Node. `scripts/prod-build.sh` copia el índice
y los textos a cada artefacto después de comprobar que existe. Esa copia no se ejecutó contra el
despliegue real: producción continúa aparcada hasta el cierre de 1.0 y `docs/ambientes.md` conserva
el hallazgo previo que impide presentar hoy el recorrido completo como reproducible desde cero.
