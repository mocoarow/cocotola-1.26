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

## OpenTelemetry トレーシング

SSR サーバーには OpenTelemetry によるトレーシングを組み込んでいる。`instrumentation.mjs`
が Node.js の `--import` フラグでサーバー起動前に読み込まれ、HTTP サーバーと外向きの
`fetch`(バックエンド API 呼び出し)を自動計装する。W3C `traceparent` が伝播するため、
cocotola-web → cocotola-question などバックエンドサービスまで 1 本のトレースとして追跡できる。

アプリ独自のスパンを追加したい場合は `app/lib/observability/tracing.server.ts` の `withSpan`
を利用する。

トレーシングは既定で無効。以下の環境変数で有効化・設定する(標準の OpenTelemetry 環境変数)。

| 環境変数 | 説明 | 既定値 |
| --- | --- | --- |
| `OTEL_TRACES_EXPORTER` | `otlp` / `console` / `none` | `none`(無効) |
| `OTEL_EXPORTER_OTLP_PROTOCOL` | `http/protobuf` / `grpc` | `http/protobuf` |
| `OTEL_EXPORTER_OTLP_ENDPOINT` | コレクターのエンドポイント | http: `http://localhost:4318` / grpc: `http://localhost:4317` |
| `OTEL_SERVICE_NAME` | `service.name` | `cocotola-web` |
| `OTEL_TRACES_SAMPLER` / `OTEL_TRACES_SAMPLER_ARG` | サンプラー設定 | `parentbased_always_on` |

`APP_ENV=test` のときと `OTEL_TRACES_EXPORTER` が未設定/`none` のときはトレーシングを無効化する
(ローカル開発・テストは既定で静かなまま)。

ローカルの Jaeger / LGTM へ送信する例:

```bash
OTEL_TRACES_EXPORTER=otlp \
OTEL_EXPORTER_OTLP_ENDPOINT=http://localhost:4318 \
APP_ENV=production \
pnpm run start
```
