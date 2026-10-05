-- Catalina-Support — esquema de la versión 1.0.0
--
-- Este archivo es la ÚNICA vía para crear el esquema, **en desarrollo y en producción**
-- (docs/ambientes.md, sección 5). Es transaccional (todo va entre BEGIN y COMMIT, de modo que un
-- fallo a mitad no deja nada a medias) e idempotente (se puede aplicar tantas veces como haga falta
-- sin romper nada), y se aplica con:
--
--   psql -v ON_ERROR_STOP=1 -U catalina_support -d catalina_support -p 11003 < backend/migrations/v1.0.0.sql
--
-- **Lleva también los datos sin los cuales la aplicación no funciona** —las dos filas de
-- configuración y las veintidós plantillas de correo—, porque eso no es un dato de ejemplo: en una
-- instalación recién desplegada, sin la fila de configuración no se puede leer ni el idioma ni el
-- nombre, y sin plantilla no sale ningún correo (docs/modules/settings.md, decisión 10).
--
-- **Los datos de ejemplo de desarrollo NO van aquí**: van en `v1.0.0_dev.sql`, que se aplica sólo en
-- desarrollo y nunca en producción.
--
-- Va por secciones, una por módulo, en el orden en que se implementan.

BEGIN;

-- ============================================================================
-- mail — las plantillas de los correos (docs/modules/mail.md)
-- ============================================================================

CREATE TABLE IF NOT EXISTS mail_templates (
    id              bigserial PRIMARY KEY,
    key             text        NOT NULL,
    language        text        NOT NULL,
    subject         text        NOT NULL,
    body            text        NOT NULL,
    -- El texto de fábrica, que no se edita nunca: es lo que permite «restaurar».
    default_subject text        NOT NULL,
    default_body    text        NOT NULL,
    updated_at      timestamptz NOT NULL DEFAULT now(),
    -- Nulo mientras el texto siga siendo el de fábrica. La referencia a `users` se añadirá cuando
    -- exista esa tabla (módulo users); hoy no puede haber una clave ajena a algo que no existe.
    updated_by_id   bigint,
    CONSTRAINT mail_templates_key_language_key UNIQUE (key, language),
    CONSTRAINT mail_templates_language_check CHECK (language IN ('es', 'en'))
);

COMMENT ON TABLE mail_templates IS
    'Los once correos del sistema, en dos idiomas: veintidós filas. Los escribe un administrador; el texto de fábrica se conserva para poder restaurarlo.';

-- Las veintidós plantillas de fábrica (docs/modules/mail.md, sección 5).
-- ON CONFLICT DO NOTHING: volver a aplicar la migración NO pisa lo que un administrador haya editado.
INSERT INTO mail_templates (key, language, subject, body, default_subject, default_body) VALUES
-- ---------------------------------------------------------------- avisos de ticket
('ticket.created', 'es',
 'Ticket nuevo: {{numero}}',
 '<p>Se ha creado el ticket <strong>{{numero}}</strong>.</p><p>{{asunto}}</p><p>Lo ha pedido {{solicitante}}.</p><p><a href="{{enlace}}">Abrir el ticket</a></p>',
 'Ticket nuevo: {{numero}}',
 '<p>Se ha creado el ticket <strong>{{numero}}</strong>.</p><p>{{asunto}}</p><p>Lo ha pedido {{solicitante}}.</p><p><a href="{{enlace}}">Abrir el ticket</a></p>'),

('ticket.created', 'en',
 'New ticket: {{numero}}',
 '<p>Ticket <strong>{{numero}}</strong> has been created.</p><p>{{asunto}}</p><p>Requested by {{solicitante}}.</p><p><a href="{{enlace}}">Open the ticket</a></p>',
 'New ticket: {{numero}}',
 '<p>Ticket <strong>{{numero}}</strong> has been created.</p><p>{{asunto}}</p><p>Requested by {{solicitante}}.</p><p><a href="{{enlace}}">Open the ticket</a></p>'),

('ticket.escalated', 'es',
 'Escalado: {{numero}}',
 '<p>El ticket <strong>{{numero}}</strong> se ha escalado a Desarrollo.</p><p>{{asunto}}</p><p><strong>Motivo:</strong> {{motivo}}</p><p><a href="{{enlace}}">Abrir el ticket interno</a></p>',
 'Escalado: {{numero}}',
 '<p>El ticket <strong>{{numero}}</strong> se ha escalado a Desarrollo.</p><p>{{asunto}}</p><p><strong>Motivo:</strong> {{motivo}}</p><p><a href="{{enlace}}">Abrir el ticket interno</a></p>'),

('ticket.escalated', 'en',
 'Escalated: {{numero}}',
 '<p>Ticket <strong>{{numero}}</strong> has been escalated to Development.</p><p>{{asunto}}</p><p><strong>Reason:</strong> {{motivo}}</p><p><a href="{{enlace}}">Open the internal ticket</a></p>',
 'Escalated: {{numero}}',
 '<p>Ticket <strong>{{numero}}</strong> has been escalated to Development.</p><p>{{asunto}}</p><p><strong>Reason:</strong> {{motivo}}</p><p><a href="{{enlace}}">Open the internal ticket</a></p>'),

('ticket.waitingUser', 'es',
 'Necesitamos algo de ti: {{numero}}',
 '<p>Hola {{nombre}}:</p><p>Para seguir con tu ticket <strong>{{numero}}</strong> ({{asunto}}) necesitamos que nos contestes.</p><p><a href="{{enlace}}">Responder en el ticket</a></p>',
 'Necesitamos algo de ti: {{numero}}',
 '<p>Hola {{nombre}}:</p><p>Para seguir con tu ticket <strong>{{numero}}</strong> ({{asunto}}) necesitamos que nos contestes.</p><p><a href="{{enlace}}">Responder en el ticket</a></p>'),

('ticket.waitingUser', 'en',
 'We need something from you: {{numero}}',
 '<p>Hi {{nombre}},</p><p>To continue with ticket <strong>{{numero}}</strong> ({{asunto}}) we need you to get back to us.</p><p><a href="{{enlace}}">Reply in the ticket</a></p>',
 'We need something from you: {{numero}}',
 '<p>Hi {{nombre}},</p><p>To continue with ticket <strong>{{numero}}</strong> ({{asunto}}) we need you to get back to us.</p><p><a href="{{enlace}}">Reply in the ticket</a></p>'),

('ticket.waitingSupport', 'es',
 'Desarrollo necesita algo: {{numero}}',
 '<p>Desarrollo necesita algo de Soporte para seguir con el ticket <strong>{{numero}}</strong>.</p><p>{{asunto}}</p><p><a href="{{enlace}}">Abrir el ticket interno</a></p>',
 'Desarrollo necesita algo: {{numero}}',
 '<p>Desarrollo necesita algo de Soporte para seguir con el ticket <strong>{{numero}}</strong>.</p><p>{{asunto}}</p><p><a href="{{enlace}}">Abrir el ticket interno</a></p>'),

('ticket.waitingSupport', 'en',
 'Development needs something: {{numero}}',
 '<p>Development needs something from Support to continue with ticket <strong>{{numero}}</strong>.</p><p>{{asunto}}</p><p><a href="{{enlace}}">Open the internal ticket</a></p>',
 'Development needs something: {{numero}}',
 '<p>Development needs something from Support to continue with ticket <strong>{{numero}}</strong>.</p><p>{{asunto}}</p><p><a href="{{enlace}}">Open the internal ticket</a></p>'),

('ticket.resolved', 'es',
 'Tu ticket {{numero}} está resuelto',
 '<p>Hola {{nombre}}:</p><p>El ticket <strong>{{numero}}</strong> ({{asunto}}) está resuelto.</p><p>Si no es así, respóndenos en el ticket o ábrelo de nuevo.</p><p><a href="{{enlace}}">Ver el ticket</a></p>',
 'Tu ticket {{numero}} está resuelto',
 '<p>Hola {{nombre}}:</p><p>El ticket <strong>{{numero}}</strong> ({{asunto}}) está resuelto.</p><p>Si no es así, respóndenos en el ticket o ábrelo de nuevo.</p><p><a href="{{enlace}}">Ver el ticket</a></p>'),

('ticket.resolved', 'en',
 'Your ticket {{numero}} is resolved',
 '<p>Hi {{nombre}},</p><p>Ticket <strong>{{numero}}</strong> ({{asunto}}) is resolved.</p><p>If that is not the case, reply in the ticket or reopen it.</p><p><a href="{{enlace}}">View the ticket</a></p>',
 'Your ticket {{numero}} is resolved',
 '<p>Hi {{nombre}},</p><p>Ticket <strong>{{numero}}</strong> ({{asunto}}) is resolved.</p><p>If that is not the case, reply in the ticket or reopen it.</p><p><a href="{{enlace}}">View the ticket</a></p>'),

