# Cloud Run デプロイ計画

## 概要

別リポジトリのアプリケーションが更新されたら Cloud Run で再デプロイする。
パターンA（アプリリポジトリから直接デプロイ）を採用。

```text
App Repo (push) → GitHub Actions → Build Image → Push to Artifact Registry → gcloud run deploy
```

## 責務分担

| リソース | 管理者 | 理由 |
| --------- | -------- | ------ |
| サービスアカウント | Terraform (この repo) | インフラ。IAM 権限と一緒に管理 |
| Secret Manager（リソース作成） | Terraform (この repo) | インフラ。IAM バインディングも必要 |
| Secret Manager（値の設定） | 手動 / 別プロセス | Terraform に平文を持たせない |
| Cloud Run サービス定義 | Terraform (この repo) | SA、スケーリング、CPU/メモリ、VPC |
| Cloud Run イメージ | App repo | デプロイごとに変わる |
| 環境変数（インフラ系） | Terraform (この repo) | DB_HOST, PROJECT_ID など |
| 環境変数（アプリ系） | App repo | Feature flags, LOG_LEVEL など |
| Secret の Cloud Run マウント | Terraform (この repo) | Secret Manager → 環境変数のマッピング |

## Terraform repo での実装内容

### 1. Artifact Registry リポジトリの作成

App repo がビルドしたコンテナイメージを格納する Docker リポジトリを作成する。

### 2. Cloud Run 用サービスアカウントの作成

Cloud Run サービスが使用するサービスアカウントを作成し、必要な IAM 権限を付与する。

### 3. Secret Manager シークレットの作成

アプリケーションが必要とするシークレット（DB パスワード等）のリソースを作成する。
値の設定は Terraform 外で行う。

### 4. Cloud Run サービスの定義

サービスの初期定義を作成する。`lifecycle.ignore_changes` で `image` を除外し、
Terraform apply 時にアプリのデプロイが巻き戻らないようにする。

### 5. App repo 用 Workload Identity (OIDC) 設定

App repo の GitHub Actions から GCP にアクセスするための Workload Identity 設定を行う。

### 6. IAM 権限の付与

- Cloud Run 用 SA: Secret Manager へのアクセス権
- App repo 用 SA: Artifact Registry への push 権限、Cloud Run のデプロイ権限

## App repo での実装内容

### GitHub Actions ワークフロー

1. コンテナイメージのビルド
2. Artifact Registry への push
3. `gcloud run services update` でデプロイ
4. アプリ系環境変数の設定

## ignore_changes の方針

```hcl
lifecycle {
  ignore_changes = [
    template[0].containers[0].image,
    # App repo が管理する環境変数があれば追加
  ]
}
```

## Secret の値の設定方法

```bash
gcloud secrets versions add <secret-id> --data-file=<path>
```

Terraform では Secret Manager のリソース（箱）のみ管理し、値は手動または別プロセスで設定する。
