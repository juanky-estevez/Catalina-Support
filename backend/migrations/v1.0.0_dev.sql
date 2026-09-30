-- Catalina-Support — datos de ejemplo de DESARROLLO
--
-- **Este archivo NO se aplica en producción.** El esquema y lo que la aplicación necesita para
-- funcionar están en `v1.0.0.sql`; esto son datos para mirar la aplicación con contenido: cuentas,
-- 25 tickets con su historia, escalados, adjuntos y avisos.
--
-- Se aplica **con su guion**, que además copia los archivos de los adjuntos:
--
--   ./scripts/dev-seed.sh
--
-- Es idempotente: empieza borrando los datos de ejemplo anteriores y los vuelve a crear. **Borra
-- todos los tickets**, también los que hayas creado a mano probando, así que deja el entorno en un
-- estado conocido: es un seeder de desarrollo, no una migración.
--
-- **Las once cuentas entran con la misma contraseña de desarrollo: `123123123`** (decisión del
-- responsable, 2026-09-25): en desarrollo se entra y se sale muchas veces al día, y una contraseña
-- que se escribe con una mano es lo que hace que probar no dé pereza.
--
-- **Cumple la política del producto**, que pide ocho caracteres como mínimo, así que este seeder no
-- se salta ninguna regla: la escribe aquí ya cifrada porque un archivo SQL no puede llamar a la
-- aplicación, y el resultado es el mismo que si se hubiera establecido desde la pantalla.

BEGIN;

-- ============================================================================
-- 1. Se limpia lo de la vez anterior
-- ============================================================================

-- Los tickets se llevan por delante sus comentarios, adjuntos, historial e internos (ON DELETE
-- CASCADE). Y van **antes** que las cuentas: un ticket apunta a quien lo pidió, y la base no deja
-- borrar una cuenta que tiene tickets detrás.
-- **Los resúmenes del motor de IA van con los tickets y hay que borrarlos aquí** (docs/modules/ai.md):
-- la tabla de los resúmenes se lleva por **número de ticket**, y este archivo **vuelve a emitir los
-- mismos números** (CS-2026-0001 y siguientes), así que una fila vieja se pegaría al ticket nuevo y
-- enseñaría el motivo de otro. No hay clave ajena que lo impida —los tickets viven en su módulo y los
-- resúmenes en el suyo—, así que se limpia a mano, aquí y en el mismo sitio donde se borran los
-- tickets.
DELETE FROM ai_insights;
-- Las etiquetas cuelgan del ticket y se van con él (ON DELETE CASCADE), pero se borran antes para
-- dejar dicho que este archivo las vuelve a crear. **El catálogo de etiquetas se borra también**:
-- este archivo deja el entorno en un estado conocido y lo vuelve a dar de alta con los tickets.
-- Las categorías van **después** de los tickets: un ticket apunta a la suya, y la base no deja
-- borrar una que esté en uso.
DELETE FROM ticket_tags;
DELETE FROM tickets;
DELETE FROM ticket_tag_names;
DELETE FROM ticket_categories;
DELETE FROM ticket_number_counters;

-- Las cuentas de ejemplo, con sus enlaces de contraseña. Las de las pruebas (`e2e-…`) no se tocan.
--
-- La comparación va **en minúsculas**: el correo se guarda tal y como lo escribe quien lo escribe, y
-- «Ana.Perez@Demo.com» es la misma persona que «ana.perez@demo.com». Sin esto, una cuenta de
-- ejemplo con mayúsculas sobrevivía a la limpieza y el entorno dejaba de estar en un estado conocido.
DELETE FROM password_tokens WHERE user_id IN (SELECT id FROM users WHERE lower(email) LIKE '%@demo.com');
-- **Las cuentas de las pruebas de interfaz también se borran** (decisión del responsable, 2026-09-28):
-- cada pasada de la capa de Playwright crea las suyas —con la marca del momento—, así que aquí no se
-- pierde nada, y dejarlas convertía el buscador de a quién etiquetar en una lista de **miles** de
-- personas que se llaman todas igual (2 521 cuentas y 636 técnicos activos, medido el 2026-09-28). El
-- entorno de ejemplo queda con las **once** cuentas de siempre.
DELETE FROM users WHERE lower(email) LIKE '%@demo.com';

-- ============================================================================
-- 2. Las cuentas: cinco que piden, tres de Soporte y tres de Desarrollo
-- ============================================================================

INSERT INTO users (name, last_name, email, password_hash, role, origin, language, is_active, created_at, updated_at) VALUES
    ('Usuario', 'Uno', 'user1@demo.com', '$2a$12$VD3neJ0zGzwWD2PmyPe55O6GRHbLfCd4fspMzk9U7WRRilLiZvFe.', 'usuario', 'local', 'es', true, (date_trunc('day', now()) - interval '90 days' + interval '8 hours 0 minutes'), (date_trunc('day', now()) - interval '90 days' + interval '8 hours 0 minutes')),
    ('Usuario', 'Dos', 'user2@demo.com', '$2a$12$VD3neJ0zGzwWD2PmyPe55O6GRHbLfCd4fspMzk9U7WRRilLiZvFe.', 'usuario', 'local', 'es', true, (date_trunc('day', now()) - interval '90 days' + interval '8 hours 0 minutes'), (date_trunc('day', now()) - interval '90 days' + interval '8 hours 0 minutes')),
    ('Usuario', 'Tres', 'user3@demo.com', '$2a$12$VD3neJ0zGzwWD2PmyPe55O6GRHbLfCd4fspMzk9U7WRRilLiZvFe.', 'usuario', 'local', 'es', true, (date_trunc('day', now()) - interval '90 days' + interval '8 hours 0 minutes'), (date_trunc('day', now()) - interval '90 days' + interval '8 hours 0 minutes')),
    ('Usuario', 'Cuatro', 'user4@demo.com', '$2a$12$VD3neJ0zGzwWD2PmyPe55O6GRHbLfCd4fspMzk9U7WRRilLiZvFe.', 'usuario', 'local', 'es', true, (date_trunc('day', now()) - interval '90 days' + interval '8 hours 0 minutes'), (date_trunc('day', now()) - interval '90 days' + interval '8 hours 0 minutes')),
    ('Usuario', 'Cinco', 'user5@demo.com', '$2a$12$VD3neJ0zGzwWD2PmyPe55O6GRHbLfCd4fspMzk9U7WRRilLiZvFe.', 'usuario', 'local', 'es', true, (date_trunc('day', now()) - interval '90 days' + interval '8 hours 0 minutes'), (date_trunc('day', now()) - interval '90 days' + interval '8 hours 0 minutes')),
    ('Soporte', 'Uno', 'support1@demo.com', '$2a$12$VD3neJ0zGzwWD2PmyPe55O6GRHbLfCd4fspMzk9U7WRRilLiZvFe.', 'soporte', 'local', 'es', true, (date_trunc('day', now()) - interval '90 days' + interval '8 hours 0 minutes'), (date_trunc('day', now()) - interval '90 days' + interval '8 hours 0 minutes')),
    ('Soporte', 'Dos', 'support2@demo.com', '$2a$12$VD3neJ0zGzwWD2PmyPe55O6GRHbLfCd4fspMzk9U7WRRilLiZvFe.', 'soporte', 'local', 'es', true, (date_trunc('day', now()) - interval '90 days' + interval '8 hours 0 minutes'), (date_trunc('day', now()) - interval '90 days' + interval '8 hours 0 minutes')),
    ('Soporte', 'Tres', 'support3@demo.com', '$2a$12$VD3neJ0zGzwWD2PmyPe55O6GRHbLfCd4fspMzk9U7WRRilLiZvFe.', 'soporte', 'local', 'es', true, (date_trunc('day', now()) - interval '90 days' + interval '8 hours 0 minutes'), (date_trunc('day', now()) - interval '90 days' + interval '8 hours 0 minutes')),
    ('Desarrollador', '1', 'dev1@demo.com', '$2a$12$VD3neJ0zGzwWD2PmyPe55O6GRHbLfCd4fspMzk9U7WRRilLiZvFe.', 'desarrollo', 'local', 'es', true, (date_trunc('day', now()) - interval '90 days' + interval '8 hours 0 minutes'), (date_trunc('day', now()) - interval '90 days' + interval '8 hours 0 minutes')),
    ('Desarrollador', '2', 'dev2@demo.com', '$2a$12$VD3neJ0zGzwWD2PmyPe55O6GRHbLfCd4fspMzk9U7WRRilLiZvFe.', 'desarrollo', 'local', 'es', true, (date_trunc('day', now()) - interval '90 days' + interval '8 hours 0 minutes'), (date_trunc('day', now()) - interval '90 days' + interval '8 hours 0 minutes')),
    ('Desarrollador', '3', 'dev3@demo.com', '$2a$12$VD3neJ0zGzwWD2PmyPe55O6GRHbLfCd4fspMzk9U7WRRilLiZvFe.', 'desarrollo', 'local', 'es', true, (date_trunc('day', now()) - interval '90 days' + interval '8 hours 0 minutes'), (date_trunc('day', now()) - interval '90 days' + interval '8 hours 0 minutes'));

-- ============================================================================
-- 3. El catálogo de categorías
-- ============================================================================

-- Las cinco del ejemplo. **«General» es la de fábrica**: la crea la migración para que ningún ticket
-- pueda quedarse sin categoría, y aquí se vuelve a poner porque este archivo borra el catálogo para
-- dejarlo en un estado conocido. Las otras cuatro son las que enseñan la función: agrupar por «qué
-- es» el ticket.
--
-- `created_by_id` apunta a Soporte Uno salvo en «General», que no la crea nadie. `ON CONFLICT` deja
-- el archivo aplicable dos veces por si alguien lo lanza sin la limpieza de arriba.
INSERT INTO ticket_categories (name, normalized, active, created_by_id) VALUES
    ('General',    'general',    true, NULL),
    ('Red',        'red',        true, (SELECT id FROM users WHERE email = 'support1@demo.com')),
    ('Software',   'software',   true, (SELECT id FROM users WHERE email = 'support1@demo.com')),
    ('Licencias',  'licencias',  true, (SELECT id FROM users WHERE email = 'support1@demo.com')),
    ('Impresoras', 'impresoras', true, (SELECT id FROM users WHERE email = 'support1@demo.com'))
ON CONFLICT (normalized) DO NOTHING;

-- ============================================================================
-- 4. Los 25 tickets, con su asignación, sus comentarios y su historia
-- ============================================================================

-- **Las descripciones y los comentarios van en HTML**, porque es lo que son desde el 2026-09-26
-- (docs/modules/tickets.md, sección 2.3): cada texto de ejemplo va dentro de su `<p>`, sin más formato
-- del que hacía falta. **Ninguno lleva adjuntos dentro del texto**: los adjuntos de ejemplo siguen
-- colgando del ticket y del comentario, y se ven en la lista del final, que es el camino que hay que
-- poder probar.

-- CS-2026-0001 · No puedo entrar en la aplicación
INSERT INTO tickets (number, number_year, number_seq, subject, description, state, requester_id, created_by_id, assignee_id, resolved_at, closed_at, created_at, updated_at, category_id)
VALUES ('CS-2026-0001', 2026, 1, 'No puedo entrar en la aplicación', '<p>Desde ayer me dice que la contraseña no es correcta y no he cambiado nada.</p>', 'nuevo',
        (SELECT id FROM users WHERE email = 'user1@demo.com'), (SELECT id FROM users WHERE email = 'user1@demo.com'), NULL,
        NULL, NULL,
        (date_trunc('day', now()) - interval '1 days' + interval '9 hours 0 minutes'), (date_trunc('day', now()) - interval '1 days' + interval '17 hours 0 minutes'), (SELECT id FROM ticket_categories WHERE normalized = 'software'));

INSERT INTO ticket_history (ticket_id, actor_id, event, from_state, to_state, detail, created_at)
VALUES ((SELECT id FROM tickets WHERE number = 'CS-2026-0001'), (SELECT id FROM users WHERE email = 'user1@demo.com'), 'creado', NULL, 'nuevo', NULL, (date_trunc('day', now()) - interval '1 days' + interval '9 hours 0 minutes'));

-- CS-2026-0002 · El informe mensual sale con las fechas cambiadas
INSERT INTO tickets (number, number_year, number_seq, subject, description, state, requester_id, created_by_id, assignee_id, resolved_at, closed_at, created_at, updated_at, category_id)
VALUES ('CS-2026-0002', 2026, 2, 'El informe mensual sale con las fechas cambiadas', '<p>Descargo el informe de agosto y las columnas de fecha salen en otro formato. Adjunto una captura.</p>', 'nuevo',
        (SELECT id FROM users WHERE email = 'user2@demo.com'), (SELECT id FROM users WHERE email = 'user2@demo.com'), NULL,
        NULL, NULL,
        (date_trunc('day', now()) - interval '2 days' + interval '9 hours 0 minutes'), (date_trunc('day', now()) - interval '1 days' + interval '17 hours 0 minutes'), (SELECT id FROM ticket_categories WHERE normalized = 'software'));

INSERT INTO ticket_history (ticket_id, actor_id, event, from_state, to_state, detail, created_at)
VALUES ((SELECT id FROM tickets WHERE number = 'CS-2026-0002'), (SELECT id FROM users WHERE email = 'user2@demo.com'), 'creado', NULL, 'nuevo', NULL, (date_trunc('day', now()) - interval '2 days' + interval '9 hours 0 minutes'));

-- El archivo de este adjunto lo copia `scripts/dev-seed.sh` desde `config/seed/captura.png`.
INSERT INTO ticket_attachments (ticket_id, uploaded_by_id, filename, stored_name, content_type, size_bytes, created_at)
VALUES ((SELECT id FROM tickets WHERE number = 'CS-2026-0002'), (SELECT id FROM users WHERE email = 'user2@demo.com'), 'captura.png', '2026/CS-2026-0002/eee4d316b5ebefd9b17df4eed40097a7.png', 'image/png', 1024, (date_trunc('day', now()) - interval '2 days' + interval '9 hours 30 minutes'));