('ticket.closed', 'es',
 'Tu ticket {{numero}} se ha cerrado',
 '<p>Hola {{nombre}}:</p><p>El ticket <strong>{{numero}}</strong> está cerrado.</p><p>Si vuelve a pasar, puedes reabrirlo.</p><p><a href="{{enlace}}">Ver el ticket</a></p>',
 'Tu ticket {{numero}} se ha cerrado',
 '<p>Hola {{nombre}}:</p><p>El ticket <strong>{{numero}}</strong> está cerrado.</p><p>Si vuelve a pasar, puedes reabrirlo.</p><p><a href="{{enlace}}">Ver el ticket</a></p>'),

('ticket.closed', 'en',
 'Your ticket {{numero}} has been closed',
 '<p>Hi {{nombre}},</p><p>Ticket <strong>{{numero}}</strong> is closed.</p><p>If it happens again, you can reopen it.</p><p><a href="{{enlace}}">View the ticket</a></p>',
 'Your ticket {{numero}} has been closed',
 '<p>Hi {{nombre}},</p><p>Ticket <strong>{{numero}}</strong> is closed.</p><p>If it happens again, you can reopen it.</p><p><a href="{{enlace}}">View the ticket</a></p>'),

('ticket.backToSupport', 'es',
 '{{numero}} ha vuelto a tu bandeja',
 '<p>El ticket <strong>{{numero}}</strong> ha vuelto a Soporte.</p><p>{{asunto}}</p><p><a href="{{enlace}}">Abrir el ticket</a></p>',
 '{{numero}} ha vuelto a tu bandeja',
 '<p>El ticket <strong>{{numero}}</strong> ha vuelto a Soporte.</p><p>{{asunto}}</p><p><a href="{{enlace}}">Abrir el ticket</a></p>'),

('ticket.backToSupport', 'en',
 '{{numero}} is back in your queue',
 '<p>Ticket <strong>{{numero}}</strong> is back with Support.</p><p>{{asunto}}</p><p><a href="{{enlace}}">Open the ticket</a></p>',
 '{{numero}} is back in your queue',
 '<p>Ticket <strong>{{numero}}</strong> is back with Support.</p><p>{{asunto}}</p><p><a href="{{enlace}}">Open the ticket</a></p>'),

-- La plantilla número once: **a quien acaban de etiquetar** en un cuerpo del ticket
-- (docs/modules/tickets.md, decisión 58). Lleva el número, el asunto y el enlace, nunca el texto.
('ticket.mentioned', 'es',
 'Te han etiquetado en {{numero}}: {{asunto}}',
 '<p>Te han etiquetado en el ticket <strong>{{numero}}</strong>.</p><p>{{asunto}}</p><p>Lo sigues desde tu bandeja, en «Observo».</p><p><a href="{{enlace}}">Abrir el ticket</a></p>',
 'Te han etiquetado en {{numero}}: {{asunto}}',
 '<p>Te han etiquetado en el ticket <strong>{{numero}}</strong>.</p><p>{{asunto}}</p><p>Lo sigues desde tu bandeja, en «Observo».</p><p><a href="{{enlace}}">Abrir el ticket</a></p>'),

('ticket.mentioned', 'en',
 'You were mentioned in {{numero}}: {{asunto}}',
 '<p>You were mentioned in ticket <strong>{{numero}}</strong>.</p><p>{{asunto}}</p><p>You are following it from your inbox, under “Watching”.</p><p><a href="{{enlace}}">Open the ticket</a></p>',
 'You were mentioned in {{numero}}: {{asunto}}',
 '<p>You were mentioned in ticket <strong>{{numero}}</strong>.</p><p>{{asunto}}</p><p>You are following it from your inbox, under “Watching”.</p><p><a href="{{enlace}}">Open the ticket</a></p>'),

-- ---------------------------------------------------------------- correos de cuenta
('auth.invitation', 'es',
 'Establece tu contraseña de Catalina Support',
 '<p>Hola {{nombre}}:</p><p>Ya tienes cuenta. Establece tu contraseña desde este enlace, que caduca en {{caducidad}}.</p><p><a href="{{enlace}}">Establecer mi contraseña</a></p>',
 'Establece tu contraseña de Catalina Support',
 '<p>Hola {{nombre}}:</p><p>Ya tienes cuenta. Establece tu contraseña desde este enlace, que caduca en {{caducidad}}.</p><p><a href="{{enlace}}">Establecer mi contraseña</a></p>'),

('auth.invitation', 'en',
 'Set your Catalina Support password',
 '<p>Hi {{nombre}},</p><p>Your account is ready. Set your password from this link; it expires in {{caducidad}}.</p><p><a href="{{enlace}}">Set my password</a></p>',
 'Set your Catalina Support password',
 '<p>Hi {{nombre}},</p><p>Your account is ready. Set your password from this link; it expires in {{caducidad}}.</p><p><a href="{{enlace}}">Set my password</a></p>'),

('auth.recovery', 'es',
 'Restablecer tu contraseña',
 '<p>Hola {{nombre}}:</p><p>Alguien ha pedido restablecer tu contraseña. Si has sido tú, hazlo desde este enlace; caduca en {{caducidad}}.</p><p>Si no has sido tú, no hagas nada: sin abrir el enlace, tu contraseña no cambia.</p><p><a href="{{enlace}}">Restablecer mi contraseña</a></p>',
 'Restablecer tu contraseña',
 '<p>Hola {{nombre}}:</p><p>Alguien ha pedido restablecer tu contraseña. Si has sido tú, hazlo desde este enlace; caduca en {{caducidad}}.</p><p>Si no has sido tú, no hagas nada: sin abrir el enlace, tu contraseña no cambia.</p><p><a href="{{enlace}}">Restablecer mi contraseña</a></p>'),

('auth.recovery', 'en',
 'Reset your password',
 '<p>Hi {{nombre}},</p><p>Someone asked to reset your password. If it was you, use this link; it expires in {{caducidad}}.</p><p>If it was not you, do nothing: your password does not change unless the link is opened.</p><p><a href="{{enlace}}">Reset my password</a></p>',
 'Reset your password',
 '<p>Hi {{nombre}},</p><p>Someone asked to reset your password. If it was you, use this link; it expires in {{caducidad}}.</p><p>If it was not you, do nothing: your password does not change unless the link is opened.</p><p><a href="{{enlace}}">Reset my password</a></p>'),

('auth.passwordChanged', 'es',
 'Tu contraseña ha cambiado',
 '<p>Hola {{nombre}}:</p><p>Tu contraseña ha cambiado el {{cuando}}, desde la dirección {{ip}}.</p><p>Si no has sido tú, avisa a Soporte cuanto antes.</p>',
 'Tu contraseña ha cambiado',
 '<p>Hola {{nombre}}:</p><p>Tu contraseña ha cambiado el {{cuando}}, desde la dirección {{ip}}.</p><p>Si no has sido tú, avisa a Soporte cuanto antes.</p>'),

('auth.passwordChanged', 'en',
 'Your password has changed',
 '<p>Hi {{nombre}},</p><p>Your password changed on {{cuando}}, from the address {{ip}}.</p><p>If it was not you, tell Support as soon as you can.</p>',
 'Your password has changed',
 '<p>Hi {{nombre}},</p><p>Your password changed on {{cuando}}, from the address {{ip}}.</p><p>If it was not you, tell Support as soon as you can.</p>')

ON CONFLICT (key, language) DO NOTHING;

-- La **plantilla número once** (`ticket.mentioned`), aparte además de dentro del bloque de arriba: el
-- bloque grande ya la trae, así que en una base nueva esto es un no-op; en una que ya tenía las veinte
-- filas, el bloque grande también la inserta, y esta sentencia deja dicho, en su sitio y sin depender
-- de leer un `INSERT` de doscientas líneas, que este correo es un añadido de la 1.0.0 y no se puede
-- pisar (docs/modules/tickets.md, sección 5, y docs/modules/mail.md, sección 5.1).
INSERT INTO mail_templates (key, language, subject, body, default_subject, default_body) VALUES
('ticket.mentioned', 'es',
 'Te han etiquetado en {{numero}}: {{asunto}}',
 '<p>Te han etiquetado en el ticket <strong>{{numero}}</strong>.</p><p>{{asunto}}</p><p>Lo sigues desde tu bandeja, en «Observo».</p><p><a href="{{enlace}}">Abrir el ticket</a></p>',
 'Te han etiquetado en {{numero}}: {{asunto}}',
 '<p>Te han etiquetado en el ticket <strong>{{numero}}</strong>.</p><p>{{asunto}}</p><p>Lo sigues desde tu bandeja, en «Observo».</p><p><a href="{{enlace}}">Abrir el ticket</a></p>'),

('ticket.mentioned', 'en',
 'You were mentioned in {{numero}}: {{asunto}}',
 '<p>You were mentioned in ticket <strong>{{numero}}</strong>.</p><p>{{asunto}}</p><p>You are following it from your inbox, under “Watching”.</p><p><a href="{{enlace}}">Open the ticket</a></p>',
 'You were mentioned in {{numero}}: {{asunto}}',
 '<p>You were mentioned in ticket <strong>{{numero}}</strong>.</p><p>{{asunto}}</p><p>You are following it from your inbox, under “Watching”.</p><p><a href="{{enlace}}">Open the ticket</a></p>')
