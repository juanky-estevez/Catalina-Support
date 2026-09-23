# Ambientes: despliegue y pruebas

> **Estado:** aprobado
> **Última actualización:** 2026-09-22
>
> Aprobado por el responsable el 2026-09-22, tras repasarlo en forma de preguntas mientras se
> escribía. Fija el despliegue (un script con la misma forma que el de Calibyou), las migraciones
> —un archivo por versión, con tres dígitos—, las copias de seguridad y las tres capas de pruebas.
>
> Con este documento, **la cadena de documentación queda completa**: no hay más documentos
> previstos, y lo que falte se decide en el documento del área que corresponda.

## 1. Alcance de este documento

Cuenta **cómo se despliega y cómo se prueba**: el runbook de producción, las migraciones, las copias
de seguridad y las tres capas de pruebas.

**No describe la arquitectura ni los contenedores**: eso está en `docs/arquitectura.md` (secciones
2, 10 y 11, con los puertos, las imágenes y los vhosts). Aquí no se repite, y cuando algo cambie en
un sitio se cambia allí.

Es una **propuesta**: no habilita escribir código hasta que esté aprobada (Regla 0). La sección 9
recoge lo que he propuesto yo.

## 2. Los dos entornos de un vistazo

| | Desarrollo | Producción |
| --- | --- | --- |
| **Dominio** | `dev-catalina-support.calibyou.com` | `catalina-support.calibyou.com` |
| **Puertos** | frontend 11001, backend 11002, base 11003 | frontend 21001, backend 21002, base 21003 |
| **Código** | Montado como volumen: se edita y el contenedor recarga | **Artefactos construidos** en `/root/prod/catalina-support` |
| **Recarga** | Angular y `air` recargan al guardar | Hay que desplegar |
| **Base de datos** | Volumen de Compose, se puede vaciar sin pensar | Datos reales: **la migración se aplica una sola vez**, al cerrar la 1.0.0 |
| **Estado hoy** | **En pie y verificado** | **Aplazado** hasta cerrar la 1.0.0 (ver `docs/arquitectura.md`, sección 13) |

Mientras el despliegue esté aplazado, el dominio de producción responde 503 con un aviso de que está
en desarrollo. No es un fallo: es lo decidido.

## 3. Desarrollo

### 3.1 Levantar, mirar y parar

```bash
docker compose -f dev.yml up -d
docker compose -f dev.yml ps
docker compose -f dev.yml logs -f backend
docker compose -f dev.yml down          # los volúmenes se conservan
```

Se entra por `https://dev-catalina-support.calibyou.com`. Los puertos 11001 y 11002 siguen publicados
en `127.0.0.1` para depurar sin pasar por nginx.

### 3.2 La base de datos

- El esquema se aplica **repetidas veces** sin miedo: `v1.0.0.sql` es transaccional e idempotente
  (`docs/arquitectura.md`, sección 7). **El archivo todavía no existe**: nace con el primer módulo,
  cuando `docs/tickets.md` se implemente.
- Aplicarlo, dentro del contenedor (comando probado el 2026-09-22):

```bash
docker compose -f dev.yml exec -T database \
  psql -v ON_ERROR_STOP=1 -U catalina_support -d catalina_support -p 11003 \
  < backend/migrations/v1.0.0.sql
```

- **Vaciar la base de desarrollo** es tan simple como borrar el volumen:

```bash
docker compose -f dev.yml down
docker volume rm catalina_support_dev_database
docker compose -f dev.yml up -d
```

Y volver a aplicar la migración. No hay nada que conservar: los datos de desarrollo son de prueba.

### 3.3 Dónde queda lo que no es código

| Carpeta | Qué guarda | ¿En git? |
| --- | --- | --- |
| `_logs` | Los archivos de `go-logs`, un archivo por día | La carpeta sí, los `.log` no |
| `_files` | Los adjuntos de los tickets | La carpeta sí, los archivos no |

Las dos son volúmenes montados en el backend: sobreviven a `down` y se copian desde la máquina sin
entrar al contenedor.

## 4. Producción

### 4.1 Cómo está montada

Contenedores «finos»: el runtime vive en imágenes base y **el código son artefactos construidos** en
`/root/prod/catalina-support` (frontend, backend, `_logs`, `_files`). Los `builder_*` de `prod.yml`
existen para construir y publicar, y **no se levantan con `up`**: están bajo el perfil `build`.

