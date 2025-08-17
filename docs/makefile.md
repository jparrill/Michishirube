# Makefile Documentation

Este documento describe todos los targets disponibles en el Makefile de Michishirube, explicando su función y uso.

## 📋 Índice

- [Build y Desarrollo](#build-y-desarrollo)
- [Testing](#testing)
- [Docker](#docker)
- [Releases](#releases)
- [CI/CD](#cicd)
- [Dependencias](#dependencias)
- [Seguridad](#seguridad)
- [Utilidades](#utilidades)

## 🏗️ Build y Desarrollo

### `make build`
**Función:** Compila la aplicación Go
**Descripción:** Crea un ejecutable binario en la carpeta `build/`
**Uso:** `make build`
**Output:** `build/michishirube`

### `make run`
**Función:** Ejecuta la aplicación en modo desarrollo
**Descripción:** Usa `go run` para ejecutar directamente el código fuente
**Uso:** `make run`
**Nota:** Ideal para desarrollo, no requiere compilación previa

### `make generate`
**Función:** Genera código automáticamente (mocks, etc.)
**Descripción:** Instala mockgen si es necesario y ejecuta `go generate`
**Uso:** `make generate`
**Dependencias:** Requiere mockgen para generar mocks de testing

### `make docs`
**Función:** Genera documentación OpenAPI/Swagger
**Descripción:** Instala swag y genera especificaciones OpenAPI
**Uso:** `make docs`
**Output:**
- `docs/` - Carpeta con documentación generada
- Endpoints disponibles:
  - `http://localhost:8080/docs` - Swagger UI personalizado
  - `http://localhost:8080/openapi.yaml` - Especificación YAML
  - `http://localhost:8080/swagger/doc.json` - Especificación JSON

## 🧪 Testing

### `make test`
**Función:** Ejecuta la suite completa de tests
**Descripción:** Ejecuta todos los tests en orden: unitarios, storage, handlers, integración
**Uso:** `make test`
**Dependencias:** `make generate`, `make fixtures-validate`
**Output:** Reporte detallado de cada categoría de tests

### `make test-coverage`
**Función:** Ejecuta tests con análisis de cobertura
**Descripción:** Genera reporte HTML de cobertura de código
**Uso:** `make test-coverage`
**Output:** `build/coverage.html`

### `make test-unit`
**Función:** Ejecuta solo tests unitarios
**Descripción:** Tests rápidos de models, config, logger y handlers
**Uso:** `make test-unit`
**Nota:** Más rápido que la suite completa

### `make test-integration`
**Función:** Ejecuta solo tests de integración
**Descripción:** Tests que verifican la integración entre componentes
**Uso:** `make test-integration`
**Dependencias:** `make fixtures-validate`

### `make test-bench`
**Función:** Ejecuta benchmarks de performance
**Descripción:** Mide rendimiento de funciones críticas
**Uso:** `make test-bench`
**Output:** Métricas de tiempo y memoria

### `make test-storage`
**Función:** Ejecuta tests específicos de storage
**Descripción:** Tests de la capa de base de datos SQLite
**Uso:** `make test-storage`

### `make test-search`
**Función:** Ejecuta tests de funcionalidad de búsqueda
**Descripción:** Tests específicos para features de búsqueda
**Uso:** `make test-search`

### `make fixtures-update`
**Función:** Actualiza fixtures con datos de test actuales
**Descripción:** Regenera datos de test basados en el estado actual
**Uso:** `make fixtures-update`
**Nota:** Útil cuando se cambian modelos o esquemas

### `make fixtures-validate`
**Función:** Valida fixtures para palabras reservadas
**Descripción:** Verifica que los fixtures no contengan problemas
**Uso:** `make fixtures-validate`

### `make dev-test`
**Función:** Helper para desarrollo
**Descripción:** Actualiza fixtures y ejecuta suite completa
**Uso:** `make dev-test`
**Equivale a:** `make fixtures-update && make test`

## 🐳 Docker

### `make docker-build`
**Función:** Construye imagen Docker (arquitectura única)
**Descripción:** Construye y hace push a registros Docker
**Uso:** `make docker-build`
**Registros:** quay.io/jparrill/michishirube, docker.io/padajuan/michishirube

### `make docker-multiarch`
**Función:** Construye imagen multi-arquitectura
**Descripción:** Construye para linux/amd64 y linux/arm64
**Uso:** `make docker-multiarch`
**Dependencias:** Docker buildx

### `make docker-up`
**Función:** Ejecuta entorno de producción con docker-compose
**Descripción:** Levanta todos los servicios de producción
**Uso:** `make docker-up`

### `make docker-dev`
**Función:** Ejecuta entorno de desarrollo con hot reload
**Descripción:** Levanta servicios con perfil de desarrollo
**Uso:** `make docker-dev`

### `make docker-down`
**Función:** Detiene todos los servicios docker-compose
**Descripción:** Para todos los contenedores
**Uso:** `make docker-down`

### `make docker-logs`
**Función:** Muestra logs de todos los servicios
**Descripción:** Logs en tiempo real con `-f`
**Uso:** `make docker-logs`

### `make docker-clean`
**Función:** Limpia recursos Docker
**Descripción:** Elimina contenedores, volúmenes y buildx
**Uso:** `make docker-clean`

## 🚀 Releases

### `make release`
**Función:** Crea y publica un release
**Descripción:** Requiere tag de git (ej: `git tag v1.0.0`)
**Uso:** `make release`
**Dependencias:** GoReleaser

### `make release-check`
**Función:** Valida configuración de GoReleaser
**Descripción:** Verifica que la configuración sea correcta
**Uso:** `make release-check`

### `make release-snapshot`
**Función:** Crea build de snapshot para desarrollo
**Descripción:** Build de prueba sin publicar
**Uso:** `make release-snapshot`

## 🔄 CI/CD

### `make ci-local`
**Función:** Simula pipeline de CI/CD localmente
**Descripción:** Ejecuta todos los pasos que se ejecutan en GitHub Actions
**Uso:** `make ci-local`
**Pasos:**
1. Gestión de dependencias
2. Generación de código y docs
3. Linting
4. Suite completa de tests
5. Análisis de cobertura
6. Build de la aplicación
7. Verificación de ejecución del binario
8. Validación de configuración de GoReleaser

### `make lint`
**Función:** Ejecuta linter de código
**Descripción:** Análisis estático con golangci-lint
**Uso:** `make lint`
**Dependencias:** golangci-lint instalado

## 📦 Dependencias

### `make deps`
**Función:** Gestiona dependencias de Go
**Descripción:** Descarga, verifica y ordena dependencias
**Uso:** `make deps`
**Pasos:** `go mod download`, `go mod verify`, `go mod tidy`

### `make deps-update`
**Función:** Actualiza dependencias a últimas versiones
**Descripción:** Actualiza todas las dependencias directas
**Uso:** `make deps-update`
**⚠️ Advertencia:** Revisar cambios y ejecutar tests antes de commit

### `make deps-clean`
**Función:** Limpia cache de dependencias
**Descripción:** Elimina cache y reinstala dependencias
**Uso:** `make deps-clean`
**Útil para:** Resolver problemas de dependencias

### `make deps-verify`
**Función:** Verifica integridad y seguridad de dependencias
**Descripción:** Verifica módulos y busca vulnerabilidades
**Uso:** `make deps-verify`
**Herramientas:** govulncheck (si está instalado)

## 🔒 Seguridad

### `make security`
**Función:** Escaneo completo de seguridad
**Descripción:** Ejecuta gosec y govulncheck
**Uso:** `make security`
**Equivale a:** `make security-install security-gosec security-govulncheck`

### `make security-install`
**Función:** Instala herramientas de seguridad
**Descripción:** Instala gosec y govulncheck
**Uso:** `make security-install`

### `make security-gosec`
**Función:** Escáner estático de seguridad
**Descripción:** Análisis de código Go para vulnerabilidades
**Uso:** `make security-gosec`
**Dependencias:** gosec instalado

### `make security-govulncheck`
**Función:** Verificador de vulnerabilidades
**Descripción:** Revisa dependencias contra base de datos de CVEs
**Uso:** `make security-govulncheck`
**Dependencias:** govulncheck instalado

### `make security-ci`
**Función:** Escaneo de seguridad para CI (modo estricto)
**Descripción:** Falla si encuentra problemas de severidad HIGH
**Uso:** `make security-ci`
**Ideal para:** Pipelines de CI/CD

### `make security-strict`
**Función:** Escaneo de seguridad estricto
**Descripción:** Falla si encuentra problemas de severidad MEDIUM+
**Uso:** `make security-strict`
**Ideal para:** Pre-releases

## 🛠️ Utilidades

### `make clean`
**Función:** Limpia artefactos de build
**Descripción:** Elimina carpetas build/, docs/, archivos .db, etc.
**Uso:** `make clean`

### `make test-help`
**Función:** Muestra ayuda sobre targets de testing
**Uso:** `make test-help`

### `make release-help`
**Función:** Muestra ayuda sobre targets de release
**Uso:** `make release-help`

### `make docker-help`
**Función:** Muestra ayuda sobre targets de Docker
**Uso:** `make docker-help`

### `make ci-help`
**Función:** Muestra ayuda sobre targets de CI/CD
**Uso:** `make ci-help`

### `make deps-help`
**Función:** Muestra ayuda sobre gestión de dependencias
**Uso:** `make deps-help`

### `make security-help`
**Función:** Muestra ayuda sobre targets de seguridad
**Uso:** `make security-help`

## 🔄 Flujos de Trabajo Comunes

### Desarrollo Diario
```bash
make deps          # Gestionar dependencias
make generate      # Generar mocks
make test-unit     # Tests rápidos
make run           # Ejecutar aplicación
```

### Antes de Commit
```bash
make lint          # Verificar calidad de código
make test          # Suite completa de tests
make security      # Escaneo de seguridad
```

### Antes de Release
```bash
make ci-local      # Simular CI completo
make release-check # Verificar configuración
make release       # Crear release
```

### Troubleshooting
```bash
make deps-clean    # Limpiar dependencias
make clean         # Limpiar build
make fixtures-update # Actualizar datos de test
```

## 📝 Variables de Entorno

### Docker
- `PORT=8080` - Puerto de la aplicación
- `LOG_LEVEL=info` - Nivel de logging
- `DB_PATH=./app.db` - Ruta de la base de datos
- `DEV_PORT=8081` - Puerto en modo desarrollo

## 🎯 Targets Principales

| Target | Descripción | Uso Frecuente |
|--------|-------------|---------------|
| `make test` | Suite completa de tests | ✅ Muy frecuente |
| `make run` | Ejecutar en desarrollo | ✅ Muy frecuente |
| `make build` | Compilar aplicación | ✅ Frecuente |
| `make lint` | Verificar código | ✅ Frecuente |
| `make ci-local` | Simular CI local | 🔄 Antes de push |
| `make release` | Crear release | 🚀 Releases |
| `make docker-up` | Levantar con Docker | 🐳 Despliegue |
| `make security` | Escaneo de seguridad | 🔒 Pre-release |

## 💡 Tips y Mejores Prácticas

1. **Siempre ejecuta `make test` antes de commit**
2. **Usa `make ci-local` para verificar que CI pasará**
3. **Ejecuta `make security` antes de cada release**
4. **Usa `make deps-update` regularmente para mantener dependencias actualizadas**
5. **`make clean` es útil cuando hay problemas de build**
6. **Los targets de ayuda (`make *-help`) proporcionan información detallada**