ON CONFLICT (key, language) DO NOTHING;

-- ============================================================================
-- users — las cuentas (docs/modules/users.md)
-- ============================================================================

CREATE TABLE IF NOT EXISTS users (
    id            bigserial PRIMARY KEY,
    name          text        NOT NULL,
    last_name     text        NOT NULL,
    email         text        NOT NULL,
    -- Nulo mientras la cuenta no tenga contraseña: en las de directorio no hay ninguna, y en las
    -- locales está así hasta que la persona la establece desde el enlace del correo de alta.
    password_hash text,
    role          text        NOT NULL,
    origin        text        NOT NULL,
    -- El identificador del directorio (el `sub` de Keycloak o el GUID de AD). Nulo en las locales.
    external_id   text,
    -- Es lo que decide en qué idioma se le escriben los correos.
    language      text        NOT NULL DEFAULT 'es',
    is_active     boolean     NOT NULL DEFAULT true,
    last_login_at timestamptz,
    created_at    timestamptz NOT NULL DEFAULT now(),
    updated_at    timestamptz NOT NULL DEFAULT now(),
    CONSTRAINT users_role_check CHECK (role IN ('usuario', 'soporte', 'desarrollo', 'administrador')),
    CONSTRAINT users_origin_check CHECK (origin IN ('local', 'ad', 'keycloak')),
    CONSTRAINT users_language_check CHECK (language IN ('es', 'en')),
    -- Si el origen es de directorio, no hay contraseña local: la regla en la base, no sólo en el
    -- código (docs/usuarios-y-permisos.md, sección 5).
    CONSTRAINT users_directory_has_no_password CHECK (origin = 'local' OR password_hash IS NULL)
);

COMMENT ON TABLE users IS
    'Las cuentas reales. La de fábrica (`admin`) NO está aquí: vive en la configuración.';

-- Un correo, una cuenta, sin distinguir mayúsculas: Ana@… y ana@… son la misma persona.
CREATE UNIQUE INDEX IF NOT EXISTS users_email_unique ON users (lower(email));

-- Dos cuentas no pueden apuntar al mismo usuario del directorio.
CREATE UNIQUE INDEX IF NOT EXISTS users_origin_external_id_unique
    ON users (origin, external_id) WHERE external_id IS NOT NULL;

-- La clave ajena que faltaba en mail_templates, ahora que existe la tabla de cuentas.
DO $$
BEGIN
    IF NOT EXISTS (SELECT 1 FROM pg_constraint WHERE conname = 'mail_templates_updated_by_id_fkey') THEN
        ALTER TABLE mail_templates
            ADD CONSTRAINT mail_templates_updated_by_id_fkey
            FOREIGN KEY (updated_by_id) REFERENCES users (id) ON DELETE SET NULL;
    END IF;
END $$;

-- ============================================================================
-- auth — los tokens de los enlaces de contraseña (docs/modules/auth.md, sección 2)
-- ============================================================================

CREATE TABLE IF NOT EXISTS password_tokens (
    id            bigserial PRIMARY KEY,
    user_id       bigint      NOT NULL REFERENCES users (id) ON DELETE CASCADE,
    purpose       text        NOT NULL,
    -- La huella del token, nunca el token: quien lea la base no puede reconstruir ningún enlace.
    token_hash    text        NOT NULL,
    expires_at    timestamptz NOT NULL,
    used_at       timestamptz,
    -- Nulo si lo pidió la propia persona desde «he olvidado mi contraseña»; con valor, si lo lanzó
    -- Soporte o un administrador.
    created_by_id bigint      REFERENCES users (id) ON DELETE SET NULL,
    created_at    timestamptz NOT NULL DEFAULT now(),
    CONSTRAINT password_tokens_purpose_check CHECK (purpose IN ('alta', 'recuperacion'))
);

COMMENT ON TABLE password_tokens IS
    'Los enlaces de alta y de recuperación: un uso cada uno, y pedir uno nuevo borra los anteriores.';

CREATE UNIQUE INDEX IF NOT EXISTS password_tokens_hash_unique ON password_tokens (token_hash);
CREATE INDEX IF NOT EXISTS password_tokens_user_idx ON password_tokens (user_id);

-- ============================================================================
-- settings — la configuración de la instalación (docs/modules/settings.md)
--
-- Dos tablas de UNA sola fila, y no una de «ajustes» con clave y valor: así se ve de un vistazo qué
-- es configurable, y añadir un ajuste obliga a tocar el esquema, que es la conversación que conviene
-- tener (docs/modules/tickets.md, sección 2.2).
-- ============================================================================

CREATE TABLE IF NOT EXISTS installation_settings (
    -- Una sola fila: la restricción lo impone, no la buena voluntad de nadie.
    id            integer     PRIMARY KEY DEFAULT 1 CHECK (id = 1),
    -- El nombre de la institución, la empresa o el equipo: **lo que sustituye a «Catalina Support»**
    -- en la pantalla de entrada, en el menú lateral y en la pestaña del navegador
    -- (docs/modules/settings.md, sección 5.7).
    installation_name text    NOT NULL DEFAULT 'Catalina Support',
    -- **El método de entrada de la instalación: uno a la vez.** `local` es el de fábrica, y con él
    -- entran las cuentas que viven aquí; `ad` y `keycloak` delegan en el directorio de la empresa.
    -- Se puede cambiar cuando se quiera, y siempre queda la cuenta de fábrica como puerta
    -- (docs/modules/settings.md, sección 5.8).
    entry_method  text        NOT NULL DEFAULT 'local',
    -- **La zona horaria de la instalación** (nombre IANA, `America/Guayaquil`): decide cómo se leen
    -- las fechas, en la interfaz y en los correos. Las guardadas siguen en UTC, así que cambiarla no
    -- mueve ningún ticket (docs/modules/settings.md, decisión 15).
    time_zone     text        NOT NULL DEFAULT 'UTC',
    -- **La dirección pública**: esquema, host y puerto. Es la base de los enlaces que salen en los
    -- correos y de la vuelta de Keycloak. Vacía es «no configurada», y entonces se usa la variable de
    -- entorno (docs/modules/settings.md, decisión 14).
    public_app_url text       NOT NULL DEFAULT '',
    -- **El motor de IA** (docs/modules/ai.md): su dirección y su modelo. Vivían en el entorno y pasan
    -- aquí, como el directorio, Keycloak y el correo, para poder integrarlo después desde Configuración.
    ai_url        text        NOT NULL DEFAULT '',
    ai_model      text        NOT NULL DEFAULT '',
    -- **El sello de instalación**: nulo mientras nadie haya terminado el asistente de primer arranque.
    -- Es lo que hace que la vista de instalación se enseñe **sólo** en una instalación sin configurar y
    -- que su API rechace configurarla dos veces (docs/primer-arranque.md, sección 2).
    installed_at  timestamptz,
    -- **El correo saliente** (docs/primer-arranque.md, sección 5): vive aquí, como el directorio y
    -- Keycloak. Las sugerencias del asistente no configuran el envío. La contraseña **no sale nunca
    -- por la API**.
    smtp_host      text NOT NULL DEFAULT '',
    smtp_port      text NOT NULL DEFAULT '',
    smtp_secure    boolean NOT NULL DEFAULT false,
    smtp_user      text NOT NULL DEFAULT '',
    smtp_password  text NOT NULL DEFAULT '',
    smtp_from_name  text NOT NULL DEFAULT '',
    smtp_from_email text NOT NULL DEFAULT '',
    language      text        NOT NULL DEFAULT 'en',
    primary_color text        NOT NULL DEFAULT '#1d4ed8',
    -- El NOMBRE del archivo del logo, no el archivo: vive en el disco, en _files/brand.
    -- Nulo es «no hay logo propio» y se usa el de fábrica que trae la aplicación.
    logo_light    text,
    logo_dark     text,
    updated_at    timestamptz NOT NULL DEFAULT now(),
    updated_by_id bigint      REFERENCES users (id) ON DELETE SET NULL,
    CONSTRAINT installation_settings_language_check CHECK (language IN ('es', 'en')),
    -- En hexadecimal de seis dígitos, en minúsculas o mayúsculas.
    CONSTRAINT installation_settings_color_check CHECK (primary_color ~ '^#[0-9a-fA-F]{6}$'),
    -- El nombre no puede quedarse en blanco —una instalación sin nombre no se puede ni nombrar— y
    -- tiene un tope, porque se enseña en un menú estrecho y en la tarjeta de la entrada.
    CONSTRAINT installation_settings_name_check CHECK (
        btrim(installation_name) <> '' AND char_length(installation_name) <= 60
    ),
    CONSTRAINT installation_settings_method_check CHECK (entry_method IN ('local', 'ad', 'keycloak')),
    -- La dirección, si se escribe, tiene que ser http o https con su host: lo demás sería un enlace
    -- roto dentro de un correo.
    CONSTRAINT installation_public_url_is_http CHECK (public_app_url = '' OR public_app_url ~ '^https?://[^[:space:]]+$')
);