### 4.2 El despliegue, paso a paso

Un script, `scripts/prod-build.sh`, con el mismo guion que el de Calibyou. **Se escribirá al aprobar
este documento** (Regla 0: primero el documento). Lo que hace:

1. **Avisa si el árbol de trabajo tiene cambios sin commitear**: un artefacto construido desde un
   árbol sucio no corresponde a ningún commit, y eso hay que saberlo antes y no después.
2. **Rota el artefacto anterior** de cada componente a `<nombre>_prev`, **por copia y no por
   movimiento**, porque el directorio está montado en los contenedores: renombrarlo los dejaría
   apuntando al inodo viejo.
3. **Construye y publica** con `docker compose -f prod.yml --profile build run --rm builder_<nombre>`.
4. **Verifica que el artefacto existe** antes de seguir: `frontend/index.html` y
   `backend/catalina-support` ejecutable. Si falta, corta sin desplegar.
5. **Escribe `BUILD_INFO`** en `/root/prod/catalina-support` con el commit, la rama, si el árbol
   estaba sucio, la fecha y qué artefactos se construyeron. Es la respuesta a «¿qué hay desplegado
   ahora mismo?» sin adivinar.
6. **Despliega**: `docker compose -f prod.yml up -d` y **reinicia el backend**. El frontend no se
   reinicia: el contenedor de nginx lee los estáticos del disco en cada petición.

```bash
./scripts/prod-build.sh                 # construye, publica y despliega
./scripts/prod-build.sh --no-deploy     # sólo construir y publicar
./scripts/prod-build.sh --only backend  # un solo componente
```

### 4.3 Después de desplegar

1. **Aplicar las migraciones** que toquen (sección 5).
2. **Comprobar que arrancó**, en este orden y sin saltarse ninguno:

```bash
docker compose -f prod.yml ps                                  # los tres arriba y la base sana
curl -s https://catalina-support.calibyou.com/api/health       # {"database":"ok","status":"ok"}
curl -sI https://catalina-support.calibyou.com/ | head -1      # 200 y la CSP de producción
```

3. **Mirar los logs** del backend, que es donde `go-logs` cuenta lo que ha pasado:

```bash
docker compose -f prod.yml logs --tail=50 backend
tail -f /root/prod/catalina-support/_logs/catalina-support_prod_$(date +%Y%m%d).log
```

### 4.4 Volver atrás

Si el despliegue sale mal, **el artefacto anterior sigue ahí**: se copia `<nombre>_prev` sobre
`<nombre>`, se reinicia el backend y se vuelve a la versión de antes. Está automatizado en la
sección 4.2 (la rotación) y a mano en el peor caso:

```bash
cp -a /root/prod/catalina-support/backend_prev/. /root/prod/catalina-support/backend/
docker compose -f prod.yml restart backend
```

**Lo que no se puede deshacer con esto es una migración**: una vez aplicada, volver atrás exige
escribir otra que lo revierta. Por eso la migración se aplica **después** de comprobar que el código
nuevo arranca, y no antes.

### 4.5 La primera vez (al cerrar la 1.0.0)

Es un despliegue normal, con tres cosas que sólo pasan una vez:

1. **La migración `v1.0.0.sql` se aplica una sola vez** y crea todo el esquema. En desarrollo se ha
   aplicado muchas veces; en producción, esa y ninguna más.
2. **Se comprueba que la cuenta `admin` entra** y que su contraseña es la que dice
   `config/env/prod.env` (la de fábrica se siembra al arrancar; `docs/usuarios-y-permisos.md`,
   sección 8).
3. **Se da de alta a Soporte y a Desarrollo** desde la aplicación, que es para lo que existe la
   cuenta de fábrica.

El vhost y el certificado **ya están puestos** desde el 2026-09-22: no hay nada que hacer ahí el día
del despliegue salvo comprobar que el certificado está renovado (sección 7).

## 5. Las migraciones

**Un archivo por versión**, y el nombre es la versión que se publica: **tres dígitos**,
`v1.0.0.sql`, `v1.1.0.sql`, `v1.2.0.sql`… El archivo y la etiqueta de git que marca esa versión
llevan el mismo número, así que mirando el tag se sabe qué migración le toca.

