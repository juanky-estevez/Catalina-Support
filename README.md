# Catalina-Support

Mesa de ayuda de dos niveles, simple y sin ruido:

```text
usuario (reporta)  ->  soporte técnico (nivel 1)  ->  desarrollo (nivel 2)
```

## Antes de tocar nada

Leer `AGENTS.md`. La **Regla 0** es obligatoria: **primero el documento aprobado, después el
código.** El estado del proyecto, lo que existe y lo que falta están en `docs/README.md` y en la
sección 13 de `docs/arquitectura.md`.

## Levantarlo en desarrollo

Todo corre en contenedores (frontend, backend y base de datos); la máquina sólo necesita Docker:

```bash
docker compose -f dev.yml up -d
docker compose -f dev.yml logs -f backend
```

Se entra por **https://dev.catalina-support.example.com** (nginx delante de los contenedores).
Para depurar sin pasar por nginx, el frontend queda en `http://127.0.0.1:11001` y el backend en
`http://127.0.0.1:11002` (`GET /api/health`). Los comandos completos están en `AGENTS.md`.