COMMENT ON TABLE installation_settings IS
    'La configuración de la instalación: el nombre, el idioma, el color institucional y la marca (los dos huecos del logo).';

-- Una instalación nueva empieza en inglés. Cambiar el valor por defecto no toca la fila de una
-- instalación existente, así que su idioma se conserva al volver a aplicar la migración.
ALTER TABLE installation_settings ALTER COLUMN language SET DEFAULT 'en';

-- El directorio de la organización (AD por LDAP). **Lo configura un Administrador desde la
-- pantalla**, no el entorno (decisión del responsable, 2026-09-25): con el servidor vacío, ese camino
-- no existe. La contraseña de la cuenta de servicio se guarda aquí y **no se devuelve nunca** por la
-- API.
CREATE TABLE IF NOT EXISTS directory_settings (
    id             integer     PRIMARY KEY DEFAULT 1 CHECK (id = 1),
    -- Vacío es «esta instalación no tiene directorio».
    host           text        NOT NULL DEFAULT '',
    port           text        NOT NULL DEFAULT '389',
    -- Por ahí viajan credenciales: en producción, sí.
    use_tls        boolean     NOT NULL DEFAULT false,
    bind_dn        text        NOT NULL DEFAULT '',
    bind_password  text        NOT NULL DEFAULT '',
    search_base    text        NOT NULL DEFAULT '',
    -- El filtro, con `%s` donde va el correo de quien entra.
    user_filter    text        NOT NULL DEFAULT '(mail=%s)',
    attr_email     text        NOT NULL DEFAULT 'mail',
    attr_name      text        NOT NULL DEFAULT 'givenName',
    attr_last_name text        NOT NULL DEFAULT 'sn',
    attr_id        text        NOT NULL DEFAULT 'objectGUID',
    updated_at     timestamptz NOT NULL DEFAULT now(),
    updated_by_id  bigint      REFERENCES users (id) ON DELETE SET NULL
);

COMMENT ON TABLE directory_settings IS
    'La configuración del directorio de la organización (AD por LDAP): una sola fila, editable desde la pantalla.';

-- Keycloak (OIDC). Igual: el emisor vacío es «no hay este camino», y el secreto del cliente no sale
-- nunca por la API.
CREATE TABLE IF NOT EXISTS keycloak_settings (
    id            integer     PRIMARY KEY DEFAULT 1 CHECK (id = 1),
    issuer        text        NOT NULL DEFAULT '',
    internal_issuer text NOT NULL DEFAULT '',
    client_id     text        NOT NULL DEFAULT '',
    client_secret text        NOT NULL DEFAULT '',
    redirect_uri  text        NOT NULL DEFAULT '',
    updated_at    timestamptz NOT NULL DEFAULT now(),
    updated_by_id bigint      REFERENCES users (id) ON DELETE SET NULL
);

COMMENT ON TABLE keycloak_settings IS
    'La configuración del camino de Keycloak: una sola fila, editable desde la pantalla.';

CREATE TABLE IF NOT EXISTS ticket_settings (
    id                    integer     PRIMARY KEY DEFAULT 1 CHECK (id = 1),
    number_prefix         text        NOT NULL DEFAULT 'CS',
    main_assignment       text        NOT NULL DEFAULT 'por_turnos',
    main_notification     text        NOT NULL DEFAULT 'al_asignado',
    internal_assignment   text        NOT NULL DEFAULT 'ninguna',
    internal_notification text        NOT NULL DEFAULT 'a_nadie',
    updated_at            timestamptz NOT NULL DEFAULT now(),
    updated_by_id         bigint      REFERENCES users (id) ON DELETE SET NULL,
    CONSTRAINT ticket_settings_prefix_check CHECK (number_prefix ~ '^[A-Z0-9]{2,8}$'),
    CONSTRAINT ticket_settings_assignment_check CHECK (
        main_assignment IN ('ninguna', 'por_turnos') AND internal_assignment IN ('ninguna', 'por_turnos')
    ),
    CONSTRAINT ticket_settings_notification_check CHECK (
        main_notification IN ('a_nadie', 'a_todo_el_equipo', 'al_asignado')
        AND internal_notification IN ('a_nadie', 'a_todo_el_equipo', 'al_asignado')
    ),
    -- Avisar «al asignado» sin repartir no tiene sentido: no hay a quién avisar. La pantalla no
    -- ofrece esa combinación y la base tampoco la acepta.
    CONSTRAINT ticket_settings_notification_needs_assignment CHECK (
        (main_assignment = 'por_turnos' OR main_notification <> 'al_asignado')
        AND (internal_assignment = 'por_turnos' OR internal_notification <> 'al_asignado')
    )
);

COMMENT ON TABLE ticket_settings IS
    'El prefijo de la numeración y el reparto y el aviso de cada tipo de ticket.';

-- ============================================================================
-- settings — puesta al día de una base que ya existía
-- ============================================================================
--
-- **Un archivo idempotente tiene que servir para las dos cosas**: crear la base desde cero y poner al
-- día una que ya estaba. `CREATE TABLE IF NOT EXISTS` no toca una tabla que existe, así que las
-- columnas y las restricciones que se han añadido **después** de crear la tabla hay que decirlas
-- aparte, con `IF NOT EXISTS`. En una base nueva son un no-op; en una que ya existía, son el cambio.

ALTER TABLE keycloak_settings ADD COLUMN IF NOT EXISTS internal_issuer text NOT NULL DEFAULT '';

ALTER TABLE installation_settings
    ADD COLUMN IF NOT EXISTS installation_name text NOT NULL DEFAULT 'Catalina Support';

ALTER TABLE installation_settings
    ADD COLUMN IF NOT EXISTS entry_method text NOT NULL DEFAULT 'local';

-- **La región horaria y la dirección pública** (docs/modules/settings.md, decisiones 14 y 15). Van aquí
-- además de en la definición de la tabla por lo mismo que las de arriba: una base que ya estaba creada
-- no se toca con `CREATE TABLE IF NOT EXISTS`, y esto es lo que la pone al día.
ALTER TABLE installation_settings
    ADD COLUMN IF NOT EXISTS time_zone text NOT NULL DEFAULT 'UTC';

ALTER TABLE installation_settings
    ADD COLUMN IF NOT EXISTS public_app_url text NOT NULL DEFAULT '';

-- **El sello de instalación, con su relleno, y en un solo bloque**: hay que saber si la columna
-- **acaba de nacer**, porque sólo entonces las filas que ya existían estaban configuradas y se sellan.
-- Así queda bien en los tres casos (docs/primer-arranque.md, sección 2):
--   · base **recién creada**: la columna nace y **la fila todavía no existe** —se inserta al final del
--     archivo—, así que no se sella nada y la instalación queda **sin instalar**;
--   · base **que ya existía** (producción): la columna nace y **su fila se sella**, así que el asistente
--     no aparece por actualizar;
--   · archivo aplicado **otra vez**: la columna ya está y **no se toca ningún sello**.
DO $$
DECLARE
    columna_nueva boolean := NOT EXISTS (
        SELECT 1 FROM information_schema.columns
         WHERE table_name = 'installation_settings' AND column_name = 'installed_at');
BEGIN
    IF columna_nueva THEN
        ALTER TABLE installation_settings ADD COLUMN installed_at timestamptz;
        UPDATE installation_settings SET installed_at = COALESCE(updated_at, now()) WHERE id = 1;
    END IF;
END
$$;

ALTER TABLE installation_settings
    ADD COLUMN IF NOT EXISTS ai_url text NOT NULL DEFAULT '',
    ADD COLUMN IF NOT EXISTS ai_model text NOT NULL DEFAULT '';

ALTER TABLE installation_settings
    ADD COLUMN IF NOT EXISTS smtp_host text NOT NULL DEFAULT '',
    ADD COLUMN IF NOT EXISTS smtp_port text NOT NULL DEFAULT '',
    ADD COLUMN IF NOT EXISTS smtp_secure boolean NOT NULL DEFAULT false,
    ADD COLUMN IF NOT EXISTS smtp_user text NOT NULL DEFAULT '',
    ADD COLUMN IF NOT EXISTS smtp_password text NOT NULL DEFAULT '',
    ADD COLUMN IF NOT EXISTS smtp_from_name text NOT NULL DEFAULT '',
    ADD COLUMN IF NOT EXISTS smtp_from_email text NOT NULL DEFAULT '';

-- Las restricciones se quitan y se vuelven a poner: no hay `ADD CONSTRAINT IF NOT EXISTS`, y quitarla
-- antes es lo que hace que se pueda aplicar el archivo dos veces.
ALTER TABLE installation_settings DROP CONSTRAINT IF EXISTS installation_settings_name_check;
ALTER TABLE installation_settings
    ADD CONSTRAINT installation_settings_name_check CHECK (
        btrim(installation_name) <> '' AND char_length(installation_name) <= 60
    );