| Momento | Qué se aplica |
| --- | --- |
| **Hasta la 1.0.0** | `backend/migrations/v1.0.0.sql`, todas las veces que haga falta en desarrollo |
| **Al cerrar la 1.0.0** | El mismo archivo, **una sola vez** en producción |
| **Después de la 1.0.0** | Un archivo por versión, aplicados **en orden**, sin saltarse ninguno |

**Un archivo por versión y no uno por cambio**: el archivo cuenta la historia completa de esa
versión, se revisa de una vez y se aplica de una vez. Un archivo por cada cambio menudea el
despliegue y llena `migrations/` de retales.

**Sin tabla de control**: no hace falta, porque los archivos son idempotentes. Aplicar dos veces el
mismo no rompe nada, así que la forma de trabajar es sencilla: **se aplican todos los archivos hasta
la versión publicada**, en orden, y los que ya estaban aplicados no hacen nada.

Y para saber qué versión está publicada no hay que adivinar: lo dicen **la etiqueta de git y el
`BUILD_INFO`** del despliegue (sección 4.2). Aplicar de más es inofensivo; aplicar de menos se ve
enseguida, porque la aplicación pedirá una tabla que no existe.

Nunca `AutoMigrate`, nunca un `ALTER` a mano, nunca un paso que no esté en un archivo del
repositorio.

## 6. Las variables de entorno

Cada grupo lo documenta quien lo usa, y aquí sólo se dice dónde vive:

| Grupo | Quién lo documenta |
| --- | --- |
| `ENVIRONMENT`, `PROJECT_NAME`, `LOGS_FOLDER`, `TIMEZONE` (los de `go-logs`) | `docs/arquitectura.md`, sección 8 |
| `SMTP_*` | `docs/arquitectura.md`, sección 9 |
| `APP_PORT`, `POSTGRES_HOST`, `PGPORT`, `POSTGRES_USER`, `POSTGRES_PASSWORD`, `POSTGRES_DB` | `docs/arquitectura.md`, sección 10 |
| `FILES_PATH` | `docs/tickets.md`, sección 2.3 |
| `ADMIN_PASSWORD` | `docs/usuarios-y-permisos.md`, sección 8 |
| `TOKEN_SECRET` | `docs/usuarios-y-permisos.md`, sección 6 |

Reglas que valen para los dos entornos:

- **Sólo lo que cambia por entorno es variable.** Lo que vale lo mismo en los dos se escribe como
  constante en el código.
- **`config/env/dev.env` se versiona** porque no tiene secretos reales: son credenciales de
  contenedores locales.
- **`config/env/prod.env` no se versiona**, y ninguno de sus valores aparece en la documentación ni
  en el código. Hay una plantilla (`prod.env.example`) con la lista de lo que hace falta.
- **Si una variable deja de tener consumidor, se borra**: del código, de los env, de los compose y de
  la documentación, en el mismo cambio.

## 7. nginx y los certificados

- Los vhosts son **copias** de `config/nginx/` en `/etc/nginx/conf.d/`: si se cambia uno en el
  repositorio, hay que volver a copiarlo y recargar (los comandos están en `AGENTS.md`).
- Los certificados se **renuevan solos** (certbot con el plugin `webroot`, un certificado por
  dominio). La renovación no necesita intervención porque los dos vhosts sirven
  `/.well-known/acme-challenge/` en el puerto 80.
- **Comprobar que sigue vivo**, de vez en cuando y siempre antes de un despliegue:

```bash
sudo certbot certificates | grep -A2 catalina
echo | openssl s_client -connect catalina-support.calibyou.com:443 \
  -servername catalina-support.calibyou.com 2>/dev/null | openssl x509 -noout -dates
```

## 8. Las pruebas

Tres capas, de la más barata a la más lenta. Las tres se ejecutan **en los contenedores**.

### 8.1 Backend: `go test ./...`

Lo que se prueba, por orden de importancia:

1. **La máquina de estados de los tickets**: qué transiciones existen, quién puede cada una y, sobre
   todo, **las ocho reglas de sincronización** (`docs/propósito-y-alcance.md`). Es la lógica más
   delicada del sistema y la que peor se ve a simple vista: un ticket que se queda en `escalado` o
   un principal que no vuelve a Soporte no salta hasta que alguien lo mira.