-- CS-2026-0003 · Pido un usuario para la persona nueva
INSERT INTO tickets (number, number_year, number_seq, subject, description, state, requester_id, created_by_id, assignee_id, resolved_at, closed_at, created_at, updated_at, category_id)
VALUES ('CS-2026-0003', 2026, 3, 'Pido un usuario para la persona nueva', '<p>Ha entrado una compañera al equipo y necesita acceso a la aplicación.</p>', 'nuevo',
        (SELECT id FROM users WHERE email = 'user3@demo.com'), (SELECT id FROM users WHERE email = 'user3@demo.com'), NULL,
        NULL, NULL,
        (date_trunc('day', now()) - interval '3 days' + interval '9 hours 0 minutes'), (date_trunc('day', now()) - interval '2 days' + interval '17 hours 0 minutes'), (SELECT id FROM ticket_categories WHERE normalized = 'licencias'));

INSERT INTO ticket_history (ticket_id, actor_id, event, from_state, to_state, detail, created_at)
VALUES ((SELECT id FROM tickets WHERE number = 'CS-2026-0003'), (SELECT id FROM users WHERE email = 'user3@demo.com'), 'creado', NULL, 'nuevo', NULL, (date_trunc('day', now()) - interval '3 days' + interval '9 hours 0 minutes'));

-- CS-2026-0004 · La impresora de la segunda planta no imprime
INSERT INTO tickets (number, number_year, number_seq, subject, description, state, requester_id, created_by_id, assignee_id, resolved_at, closed_at, created_at, updated_at, category_id)
VALUES ('CS-2026-0004', 2026, 4, 'La impresora de la segunda planta no imprime', '<p>Manda el trabajo pero no sale nada. Ya la hemos apagado y encendido.</p>', 'escalado',
        (SELECT id FROM users WHERE email = 'user4@demo.com'), (SELECT id FROM users WHERE email = 'user4@demo.com'), (SELECT id FROM users WHERE email = 'support1@demo.com'),
        NULL, NULL,
        (date_trunc('day', now()) - interval '6 days' + interval '9 hours 0 minutes'), (date_trunc('day', now()) - interval '5 days' + interval '17 hours 0 minutes'), (SELECT id FROM ticket_categories WHERE normalized = 'impresoras'));

INSERT INTO ticket_history (ticket_id, actor_id, event, from_state, to_state, detail, created_at)
VALUES ((SELECT id FROM tickets WHERE number = 'CS-2026-0004'), (SELECT id FROM users WHERE email = 'support1@demo.com'), 'escalado', 'en progreso', 'escalado', 'La impresora necesita que alguien mire el servidor de colas, no es cosa del equipo.', (date_trunc('day', now()) - interval '4 days' + interval '11 hours 0 minutes'));
INSERT INTO ticket_history (ticket_id, actor_id, event, from_state, to_state, detail, created_at)
VALUES ((SELECT id FROM tickets WHERE number = 'CS-2026-0004'), (SELECT id FROM users WHERE email = 'support1@demo.com'), 'estado', 'nuevo', 'en progreso', NULL, (date_trunc('day', now()) - interval '5 days' + interval '10 hours 0 minutes'));
INSERT INTO ticket_history (ticket_id, actor_id, event, from_state, to_state, detail, created_at)
VALUES ((SELECT id FROM tickets WHERE number = 'CS-2026-0004'), (SELECT id FROM users WHERE email = 'user4@demo.com'), 'creado', NULL, 'nuevo', NULL, (date_trunc('day', now()) - interval '6 days' + interval '9 hours 0 minutes'));
INSERT INTO ticket_history (ticket_id, actor_id, event, from_state, to_state, detail, created_at)
VALUES ((SELECT id FROM tickets WHERE number = 'CS-2026-0004'), (SELECT id FROM users WHERE email = 'support1@demo.com'), 'asignado', NULL, NULL, 'Soporte Uno', (date_trunc('day', now()) - interval '6 days' + interval '10 hours 0 minutes'));

INSERT INTO ticket_comments (ticket_id, author_id, body, created_at)
VALUES ((SELECT id FROM tickets WHERE number = 'CS-2026-0004'), (SELECT id FROM users WHERE email = 'support1@demo.com'), '<p>¿Le pasa a todo el mundo o sólo a tu equipo?</p>', (date_trunc('day', now()) - interval '5 days' + interval '11 hours 15 minutes'));
INSERT INTO ticket_comments (ticket_id, author_id, body, created_at)
VALUES ((SELECT id FROM tickets WHERE number = 'CS-2026-0004'), (SELECT id FROM users WHERE email = 'user4@demo.com'), '<p>A todos: lo hemos probado entre tres.</p>', (date_trunc('day', now()) - interval '4 days' + interval '12 hours 15 minutes'));

-- INT-CS-2026-0004: el ticket interno que cuelga de CS-2026-0004
INSERT INTO internal_tickets (ticket_id, number, state, escalation_reason, created_by_id, assignee_id, resolved_at, closed_at, created_at, updated_at)
VALUES ((SELECT id FROM tickets WHERE number = 'CS-2026-0004'), 'INT-CS-2026-0004', 'nuevo', 'La impresora necesita que alguien mire el servidor de colas, no es cosa del equipo.', (SELECT id FROM users WHERE email = 'support1@demo.com'), NULL,
        NULL, NULL,
        (date_trunc('day', now()) - interval '4 days' + interval '11 hours 0 minutes'), (date_trunc('day', now()) - interval '3 days' + interval '16 hours 0 minutes'));

INSERT INTO ticket_history (internal_ticket_id, actor_id, event, from_state, to_state, detail, created_at)
VALUES ((SELECT id FROM internal_tickets WHERE number = 'INT-CS-2026-0004'), (SELECT id FROM users WHERE email = 'support1@demo.com'), 'creado', NULL, 'nuevo', 'La impresora necesita que alguien mire el servidor de colas, no es cosa del equipo.', (date_trunc('day', now()) - interval '4 days' + interval '11 hours 0 minutes'));
INSERT INTO ticket_comments (internal_ticket_id, author_id, body, created_at)
VALUES ((SELECT id FROM internal_tickets WHERE number = 'INT-CS-2026-0004'), (SELECT id FROM users WHERE email = 'support1@demo.com'), '<p>Te lo paso con todo lo que he probado hasta ahora.</p>', (date_trunc('day', now()) - interval '4 days' + interval '12 hours 20 minutes'));

-- CS-2026-0005 · Los adjuntos de más de 10 MB no se suben
INSERT INTO tickets (number, number_year, number_seq, subject, description, state, requester_id, created_by_id, assignee_id, resolved_at, closed_at, created_at, updated_at, category_id)
VALUES ('CS-2026-0005', 2026, 5, 'Los adjuntos de más de 10 MB no se suben', '<p>Intento adjuntar un PDF de 12 MB y se queda pensando y no termina.</p>', 'en progreso',
        (SELECT id FROM users WHERE email = 'user5@demo.com'), (SELECT id FROM users WHERE email = 'user5@demo.com'), (SELECT id FROM users WHERE email = 'support2@demo.com'),
        NULL, NULL,
        (date_trunc('day', now()) - interval '9 days' + interval '9 hours 0 minutes'), (date_trunc('day', now()) - interval '8 days' + interval '17 hours 0 minutes'), (SELECT id FROM ticket_categories WHERE normalized = 'software'));

INSERT INTO ticket_history (ticket_id, actor_id, event, from_state, to_state, detail, created_at)
VALUES ((SELECT id FROM tickets WHERE number = 'CS-2026-0005'), (SELECT id FROM users WHERE email = 'support2@demo.com'), 'estado', 'nuevo', 'en progreso', NULL, (date_trunc('day', now()) - interval '8 days' + interval '10 hours 0 minutes'));
INSERT INTO ticket_history (ticket_id, actor_id, event, from_state, to_state, detail, created_at)
VALUES ((SELECT id FROM tickets WHERE number = 'CS-2026-0005'), (SELECT id FROM users WHERE email = 'user5@demo.com'), 'creado', NULL, 'nuevo', NULL, (date_trunc('day', now()) - interval '9 days' + interval '9 hours 0 minutes'));
INSERT INTO ticket_history (ticket_id, actor_id, event, from_state, to_state, detail, created_at)
VALUES ((SELECT id FROM tickets WHERE number = 'CS-2026-0005'), (SELECT id FROM users WHERE email = 'support2@demo.com'), 'asignado', NULL, NULL, 'Soporte Dos', (date_trunc('day', now()) - interval '9 days' + interval '10 hours 0 minutes'));

INSERT INTO ticket_comments (ticket_id, author_id, body, created_at)
VALUES ((SELECT id FROM tickets WHERE number = 'CS-2026-0005'), (SELECT id FROM users WHERE email = 'support2@demo.com'), '<p>Estamos mirándolo. Mientras tanto, ¿puedes dejarlo en la carpeta compartida?</p>', (date_trunc('day', now()) - interval '8 days' + interval '11 hours 15 minutes'));
INSERT INTO ticket_comments (ticket_id, author_id, body, created_at)
VALUES ((SELECT id FROM tickets WHERE number = 'CS-2026-0005'), (SELECT id FROM users WHERE email = 'user5@demo.com'), '<p>Hecho, ahí está.</p>', (date_trunc('day', now()) - interval '7 days' + interval '12 hours 15 minutes'));

-- CS-2026-0006 · Cambio de correo en mi cuenta
INSERT INTO tickets (number, number_year, number_seq, subject, description, state, requester_id, created_by_id, assignee_id, resolved_at, closed_at, created_at, updated_at, category_id)
VALUES ('CS-2026-0006', 2026, 6, 'Cambio de correo en mi cuenta', '<p>Me han cambiado el correo de la empresa y quiero seguir entrando con él.</p>', 'en progreso',
        (SELECT id FROM users WHERE email = 'user1@demo.com'), (SELECT id FROM users WHERE email = 'user1@demo.com'), (SELECT id FROM users WHERE email = 'support3@demo.com'),
        NULL, NULL,
        (date_trunc('day', now()) - interval '12 days' + interval '9 hours 0 minutes'), (date_trunc('day', now()) - interval '11 days' + interval '17 hours 0 minutes'), (SELECT id FROM ticket_categories WHERE normalized = 'licencias'));

INSERT INTO ticket_history (ticket_id, actor_id, event, from_state, to_state, detail, created_at)
VALUES ((SELECT id FROM tickets WHERE number = 'CS-2026-0006'), (SELECT id FROM users WHERE email = 'support3@demo.com'), 'estado', 'nuevo', 'en progreso', NULL, (date_trunc('day', now()) - interval '11 days' + interval '10 hours 0 minutes'));
INSERT INTO ticket_history (ticket_id, actor_id, event, from_state, to_state, detail, created_at)
VALUES ((SELECT id FROM tickets WHERE number = 'CS-2026-0006'), (SELECT id FROM users WHERE email = 'user1@demo.com'), 'creado', NULL, 'nuevo', NULL, (date_trunc('day', now()) - interval '12 days' + interval '9 hours 0 minutes'));
INSERT INTO ticket_history (ticket_id, actor_id, event, from_state, to_state, detail, created_at)
VALUES ((SELECT id FROM tickets WHERE number = 'CS-2026-0006'), (SELECT id FROM users WHERE email = 'support3@demo.com'), 'asignado', NULL, NULL, 'Soporte Tres', (date_trunc('day', now()) - interval '12 days' + interval '10 hours 0 minutes'));

-- CS-2026-0007 · La aplicación va lenta por la mañana
INSERT INTO tickets (number, number_year, number_seq, subject, description, state, requester_id, created_by_id, assignee_id, resolved_at, closed_at, created_at, updated_at, category_id)
VALUES ('CS-2026-0007', 2026, 7, 'La aplicación va lenta por la mañana', '<p>Entre las nueve y las diez tarda mucho en abrir la bandeja.</p>', 'escalado',
        (SELECT id FROM users WHERE email = 'user2@demo.com'), (SELECT id FROM users WHERE email = 'user2@demo.com'), (SELECT id FROM users WHERE email = 'support1@demo.com'),
        NULL, NULL,
        (date_trunc('day', now()) - interval '14 days' + interval '9 hours 0 minutes'), (date_trunc('day', now()) - interval '13 days' + interval '17 hours 0 minutes'), (SELECT id FROM ticket_categories WHERE normalized = 'red'));

INSERT INTO ticket_history (ticket_id, actor_id, event, from_state, to_state, detail, created_at)
VALUES ((SELECT id FROM tickets WHERE number = 'CS-2026-0007'), (SELECT id FROM users WHERE email = 'support1@demo.com'), 'escalado', 'en progreso', 'escalado', 'La lentitud es de la base de datos: hace falta Desarrollo.', (date_trunc('day', now()) - interval '12 days' + interval '11 hours 0 minutes'));
INSERT INTO ticket_history (ticket_id, actor_id, event, from_state, to_state, detail, created_at)
VALUES ((SELECT id FROM tickets WHERE number = 'CS-2026-0007'), (SELECT id FROM users WHERE email = 'support1@demo.com'), 'estado', 'nuevo', 'en progreso', NULL, (date_trunc('day', now()) - interval '13 days' + interval '10 hours 0 minutes'));
INSERT INTO ticket_history (ticket_id, actor_id, event, from_state, to_state, detail, created_at)
VALUES ((SELECT id FROM tickets WHERE number = 'CS-2026-0007'), (SELECT id FROM users WHERE email = 'user2@demo.com'), 'creado', NULL, 'nuevo', NULL, (date_trunc('day', now()) - interval '14 days' + interval '9 hours 0 minutes'));
INSERT INTO ticket_history (ticket_id, actor_id, event, from_state, to_state, detail, created_at)
VALUES ((SELECT id FROM tickets WHERE number = 'CS-2026-0007'), (SELECT id FROM users WHERE email = 'support1@demo.com'), 'asignado', NULL, NULL, 'Soporte Uno', (date_trunc('day', now()) - interval '14 days' + interval '10 hours 0 minutes'));

