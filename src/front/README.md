# Front - BFF Storage

Tela React inspirada no Google Drive para listar a raiz do bucket via BFF e baixar arquivos.

## Requisitos

- Node.js 18+
- pnpm
- BFF rodando (por padrao em `http://localhost:8080`)

## Configuracao

```bash
cp .env.example .env
```

## Executar

```bash
pnpm install
pnpm dev
```

Acesse `http://localhost:5173`.

## Build

```bash
pnpm build
```

## Endpoints usados

- `GET /bucket/items`
- `GET /bucket/items/file?id=<object_id>`

Por padrao, o front chama `/api/*` e o Vite faz proxy para o BFF.
