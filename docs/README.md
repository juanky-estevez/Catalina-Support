# Documentación de Catalina-Support

> **Estado:** as-built
> **Última actualización:** 2026-09-22

Índice de la documentación del proyecto. La **Regla 0** de `AGENTS.md` es obligatoria: no se
escribe código sin un documento aprobado antes.

## Documentos

| Documento | Estado | Cubre |
| --- | --- | --- |
| `README.md` (este archivo) | as-built | Índice, convenciones y estados |
| `arquitectura.md` | **as-built** | Stack, versiones, forma del repositorio, contenedores y regla de modularidad (frontend ↔ backend). Su sección 13 dice qué está construido, qué está sin verificar y qué falta |
| `propósito-y-alcance.md` | **aprobado** (enmendado) | Qué problema resuelve la mesa de ayuda, los dos equipos y los cuatro papeles, el modelo de tickets (principal e interno, numeración y estados) y qué queda fuera. Sin decisiones abiertas |
| `usuarios-y-permisos.md` | **aprobado** (2 enmiendas) | La matriz papel × acción, los tres caminos de entrada (correo, AD, Keycloak), las reglas de convivencia, la sesión, el ciclo de vida de las cuentas y la cuenta de administrador de fábrica |
| `flujos.md` | **aprobado** | Los seis recorridos paso a paso (alta, triaje, escalado, trabajo de Desarrollo, cierre y reapertura), con los siete avisos por correo |
| `interfaz-y-experiencia.md` | **aprobado** | Principios, forma de la aplicación, los cuatro enfoques por papel, la vista doble, el lenguaje visual (Tailwind v4 y 19 componentes propios), los ocho temas, multi-dispositivo y accesibilidad |
| `ambientes.md` | **aprobado** | El runbook de despliegue, las migraciones, las copias de seguridad y las tres capas de pruebas |
| `tickets.md` | **aprobado** | Modelo de datos, numeración, transiciones de los dos ciclos de vida, adjuntos y la lista cerrada de módulos y endpoints. Aprobado tras dos repasos, con los seis huecos técnicos ya aplicados |

La documentación **está completa**: siete documentos que cubren qué se construye, cómo se comporta,
cómo se ve y cómo se despliega. **El código puede empezar** por `auth` (`docs/tickets.md`).

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

### Nombres

En español, en minúsculas, sin acentos ni espacios (guiones): `usuarios-y-permisos.md`.

### Correspondencia con el código

La tabla documento ↔ código vive en `AGENTS.md`. Cuando un documento pase a `as-built`, hay que
añadir o actualizar su fila allí en el mismo cambio.

### Reglas de escritura

- Español, frases cortas, sin relleno y sin adornos de marketing.
- Cada documento dice **qué está decidido**, **qué está pendiente** y **qué queda fuera de alcance**.
- Lo que no esté decidido se marca como pendiente; no se rellena con suposiciones.
- Cuando un documento describa algo ya implementado, describe lo que el código hace de verdad,
  no lo que se pretendía.