INSERT INTO ticket_comments (ticket_id, author_id, body, created_at)
VALUES ((SELECT id FROM tickets WHERE number = 'CS-2026-0007'), (SELECT id FROM users WHERE email = 'support1@demo.com'), '<p>¿Pasa siempre en el mismo equipo?</p>', (date_trunc('day', now()) - interval '13 days' + interval '11 hours 15 minutes'));
INSERT INTO ticket_comments (ticket_id, author_id, body, created_at)
VALUES ((SELECT id FROM tickets WHERE number = 'CS-2026-0007'), (SELECT id FROM users WHERE email = 'support2@demo.com'), '<p>Sí, en el mío, y en el de al lado no.</p>', (date_trunc('day', now()) - interval '12 days' + interval '12 hours 15 minutes'));

-- El archivo de este adjunto lo copia `scripts/dev-seed.sh` desde `config/seed/captura.png`.
INSERT INTO ticket_attachments (ticket_id, uploaded_by_id, filename, stored_name, content_type, size_bytes, created_at)
VALUES ((SELECT id FROM tickets WHERE number = 'CS-2026-0007'), (SELECT id FROM users WHERE email = 'user2@demo.com'), 'captura.png', '2026/CS-2026-0007/c9385c11ec0efd571d5faa7118635a20.png', 'image/png', 1024, (date_trunc('day', now()) - interval '14 days' + interval '9 hours 30 minutes'));

-- INT-CS-2026-0007: el ticket interno que cuelga de CS-2026-0007
INSERT INTO internal_tickets (ticket_id, number, state, escalation_reason, created_by_id, assignee_id, resolved_at, closed_at, created_at, updated_at)
VALUES ((SELECT id FROM tickets WHERE number = 'CS-2026-0007'), 'INT-CS-2026-0007', 'en progreso', 'La lentitud es de la base de datos: hace falta Desarrollo.', (SELECT id FROM users WHERE email = 'support1@demo.com'), (SELECT id FROM users WHERE email = 'dev1@demo.com'),
        NULL, NULL,
        (date_trunc('day', now()) - interval '12 days' + interval '11 hours 0 minutes'), (date_trunc('day', now()) - interval '11 days' + interval '16 hours 0 minutes'));

INSERT INTO ticket_history (internal_ticket_id, actor_id, event, from_state, to_state, detail, created_at)
VALUES ((SELECT id FROM internal_tickets WHERE number = 'INT-CS-2026-0007'), (SELECT id FROM users WHERE email = 'support1@demo.com'), 'creado', NULL, 'nuevo', 'La lentitud es de la base de datos: hace falta Desarrollo.', (date_trunc('day', now()) - interval '12 days' + interval '11 hours 0 minutes'));
INSERT INTO ticket_history (internal_ticket_id, actor_id, event, from_state, to_state, detail, created_at)
VALUES ((SELECT id FROM internal_tickets WHERE number = 'INT-CS-2026-0007'), (SELECT id FROM users WHERE email = 'dev1@demo.com'), 'asignado', NULL, NULL, 'Desarrollo Uno', (date_trunc('day', now()) - interval '12 days' + interval '12 hours 0 minutes'));
INSERT INTO ticket_comments (internal_ticket_id, author_id, body, created_at)
VALUES ((SELECT id FROM internal_tickets WHERE number = 'INT-CS-2026-0007'), (SELECT id FROM users WHERE email = 'support1@demo.com'), '<p>Te lo paso con todo lo que he probado hasta ahora.</p>', (date_trunc('day', now()) - interval '12 days' + interval '12 hours 20 minutes'));
INSERT INTO ticket_comments (internal_ticket_id, author_id, body, created_at)
VALUES ((SELECT id FROM internal_tickets WHERE number = 'INT-CS-2026-0007'), (SELECT id FROM users WHERE email = 'dev1@demo.com'), '<p>Lo miro hoy y te digo algo.</p>', (date_trunc('day', now()) - interval '11 days' + interval '10 hours 0 minutes'));

-- CS-2026-0008 · No me llegan los avisos de ticket nuevo
INSERT INTO tickets (number, number_year, number_seq, subject, description, state, requester_id, created_by_id, assignee_id, resolved_at, closed_at, created_at, updated_at, category_id)
VALUES ('CS-2026-0008', 2026, 8, 'No me llegan los avisos de ticket nuevo', '<p>Me asignan tickets y no recibo el correo de aviso.</p>', 'en progreso',
        (SELECT id FROM users WHERE email = 'user3@demo.com'), (SELECT id FROM users WHERE email = 'user3@demo.com'), (SELECT id FROM users WHERE email = 'support2@demo.com'),
        NULL, NULL,
        (date_trunc('day', now()) - interval '17 days' + interval '9 hours 0 minutes'), (date_trunc('day', now()) - interval '16 days' + interval '17 hours 0 minutes'), (SELECT id FROM ticket_categories WHERE normalized = 'software'));

INSERT INTO ticket_history (ticket_id, actor_id, event, from_state, to_state, detail, created_at)
VALUES ((SELECT id FROM tickets WHERE number = 'CS-2026-0008'), (SELECT id FROM users WHERE email = 'support2@demo.com'), 'estado', 'nuevo', 'en progreso', NULL, (date_trunc('day', now()) - interval '16 days' + interval '10 hours 0 minutes'));
INSERT INTO ticket_history (ticket_id, actor_id, event, from_state, to_state, detail, created_at)
VALUES ((SELECT id FROM tickets WHERE number = 'CS-2026-0008'), (SELECT id FROM users WHERE email = 'user3@demo.com'), 'creado', NULL, 'nuevo', NULL, (date_trunc('day', now()) - interval '17 days' + interval '9 hours 0 minutes'));
INSERT INTO ticket_history (ticket_id, actor_id, event, from_state, to_state, detail, created_at)
VALUES ((SELECT id FROM tickets WHERE number = 'CS-2026-0008'), (SELECT id FROM users WHERE email = 'support2@demo.com'), 'asignado', NULL, NULL, 'Soporte Dos', (date_trunc('day', now()) - interval '17 days' + interval '10 hours 0 minutes'));

-- CS-2026-0009 · Falta el teléfono de contacto en la ficha
INSERT INTO tickets (number, number_year, number_seq, subject, description, state, requester_id, created_by_id, assignee_id, resolved_at, closed_at, created_at, updated_at, category_id)
VALUES ('CS-2026-0009', 2026, 9, 'Falta el teléfono de contacto en la ficha', '<p>Necesitamos el teléfono en la ficha del cliente para llamarle.</p>', 'escalado',
        (SELECT id FROM users WHERE email = 'user4@demo.com'), (SELECT id FROM users WHERE email = 'user4@demo.com'), (SELECT id FROM users WHERE email = 'support1@demo.com'),
        NULL, NULL,
        (date_trunc('day', now()) - interval '20 days' + interval '9 hours 0 minutes'), (date_trunc('day', now()) - interval '19 days' + interval '17 hours 0 minutes'), (SELECT id FROM ticket_categories WHERE normalized = 'software'));

INSERT INTO ticket_history (ticket_id, actor_id, event, from_state, to_state, detail, created_at)
VALUES ((SELECT id FROM tickets WHERE number = 'CS-2026-0009'), (SELECT id FROM users WHERE email = 'support1@demo.com'), 'escalado', 'en progreso', 'escalado', 'Desarrollo tiene que decir si el campo se puede añadir sin tocar la importación.', (date_trunc('day', now()) - interval '18 days' + interval '11 hours 0 minutes'));
INSERT INTO ticket_history (ticket_id, actor_id, event, from_state, to_state, detail, created_at)
VALUES ((SELECT id FROM tickets WHERE number = 'CS-2026-0009'), (SELECT id FROM users WHERE email = 'support1@demo.com'), 'estado', 'nuevo', 'en progreso', NULL, (date_trunc('day', now()) - interval '19 days' + interval '10 hours 0 minutes'));
INSERT INTO ticket_history (ticket_id, actor_id, event, from_state, to_state, detail, created_at)
VALUES ((SELECT id FROM tickets WHERE number = 'CS-2026-0009'), (SELECT id FROM users WHERE email = 'user4@demo.com'), 'creado', NULL, 'nuevo', NULL, (date_trunc('day', now()) - interval '20 days' + interval '9 hours 0 minutes'));
INSERT INTO ticket_history (ticket_id, actor_id, event, from_state, to_state, detail, created_at)
VALUES ((SELECT id FROM tickets WHERE number = 'CS-2026-0009'), (SELECT id FROM users WHERE email = 'support1@demo.com'), 'asignado', NULL, NULL, 'Soporte Uno', (date_trunc('day', now()) - interval '20 days' + interval '10 hours 0 minutes'));

INSERT INTO ticket_comments (ticket_id, author_id, body, created_at)
VALUES ((SELECT id FROM tickets WHERE number = 'CS-2026-0009'), (SELECT id FROM users WHERE email = 'support1@demo.com'), '<p>Se lo he pasado a Desarrollo. Te digo algo en cuanto sepa.</p>', (date_trunc('day', now()) - interval '19 days' + interval '11 hours 15 minutes'));

-- INT-CS-2026-0009: el ticket interno que cuelga de CS-2026-0009
INSERT INTO internal_tickets (ticket_id, number, state, escalation_reason, created_by_id, assignee_id, resolved_at, closed_at, created_at, updated_at)
VALUES ((SELECT id FROM tickets WHERE number = 'CS-2026-0009'), 'INT-CS-2026-0009', 'en espera', 'Desarrollo tiene que decir si el campo se puede añadir sin tocar la importación.', (SELECT id FROM users WHERE email = 'support1@demo.com'), (SELECT id FROM users WHERE email = 'dev2@demo.com'),
        NULL, NULL,
        (date_trunc('day', now()) - interval '18 days' + interval '11 hours 0 minutes'), (date_trunc('day', now()) - interval '17 days' + interval '16 hours 0 minutes'));

INSERT INTO ticket_history (internal_ticket_id, actor_id, event, from_state, to_state, detail, created_at)
VALUES ((SELECT id FROM internal_tickets WHERE number = 'INT-CS-2026-0009'), (SELECT id FROM users WHERE email = 'support1@demo.com'), 'creado', NULL, 'nuevo', 'Desarrollo tiene que decir si el campo se puede añadir sin tocar la importación.', (date_trunc('day', now()) - interval '18 days' + interval '11 hours 0 minutes'));
INSERT INTO ticket_history (internal_ticket_id, actor_id, event, from_state, to_state, detail, created_at)
VALUES ((SELECT id FROM internal_tickets WHERE number = 'INT-CS-2026-0009'), (SELECT id FROM users WHERE email = 'dev2@demo.com'), 'asignado', NULL, NULL, 'Desarrollo Dos', (date_trunc('day', now()) - interval '18 days' + interval '12 hours 0 minutes'));
INSERT INTO ticket_comments (internal_ticket_id, author_id, body, created_at)
VALUES ((SELECT id FROM internal_tickets WHERE number = 'INT-CS-2026-0009'), (SELECT id FROM users WHERE email = 'support1@demo.com'), '<p>Te lo paso con todo lo que he probado hasta ahora.</p>', (date_trunc('day', now()) - interval '18 days' + interval '12 hours 20 minutes'));
INSERT INTO ticket_comments (internal_ticket_id, author_id, body, created_at)
VALUES ((SELECT id FROM internal_tickets WHERE number = 'INT-CS-2026-0009'), (SELECT id FROM users WHERE email = 'dev2@demo.com'), '<p>Lo miro hoy y te digo algo.</p>', (date_trunc('day', now()) - interval '17 days' + interval '10 hours 0 minutes'));

-- CS-2026-0010 · Se ha borrado un comentario
INSERT INTO tickets (number, number_year, number_seq, subject, description, state, requester_id, created_by_id, assignee_id, resolved_at, closed_at, created_at, updated_at, category_id)
VALUES ('CS-2026-0010', 2026, 10, 'Se ha borrado un comentario', '<p>Ayer escribí un comentario en un ticket y hoy ya no está.</p>', 'en espera',
        (SELECT id FROM users WHERE email = 'user5@demo.com'), (SELECT id FROM users WHERE email = 'user5@demo.com'), (SELECT id FROM users WHERE email = 'support3@demo.com'),
        NULL, NULL,
        (date_trunc('day', now()) - interval '22 days' + interval '9 hours 0 minutes'), (date_trunc('day', now()) - interval '21 days' + interval '17 hours 0 minutes'), (SELECT id FROM ticket_categories WHERE normalized = 'general'));

INSERT INTO ticket_history (ticket_id, actor_id, event, from_state, to_state, detail, created_at)
VALUES ((SELECT id FROM tickets WHERE number = 'CS-2026-0010'), (SELECT id FROM users WHERE email = 'support3@demo.com'), 'estado', 'en progreso', 'en espera', NULL, (date_trunc('day', now()) - interval '20 days' + interval '12 hours 0 minutes'));
INSERT INTO ticket_history (ticket_id, actor_id, event, from_state, to_state, detail, created_at)
VALUES ((SELECT id FROM tickets WHERE number = 'CS-2026-0010'), (SELECT id FROM users WHERE email = 'support3@demo.com'), 'estado', 'nuevo', 'en progreso', NULL, (date_trunc('day', now()) - interval '21 days' + interval '10 hours 0 minutes'));
INSERT INTO ticket_history (ticket_id, actor_id, event, from_state, to_state, detail, created_at)
VALUES ((SELECT id FROM tickets WHERE number = 'CS-2026-0010'), (SELECT id FROM users WHERE email = 'user5@demo.com'), 'creado', NULL, 'nuevo', NULL, (date_trunc('day', now()) - interval '22 days' + interval '9 hours 0 minutes'));
INSERT INTO ticket_history (ticket_id, actor_id, event, from_state, to_state, detail, created_at)
VALUES ((SELECT id FROM tickets WHERE number = 'CS-2026-0010'), (SELECT id FROM users WHERE email = 'support3@demo.com'), 'asignado', NULL, NULL, 'Soporte Tres', (date_trunc('day', now()) - interval '22 days' + interval '10 hours 0 minutes'));

INSERT INTO ticket_comments (ticket_id, author_id, body, created_at)
VALUES ((SELECT id FROM tickets WHERE number = 'CS-2026-0010'), (SELECT id FROM users WHERE email = 'support3@demo.com'), '<p>¿Recuerdas en qué ticket era? Vamos a mirarlo.</p>', (date_trunc('day', now()) - interval '21 days' + interval '11 hours 15 minutes'));

