# Contribuir a Catalina Support

Gracias por querer mejorar Catalina Support. Las conversaciones generales y preguntas van en
GitHub Discussions; los defectos y propuestas concretas, en Issues. Las vulnerabilidades se envían
por el reporte privado descrito en `SECURITY.md`.

## Documentación antes que código

Este proyecto sigue `AGENTS.md`: cualquier cambio de comportamiento, interfaz, modelo, permisos o
flujo necesita primero un documento en `docs/`, decisiones cerradas y aprobación explícita. Una
incidencia o discusión puede explorar una idea, pero no sustituye ese documento.

Antes de programar:

1. busca una incidencia y el documento del área;
2. explica el problema con un ejemplo reproducible;
3. acuerda el alcance y consigue que el documento quede `aprobado`;
4. implementa sólo ese alcance;
5. actualiza el documento a `as-built` con las pruebas reales.

Las erratas y ajustes puramente editoriales no necesitan el ciclo completo.

## Preparar el entorno

Sigue el recorrido de instalación vacía de `README.md`. El proyecto usa contenedores; no exige Go,
Node o PostgreSQL instalados en la máquina. Los datos de ejemplo son opcionales.

Antes de enviar un pull request ejecuta, según el área afectada:

```bash
docker compose -f dev.yml exec backend go test ./...
docker compose -f dev.yml exec backend go vet ./...
docker compose -f dev.yml exec frontend npm test -- --watch=false
docker compose -f dev.yml exec frontend npm run build
```

Si cambias `go.mod`, `go.sum`, `package.json` o `package-lock.json`, regenera y comprueba el
inventario de licencias de producción:

```bash
./scripts/third-party-notices.sh write
./scripts/third-party-notices.sh check
```

El guion falla si una dependencia distribuida no publica una licencia reconocida. Incluye en el
pull request `THIRD_PARTY_NOTICES.md` y cualquier cambio de `third_party_licenses/`.

Los cambios de recorridos visibles deben ejecutar también la suite desechable de Playwright indicada
en el README. No uses datos ni credenciales reales en pruebas, capturas o commits.

## Pull requests

Mantén cada pull request enfocado. Describe el problema, el comportamiento final, los documentos que
lo autorizan y la validación ejecutada. No incluyas artefactos generados, secretos ni cambios de
producción que no pertenezcan al alcance.

Los títulos y commits deben explicar una unidad de cambio. Se acepta español o inglés. La revisión
puede pedir dividir un cambio que mezcle áreas independientes.

## Acuerdo de contribución

Código, recursos y documentación sustancial requieren aceptar `CLA.md`. El bot lo solicita en el
pull request y registra la aceptación mínima en `signatures/version1/cla.json`. Las incidencias y las
correcciones triviales de erratas no requieren CLA.

Al participar aceptas `CODE_OF_CONDUCT.md` y confirmas que tienes derecho a aportar el contenido.