ALTER TABLE installation_settings DROP CONSTRAINT IF EXISTS installation_public_url_is_http;
ALTER TABLE installation_settings
    ADD CONSTRAINT installation_public_url_is_http
    CHECK (public_app_url = '' OR public_app_url ~ '^https?://[^[:space:]]+$');

ALTER TABLE installation_settings DROP CONSTRAINT IF EXISTS installation_settings_method_check;
ALTER TABLE installation_settings
    ADD CONSTRAINT installation_settings_method_check CHECK (entry_method IN ('local', 'ad', 'keycloak'));

COMMENT ON COLUMN installation_settings.installation_name IS
    'El nombre de la institución, la empresa o el equipo. Es lo que se enseña donde antes decía «Catalina Support».';

COMMENT ON COLUMN installation_settings.entry_method IS
    'El método de entrada de la instalación: local, ad o keycloak. Uno a la vez, y se puede cambiar.';

COMMENT ON COLUMN installation_settings.time_zone IS
    'La zona horaria de la instalación (nombre IANA). Decide cómo se leen las fechas; las guardadas siguen en UTC.';

COMMENT ON COLUMN installation_settings.ai_url IS
    'La dirección del motor de IA (http://ai:8080). Vacío es «no integrado»: la mesa de ayuda funciona sin él.';

COMMENT ON COLUMN installation_settings.ai_model IS
    'El modelo que sirve el motor, para poder decirlo en la pantalla.';

COMMENT ON COLUMN installation_settings.installed_at IS
    'Cuándo se terminó el asistente de primer arranque. Nulo es «sin instalar»: sólo entonces se enseña la vista de instalación.';

COMMENT ON COLUMN installation_settings.smtp_host IS
    'El servidor de correo saliente. Vacío usa SMTP_HOST del entorno.';

COMMENT ON COLUMN installation_settings.public_app_url IS
    'La dirección pública: base de los enlaces de los correos y de la vuelta de Keycloak. Vacío usa PUBLIC_APP_URL.';

-- La fila de cada tabla, con los valores de fábrica: la aplicación tiene que poder arrancar y crear
-- el primer ticket sin pasar por la pantalla de configuración. Las dos de los caminos de directorio
-- nacen **vacías**, que es «esta instalación no los tiene».
INSERT INTO installation_settings (id) VALUES (1) ON CONFLICT (id) DO NOTHING;
INSERT INTO ticket_settings (id) VALUES (1) ON CONFLICT (id) DO NOTHING;
INSERT INTO directory_settings (id) VALUES (1) ON CONFLICT (id) DO NOTHING;
INSERT INTO keycloak_settings (id) VALUES (1) ON CONFLICT (id) DO NOTHING;

-- ============================================================================
-- tickets — los principales, los internos y lo que cuelga de ellos
-- (docs/modules/tickets.md, secciones 2 y 6)
--
-- Dos tablas de tickets y no una, y a cambio las tres tablas que cuelgan llevan DOS columnas de
-- destino —una por tipo— con la restricción de que exactamente una esté rellena. El precio se paga
-- aquí y sólo aquí; a cambio, cada tabla tiene sólo sus columnas.
-- ============================================================================

CREATE TABLE IF NOT EXISTS tickets (
    id                    bigserial   PRIMARY KEY,
    -- El número va entero y NO se recompone desde el prefijo actual: si mañana cambia el prefijo,
    -- los tickets viejos conservan el suyo.
    number                text        NOT NULL UNIQUE,
    number_year           integer     NOT NULL,
    number_seq            integer     NOT NULL,
    subject               text        NOT NULL,
    description           text        NOT NULL,
    state                 text        NOT NULL DEFAULT 'nuevo',
    requester_id          bigint      NOT NULL REFERENCES users (id) ON DELETE RESTRICT,
    created_by_id         bigint      NOT NULL REFERENCES users (id) ON DELETE RESTRICT,
    -- Nulo sólo si no había ningún técnico activo al repartir.
    assignee_id           bigint      REFERENCES users (id) ON DELETE SET NULL,
    -- Nulo mientras nadie haya editado ese campo: una marca por campo, para saber qué se cambió.
    subject_edited_at     timestamptz,
    description_edited_at timestamptz,
    resolved_at           timestamptz,
    closed_at             timestamptz,
    created_at            timestamptz NOT NULL DEFAULT now(),
    updated_at            timestamptz NOT NULL DEFAULT now(),
    CONSTRAINT tickets_state_check CHECK (
        state IN ('nuevo', 'en progreso', 'en espera', 'escalado', 'resuelto', 'cerrado')
    ),
    CONSTRAINT tickets_subject_check CHECK (btrim(subject) <> ''),
    CONSTRAINT tickets_description_check CHECK (btrim(description) <> ''),
    CONSTRAINT tickets_number_seq_check CHECK (number_seq > 0),
    UNIQUE (number_year, number_seq)
);

COMMENT ON TABLE tickets IS
    'Los tickets principales. La API se dirige por `number`, nunca por `id`.';

CREATE TABLE IF NOT EXISTS internal_tickets (
    id                bigserial   PRIMARY KEY,
    -- Único: es lo que garantiza un solo interno por principal, y lo garantiza la base.
    ticket_id         bigint      NOT NULL UNIQUE REFERENCES tickets (id) ON DELETE CASCADE,
    number            text        NOT NULL UNIQUE,
    state             text        NOT NULL DEFAULT 'nuevo',
    -- El motivo es obligatorio y no se edita: reescribirlo es reescribir la historia.
    escalation_reason text        NOT NULL,
    created_by_id     bigint      NOT NULL REFERENCES users (id) ON DELETE RESTRICT,
    assignee_id       bigint      REFERENCES users (id) ON DELETE SET NULL,
    resolved_at       timestamptz,
    closed_at         timestamptz,
    created_at        timestamptz NOT NULL DEFAULT now(),
    updated_at        timestamptz NOT NULL DEFAULT now(),
    -- El interno no puede estar `escalado`: el interno ES la escalación.
    CONSTRAINT internal_tickets_state_check CHECK (
        state IN ('nuevo', 'en progreso', 'en espera', 'resuelto', 'cerrado')
    ),
    CONSTRAINT internal_tickets_reason_check CHECK (btrim(escalation_reason) <> '')
);

COMMENT ON TABLE internal_tickets IS
    'Los tickets internos, uno por principal. Su contenido es el motivo del escalado: lo demás se lee del principal.';

-- El secuencial de cada año, que se incrementa en la misma transacción que crea el ticket: con
-- `MAX(number) + 1` dos tickets a la vez sacarían el mismo número (docs/modules/tickets.md, 2.2).
CREATE TABLE IF NOT EXISTS ticket_number_counters (
    year        integer PRIMARY KEY,
    last_number integer NOT NULL DEFAULT 0,
    CONSTRAINT ticket_number_counters_last_check CHECK (last_number >= 0)
);

COMMENT ON TABLE ticket_number_counters IS
    'El último número emitido en cada año. Sólo los tickets principales consumen secuencial.';

CREATE TABLE IF NOT EXISTS ticket_comments (
    id                bigserial   PRIMARY KEY,
    -- Exactamente una de las dos: es un comentario de un principal o de un interno, nunca de los dos.
    ticket_id         bigint      REFERENCES tickets (id) ON DELETE CASCADE,
    internal_ticket_id bigint     REFERENCES internal_tickets (id) ON DELETE CASCADE,
    author_id         bigint      NOT NULL REFERENCES users (id) ON DELETE RESTRICT,
    body              text        NOT NULL,
    edited_at         timestamptz,
    -- Borrado LÓGICO: la fila se queda con su marca y el texto se vacía de verdad.
    deleted_at        timestamptz,
    created_at        timestamptz NOT NULL DEFAULT now(),
    CONSTRAINT ticket_comments_one_target CHECK (
        (ticket_id IS NOT NULL AND internal_ticket_id IS NULL)
        OR (ticket_id IS NULL AND internal_ticket_id IS NOT NULL)
    )
);

COMMENT ON TABLE ticket_comments IS
    'La conversación de los dos tipos de ticket. Editables por su autor; borrar vacía el texto y deja la marca.';

CREATE TABLE IF NOT EXISTS ticket_attachments (
    id                 bigserial   PRIMARY KEY,
    ticket_id          bigint      REFERENCES tickets (id) ON DELETE CASCADE,
    internal_ticket_id bigint      REFERENCES internal_tickets (id) ON DELETE CASCADE,
    -- Nulo cuando el adjunto va en la descripción inicial, y no en un comentario.
    comment_id         bigint      REFERENCES ticket_comments (id) ON DELETE SET NULL,
    uploaded_by_id     bigint      NOT NULL REFERENCES users (id) ON DELETE RESTRICT,
    filename           text        NOT NULL,
    -- El nombre en disco, generado y único: el original puede repetirse y puede traer de todo.
    stored_name        text        NOT NULL UNIQUE,
    content_type       text        NOT NULL,
    size_bytes         bigint      NOT NULL,
    created_at         timestamptz NOT NULL DEFAULT now(),
    CONSTRAINT ticket_attachments_one_target CHECK (
        (ticket_id IS NOT NULL AND internal_ticket_id IS NULL)
        OR (ticket_id IS NULL AND internal_ticket_id IS NOT NULL)
    ),
    CONSTRAINT ticket_attachments_size_check CHECK (size_bytes > 0)
);