-- CS-2026-0011 · Quiero cambiar el nombre de la instalación
INSERT INTO tickets (number, number_year, number_seq, subject, description, state, requester_id, created_by_id, assignee_id, resolved_at, closed_at, created_at, updated_at, category_id)
VALUES ('CS-2026-0011', 2026, 11, 'Quiero cambiar el nombre de la instalación', '<p>En la cabecera pone el nombre de la aplicación y debería poner el nuestro.</p>', 'en espera',
        (SELECT id FROM users WHERE email = 'user1@demo.com'), (SELECT id FROM users WHERE email = 'user1@demo.com'), (SELECT id FROM users WHERE email = 'support2@demo.com'),
        NULL, NULL,
        (date_trunc('day', now()) - interval '25 days' + interval '9 hours 0 minutes'), (date_trunc('day', now()) - interval '24 days' + interval '17 hours 0 minutes'), (SELECT id FROM ticket_categories WHERE normalized = 'software'));

INSERT INTO ticket_history (ticket_id, actor_id, event, from_state, to_state, detail, created_at)
VALUES ((SELECT id FROM tickets WHERE number = 'CS-2026-0011'), (SELECT id FROM users WHERE email = 'support2@demo.com'), 'estado', 'en progreso', 'en espera', NULL, (date_trunc('day', now()) - interval '23 days' + interval '12 hours 0 minutes'));
INSERT INTO ticket_history (ticket_id, actor_id, event, from_state, to_state, detail, created_at)
VALUES ((SELECT id FROM tickets WHERE number = 'CS-2026-0011'), (SELECT id FROM users WHERE email = 'support2@demo.com'), 'estado', 'nuevo', 'en progreso', NULL, (date_trunc('day', now()) - interval '24 days' + interval '10 hours 0 minutes'));
INSERT INTO ticket_history (ticket_id, actor_id, event, from_state, to_state, detail, created_at)
VALUES ((SELECT id FROM tickets WHERE number = 'CS-2026-0011'), (SELECT id FROM users WHERE email = 'user1@demo.com'), 'creado', NULL, 'nuevo', NULL, (date_trunc('day', now()) - interval '25 days' + interval '9 hours 0 minutes'));
INSERT INTO ticket_history (ticket_id, actor_id, event, from_state, to_state, detail, created_at)
VALUES ((SELECT id FROM tickets WHERE number = 'CS-2026-0011'), (SELECT id FROM users WHERE email = 'support2@demo.com'), 'asignado', NULL, NULL, 'Soporte Dos', (date_trunc('day', now()) - interval '25 days' + interval '10 hours 0 minutes'));

-- CS-2026-0012 · El buscador no encuentra por número
INSERT INTO tickets (number, number_year, number_seq, subject, description, state, requester_id, created_by_id, assignee_id, resolved_at, closed_at, created_at, updated_at, category_id)
VALUES ('CS-2026-0012', 2026, 12, 'El buscador no encuentra por número', '<p>Si pego el número del ticket no sale nada; buscando por texto sí aparece.</p>', 'en progreso',
        (SELECT id FROM users WHERE email = 'user2@demo.com'), (SELECT id FROM users WHERE email = 'user2@demo.com'), (SELECT id FROM users WHERE email = 'support1@demo.com'),
        NULL, NULL,
        (date_trunc('day', now()) - interval '28 days' + interval '9 hours 0 minutes'), (date_trunc('day', now()) - interval '27 days' + interval '17 hours 0 minutes'), (SELECT id FROM ticket_categories WHERE normalized = 'software'));

INSERT INTO ticket_history (ticket_id, actor_id, event, from_state, to_state, detail, created_at)
VALUES ((SELECT id FROM tickets WHERE number = 'CS-2026-0012'), NULL, 'estado', 'escalado', 'en progreso', 'Desarrollo ha resuelto el interno: el ticket vuelve a Soporte.', (date_trunc('day', now()) - interval '24 days' + interval '12 hours 0 minutes'));
INSERT INTO ticket_history (ticket_id, actor_id, event, from_state, to_state, detail, created_at)
VALUES ((SELECT id FROM tickets WHERE number = 'CS-2026-0012'), (SELECT id FROM users WHERE email = 'support1@demo.com'), 'escalado', 'en progreso', 'escalado', 'La búsqueda por número necesita un cambio en la consulta.', (date_trunc('day', now()) - interval '26 days' + interval '11 hours 0 minutes'));
INSERT INTO ticket_history (ticket_id, actor_id, event, from_state, to_state, detail, created_at)
VALUES ((SELECT id FROM tickets WHERE number = 'CS-2026-0012'), (SELECT id FROM users WHERE email = 'support1@demo.com'), 'estado', 'nuevo', 'en progreso', NULL, (date_trunc('day', now()) - interval '27 days' + interval '10 hours 0 minutes'));
INSERT INTO ticket_history (ticket_id, actor_id, event, from_state, to_state, detail, created_at)
VALUES ((SELECT id FROM tickets WHERE number = 'CS-2026-0012'), (SELECT id FROM users WHERE email = 'user2@demo.com'), 'creado', NULL, 'nuevo', NULL, (date_trunc('day', now()) - interval '28 days' + interval '9 hours 0 minutes'));
INSERT INTO ticket_history (ticket_id, actor_id, event, from_state, to_state, detail, created_at)
VALUES ((SELECT id FROM tickets WHERE number = 'CS-2026-0012'), (SELECT id FROM users WHERE email = 'support1@demo.com'), 'asignado', NULL, NULL, 'Soporte Uno', (date_trunc('day', now()) - interval '28 days' + interval '10 hours 0 minutes'));

INSERT INTO ticket_comments (ticket_id, author_id, body, created_at)
VALUES ((SELECT id FROM tickets WHERE number = 'CS-2026-0012'), (SELECT id FROM users WHERE email = 'support2@demo.com'), '<p>Lo he probado con el 4 y tampoco sale.</p>', (date_trunc('day', now()) - interval '27 days' + interval '11 hours 15 minutes'));

-- INT-CS-2026-0012: el ticket interno que cuelga de CS-2026-0012
INSERT INTO internal_tickets (ticket_id, number, state, escalation_reason, created_by_id, assignee_id, resolved_at, closed_at, created_at, updated_at)
VALUES ((SELECT id FROM tickets WHERE number = 'CS-2026-0012'), 'INT-CS-2026-0012', 'resuelto', 'La búsqueda por número necesita un cambio en la consulta.', (SELECT id FROM users WHERE email = 'support1@demo.com'), (SELECT id FROM users WHERE email = 'dev1@demo.com'),
        (date_trunc('day', now()) - interval '24 days' + interval '12 hours 0 minutes'), NULL,
        (date_trunc('day', now()) - interval '26 days' + interval '11 hours 0 minutes'), (date_trunc('day', now()) - interval '25 days' + interval '16 hours 0 minutes'));

INSERT INTO ticket_history (internal_ticket_id, actor_id, event, from_state, to_state, detail, created_at)
VALUES ((SELECT id FROM internal_tickets WHERE number = 'INT-CS-2026-0012'), (SELECT id FROM users WHERE email = 'support1@demo.com'), 'creado', NULL, 'nuevo', 'La búsqueda por número necesita un cambio en la consulta.', (date_trunc('day', now()) - interval '26 days' + interval '11 hours 0 minutes'));
INSERT INTO ticket_history (internal_ticket_id, actor_id, event, from_state, to_state, detail, created_at)
VALUES ((SELECT id FROM internal_tickets WHERE number = 'INT-CS-2026-0012'), (SELECT id FROM users WHERE email = 'dev1@demo.com'), 'asignado', NULL, NULL, 'Desarrollo Uno', (date_trunc('day', now()) - interval '26 days' + interval '12 hours 0 minutes'));
INSERT INTO ticket_history (internal_ticket_id, actor_id, event, from_state, to_state, detail, created_at)
VALUES ((SELECT id FROM internal_tickets WHERE number = 'INT-CS-2026-0012'), (SELECT id FROM users WHERE email = 'dev1@demo.com'), 'resuelto', 'en progreso', 'resuelto', NULL, (date_trunc('day', now()) - interval '24 days' + interval '12 hours 0 minutes'));
INSERT INTO ticket_comments (internal_ticket_id, author_id, body, created_at)
VALUES ((SELECT id FROM internal_tickets WHERE number = 'INT-CS-2026-0012'), (SELECT id FROM users WHERE email = 'support1@demo.com'), '<p>Te lo paso con todo lo que he probado hasta ahora.</p>', (date_trunc('day', now()) - interval '26 days' + interval '12 hours 20 minutes'));
INSERT INTO ticket_comments (internal_ticket_id, author_id, body, created_at)
VALUES ((SELECT id FROM internal_tickets WHERE number = 'INT-CS-2026-0012'), (SELECT id FROM users WHERE email = 'dev1@demo.com'), '<p>Lo miro hoy y te digo algo.</p>', (date_trunc('day', now()) - interval '25 days' + interval '10 hours 0 minutes'));

-- CS-2026-0013 · Pedir acceso a la carpeta compartida
INSERT INTO tickets (number, number_year, number_seq, subject, description, state, requester_id, created_by_id, assignee_id, resolved_at, closed_at, created_at, updated_at, category_id)
VALUES ('CS-2026-0013', 2026, 13, 'Pedir acceso a la carpeta compartida', '<p>Necesito permiso de escritura en la carpeta de trabajo del equipo.</p>', 'resuelto',
        (SELECT id FROM users WHERE email = 'user3@demo.com'), (SELECT id FROM users WHERE email = 'user3@demo.com'), (SELECT id FROM users WHERE email = 'support1@demo.com'),
        (date_trunc('day', now()) - interval '29 days' + interval '13 hours 0 minutes'), NULL,
        (date_trunc('day', now()) - interval '32 days' + interval '9 hours 0 minutes'), (date_trunc('day', now()) - interval '31 days' + interval '17 hours 0 minutes'), (SELECT id FROM ticket_categories WHERE normalized = 'red'));

INSERT INTO ticket_history (ticket_id, actor_id, event, from_state, to_state, detail, created_at)
VALUES ((SELECT id FROM tickets WHERE number = 'CS-2026-0013'), (SELECT id FROM users WHERE email = 'support1@demo.com'), 'resuelto', 'en progreso', 'resuelto', NULL, (date_trunc('day', now()) - interval '29 days' + interval '13 hours 0 minutes'));
INSERT INTO ticket_history (ticket_id, actor_id, event, from_state, to_state, detail, created_at)
VALUES ((SELECT id FROM tickets WHERE number = 'CS-2026-0013'), (SELECT id FROM users WHERE email = 'support1@demo.com'), 'estado', 'nuevo', 'en progreso', NULL, (date_trunc('day', now()) - interval '31 days' + interval '10 hours 0 minutes'));
INSERT INTO ticket_history (ticket_id, actor_id, event, from_state, to_state, detail, created_at)
VALUES ((SELECT id FROM tickets WHERE number = 'CS-2026-0013'), (SELECT id FROM users WHERE email = 'user3@demo.com'), 'creado', NULL, 'nuevo', NULL, (date_trunc('day', now()) - interval '32 days' + interval '9 hours 0 minutes'));
INSERT INTO ticket_history (ticket_id, actor_id, event, from_state, to_state, detail, created_at)
VALUES ((SELECT id FROM tickets WHERE number = 'CS-2026-0013'), (SELECT id FROM users WHERE email = 'support1@demo.com'), 'asignado', NULL, NULL, 'Soporte Uno', (date_trunc('day', now()) - interval '32 days' + interval '10 hours 0 minutes'));

INSERT INTO ticket_comments (ticket_id, author_id, body, created_at)
VALUES ((SELECT id FROM tickets WHERE number = 'CS-2026-0013'), (SELECT id FROM users WHERE email = 'support3@demo.com'), '<p>Ya puedo escribir, muchas gracias.</p>', (date_trunc('day', now()) - interval '31 days' + interval '11 hours 15 minutes'));

-- CS-2026-0014 · El correo de alta llegó a spam
INSERT INTO tickets (number, number_year, number_seq, subject, description, state, requester_id, created_by_id, assignee_id, resolved_at, closed_at, created_at, updated_at, category_id)
VALUES ('CS-2026-0014', 2026, 14, 'El correo de alta llegó a spam', '<p>Aviso por si le pasa a más gente: el correo de alta me llegó a spam.</p>', 'resuelto',
        (SELECT id FROM users WHERE email = 'user4@demo.com'), (SELECT id FROM users WHERE email = 'user4@demo.com'), (SELECT id FROM users WHERE email = 'support2@demo.com'),
        (date_trunc('day', now()) - interval '32 days' + interval '13 hours 0 minutes'), NULL,
        (date_trunc('day', now()) - interval '35 days' + interval '9 hours 0 minutes'), (date_trunc('day', now()) - interval '34 days' + interval '17 hours 0 minutes'), (SELECT id FROM ticket_categories WHERE normalized = 'general'));

INSERT INTO ticket_history (ticket_id, actor_id, event, from_state, to_state, detail, created_at)
VALUES ((SELECT id FROM tickets WHERE number = 'CS-2026-0014'), (SELECT id FROM users WHERE email = 'support2@demo.com'), 'resuelto', 'en progreso', 'resuelto', NULL, (date_trunc('day', now()) - interval '32 days' + interval '13 hours 0 minutes'));
INSERT INTO ticket_history (ticket_id, actor_id, event, from_state, to_state, detail, created_at)
VALUES ((SELECT id FROM tickets WHERE number = 'CS-2026-0014'), (SELECT id FROM users WHERE email = 'support2@demo.com'), 'estado', 'nuevo', 'en progreso', NULL, (date_trunc('day', now()) - interval '34 days' + interval '10 hours 0 minutes'));
INSERT INTO ticket_history (ticket_id, actor_id, event, from_state, to_state, detail, created_at)
VALUES ((SELECT id FROM tickets WHERE number = 'CS-2026-0014'), (SELECT id FROM users WHERE email = 'user4@demo.com'), 'creado', NULL, 'nuevo', NULL, (date_trunc('day', now()) - interval '35 days' + interval '9 hours 0 minutes'));
INSERT INTO ticket_history (ticket_id, actor_id, event, from_state, to_state, detail, created_at)
VALUES ((SELECT id FROM tickets WHERE number = 'CS-2026-0014'), (SELECT id FROM users WHERE email = 'support2@demo.com'), 'asignado', NULL, NULL, 'Soporte Dos', (date_trunc('day', now()) - interval '35 days' + interval '10 hours 0 minutes'));

