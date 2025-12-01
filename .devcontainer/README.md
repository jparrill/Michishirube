# Michishirube Development Container

Este directorio contiene la configuración para desarrollar Michishirube usando VS Code/Cursor dev containers.

## 🚀 Quick Start

1. Abre el proyecto en Cursor/VS Code
2. `Cmd+Shift+P` → `Remote-Containers: Reopen in Container`
3. Espera a que se construya el container (2-3 minutos la primera vez)

## 📁 Configuraciones Disponibles

### Default: Configuración Mínima (`devcontainer.json`)
- **Base**: `golang:1.24-bookworm`
- **Enfoque**: Rápido de construir, instalación de herramientas al crear
- **Herramientas incluidas**:
  - Go 1.24 con todas las dependencias de Michishirube
  - SQLite + CGO support
  - Tools específicas: `gopls`, `mockgen`, `swag`, `dlv`
  - Extensiones: Go + Makefile tools

### Completo: Full-Featured (`Dockerfile.full`)
- **Base**: Custom Dockerfile con todo pre-instalado
- **Enfoque**: Ambiente completo para desarrollo avanzado
- **Herramientas adicionales**:
  - ZSH + Oh My Zsh
  - git-delta para mejores diffs
  - fzf para búsquedas
  - golangci-lint
  - Usuario `developer` no-root
  - Firewall de seguridad configurado

## 🔄 Cambiar Configuración

Para usar la configuración completa, edita `devcontainer.json`:

```json
{
  "build": {
    "dockerfile": "Dockerfile.full"
  },
  "remoteUser": "developer"
}
```

## 🛠️ Herramientas Disponibles

### Go Tools
- `gopls` - Language Server Protocol
- `mockgen` - Mock generator (para tus tests)
- `swag` - Swagger documentation generator
- `dlv` - Go debugger
- `golangci-lint` - Linter (solo en Dockerfile.full)

### Makefile Commands
Una vez en el container:
```bash
make build          # Compilar aplicación
make run            # Ejecutar en desarrollo
make test           # Ejecutar tests
make docs           # Generar documentación API
make lint           # Linting del código
```

## 🔒 Seguridad

- El container está aislado de tu sistema
- Configuración de firewall en `Dockerfile.full`
- Usuario no-root en configuración completa
- Mounts de solo los archivos necesarios

## 🐳 Arquitectura Soportada

- **amd64** (Intel)
- **arm64** (Apple Silicon)

## 📝 Notas de Desarrollo

- Tu código se monta en `/workspace`
- Configuraciones de Go optimizadas para Michishirube
- Extensions de Cursor/VS Code pre-configuradas
- Historial y configuraciones persistentes entre reinicios