COMMENT ON TABLE ticket_attachments IS
    'Los adjuntos de los dos tipos. El archivo vive en disco (FILES_PATH), no en la base.';

CREATE TABLE IF NOT EXISTS ticket_history (
    id                 bigserial   PRIMARY KEY,
    ticket_id          bigint      REFERENCES tickets (id) ON DELETE CASCADE,
    internal_ticket_id bigint      REFERENCES internal_tickets (id) ON DELETE CASCADE,
    -- Nulo cuando lo hizo el sistema: pasar el principal a `en progreso` porque el interno se resolvió.
    actor_id           bigint      REFERENCES users (id) ON DELETE SET NULL,
    event              text        NOT NULL,
    from_state         text,
    to_state           text,
    detail             text,
    created_at         timestamptz NOT NULL DEFAULT now(),
    CONSTRAINT ticket_history_one_target CHECK (
        (ticket_id IS NOT NULL AND internal_ticket_id IS NULL)
        OR (ticket_id IS NULL AND internal_ticket_id IS NOT NULL)
    ),
    CONSTRAINT ticket_history_event_check CHECK (
        event IN ('creado', 'editado', 'asignado', 'estado', 'escalado', 'resuelto', 'cerrado', 'reabierto', 'observador', 'observador_anadido')
    )
);

COMMENT ON TABLE ticket_history IS
    'Lo que hizo el sistema en los dos tipos de ticket. No lo escribe nadie a mano: lo genera cada transición.';

-- La puesta al día de la restricción de eventos: `CREATE TABLE IF NOT EXISTS` no la toca en una base
-- que ya existía. `observador` se añadió el 2026-09-27 con los observadores, y `observador_anadido` el
-- 2026-09-28 al poder **añadirlos a mano** desde la ficha (docs/modules/tickets.md, sección 2.3.1 y
-- decisión 74): la misma lista se llena de dos maneras, y el historial dice cuál fue. Se quita y se
-- vuelve a poner, que es lo que hace que el archivo se pueda aplicar dos veces.
ALTER TABLE ticket_history DROP CONSTRAINT IF EXISTS ticket_history_event_check;
ALTER TABLE ticket_history
    ADD CONSTRAINT ticket_history_event_check CHECK (
        event IN ('creado', 'editado', 'asignado', 'estado', 'escalado', 'resuelto', 'cerrado', 'reabierto', 'observador', 'observador_anadido')
    );

-- Los observadores de un ticket: **quien sigue el ticket sin atenderlo**, y no son lo mismo que el
-- asignado —como mucho uno, y sigue siendo opcional— (docs/modules/tickets.md, sección 2.3.1,
-- decisión 60). Van en su tabla porque son muchos por ticket y porque la lista tiene que poder
-- cambiar sin tocar el texto de nadie: **quitar a mano manda sobre lo ya escrito** (decisión 63).
CREATE TABLE IF NOT EXISTS ticket_observers (
    id                 bigserial   PRIMARY KEY,
    -- Uno de los dos, y sólo uno: cada hilo tiene sus observadores, como sus comentarios.
    ticket_id          bigint      REFERENCES tickets (id) ON DELETE CASCADE,
    internal_ticket_id bigint      REFERENCES internal_tickets (id) ON DELETE CASCADE,
    -- Quien observa. Clave ajena a `users`: es una cuenta real, y desactivarla no la quita de aquí.
    account_id         bigint      NOT NULL REFERENCES users (id) ON DELETE CASCADE,
    -- Quien lo añadió: hoy, el autor de la mención.
    added_by_id        bigint      NOT NULL REFERENCES users (id) ON DELETE RESTRICT,
    created_at         timestamptz NOT NULL DEFAULT now(),
    CONSTRAINT ticket_observers_one_target CHECK (
        (ticket_id IS NOT NULL AND internal_ticket_id IS NULL)
        OR (ticket_id IS NULL AND internal_ticket_id IS NOT NULL)
    )
);

COMMENT ON TABLE ticket_observers IS
    'Quien sigue el ticket sin atenderlo. Se añade al etiquetar y se quita a mano —lo puede hacer cualquier técnico o desarrollador—, y quitar manda sobre lo ya escrito.';

-- Mencionar dos veces a la misma persona no la pone dos veces. Con `NULL` de por medio —una columna
-- nula siempre es distinta de otra— hacen falta **dos índices únicos parciales**, uno por cada lado,
-- como en los comentarios (docs/modules/tickets.md, sección 2.3.1).
CREATE UNIQUE INDEX IF NOT EXISTS ticket_observers_ticket_unique
    ON ticket_observers (ticket_id, account_id) WHERE ticket_id IS NOT NULL;
CREATE UNIQUE INDEX IF NOT EXISTS ticket_observers_internal_unique
    ON ticket_observers (internal_ticket_id, account_id) WHERE internal_ticket_id IS NOT NULL;

-- Y un índice por persona, que es como se pregunta «los tickets que observo» en la bandeja.
CREATE INDEX IF NOT EXISTS ticket_observers_account_idx ON ticket_observers (account_id);

-- Los índices de la sección 2.5, para lo que de verdad se pregunta: las bandejas, «mis tickets»,
-- los tickets de una persona y la conversación en orden.
CREATE INDEX IF NOT EXISTS tickets_state_idx ON tickets (state);
CREATE INDEX IF NOT EXISTS tickets_requester_idx ON tickets (requester_id);
CREATE INDEX IF NOT EXISTS tickets_assignee_idx ON tickets (assignee_id);
CREATE INDEX IF NOT EXISTS internal_tickets_state_idx ON internal_tickets (state);
CREATE INDEX IF NOT EXISTS internal_tickets_assignee_idx ON internal_tickets (assignee_id);
CREATE INDEX IF NOT EXISTS ticket_comments_ticket_idx ON ticket_comments (ticket_id, created_at);
CREATE INDEX IF NOT EXISTS ticket_comments_internal_idx ON ticket_comments (internal_ticket_id, created_at);
CREATE INDEX IF NOT EXISTS ticket_history_ticket_idx ON ticket_history (ticket_id, created_at);
CREATE INDEX IF NOT EXISTS ticket_history_internal_idx ON ticket_history (internal_ticket_id, created_at);
CREATE INDEX IF NOT EXISTS ticket_attachments_ticket_idx ON ticket_attachments (ticket_id);
CREATE INDEX IF NOT EXISTS ticket_attachments_internal_idx ON ticket_attachments (internal_ticket_id);

-- ============================================================================
-- tickets — puesta al día del texto: de texto plano a HTML
-- ============================================================================
--
-- El 2026-09-26 el texto de un ticket y el de un comentario **dejó de ser texto plano y pasó a ser HTML
-- con formato** (docs/modules/tickets.md, sección 2.3), así que una base que ya tenía tickets tiene sus
-- descripciones y sus comentarios con el texto a secas. Esta sección los pone al día: envuelve cada
-- texto en su párrafo, escapa lo que el HTML se come y cambia los saltos de línea por `<br>`, que es lo
-- que hace que un ticket viejo se lea **igual que antes** —ahora con formato— y no como una sola línea.
--
-- **El `&` va primero**: si se escapara después, los `&` de los `&lt;` y los `&gt;` recién escritos se
-- volverían a escapar y un `<` acabaría leyéndose `&amp;lt;`. Y los saltos van **después de escapar**,
-- porque el `<br>` que se escribe lleva un `<` que no hay que escapar.
--
-- Los tres saltos se cubren por separado y el de Windows (`\r\n`) antes que el suelto: al revés, el
-- `\r\n` quedaría como `<br>` y un `\r` colgando.
--
-- **La condición es «no tiene ni `<` ni `&`», y no «no empieza por `<`»** (corregido el 2026-09-26, en
-- el mismo trabajo que metió los adjuntos dentro del texto). Con la condición vieja, un cuerpo escrito
-- por el editor **empezaba por su texto y llevaba etiquetas dentro** —«Mira lo que me sale.<img
-- data-adjunto="captura.png">»—, así que volvía a entrar, se escapaba entero y las referencias a los
-- adjuntos se convertían en texto literal: aplicar el archivo dos veces **rompía lo que el editor
-- escribe**. Ahora se convierte **sólo lo que puede ser texto plano de antes**:
--
--   - un texto escrito por el editor **siempre lleva `&` o `<`**: escapa el `&` y el `<` al guardarlos
--     (`&amp;`, `&lt;`) y mete etiquetas en cuanto hay formato o un adjunto;
--   - un texto convertido por esta misma sección empieza por `<p>`, así que tampoco entra.
--
-- Lo que la condición deja fuera, y se dice: un texto antiguo que **llevara un `<` o un `&` a mano** no
-- se toca, y se lee como lo que es, sin formato. Es el precio de que la conversión se pueda aplicar
-- todas las veces que haga falta, que es lo que `docs/ambientes.md`, sección 5 pide.
--
-- **El `<p>` sigue puesto**, aunque ya no sea lo que sostiene la idempotencia: es cómo se lee un texto
-- como un párrafo.
--
-- **En producción no toca ninguna fila**, porque allí no hay todavía ningún ticket
-- (docs/ambientes.md, sección 5). Lo que sí hay que saber es que **las copias anteriores a ese día
-- tienen el texto plano**: una restauración vieja se ve igual, sin formato, y hay que volver a aplicar
-- este archivo para ponerla al día.