-- CS-2026-0015 · Duplicado: la impresora de la segunda planta
INSERT INTO tickets (number, number_year, number_seq, subject, description, state, requester_id, created_by_id, assignee_id, resolved_at, closed_at, created_at, updated_at, category_id)
VALUES ('CS-2026-0015', 2026, 15, 'Duplicado: la impresora de la segunda planta', '<p>Abro esto por error, ya había un ticket de la impresora.</p>', 'resuelto',
        (SELECT id FROM users WHERE email = 'user5@demo.com'), (SELECT id FROM users WHERE email = 'user5@demo.com'), (SELECT id FROM users WHERE email = 'support3@demo.com'),
        (date_trunc('day', now()) - interval '35 days' + interval '13 hours 0 minutes'), NULL,
        (date_trunc('day', now()) - interval '38 days' + interval '9 hours 0 minutes'), (date_trunc('day', now()) - interval '37 days' + interval '17 hours 0 minutes'), (SELECT id FROM ticket_categories WHERE normalized = 'impresoras'));

INSERT INTO ticket_history (ticket_id, actor_id, event, from_state, to_state, detail, created_at)
VALUES ((SELECT id FROM tickets WHERE number = 'CS-2026-0015'), (SELECT id FROM users WHERE email = 'support3@demo.com'), 'resuelto', 'en progreso', 'resuelto', NULL, (date_trunc('day', now()) - interval '35 days' + interval '13 hours 0 minutes'));
INSERT INTO ticket_history (ticket_id, actor_id, event, from_state, to_state, detail, created_at)
VALUES ((SELECT id FROM tickets WHERE number = 'CS-2026-0015'), (SELECT id FROM users WHERE email = 'support3@demo.com'), 'estado', 'nuevo', 'en progreso', NULL, (date_trunc('day', now()) - interval '37 days' + interval '10 hours 0 minutes'));
INSERT INTO ticket_history (ticket_id, actor_id, event, from_state, to_state, detail, created_at)
VALUES ((SELECT id FROM tickets WHERE number = 'CS-2026-0015'), (SELECT id FROM users WHERE email = 'user5@demo.com'), 'creado', NULL, 'nuevo', NULL, (date_trunc('day', now()) - interval '38 days' + interval '9 hours 0 minutes'));
INSERT INTO ticket_history (ticket_id, actor_id, event, from_state, to_state, detail, created_at)
VALUES ((SELECT id FROM tickets WHERE number = 'CS-2026-0015'), (SELECT id FROM users WHERE email = 'support3@demo.com'), 'asignado', NULL, NULL, 'Soporte Tres', (date_trunc('day', now()) - interval '38 days' + interval '10 hours 0 minutes'));

INSERT INTO ticket_comments (ticket_id, author_id, body, created_at)
VALUES ((SELECT id FROM tickets WHERE number = 'CS-2026-0015'), (SELECT id FROM users WHERE email = 'support3@demo.com'), '<p>Es el mismo, sí: seguimos en el otro.</p>', (date_trunc('day', now()) - interval '37 days' + interval '11 hours 15 minutes'));

-- CS-2026-0016 · La contraseña caducaba y no lo sabía
INSERT INTO tickets (number, number_year, number_seq, subject, description, state, requester_id, created_by_id, assignee_id, resolved_at, closed_at, created_at, updated_at, category_id)
VALUES ('CS-2026-0016', 2026, 16, 'La contraseña caducaba y no lo sabía', '<p>Debería avisar antes de que caduque, no al intentar entrar.</p>', 'resuelto',
        (SELECT id FROM users WHERE email = 'user1@demo.com'), (SELECT id FROM users WHERE email = 'user1@demo.com'), (SELECT id FROM users WHERE email = 'support1@demo.com'),
        (date_trunc('day', now()) - interval '38 days' + interval '13 hours 0 minutes'), NULL,
        (date_trunc('day', now()) - interval '41 days' + interval '9 hours 0 minutes'), (date_trunc('day', now()) - interval '40 days' + interval '17 hours 0 minutes'), (SELECT id FROM ticket_categories WHERE normalized = 'software'));

INSERT INTO ticket_history (ticket_id, actor_id, event, from_state, to_state, detail, created_at)
VALUES ((SELECT id FROM tickets WHERE number = 'CS-2026-0016'), NULL, 'estado', 'escalado', 'en progreso', 'Desarrollo ha devuelto el ticket a Soporte.', (date_trunc('day', now()) - interval '37 days' + interval '12 hours 0 minutes'));
INSERT INTO ticket_history (ticket_id, actor_id, event, from_state, to_state, detail, created_at)
VALUES ((SELECT id FROM tickets WHERE number = 'CS-2026-0016'), (SELECT id FROM users WHERE email = 'support1@demo.com'), 'resuelto', 'en progreso', 'resuelto', NULL, (date_trunc('day', now()) - interval '38 days' + interval '13 hours 0 minutes'));
INSERT INTO ticket_history (ticket_id, actor_id, event, from_state, to_state, detail, created_at)
VALUES ((SELECT id FROM tickets WHERE number = 'CS-2026-0016'), (SELECT id FROM users WHERE email = 'support1@demo.com'), 'escalado', 'en progreso', 'escalado', 'El aviso de caducidad no existe: hay que añadirlo al módulo de acceso.', (date_trunc('day', now()) - interval '39 days' + interval '11 hours 0 minutes'));
INSERT INTO ticket_history (ticket_id, actor_id, event, from_state, to_state, detail, created_at)
VALUES ((SELECT id FROM tickets WHERE number = 'CS-2026-0016'), (SELECT id FROM users WHERE email = 'support1@demo.com'), 'estado', 'nuevo', 'en progreso', NULL, (date_trunc('day', now()) - interval '40 days' + interval '10 hours 0 minutes'));
INSERT INTO ticket_history (ticket_id, actor_id, event, from_state, to_state, detail, created_at)
VALUES ((SELECT id FROM tickets WHERE number = 'CS-2026-0016'), (SELECT id FROM users WHERE email = 'user1@demo.com'), 'creado', NULL, 'nuevo', NULL, (date_trunc('day', now()) - interval '41 days' + interval '9 hours 0 minutes'));
INSERT INTO ticket_history (ticket_id, actor_id, event, from_state, to_state, detail, created_at)
VALUES ((SELECT id FROM tickets WHERE number = 'CS-2026-0016'), (SELECT id FROM users WHERE email = 'support1@demo.com'), 'asignado', NULL, NULL, 'Soporte Uno', (date_trunc('day', now()) - interval '41 days' + interval '10 hours 0 minutes'));

-- El archivo de este adjunto lo copia `scripts/dev-seed.sh` desde `config/seed/informe.pdf`.
INSERT INTO ticket_attachments (ticket_id, uploaded_by_id, filename, stored_name, content_type, size_bytes, created_at)
VALUES ((SELECT id FROM tickets WHERE number = 'CS-2026-0016'), (SELECT id FROM users WHERE email = 'user1@demo.com'), 'informe.pdf', '2026/CS-2026-0016/748c64619c90ab6963f56c425ef9150b.pdf', 'application/pdf', 1024, (date_trunc('day', now()) - interval '41 days' + interval '9 hours 30 minutes'));

-- INT-CS-2026-0016: el ticket interno que cuelga de CS-2026-0016
INSERT INTO internal_tickets (ticket_id, number, state, escalation_reason, created_by_id, assignee_id, resolved_at, closed_at, created_at, updated_at)
VALUES ((SELECT id FROM tickets WHERE number = 'CS-2026-0016'), 'INT-CS-2026-0016', 'cerrado', 'El aviso de caducidad no existe: hay que añadirlo al módulo de acceso.', (SELECT id FROM users WHERE email = 'support1@demo.com'), (SELECT id FROM users WHERE email = 'dev3@demo.com'),
        (date_trunc('day', now()) - interval '37 days' + interval '12 hours 0 minutes'), (date_trunc('day', now()) - interval '36 days' + interval '15 hours 0 minutes'),
        (date_trunc('day', now()) - interval '39 days' + interval '11 hours 0 minutes'), (date_trunc('day', now()) - interval '38 days' + interval '16 hours 0 minutes'));

INSERT INTO ticket_history (internal_ticket_id, actor_id, event, from_state, to_state, detail, created_at)
VALUES ((SELECT id FROM internal_tickets WHERE number = 'INT-CS-2026-0016'), (SELECT id FROM users WHERE email = 'support1@demo.com'), 'creado', NULL, 'nuevo', 'El aviso de caducidad no existe: hay que añadirlo al módulo de acceso.', (date_trunc('day', now()) - interval '39 days' + interval '11 hours 0 minutes'));
INSERT INTO ticket_history (internal_ticket_id, actor_id, event, from_state, to_state, detail, created_at)
VALUES ((SELECT id FROM internal_tickets WHERE number = 'INT-CS-2026-0016'), (SELECT id FROM users WHERE email = 'dev3@demo.com'), 'asignado', NULL, NULL, 'Desarrollo Tres', (date_trunc('day', now()) - interval '39 days' + interval '12 hours 0 minutes'));
INSERT INTO ticket_history (internal_ticket_id, actor_id, event, from_state, to_state, detail, created_at)
VALUES ((SELECT id FROM internal_tickets WHERE number = 'INT-CS-2026-0016'), (SELECT id FROM users WHERE email = 'dev3@demo.com'), 'resuelto', 'en progreso', 'resuelto', NULL, (date_trunc('day', now()) - interval '37 days' + interval '12 hours 0 minutes'));
INSERT INTO ticket_history (internal_ticket_id, actor_id, event, from_state, to_state, detail, created_at)
VALUES ((SELECT id FROM internal_tickets WHERE number = 'INT-CS-2026-0016'), (SELECT id FROM users WHERE email = 'support1@demo.com'), 'cerrado', 'resuelto', 'cerrado', NULL, (date_trunc('day', now()) - interval '36 days' + interval '15 hours 0 minutes'));
INSERT INTO ticket_comments (internal_ticket_id, author_id, body, created_at)
VALUES ((SELECT id FROM internal_tickets WHERE number = 'INT-CS-2026-0016'), (SELECT id FROM users WHERE email = 'support1@demo.com'), '<p>Te lo paso con todo lo que he probado hasta ahora.</p>', (date_trunc('day', now()) - interval '39 days' + interval '12 hours 20 minutes'));
INSERT INTO ticket_comments (internal_ticket_id, author_id, body, created_at)
VALUES ((SELECT id FROM internal_tickets WHERE number = 'INT-CS-2026-0016'), (SELECT id FROM users WHERE email = 'dev3@demo.com'), '<p>Lo miro hoy y te digo algo.</p>', (date_trunc('day', now()) - interval '38 days' + interval '10 hours 0 minutes'));

-- CS-2026-0017 · Quiero ordenar la bandeja por fecha
INSERT INTO tickets (number, number_year, number_seq, subject, description, state, requester_id, created_by_id, assignee_id, resolved_at, closed_at, created_at, updated_at, category_id)
VALUES ('CS-2026-0017', 2026, 17, 'Quiero ordenar la bandeja por fecha', '<p>Por defecto sale por lo más antiguo y prefiero lo más nuevo.</p>', 'resuelto',
        (SELECT id FROM users WHERE email = 'user2@demo.com'), (SELECT id FROM users WHERE email = 'user2@demo.com'), (SELECT id FROM users WHERE email = 'support3@demo.com'),
        (date_trunc('day', now()) - interval '42 days' + interval '13 hours 0 minutes'), NULL,
        (date_trunc('day', now()) - interval '45 days' + interval '9 hours 0 minutes'), (date_trunc('day', now()) - interval '44 days' + interval '17 hours 0 minutes'), (SELECT id FROM ticket_categories WHERE normalized = 'software'));

INSERT INTO ticket_history (ticket_id, actor_id, event, from_state, to_state, detail, created_at)
VALUES ((SELECT id FROM tickets WHERE number = 'CS-2026-0017'), (SELECT id FROM users WHERE email = 'support3@demo.com'), 'resuelto', 'en progreso', 'resuelto', NULL, (date_trunc('day', now()) - interval '42 days' + interval '13 hours 0 minutes'));
INSERT INTO ticket_history (ticket_id, actor_id, event, from_state, to_state, detail, created_at)
VALUES ((SELECT id FROM tickets WHERE number = 'CS-2026-0017'), (SELECT id FROM users WHERE email = 'support3@demo.com'), 'estado', 'nuevo', 'en progreso', NULL, (date_trunc('day', now()) - interval '44 days' + interval '10 hours 0 minutes'));
INSERT INTO ticket_history (ticket_id, actor_id, event, from_state, to_state, detail, created_at)
VALUES ((SELECT id FROM tickets WHERE number = 'CS-2026-0017'), (SELECT id FROM users WHERE email = 'user2@demo.com'), 'creado', NULL, 'nuevo', NULL, (date_trunc('day', now()) - interval '45 days' + interval '9 hours 0 minutes'));
INSERT INTO ticket_history (ticket_id, actor_id, event, from_state, to_state, detail, created_at)
VALUES ((SELECT id FROM tickets WHERE number = 'CS-2026-0017'), (SELECT id FROM users WHERE email = 'support3@demo.com'), 'asignado', NULL, NULL, 'Soporte Tres', (date_trunc('day', now()) - interval '45 days' + interval '10 hours 0 minutes'));

-- CS-2026-0018 · El ticket se cerró sin querer
INSERT INTO tickets (number, number_year, number_seq, subject, description, state, requester_id, created_by_id, assignee_id, resolved_at, closed_at, created_at, updated_at, category_id)
VALUES ('CS-2026-0018', 2026, 18, 'El ticket se cerró sin querer', '<p>Se cerró solo al quedarse sin actividad y el problema sigue.</p>', 'cerrado',
        (SELECT id FROM users WHERE email = 'user3@demo.com'), (SELECT id FROM users WHERE email = 'user3@demo.com'), (SELECT id FROM users WHERE email = 'support1@demo.com'),
        (date_trunc('day', now()) - interval '47 days' + interval '13 hours 0 minutes'), (date_trunc('day', now()) - interval '45 days' + interval '16 hours 0 minutes'),
        (date_trunc('day', now()) - interval '50 days' + interval '9 hours 0 minutes'), (date_trunc('day', now()) - interval '49 days' + interval '17 hours 0 minutes'), (SELECT id FROM ticket_categories WHERE normalized = 'general'));

