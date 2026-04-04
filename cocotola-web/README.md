# cocotola-web

Webフロントエンド。React Router 7 (SSR) + Tailwind CSS + Vite で構成。

## Getting Started

```bash
pnpm install
pnpm run dev
```

## Building for Production

```bash
pnpm run build
```

## Docker Deployment

```bash
docker build -t cocotola-web .
docker run -p 3000:3000 cocotola-web
```