UPDATE tickets
SET description = '<p>' || replace(
        replace(
            replace(
                replace(
                    replace(replace(description, '&', '&amp;'), '<', '&lt;'), '>', '&gt;'
                ),
                E'\r\n', '<br>'
            ),
            E'\n', '<br>'
        ),
        E'\r', '<br>'
    ) || '</p>'
WHERE btrim(description) <> '' AND description !~ '[<&]';

UPDATE ticket_comments
SET body = '<p>' || replace(
        replace(
            replace(
                replace(
                    replace(replace(body, '&', '&amp;'), '<', '&lt;'), '>', '&gt;'
                ),
                E'\r\n', '<br>'
            ),
            E'\n', '<br>'
        ),
        E'\r', '<br>'
    ) || '</p>'
-- Un comentario borrado tiene el texto vacío a propósito y se queda vacío: no hay nada que convertir.
WHERE btrim(body) <> '' AND body !~ '[<&]';

-- ============================================================================
-- ai — puesta al día: la tabla de los dos resúmenes del motor de IA
-- (docs/modules/ai.md, secciones 3 y 4)
-- ============================================================================
--
-- La tabla nace **aquí, en la puesta al día de la 1.0.0**, y no en un archivo de versión nuevo: la
-- 1.0.0 no está cerrada, así que abrir una versión por una tabla sería inventar una migración para
-- algo que todavía no ha salido (docs/modules/ai.md, sección 4). `CREATE TABLE IF NOT EXISTS` la crea
-- en una base nueva y **no toca nada** en una que ya la tenga, que es lo que hace que este archivo
-- siga sirviendo para las dos cosas (docs/ambientes.md, sección 5).
--
-- Una fila por ticket y por campo, y **la fila se reescribe**: lo que interesa es lo último, y
-- guardar cada versión sería llenar la base de textos que nadie lee.

CREATE TABLE IF NOT EXISTS ai_insights (
    id            bigserial   PRIMARY KEY,
    -- El número del ticket, `ACME-2026-0042` o `INT-ACME-2026-0042`, **tal y como se lee en los
    -- correos**. No lleva clave ajena a propósito: el mismo hueco guarda el número de un principal y
    -- el de un interno, y esos dos números viven en dos tablas distintas.
    ticket_number text        NOT NULL,
    -- Los dos encargos que se le hacen al motor: de qué va el ticket y qué fue lo último que pasó.
    -- Son dos peticiones distintas con la misma forma (docs/modules/ai.md, decisión 7).
    kind          text        NOT NULL,
    -- `pendiente` mientras el motor no haya contestado, `listo` cuando hay texto, `error` cuando la
    -- respuesta no valía y `sin_motor` cuando hay motor configurado pero no contesta. Los dos fallos
    -- se guardan por separado porque **no significan lo mismo**: uno es culpa del modelo y el otro de
    -- que el contenedor no está (docs/modules/ai.md, decisión 8).
    state         text        NOT NULL DEFAULT 'pendiente',
    -- Las dos redacciones en la misma fila: el motor las devuelve juntas (decisión 3) y la pantalla
    -- enseña la del idioma de quien mira, así que traducir al vuelo sería pedir dos veces lo mismo.
    -- Nulas mientras no haya texto: la fila de un campo `pendiente` no tiene nada que enseñar.
    text_es       text,
    text_en       text,
    -- La clave del error (`ai.unavailable` o `ai.invalid`) cuando el estado no es `listo`. Se guarda
    -- la clave y no el texto del fallo, como en el resto de la API.
    error_key     text,
    -- El modelo que lo escribió, para saber con qué se generó: un texto viejo puede venir de otro
    -- modelo, y la calidad del resumen depende de él (docs/modules/ai.md, sección 7).
    model         text,
    -- Cuántos intentos lleva. Es lo que corta el reintento infinito: un bucle de reintentos contra un
    -- contenedor caído es ruido, y sin este número no habría forma de pararlo (decisión 8).
    attempts      integer     NOT NULL DEFAULT 0,
    -- Cuándo se pidió y cuándo se escribió. `requested_at` es lo que ordena la puesta al día del
    -- arranque; `generated_at` queda nulo hasta que hay texto.
    requested_at  timestamptz NOT NULL DEFAULT now(),
    generated_at  timestamptz,
    CONSTRAINT ai_insights_kind_check CHECK (kind IN ('motivo', 'ultima_accion')),
    CONSTRAINT ai_insights_state_check CHECK (state IN ('pendiente', 'listo', 'error', 'sin_motor')),
    CONSTRAINT ai_insights_attempts_check CHECK (attempts >= 0),
    -- **Una fila por ticket y por campo**, y la base lo garantiza: el campo se reescribe en cada
    -- movimiento del ticket y no se acumula historial (docs/modules/ai.md, sección 4).
    CONSTRAINT ai_insights_ticket_kind_key UNIQUE (ticket_number, kind)
);

COMMENT ON TABLE ai_insights IS
    'Los dos resúmenes que redacta el motor de IA: motivo y última acción, en español y en inglés. No sustituyen a nada.';

COMMENT ON COLUMN ai_insights.state IS
    'pendiente, listo, error o sin_motor. `error` es una respuesta que no valía y `sin_motor` un motor que no contesta.';

-- Se indexa por el número del ticket porque es como se piden: la lista de tickets pide los suyos **en
-- bloque**, no uno a uno (docs/modules/ai.md, decisión 9).
CREATE INDEX IF NOT EXISTS ai_insights_ticket_number_idx ON ai_insights (ticket_number);

-- ============================================================================
-- tickets — puesta al día: las categorías y las etiquetas
-- (docs/modules/tickets.md, sección 2.3.2, decisiones 64 a 68)
-- ============================================================================
--
-- **La categoría es obligatoria y la garantiza la base, no la pantalla** (decisión 65): la columna
-- `tickets.category_id` es `NOT NULL` con clave ajena al catálogo. Para que eso valga desde el primer
-- día, la migración **crea «General»** y se la pone a todo lo que ya exista —en producción no hay
-- tickets y no toca ninguna fila; en desarrollo sí—.
--
-- **Retirar una categoría es desactivarla** (decisión 66), como se desactiva una cuenta: deja de
-- ofrecerse al crear un ticket y **los tickets que la tienen la conservan**. No se borra ninguna fila,
-- y es lo que hace que «no hay ticket sin categoría» se pueda cumplir siempre.

CREATE TABLE IF NOT EXISTS ticket_categories (
    id             bigserial   PRIMARY KEY,
    -- El nombre tal cual se lee en la pantalla.
    name           text        NOT NULL,
    -- El nombre en minúsculas y sin acentos: es lo que impide «Red» y «red» en el mismo catálogo.
    normalized     text        NOT NULL UNIQUE,
    -- Retirada deja de ofrecerse; los tickets que la tienen la siguen enseñando.
    active         boolean     NOT NULL DEFAULT true,
    -- Nulo en la categoría de fábrica («General»), que no la crea nadie.
    created_by_id  bigint      REFERENCES users (id) ON DELETE SET NULL,
    created_at     timestamptz NOT NULL DEFAULT now(),
    -- Cuándo se retiró. Volver a activarla lo limpia.
    deactivated_at timestamptz
);

COMMENT ON TABLE ticket_categories IS
    'El catálogo de categorías: una por ticket y obligatoria. Retirar es desactivar, y no se borra ninguna fila.';

COMMENT ON COLUMN ticket_categories.normalized IS
    'El nombre en minúsculas y sin acentos. Único: es lo que impide «Red» y «red» en el mismo catálogo.';

-- La categoría de fábrica, para que ningún ticket pueda quedarse sin ella. `ON CONFLICT` es lo que
-- hace que volver a aplicar el archivo no falle.
INSERT INTO ticket_categories (name, normalized, active) VALUES ('General', 'general', true)
ON CONFLICT (normalized) DO NOTHING;

-- La columna nace **nula** para poder rellenarla antes de exigirla: primero se pone «General» a todo
-- lo que ya exista y sólo después se pone el `NOT NULL`. Al revés, una base que ya tuviera tickets no
-- podría aplicar la migración.
ALTER TABLE tickets ADD COLUMN IF NOT EXISTS category_id bigint;