INSERT INTO ticket_history (ticket_id, actor_id, event, from_state, to_state, detail, created_at)
VALUES ((SELECT id FROM tickets WHERE number = 'CS-2026-0018'), (SELECT id FROM users WHERE email = 'support1@demo.com'), 'cerrado', 'resuelto', 'cerrado', NULL, (date_trunc('day', now()) - interval '45 days' + interval '16 hours 0 minutes'));
INSERT INTO ticket_history (ticket_id, actor_id, event, from_state, to_state, detail, created_at)
VALUES ((SELECT id FROM tickets WHERE number = 'CS-2026-0018'), (SELECT id FROM users WHERE email = 'user3@demo.com'), 'reabierto', 'cerrado', 'en progreso', NULL, (date_trunc('day', now()) - interval '46 days' + interval '15 hours 0 minutes'));
INSERT INTO ticket_history (ticket_id, actor_id, event, from_state, to_state, detail, created_at)
VALUES ((SELECT id FROM tickets WHERE number = 'CS-2026-0018'), (SELECT id FROM users WHERE email = 'support1@demo.com'), 'resuelto', 'en progreso', 'resuelto', NULL, (date_trunc('day', now()) - interval '47 days' + interval '13 hours 0 minutes'));
INSERT INTO ticket_history (ticket_id, actor_id, event, from_state, to_state, detail, created_at)
VALUES ((SELECT id FROM tickets WHERE number = 'CS-2026-0018'), (SELECT id FROM users WHERE email = 'support1@demo.com'), 'resuelto', 'en progreso', 'resuelto', NULL, (date_trunc('day', now()) - interval '48 days' + interval '11 hours 0 minutes'));
INSERT INTO ticket_history (ticket_id, actor_id, event, from_state, to_state, detail, created_at)
VALUES ((SELECT id FROM tickets WHERE number = 'CS-2026-0018'), (SELECT id FROM users WHERE email = 'support1@demo.com'), 'estado', 'nuevo', 'en progreso', NULL, (date_trunc('day', now()) - interval '49 days' + interval '10 hours 0 minutes'));
INSERT INTO ticket_history (ticket_id, actor_id, event, from_state, to_state, detail, created_at)
VALUES ((SELECT id FROM tickets WHERE number = 'CS-2026-0018'), (SELECT id FROM users WHERE email = 'user3@demo.com'), 'creado', NULL, 'nuevo', NULL, (date_trunc('day', now()) - interval '50 days' + interval '9 hours 0 minutes'));
INSERT INTO ticket_history (ticket_id, actor_id, event, from_state, to_state, detail, created_at)
VALUES ((SELECT id FROM tickets WHERE number = 'CS-2026-0018'), (SELECT id FROM users WHERE email = 'support1@demo.com'), 'asignado', NULL, NULL, 'Soporte Uno', (date_trunc('day', now()) - interval '50 days' + interval '10 hours 0 minutes'));

INSERT INTO ticket_comments (ticket_id, author_id, body, created_at)
VALUES ((SELECT id FROM tickets WHERE number = 'CS-2026-0018'), (SELECT id FROM users WHERE email = 'support3@demo.com'), '<p>¿Lo podéis reabrir? Sigue pasando.</p>', (date_trunc('day', now()) - interval '49 days' + interval '11 hours 15 minutes'));
INSERT INTO ticket_comments (ticket_id, author_id, body, created_at)
VALUES ((SELECT id FROM tickets WHERE number = 'CS-2026-0018'), (SELECT id FROM users WHERE email = 'support1@demo.com'), '<p>Reabierto, perdona.</p>', (date_trunc('day', now()) - interval '48 days' + interval '12 hours 15 minutes'));

-- CS-2026-0019 · No puedo subir un adjunto en ZIP
INSERT INTO tickets (number, number_year, number_seq, subject, description, state, requester_id, created_by_id, assignee_id, resolved_at, closed_at, created_at, updated_at, category_id)
VALUES ('CS-2026-0019', 2026, 19, 'No puedo subir un adjunto en ZIP', '<p>Adjunto un zip con los registros y me lo rechaza.</p>', 'cerrado',
        (SELECT id FROM users WHERE email = 'user4@demo.com'), (SELECT id FROM users WHERE email = 'user4@demo.com'), (SELECT id FROM users WHERE email = 'support2@demo.com'),
        (date_trunc('day', now()) - interval '51 days' + interval '13 hours 0 minutes'), (date_trunc('day', now()) - interval '49 days' + interval '16 hours 0 minutes'),
        (date_trunc('day', now()) - interval '54 days' + interval '9 hours 0 minutes'), (date_trunc('day', now()) - interval '53 days' + interval '17 hours 0 minutes'), (SELECT id FROM ticket_categories WHERE normalized = 'software'));

INSERT INTO ticket_history (ticket_id, actor_id, event, from_state, to_state, detail, created_at)
VALUES ((SELECT id FROM tickets WHERE number = 'CS-2026-0019'), (SELECT id FROM users WHERE email = 'support2@demo.com'), 'cerrado', 'resuelto', 'cerrado', NULL, (date_trunc('day', now()) - interval '49 days' + interval '16 hours 0 minutes'));
INSERT INTO ticket_history (ticket_id, actor_id, event, from_state, to_state, detail, created_at)
VALUES ((SELECT id FROM tickets WHERE number = 'CS-2026-0019'), (SELECT id FROM users WHERE email = 'support2@demo.com'), 'resuelto', 'en progreso', 'resuelto', NULL, (date_trunc('day', now()) - interval '51 days' + interval '13 hours 0 minutes'));
INSERT INTO ticket_history (ticket_id, actor_id, event, from_state, to_state, detail, created_at)
VALUES ((SELECT id FROM tickets WHERE number = 'CS-2026-0019'), (SELECT id FROM users WHERE email = 'support2@demo.com'), 'estado', 'nuevo', 'en progreso', NULL, (date_trunc('day', now()) - interval '53 days' + interval '10 hours 0 minutes'));
INSERT INTO ticket_history (ticket_id, actor_id, event, from_state, to_state, detail, created_at)
VALUES ((SELECT id FROM tickets WHERE number = 'CS-2026-0019'), (SELECT id FROM users WHERE email = 'user4@demo.com'), 'creado', NULL, 'nuevo', NULL, (date_trunc('day', now()) - interval '54 days' + interval '9 hours 0 minutes'));
INSERT INTO ticket_history (ticket_id, actor_id, event, from_state, to_state, detail, created_at)
VALUES ((SELECT id FROM tickets WHERE number = 'CS-2026-0019'), (SELECT id FROM users WHERE email = 'support2@demo.com'), 'asignado', NULL, NULL, 'Soporte Dos', (date_trunc('day', now()) - interval '54 days' + interval '10 hours 0 minutes'));

-- El archivo de este adjunto lo copia `scripts/dev-seed.sh` desde `config/seed/notas.txt`.
INSERT INTO ticket_attachments (ticket_id, uploaded_by_id, filename, stored_name, content_type, size_bytes, created_at)
VALUES ((SELECT id FROM tickets WHERE number = 'CS-2026-0019'), (SELECT id FROM users WHERE email = 'user4@demo.com'), 'notas.txt', '2026/CS-2026-0019/fb3af71bf3908eb8dbc6105c1edff4e2.txt', 'text/plain', 1024, (date_trunc('day', now()) - interval '54 days' + interval '9 hours 30 minutes'));

-- CS-2026-0020 · Cambiar el idioma de los correos
INSERT INTO tickets (number, number_year, number_seq, subject, description, state, requester_id, created_by_id, assignee_id, resolved_at, closed_at, created_at, updated_at, category_id)
VALUES ('CS-2026-0020', 2026, 20, 'Cambiar el idioma de los correos', '<p>Los avisos me llegan en español y los quiero en inglés.</p>', 'cerrado',
        (SELECT id FROM users WHERE email = 'user5@demo.com'), (SELECT id FROM users WHERE email = 'user5@demo.com'), (SELECT id FROM users WHERE email = 'support3@demo.com'),
        (date_trunc('day', now()) - interval '55 days' + interval '13 hours 0 minutes'), (date_trunc('day', now()) - interval '53 days' + interval '16 hours 0 minutes'),
        (date_trunc('day', now()) - interval '58 days' + interval '9 hours 0 minutes'), (date_trunc('day', now()) - interval '57 days' + interval '17 hours 0 minutes'), (SELECT id FROM ticket_categories WHERE normalized = 'software'));

INSERT INTO ticket_history (ticket_id, actor_id, event, from_state, to_state, detail, created_at)
VALUES ((SELECT id FROM tickets WHERE number = 'CS-2026-0020'), (SELECT id FROM users WHERE email = 'support3@demo.com'), 'cerrado', 'resuelto', 'cerrado', NULL, (date_trunc('day', now()) - interval '53 days' + interval '16 hours 0 minutes'));
INSERT INTO ticket_history (ticket_id, actor_id, event, from_state, to_state, detail, created_at)
VALUES ((SELECT id FROM tickets WHERE number = 'CS-2026-0020'), (SELECT id FROM users WHERE email = 'support3@demo.com'), 'resuelto', 'en progreso', 'resuelto', NULL, (date_trunc('day', now()) - interval '55 days' + interval '13 hours 0 minutes'));
INSERT INTO ticket_history (ticket_id, actor_id, event, from_state, to_state, detail, created_at)
VALUES ((SELECT id FROM tickets WHERE number = 'CS-2026-0020'), (SELECT id FROM users WHERE email = 'support3@demo.com'), 'estado', 'nuevo', 'en progreso', NULL, (date_trunc('day', now()) - interval '57 days' + interval '10 hours 0 minutes'));
INSERT INTO ticket_history (ticket_id, actor_id, event, from_state, to_state, detail, created_at)
VALUES ((SELECT id FROM tickets WHERE number = 'CS-2026-0020'), (SELECT id FROM users WHERE email = 'user5@demo.com'), 'creado', NULL, 'nuevo', NULL, (date_trunc('day', now()) - interval '58 days' + interval '9 hours 0 minutes'));
INSERT INTO ticket_history (ticket_id, actor_id, event, from_state, to_state, detail, created_at)
VALUES ((SELECT id FROM tickets WHERE number = 'CS-2026-0020'), (SELECT id FROM users WHERE email = 'support3@demo.com'), 'asignado', NULL, NULL, 'Soporte Tres', (date_trunc('day', now()) - interval '58 days' + interval '10 hours 0 minutes'));

-- CS-2026-0021 · Informe de tickets de agosto
INSERT INTO tickets (number, number_year, number_seq, subject, description, state, requester_id, created_by_id, assignee_id, resolved_at, closed_at, created_at, updated_at, category_id)
VALUES ('CS-2026-0021', 2026, 21, 'Informe de tickets de agosto', '<p>Necesito el listado de agosto para la reunión de equipo.</p>', 'cerrado',
        (SELECT id FROM users WHERE email = 'user1@demo.com'), (SELECT id FROM users WHERE email = 'user1@demo.com'), (SELECT id FROM users WHERE email = 'support1@demo.com'),
        (date_trunc('day', now()) - interval '59 days' + interval '13 hours 0 minutes'), (date_trunc('day', now()) - interval '57 days' + interval '16 hours 0 minutes'),
        (date_trunc('day', now()) - interval '62 days' + interval '9 hours 0 minutes'), (date_trunc('day', now()) - interval '61 days' + interval '17 hours 0 minutes'), (SELECT id FROM ticket_categories WHERE normalized = 'software'));

INSERT INTO ticket_history (ticket_id, actor_id, event, from_state, to_state, detail, created_at)
VALUES ((SELECT id FROM tickets WHERE number = 'CS-2026-0021'), (SELECT id FROM users WHERE email = 'support1@demo.com'), 'cerrado', 'resuelto', 'cerrado', NULL, (date_trunc('day', now()) - interval '57 days' + interval '16 hours 0 minutes'));
INSERT INTO ticket_history (ticket_id, actor_id, event, from_state, to_state, detail, created_at)
VALUES ((SELECT id FROM tickets WHERE number = 'CS-2026-0021'), (SELECT id FROM users WHERE email = 'support1@demo.com'), 'resuelto', 'en progreso', 'resuelto', NULL, (date_trunc('day', now()) - interval '59 days' + interval '13 hours 0 minutes'));
INSERT INTO ticket_history (ticket_id, actor_id, event, from_state, to_state, detail, created_at)
VALUES ((SELECT id FROM tickets WHERE number = 'CS-2026-0021'), (SELECT id FROM users WHERE email = 'support1@demo.com'), 'estado', 'nuevo', 'en progreso', NULL, (date_trunc('day', now()) - interval '61 days' + interval '10 hours 0 minutes'));
INSERT INTO ticket_history (ticket_id, actor_id, event, from_state, to_state, detail, created_at)
VALUES ((SELECT id FROM tickets WHERE number = 'CS-2026-0021'), (SELECT id FROM users WHERE email = 'user1@demo.com'), 'creado', NULL, 'nuevo', NULL, (date_trunc('day', now()) - interval '62 days' + interval '9 hours 0 minutes'));
INSERT INTO ticket_history (ticket_id, actor_id, event, from_state, to_state, detail, created_at)
VALUES ((SELECT id FROM tickets WHERE number = 'CS-2026-0021'), (SELECT id FROM users WHERE email = 'support1@demo.com'), 'asignado', NULL, NULL, 'Soporte Uno', (date_trunc('day', now()) - interval '62 days' + interval '10 hours 0 minutes'));