2. **La numeración**: que dos tickets creados a la vez no saquen el mismo número, y que el número sea
   inmutable aunque cambie el prefijo.
3. **Los permisos**: que cada papel sólo pueda lo suyo, probado en los servicios y no sólo en la
   interfaz.
4. **La resolución del origen de una cuenta** al entrar: `local` contra la base, `ad` contra el
   directorio, y las reglas de convivencia (`docs/usuarios-y-permisos.md`, sección 5).

```bash
docker compose -f dev.yml exec backend go test ./...
docker compose -f dev.yml exec backend go vet ./...
```

### 8.2 Frontend: `npm test`

Vitest, que es lo que trae Angular 22. Para lo que tiene lógica: servicios, guardas, y los
componentes que calculan algo. **No se prueba el aspecto**: para eso están las pruebas de 8.3.

```bash
docker compose -f dev.yml exec frontend npm test -- --watch=false
```

### 8.3 Interfaz: Playwright contra desarrollo

Como en Calibyou: `tests/e2e/` con Playwright, **contra el entorno de desarrollo**, con un caso por
flujo crítico. Un recorrido completo vale más que veinte pruebas de detalle:

| Caso | Qué recorre |
| --- | --- |
| Entrar y salir | Sesión, cabecera `Authorization` y `GET /api/auth/me` |
| Crear un ticket | Alta, numeración y bandeja de Soporte |
| Escalar y resolver | El escalado, el interno, y que **el principal vuelve a Soporte** |
| Cerrar y reabrir | Cierre, reapertura y limpieza de fechas |
| Permisos | Que un usuario no ve el ticket de otro ni la lista de usuarios |

**Cuándo se ejecutan**: cuando el cambio toca la interfaz, antes de darlo por terminado. Si el cambio
afecta a un flujo que todavía no tiene prueba, **se añade el caso en el mismo trabajo**.

```bash
docker compose -f dev.yml up -d
cd tests/e2e && npm install && npm test          # el contenedor de desarrollo tiene que estar arriba
```

## 9. Lo que se decidió al repasar este documento

| # | Decisión | Quedó así |
| --- | --- | --- |
| 1 | **Las tres capas de pruebas** | Sí: unitarias de backend, unitarias de frontend y los cinco recorridos con **Playwright** contra desarrollo |
| 2 | **La base de las pruebas del backend** | Una base aparte, `catalina_support_test`, en el mismo contenedor: las pruebas la vacían y no tocan los datos de desarrollo |
| 3 | **Copias de seguridad** | `pg_dump` diario más copia de `_files`, con retención (14 días) y un script en `scripts/`, programado en el `cron` del servidor |
| 4 | **Integración continua** | **Ninguna** en la 1.0.0: las pruebas se ejecutan a mano en los contenedores, como parte del trabajo |
| 5 | **Las migraciones** | **Un archivo por versión**, con tres dígitos (`v1.0.0.sql`, `v1.1.0.sql`…), aplicados **en orden** y **sin tabla de control**: son idempotentes y se aplican todos hasta la versión publicada |
| 6 | **La limpieza de los logs** | **Manual**: `go-logs` no rota y no se le añade un script |

## 10. Qué NO entra en la 1.0.0

- **Integración continua y despliegue automático**: todo a mano, a propósito.
- **Un entorno de staging**: desarrollo y producción, y nada en medio.
- **Monitorización y alertas**: nadie avisa de que la web se ha caído. Enterarse es cosa de mirar.
- **Alta disponibilidad**: un servidor, sin réplicas ni balanceo.
- **Despliegue sin cortar el servicio**: reiniciar el backend corta lo que esté en curso, y se acepta.
- **Copias de seguridad fuera del servidor**: el script las deja en el servidor; sacarlas es aparte.
- **Rotación automática de logs** (ni por tamaño ni por antigüedad): la limpieza es manual, y la
  carpeta `_logs` crece un archivo por día. Está escrito para que no sorprenda.
- **Envío de logs a un servidor centralizado.**

## 11. Qué habilita este documento

Con `ambientes.md` aprobado, **la cadena de documentos está completa** y se puede empezar a
implementar, en el orden que fija `docs/tickets.md`: `auth`, `users` y `tickets`, con sus pruebas. El
`scripts/prod-build.sh` y el de copias se escriben entonces, no antes.