UPDATE tickets
SET category_id = (SELECT id FROM ticket_categories WHERE normalized = 'general')
WHERE category_id IS NULL;

-- `SET NOT NULL` es idempotente: sobre una columna que ya lo es, no hace nada.
ALTER TABLE tickets ALTER COLUMN category_id SET NOT NULL;

-- La clave ajena, con el nombre que le pone PostgreSQL y buscándola antes: no hay
-- `ADD CONSTRAINT IF NOT EXISTS`, así que sin esta comprobación el archivo no se podría aplicar dos
-- veces. `ON DELETE RESTRICT` porque una categoría en uso no se puede borrar: se retira.
DO $$
BEGIN
    IF NOT EXISTS (SELECT 1 FROM pg_constraint WHERE conname = 'tickets_category_id_fkey') THEN
        ALTER TABLE tickets
            ADD CONSTRAINT tickets_category_id_fkey
            FOREIGN KEY (category_id) REFERENCES ticket_categories (id) ON DELETE RESTRICT;
    END IF;
END $$;

COMMENT ON COLUMN tickets.category_id IS
    'La categoría del ticket: obligatoria, una sola, con clave ajena al catálogo. No se puede quedar sin ella.';

-- El índice de la categoría: es como se filtra la lista por su chip.
CREATE INDEX IF NOT EXISTS tickets_category_id_idx ON tickets (category_id);

-- **Las etiquetas se mantienen y siguen naciendo solas** (decisión 72): se crean, se renombran —y el
-- cambio vale para **todos** los tickets que la llevan— y se retiran —quitándolas de esos tickets—.
-- Para que una etiqueta pueda existir **sin que ningún ticket la lleve**, el catálogo pasa a ser su
-- propia tabla y `ticket_tags` deja de guardar el texto: guarda una clave ajena a él. **Una sola
-- fuente de verdad**: el nombre vive en el catálogo y `ticket_tags` dice qué ticket la lleva. Con eso
-- renombrar es cambiar una fila y retirar es borrar las que la usan.
CREATE TABLE IF NOT EXISTS ticket_tag_names (
    id            bigserial   PRIMARY KEY,
    -- El nombre tal cual se lee. La normalización de una etiqueta ya la deja en minúsculas y con
    -- guiones, así que hoy coincide con `normalized`: la columna está para que la comparación sea
    -- siempre con la misma forma, como en las categorías.
    tag           text        NOT NULL,
    -- En minúsculas y sin acentos: **único**, y es lo que impide `red-wifi` y `Red-Wifi` a la vez.
    normalized    text        NOT NULL UNIQUE,
    -- **Puede quedar nulo**: una etiqueta nacida al escribirla en un ticket la crea esa persona, pero
    -- una etiqueta tiene que poder existir sin nadie detrás —la deja la migración al poner al día una
    -- base vieja, o el seeder de desarrollo—.
    created_by_id bigint      REFERENCES users (id) ON DELETE SET NULL,
    created_at    timestamptz NOT NULL DEFAULT now()
);

COMMENT ON TABLE ticket_tag_names IS
    'El catálogo de etiquetas: pueden existir sin que ningún ticket las lleve. Una sola fuente de verdad: `ticket_tags` apunta aquí.';

COMMENT ON COLUMN ticket_tag_names.normalized IS
    'El nombre en minúsculas y sin acentos. Único: es lo que impide `red-wifi` y `Red-Wifi` en el catálogo.';

COMMENT ON COLUMN ticket_tag_names.created_by_id IS
    'Quién la creó. Nulo en una etiqueta que no nació de nadie: puede existir sin que ningún ticket la lleve.';

-- La tabla puente nace ya con la clave ajena cuando la base es nueva. En una base que ya existía,
-- `CREATE TABLE IF NOT EXISTS` no toca nada y la pone al día el bloque de más abajo.
CREATE TABLE IF NOT EXISTS ticket_tags (
    id            bigserial   PRIMARY KEY,
    -- Con borrado en cascada: un ticket que desaparece no deja etiquetas huérfanas.
    ticket_id     bigint      NOT NULL REFERENCES tickets (id) ON DELETE CASCADE,
    -- La clave ajena al catálogo. Su borrado en cascada es lo que hace que **retirar una etiqueta la
    -- quite de todos sus tickets** en un solo movimiento (decisión 72).
    tag_id        bigint      NOT NULL,
    created_by_id bigint      NOT NULL REFERENCES users (id) ON DELETE RESTRICT,
    created_at    timestamptz NOT NULL DEFAULT now()
);

COMMENT ON TABLE ticket_tags IS
    'Qué ticket lleva qué etiqueta: muchas por ticket y únicas por ticket. El nombre vive en `ticket_tag_names` y aquí sólo su clave ajena.';

-- Puesta al día: de `ticket_tags(ticket_id, tag)` a `ticket_tags(ticket_id, tag_id)`.
--
-- Una base que ya existía tiene la columna de texto con filas dentro. El orden importa: **primero se
-- meten en el catálogo todas las etiquetas distintas que ya existan**, después se rellena `tag_id`
-- desde ellas y **sólo entonces se quita `tag`**; al revés se perderían.
--
-- Todo va dentro de un `DO` que sólo entra si la columna `tag` sigue ahí: es lo que hace que aplicar
-- el archivo dos veces seguidas no falle, porque la segunda vez no hay nada que convertir.
DO $$
BEGIN
    IF EXISTS (
        SELECT 1 FROM information_schema.columns
        WHERE table_schema = current_schema()
          AND table_name = 'ticket_tags'
          AND column_name = 'tag'
    ) THEN
        -- 1. Las etiquetas que ya existían, al catálogo. Se comparan por su forma normalizada —que es
        --    como se guardaban— y `ON CONFLICT` deja fuera la que ya estuviera. **Sin autor**: las que
        --    ya existían pueden quedar con `created_by_id` nulo (sección 2.3.2).
        INSERT INTO ticket_tag_names (tag, normalized)
        SELECT DISTINCT ON (lower(tag)) tag, lower(tag)
        FROM ticket_tags
        ON CONFLICT (normalized) DO NOTHING;

        -- 2. La columna nueva nace **nula** para poder rellenarla antes de exigirla: al revés, una
        --    base que ya tuviera filas no podría aplicar la migración.
        ALTER TABLE ticket_tags ADD COLUMN IF NOT EXISTS tag_id bigint;

        -- 3. Y se rellena desde el catálogo.
        UPDATE ticket_tags tt
        SET tag_id = n.id
        FROM ticket_tag_names n
        WHERE n.normalized = lower(tt.tag)
          AND tt.tag_id IS NULL;

        -- 4. Si dos textos distintos de un mismo ticket acabaran apuntando a la misma etiqueta —dos
        --    formas de escribir lo mismo—, sobra una: la misma etiqueta dos veces en un ticket es una.
        DELETE FROM ticket_tags a
        USING ticket_tags b
        WHERE a.ticket_id = b.ticket_id
          AND a.tag_id = b.tag_id
          AND a.id > b.id;

        -- 5. Con todas las filas rellenas, ya se puede exigir.
        ALTER TABLE ticket_tags ALTER COLUMN tag_id SET NOT NULL;

        -- 6. Y ahora sí, fuera el texto y su índice: el nombre vive en el catálogo.
        ALTER TABLE ticket_tags DROP CONSTRAINT IF EXISTS ticket_tags_ticket_tag_key;
        DROP INDEX IF EXISTS ticket_tags_tag_idx;
        ALTER TABLE ticket_tags DROP COLUMN IF EXISTS tag;
    END IF;
END $$;

-- La clave ajena al catálogo, con el nombre que le pone PostgreSQL y buscándola antes: no hay
-- `ADD CONSTRAINT IF NOT EXISTS`, así que sin esta comprobación el archivo no se podría aplicar dos
-- veces. **`ON DELETE CASCADE`**: retirar una etiqueta la quita de todos sus tickets.
DO $$
BEGIN
    IF NOT EXISTS (SELECT 1 FROM pg_constraint WHERE conname = 'ticket_tags_tag_id_fkey') THEN
        ALTER TABLE ticket_tags
            ADD CONSTRAINT ticket_tags_tag_id_fkey
            FOREIGN KEY (tag_id) REFERENCES ticket_tag_names (id) ON DELETE CASCADE;
    END IF;
END $$;

-- La misma etiqueta dos veces en un ticket es la misma etiqueta: la base no deja repetirla.
DO $$
BEGIN
    IF NOT EXISTS (SELECT 1 FROM pg_constraint WHERE conname = 'ticket_tags_ticket_tag_key') THEN
        ALTER TABLE ticket_tags
            ADD CONSTRAINT ticket_tags_ticket_tag_key UNIQUE (ticket_id, tag_id);
    END IF;
END $$;

-- El índice por `tag_id`: es como se cuentan los tickets que llevan una etiqueta y como se filtran.
CREATE INDEX IF NOT EXISTS ticket_tags_tag_id_idx ON ticket_tags (tag_id);

COMMIT;