-- CS-2026-0022 · El logo se ve pequeño en móvil
INSERT INTO tickets (number, number_year, number_seq, subject, description, state, requester_id, created_by_id, assignee_id, resolved_at, closed_at, created_at, updated_at, category_id)
VALUES ('CS-2026-0022', 2026, 22, 'El logo se ve pequeño en móvil', '<p>En la pantalla de entrada el logo se ve muy pequeño en el móvil.</p>', 'cerrado',
        (SELECT id FROM users WHERE email = 'user2@demo.com'), (SELECT id FROM users WHERE email = 'user2@demo.com'), (SELECT id FROM users WHERE email = 'support2@demo.com'),
        (date_trunc('day', now()) - interval '63 days' + interval '13 hours 0 minutes'), (date_trunc('day', now()) - interval '61 days' + interval '16 hours 0 minutes'),
        (date_trunc('day', now()) - interval '66 days' + interval '9 hours 0 minutes'), (date_trunc('day', now()) - interval '65 days' + interval '17 hours 0 minutes'), (SELECT id FROM ticket_categories WHERE normalized = 'software'));

INSERT INTO ticket_history (ticket_id, actor_id, event, from_state, to_state, detail, created_at)
VALUES ((SELECT id FROM tickets WHERE number = 'CS-2026-0022'), (SELECT id FROM users WHERE email = 'support2@demo.com'), 'cerrado', 'resuelto', 'cerrado', NULL, (date_trunc('day', now()) - interval '61 days' + interval '16 hours 0 minutes'));
INSERT INTO ticket_history (ticket_id, actor_id, event, from_state, to_state, detail, created_at)
VALUES ((SELECT id FROM tickets WHERE number = 'CS-2026-0022'), (SELECT id FROM users WHERE email = 'support2@demo.com'), 'resuelto', 'en progreso', 'resuelto', NULL, (date_trunc('day', now()) - interval '63 days' + interval '13 hours 0 minutes'));
INSERT INTO ticket_history (ticket_id, actor_id, event, from_state, to_state, detail, created_at)
VALUES ((SELECT id FROM tickets WHERE number = 'CS-2026-0022'), (SELECT id FROM users WHERE email = 'support2@demo.com'), 'estado', 'nuevo', 'en progreso', NULL, (date_trunc('day', now()) - interval '65 days' + interval '10 hours 0 minutes'));
INSERT INTO ticket_history (ticket_id, actor_id, event, from_state, to_state, detail, created_at)
VALUES ((SELECT id FROM tickets WHERE number = 'CS-2026-0022'), (SELECT id FROM users WHERE email = 'user2@demo.com'), 'creado', NULL, 'nuevo', NULL, (date_trunc('day', now()) - interval '66 days' + interval '9 hours 0 minutes'));
INSERT INTO ticket_history (ticket_id, actor_id, event, from_state, to_state, detail, created_at)
VALUES ((SELECT id FROM tickets WHERE number = 'CS-2026-0022'), (SELECT id FROM users WHERE email = 'support2@demo.com'), 'asignado', NULL, NULL, 'Soporte Dos', (date_trunc('day', now()) - interval '66 days' + interval '10 hours 0 minutes'));

-- CS-2026-0023 · Falta el aviso cuando alguien escala
INSERT INTO tickets (number, number_year, number_seq, subject, description, state, requester_id, created_by_id, assignee_id, resolved_at, closed_at, created_at, updated_at, category_id)
VALUES ('CS-2026-0023', 2026, 23, 'Falta el aviso cuando alguien escala', '<p>Desarrollo no se entera de que le toca un ticket.</p>', 'cerrado',
        (SELECT id FROM users WHERE email = 'user3@demo.com'), (SELECT id FROM users WHERE email = 'user3@demo.com'), (SELECT id FROM users WHERE email = 'support3@demo.com'),
        (date_trunc('day', now()) - interval '67 days' + interval '13 hours 0 minutes'), (date_trunc('day', now()) - interval '65 days' + interval '16 hours 0 minutes'),
        (date_trunc('day', now()) - interval '70 days' + interval '9 hours 0 minutes'), (date_trunc('day', now()) - interval '69 days' + interval '17 hours 0 minutes'), (SELECT id FROM ticket_categories WHERE normalized = 'software'));

INSERT INTO ticket_history (ticket_id, actor_id, event, from_state, to_state, detail, created_at)
VALUES ((SELECT id FROM tickets WHERE number = 'CS-2026-0023'), (SELECT id FROM users WHERE email = 'support3@demo.com'), 'cerrado', 'resuelto', 'cerrado', NULL, (date_trunc('day', now()) - interval '65 days' + interval '16 hours 0 minutes'));
INSERT INTO ticket_history (ticket_id, actor_id, event, from_state, to_state, detail, created_at)
VALUES ((SELECT id FROM tickets WHERE number = 'CS-2026-0023'), NULL, 'estado', 'escalado', 'en progreso', 'Desarrollo ha devuelto el ticket a Soporte.', (date_trunc('day', now()) - interval '66 days' + interval '12 hours 0 minutes'));
INSERT INTO ticket_history (ticket_id, actor_id, event, from_state, to_state, detail, created_at)
VALUES ((SELECT id FROM tickets WHERE number = 'CS-2026-0023'), (SELECT id FROM users WHERE email = 'support3@demo.com'), 'resuelto', 'en progreso', 'resuelto', NULL, (date_trunc('day', now()) - interval '67 days' + interval '13 hours 0 minutes'));
INSERT INTO ticket_history (ticket_id, actor_id, event, from_state, to_state, detail, created_at)
VALUES ((SELECT id FROM tickets WHERE number = 'CS-2026-0023'), (SELECT id FROM users WHERE email = 'support1@demo.com'), 'escalado', 'en progreso', 'escalado', 'El aviso de escalado no salía por un error de configuración del reparto.', (date_trunc('day', now()) - interval '68 days' + interval '11 hours 0 minutes'));
INSERT INTO ticket_history (ticket_id, actor_id, event, from_state, to_state, detail, created_at)
VALUES ((SELECT id FROM tickets WHERE number = 'CS-2026-0023'), (SELECT id FROM users WHERE email = 'support3@demo.com'), 'estado', 'nuevo', 'en progreso', NULL, (date_trunc('day', now()) - interval '69 days' + interval '10 hours 0 minutes'));
INSERT INTO ticket_history (ticket_id, actor_id, event, from_state, to_state, detail, created_at)
VALUES ((SELECT id FROM tickets WHERE number = 'CS-2026-0023'), (SELECT id FROM users WHERE email = 'user3@demo.com'), 'creado', NULL, 'nuevo', NULL, (date_trunc('day', now()) - interval '70 days' + interval '9 hours 0 minutes'));
INSERT INTO ticket_history (ticket_id, actor_id, event, from_state, to_state, detail, created_at)
VALUES ((SELECT id FROM tickets WHERE number = 'CS-2026-0023'), (SELECT id FROM users WHERE email = 'support3@demo.com'), 'asignado', NULL, NULL, 'Soporte Tres', (date_trunc('day', now()) - interval '70 days' + interval '10 hours 0 minutes'));

INSERT INTO ticket_comments (ticket_id, author_id, body, created_at)
VALUES ((SELECT id FROM tickets WHERE number = 'CS-2026-0023'), (SELECT id FROM users WHERE email = 'support3@demo.com'), '<p>Al final era cosa de la configuración del aviso.</p>', (date_trunc('day', now()) - interval '69 days' + interval '11 hours 15 minutes'));

-- INT-CS-2026-0023: el ticket interno que cuelga de CS-2026-0023
INSERT INTO internal_tickets (ticket_id, number, state, escalation_reason, created_by_id, assignee_id, resolved_at, closed_at, created_at, updated_at)
VALUES ((SELECT id FROM tickets WHERE number = 'CS-2026-0023'), 'INT-CS-2026-0023', 'cerrado', 'El aviso de escalado no salía por un error de configuración del reparto.', (SELECT id FROM users WHERE email = 'support1@demo.com'), (SELECT id FROM users WHERE email = 'dev2@demo.com'),
        (date_trunc('day', now()) - interval '66 days' + interval '12 hours 0 minutes'), (date_trunc('day', now()) - interval '65 days' + interval '15 hours 0 minutes'),
        (date_trunc('day', now()) - interval '68 days' + interval '11 hours 0 minutes'), (date_trunc('day', now()) - interval '67 days' + interval '16 hours 0 minutes'));

INSERT INTO ticket_history (internal_ticket_id, actor_id, event, from_state, to_state, detail, created_at)
VALUES ((SELECT id FROM internal_tickets WHERE number = 'INT-CS-2026-0023'), (SELECT id FROM users WHERE email = 'support1@demo.com'), 'creado', NULL, 'nuevo', 'El aviso de escalado no salía por un error de configuración del reparto.', (date_trunc('day', now()) - interval '68 days' + interval '11 hours 0 minutes'));
INSERT INTO ticket_history (internal_ticket_id, actor_id, event, from_state, to_state, detail, created_at)
VALUES ((SELECT id FROM internal_tickets WHERE number = 'INT-CS-2026-0023'), (SELECT id FROM users WHERE email = 'dev2@demo.com'), 'asignado', NULL, NULL, 'Desarrollo Dos', (date_trunc('day', now()) - interval '68 days' + interval '12 hours 0 minutes'));
INSERT INTO ticket_history (internal_ticket_id, actor_id, event, from_state, to_state, detail, created_at)
VALUES ((SELECT id FROM internal_tickets WHERE number = 'INT-CS-2026-0023'), (SELECT id FROM users WHERE email = 'dev2@demo.com'), 'resuelto', 'en progreso', 'resuelto', NULL, (date_trunc('day', now()) - interval '66 days' + interval '12 hours 0 minutes'));
INSERT INTO ticket_history (internal_ticket_id, actor_id, event, from_state, to_state, detail, created_at)
VALUES ((SELECT id FROM internal_tickets WHERE number = 'INT-CS-2026-0023'), (SELECT id FROM users WHERE email = 'support1@demo.com'), 'cerrado', 'resuelto', 'cerrado', NULL, (date_trunc('day', now()) - interval '65 days' + interval '15 hours 0 minutes'));
INSERT INTO ticket_comments (internal_ticket_id, author_id, body, created_at)
VALUES ((SELECT id FROM internal_tickets WHERE number = 'INT-CS-2026-0023'), (SELECT id FROM users WHERE email = 'support1@demo.com'), '<p>Te lo paso con todo lo que he probado hasta ahora.</p>', (date_trunc('day', now()) - interval '68 days' + interval '12 hours 20 minutes'));
INSERT INTO ticket_comments (internal_ticket_id, author_id, body, created_at)
VALUES ((SELECT id FROM internal_tickets WHERE number = 'INT-CS-2026-0023'), (SELECT id FROM users WHERE email = 'dev2@demo.com'), '<p>Lo miro hoy y te digo algo.</p>', (date_trunc('day', now()) - interval '67 days' + interval '10 hours 0 minutes'));

-- CS-2026-0024 · Dos tickets con el mismo número
INSERT INTO tickets (number, number_year, number_seq, subject, description, state, requester_id, created_by_id, assignee_id, resolved_at, closed_at, created_at, updated_at, category_id)
VALUES ('CS-2026-0024', 2026, 24, 'Dos tickets con el mismo número', '<p>He visto dos tickets seguidos con el mismo número.</p>', 'cerrado',
        (SELECT id FROM users WHERE email = 'user4@demo.com'), (SELECT id FROM users WHERE email = 'user4@demo.com'), (SELECT id FROM users WHERE email = 'support1@demo.com'),
        (date_trunc('day', now()) - interval '71 days' + interval '13 hours 0 minutes'), (date_trunc('day', now()) - interval '69 days' + interval '16 hours 0 minutes'),
        (date_trunc('day', now()) - interval '74 days' + interval '9 hours 0 minutes'), (date_trunc('day', now()) - interval '73 days' + interval '17 hours 0 minutes'), (SELECT id FROM ticket_categories WHERE normalized = 'software'));

INSERT INTO ticket_history (ticket_id, actor_id, event, from_state, to_state, detail, created_at)
VALUES ((SELECT id FROM tickets WHERE number = 'CS-2026-0024'), (SELECT id FROM users WHERE email = 'support1@demo.com'), 'cerrado', 'resuelto', 'cerrado', NULL, (date_trunc('day', now()) - interval '69 days' + interval '16 hours 0 minutes'));
INSERT INTO ticket_history (ticket_id, actor_id, event, from_state, to_state, detail, created_at)
VALUES ((SELECT id FROM tickets WHERE number = 'CS-2026-0024'), (SELECT id FROM users WHERE email = 'support1@demo.com'), 'resuelto', 'en progreso', 'resuelto', NULL, (date_trunc('day', now()) - interval '71 days' + interval '13 hours 0 minutes'));
INSERT INTO ticket_history (ticket_id, actor_id, event, from_state, to_state, detail, created_at)
VALUES ((SELECT id FROM tickets WHERE number = 'CS-2026-0024'), (SELECT id FROM users WHERE email = 'support1@demo.com'), 'estado', 'nuevo', 'en progreso', NULL, (date_trunc('day', now()) - interval '73 days' + interval '10 hours 0 minutes'));
INSERT INTO ticket_history (ticket_id, actor_id, event, from_state, to_state, detail, created_at)
VALUES ((SELECT id FROM tickets WHERE number = 'CS-2026-0024'), (SELECT id FROM users WHERE email = 'user4@demo.com'), 'creado', NULL, 'nuevo', NULL, (date_trunc('day', now()) - interval '74 days' + interval '9 hours 0 minutes'));
INSERT INTO ticket_history (ticket_id, actor_id, event, from_state, to_state, detail, created_at)
VALUES ((SELECT id FROM tickets WHERE number = 'CS-2026-0024'), (SELECT id FROM users WHERE email = 'support1@demo.com'), 'asignado', NULL, NULL, 'Soporte Uno', (date_trunc('day', now()) - interval '74 days' + interval '10 hours 0 minutes'));

-- El archivo de este adjunto lo copia `scripts/dev-seed.sh` desde `config/seed/captura.png`.
INSERT INTO ticket_attachments (ticket_id, uploaded_by_id, filename, stored_name, content_type, size_bytes, created_at)
VALUES ((SELECT id FROM tickets WHERE number = 'CS-2026-0024'), (SELECT id FROM users WHERE email = 'user4@demo.com'), 'captura.png', '2026/CS-2026-0024/04cd1e6b4d81a038f95ec51a6438abba.png', 'image/png', 1024, (date_trunc('day', now()) - interval '74 days' + interval '9 hours 30 minutes'));

-- CS-2026-0025 · Pedir la baja de una cuenta antigua
INSERT INTO tickets (number, number_year, number_seq, subject, description, state, requester_id, created_by_id, assignee_id, resolved_at, closed_at, created_at, updated_at, category_id)
VALUES ('CS-2026-0025', 2026, 25, 'Pedir la baja de una cuenta antigua', '<p>Se ha ido una persona y su cuenta sigue activa.</p>', 'escalado',
        (SELECT id FROM users WHERE email = 'user5@demo.com'), (SELECT id FROM users WHERE email = 'user5@demo.com'), (SELECT id FROM users WHERE email = 'support2@demo.com'),
        NULL, NULL,
        (date_trunc('day', now()) - interval '80 days' + interval '9 hours 0 minutes'), (date_trunc('day', now()) - interval '79 days' + interval '17 hours 0 minutes'), (SELECT id FROM ticket_categories WHERE normalized = 'licencias'));

INSERT INTO ticket_history (ticket_id, actor_id, event, from_state, to_state, detail, created_at)
VALUES ((SELECT id FROM tickets WHERE number = 'CS-2026-0025'), NULL, 'estado', 'escalado', 'en progreso', 'Desarrollo ha devuelto el ticket a Soporte.', (date_trunc('day', now()) - interval '76 days' + interval '12 hours 0 minutes'));
INSERT INTO ticket_history (ticket_id, actor_id, event, from_state, to_state, detail, created_at)
VALUES ((SELECT id FROM tickets WHERE number = 'CS-2026-0025'), (SELECT id FROM users WHERE email = 'support1@demo.com'), 'escalado', 'en progreso', 'escalado', 'La baja de cuentas no está prevista: hay que desactivarla a mano.', (date_trunc('day', now()) - interval '78 days' + interval '11 hours 0 minutes'));
INSERT INTO ticket_history (ticket_id, actor_id, event, from_state, to_state, detail, created_at)
VALUES ((SELECT id FROM tickets WHERE number = 'CS-2026-0025'), (SELECT id FROM users WHERE email = 'support2@demo.com'), 'estado', 'nuevo', 'en progreso', NULL, (date_trunc('day', now()) - interval '79 days' + interval '10 hours 0 minutes'));
INSERT INTO ticket_history (ticket_id, actor_id, event, from_state, to_state, detail, created_at)
VALUES ((SELECT id FROM tickets WHERE number = 'CS-2026-0025'), (SELECT id FROM users WHERE email = 'user5@demo.com'), 'creado', NULL, 'nuevo', NULL, (date_trunc('day', now()) - interval '80 days' + interval '9 hours 0 minutes'));
INSERT INTO ticket_history (ticket_id, actor_id, event, from_state, to_state, detail, created_at)
VALUES ((SELECT id FROM tickets WHERE number = 'CS-2026-0025'), (SELECT id FROM users WHERE email = 'support2@demo.com'), 'asignado', NULL, NULL, 'Soporte Dos', (date_trunc('day', now()) - interval '80 days' + interval '10 hours 0 minutes'));

-- INT-CS-2026-0025: el ticket interno que cuelga de CS-2026-0025
INSERT INTO internal_tickets (ticket_id, number, state, escalation_reason, created_by_id, assignee_id, resolved_at, closed_at, created_at, updated_at)
VALUES ((SELECT id FROM tickets WHERE number = 'CS-2026-0025'), 'INT-CS-2026-0025', 'cerrado', 'La baja de cuentas no está prevista: hay que desactivarla a mano.', (SELECT id FROM users WHERE email = 'support1@demo.com'), (SELECT id FROM users WHERE email = 'dev1@demo.com'),
        (date_trunc('day', now()) - interval '76 days' + interval '12 hours 0 minutes'), (date_trunc('day', now()) - interval '75 days' + interval '15 hours 0 minutes'),
        (date_trunc('day', now()) - interval '78 days' + interval '11 hours 0 minutes'), (date_trunc('day', now()) - interval '77 days' + interval '16 hours 0 minutes'));

INSERT INTO ticket_history (internal_ticket_id, actor_id, event, from_state, to_state, detail, created_at)
VALUES ((SELECT id FROM internal_tickets WHERE number = 'INT-CS-2026-0025'), (SELECT id FROM users WHERE email = 'support1@demo.com'), 'creado', NULL, 'nuevo', 'La baja de cuentas no está prevista: hay que desactivarla a mano.', (date_trunc('day', now()) - interval '78 days' + interval '11 hours 0 minutes'));
INSERT INTO ticket_history (internal_ticket_id, actor_id, event, from_state, to_state, detail, created_at)
VALUES ((SELECT id FROM internal_tickets WHERE number = 'INT-CS-2026-0025'), (SELECT id FROM users WHERE email = 'dev1@demo.com'), 'asignado', NULL, NULL, 'Desarrollo Uno', (date_trunc('day', now()) - interval '78 days' + interval '12 hours 0 minutes'));
INSERT INTO ticket_history (internal_ticket_id, actor_id, event, from_state, to_state, detail, created_at)
VALUES ((SELECT id FROM internal_tickets WHERE number = 'INT-CS-2026-0025'), (SELECT id FROM users WHERE email = 'dev1@demo.com'), 'resuelto', 'en progreso', 'resuelto', NULL, (date_trunc('day', now()) - interval '76 days' + interval '12 hours 0 minutes'));
INSERT INTO ticket_history (internal_ticket_id, actor_id, event, from_state, to_state, detail, created_at)
VALUES ((SELECT id FROM internal_tickets WHERE number = 'INT-CS-2026-0025'), (SELECT id FROM users WHERE email = 'support1@demo.com'), 'cerrado', 'resuelto', 'cerrado', NULL, (date_trunc('day', now()) - interval '75 days' + interval '15 hours 0 minutes'));
INSERT INTO ticket_comments (internal_ticket_id, author_id, body, created_at)
VALUES ((SELECT id FROM internal_tickets WHERE number = 'INT-CS-2026-0025'), (SELECT id FROM users WHERE email = 'support1@demo.com'), '<p>Te lo paso con todo lo que he probado hasta ahora.</p>', (date_trunc('day', now()) - interval '78 days' + interval '12 hours 20 minutes'));
INSERT INTO ticket_comments (internal_ticket_id, author_id, body, created_at)
VALUES ((SELECT id FROM internal_tickets WHERE number = 'INT-CS-2026-0025'), (SELECT id FROM users WHERE email = 'dev1@demo.com'), '<p>Lo miro hoy y te digo algo.</p>', (date_trunc('day', now()) - interval '77 days' + interval '10 hours 0 minutes'));

-- ============================================================================
-- 5. El contador de la numeración
-- ============================================================================

-- Los números que se han usado: el siguiente ticket de ejemplo que cree alguien será el 26.
INSERT INTO ticket_number_counters (year, last_number) VALUES (2026, 25);

-- Las etiquetas: **ahora viven en su propio catálogo** (`ticket_tag_names`) y los tickets las llevan
-- por clave ajena (docs/modules/tickets.md, decisión 72). Se dan de alta primero las que hacen falta
-- —distintas, que varias se repiten entre tickets— y después se enganchan. Siguen en minúsculas y con
-- guion medio en lugar de espacios, que es como se guardan.
--
-- Se dan de alta a nombre de Soporte Uno. `ON CONFLICT` deja el archivo aplicable dos veces por si
-- alguien lo lanza sin la limpieza de arriba.
INSERT INTO ticket_tag_names (tag, normalized, created_by_id)
SELECT v.tag, v.tag, u.id
FROM (VALUES
    ('acceso'),
    ('adjuntos'),
    ('alta-usuario'),
    ('avisos'),
    ('baja'),
    ('bandeja'),
    ('borrado'),
    ('buscador'),
    ('caducidad'),
    ('carpeta-compartida'),
    ('cerrado'),
    ('comentario'),
    ('contrasena'),
    ('correo'),
    ('cuenta'),
    ('duplicado'),
    ('escalado'),
    ('ficha'),
    ('formato-fecha'),
    ('idioma'),
    ('impresora'),
    ('informe'),
    ('instalacion'),
    ('lentitud'),
    ('logo'),
    ('manana'),
    ('movil'),
    ('nombre'),
    ('numeracion'),
    ('numero'),
    ('orden'),
    ('permisos'),
    ('reabrir'),
    ('segunda-planta'),
    ('spam'),
    ('tamano'),
    ('telefono'),
    ('tickets'),
    ('zip')
) AS v(tag)
JOIN users u ON u.email = 'support1@demo.com'
ON CONFLICT (normalized) DO NOTHING;

-- Y ahora cada ticket con las suyas: **dos por ticket**, enganchadas por el número, que es como se
-- lee aquí.
INSERT INTO ticket_tags (ticket_id, tag_id, created_by_id)
SELECT t.id, n.id, u.id
FROM (VALUES
    ('CS-2026-0001', 'contrasena'),
    ('CS-2026-0001', 'acceso'),
    ('CS-2026-0002', 'informe'),
    ('CS-2026-0002', 'formato-fecha'),
    ('CS-2026-0003', 'alta-usuario'),
    ('CS-2026-0003', 'permisos'),
    ('CS-2026-0004', 'impresora'),
    ('CS-2026-0004', 'segunda-planta'),
    ('CS-2026-0005', 'adjuntos'),
    ('CS-2026-0005', 'tamano'),
    ('CS-2026-0006', 'correo'),
    ('CS-2026-0006', 'cuenta'),
    ('CS-2026-0007', 'lentitud'),
    ('CS-2026-0007', 'manana'),
    ('CS-2026-0008', 'avisos'),
    ('CS-2026-0008', 'correo'),
    ('CS-2026-0009', 'ficha'),
    ('CS-2026-0009', 'telefono'),
    ('CS-2026-0010', 'comentario'),
    ('CS-2026-0010', 'borrado'),
    ('CS-2026-0011', 'nombre'),
    ('CS-2026-0011', 'instalacion'),
    ('CS-2026-0012', 'buscador'),
    ('CS-2026-0012', 'numero'),
    ('CS-2026-0013', 'carpeta-compartida'),
    ('CS-2026-0013', 'permisos'),
    ('CS-2026-0014', 'correo'),
    ('CS-2026-0014', 'spam'),
    ('CS-2026-0015', 'duplicado'),
    ('CS-2026-0015', 'impresora'),
    ('CS-2026-0016', 'contrasena'),
    ('CS-2026-0016', 'caducidad'),
    ('CS-2026-0017', 'bandeja'),
    ('CS-2026-0017', 'orden'),
    ('CS-2026-0018', 'cerrado'),
    ('CS-2026-0018', 'reabrir'),
    ('CS-2026-0019', 'zip'),
    ('CS-2026-0019', 'adjuntos'),
    ('CS-2026-0020', 'idioma'),
    ('CS-2026-0020', 'correo'),
    ('CS-2026-0021', 'informe'),
    ('CS-2026-0021', 'tickets'),
    ('CS-2026-0022', 'logo'),
    ('CS-2026-0022', 'movil'),
    ('CS-2026-0023', 'escalado'),
    ('CS-2026-0023', 'avisos'),
    ('CS-2026-0024', 'numeracion'),
    ('CS-2026-0024', 'duplicado'),
    ('CS-2026-0025', 'baja'),
    ('CS-2026-0025', 'cuenta')
) AS v(numero, tag)
JOIN tickets t ON t.number = v.numero
JOIN ticket_tag_names n ON n.normalized = v.tag
JOIN users u ON u.email = 'support1@demo.com'
ON CONFLICT (ticket_id, tag_id) DO NOTHING;

-- ============================================================================
-- 6. Los dos caminos de directorio, configurados pero apagados
-- ============================================================================

-- Las dos configuraciones quedan puestas y **el método se queda en `local`**: el entorno arranca
-- entrando con las once cuentas de ejemplo, y cambiar de método se hace desde Configuración, que es
-- justo lo que hay que poder probar a mano (docs/ambientes.md, sección 3.3).
--
-- Los dos servicios viven en `dev.yml` detrás del perfil `auth`. Con el perfil levantado, «Probar la
-- conexión» contesta que sí y se puede cambiar el método; sin él, contesta que no, que es la verdad.
--
-- La contraseña de la cuenta de servicio es la del LDIF de pruebas (`config/ldap/`): es un directorio
-- de mentira en un contenedor de desarrollo, no un secreto.

UPDATE directory_settings SET
    host = 'ldap',
    port = '389',
    use_tls = false,
    bind_dn = 'cn=admin,dc=ejemplo,dc=com',
    bind_password = 'admin-directorio',
    search_base = 'ou=personas,dc=ejemplo,dc=com',
    user_filter = '(mail=%s)',
    attr_email = 'mail',
    attr_name = 'givenName',
    attr_last_name = 'sn',
    attr_id = 'uid',
    updated_at = now()
WHERE id = 1;

UPDATE keycloak_settings SET
    issuer = 'https://dev-catalina-support.calibyou.com/sso/realms/catalina-support',
    client_id = 'catalina-support',
    client_secret = 'el-secreto-de-desarrollo',
    redirect_uri = 'https://dev-catalina-support.calibyou.com/api/auth/keycloak/callback',
    updated_at = now()
WHERE id = 1;

UPDATE installation_settings
   SET installation_name = 'Catalina Support',
       -- **El correo saliente de desarrollo vive aquí**, no en el entorno (decisión del responsable,
       -- 2026-09-30): es el buzón de pruebas, y sin él no saldría ningún correo en el entorno.
       smtp_host       = 'mail',
       smtp_port       = '1025',
       smtp_secure     = false,
       smtp_user       = '',
       smtp_password   = '',
       smtp_from_name  = 'Catalina Support',
       smtp_from_email = 'no-responder@catalina-support.local',
       time_zone      = 'America/Guayaquil',
       public_app_url = 'https://dev-catalina-support.calibyou.com',
       -- **Desarrollo queda instalado**: el asistente de primer arranque se prueba a mano, quitándole
       -- el sello, y no puede salir en cada arranque del entorno ni en las pruebas.
       installed_at   = COALESCE(installed_at, now()),
       updated_at     = now()
 WHERE id = 1;

COMMIT;
