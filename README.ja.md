# syncenv

クラウドストレージを使ったバージョン管理された環境設定ファイルの管理ツール

[English](./README.md) | [日本語](#日本語)

---

## 概要

**syncenv**は、Gitのブランチやタグに応じた環境設定ファイルをクラウドストレージから自動で取得・保存できるCLIツールです。暗号化にも対応し、チーム全体で安全にバージョン管理された環境設定ファイルを共有できます。

## エレベーターピッチ（30秒版）

「v1.5で動作確認したいのに、新しいメンバーが環境設定ファイルのバージョン違いで動かせない」そんな経験ありませんか?

**syncenv**は、Gitのブランチやタグに応じた環境設定ファイルを、AWS/Azure/GCPのストレージから自動で取得・保存できるCLIツールです。暗号化にも対応し、チーム全体で安全にバージョン管理された環境設定ファイルを共有できます。

`git checkout v1.5` → `syncenv pull` これだけで、そのバージョンの環境設定ファイルが手に入ります。

## 特徴

- **Git連携**: タグ/ブランチから自動判定
- **マルチクラウド**: AWS S3、Azure Blob、Google Cloud対応
- **複数ファイル対応**: 複数の環境ファイルを同時に管理（`.env`、`.env.local`、`config/*.json`など）
- **パスサポート**: サブディレクトリ内のファイルにも対応、ディレクトリ自動作成
- **セキュア**: AES-256-GCM暗号化サポート
- **シンプル**: Go製シングルバイナリ、設定は1ファイル

## インストール

```bash
go install github.com/O6lvl4/syncenv/cmd/syncenv@latest
```

または、ソースからビルド:

```bash
git clone https://github.com/O6lvl4/syncenv.git
cd syncenv
go build -o syncenv ./cmd/syncenv

# またはビルドスクリプトを使用
./build.sh
```

## 使用例

```bash
# 初期化
$ syncenv init

# プッシュ
$ git checkout v1.5
$ syncenv push  # 自動的にv1.5として保存

$ syncenv push --tag v1.6  # 明示的にタグ指定

# プル
$ git checkout v1.5
$ syncenv pull  # 自動的にv1.5を取得

$ syncenv pull --tag v1.6  # 明示的にタグ指定

# 一覧
$ syncenv list

# 差分
$ syncenv diff v1.5 v1.6
```

## 設定

初回実行時に`syncenv init`を実行すると、`.syncenv.yml`が作成されます。

設定ファイルの例:

**単一ファイル:**
```yaml
storage:
  type: s3
  bucket: my-syncenv-bucket
  region: us-west-2
  prefix: envs/  # オプション

encryption:
  enabled: true
  key: <自動生成>  # init時に自動生成

env_file: .env
```

**複数ファイル（パスあり）:**
```yaml
storage:
  type: gcs
  project_id: my-project
  bucket_name: my-syncenv-bucket

encryption:
  enabled: true

env_files:
  - .env
  - .env.local
  - config/database/settings.json
  - secrets/api.conf
```

## 暗号化キーの管理

暗号化を有効にすると、暗号化キーが**自動生成**されて `.syncenv.yml` 設定ファイル内に保存されます。

**チーム共有**: `.syncenv.yml` ファイルをチームメンバーと共有するだけでOK！別途キーファイルを管理する必要はありません。

**セキュリティについて**:
- 暗号化キーは利便性のため設定ファイルに保存されます
- クラウドストレージ上で平文保存されないようにしつつ、セットアップをシンプルに保ちます
- セキュリティを強化するには、`.syncenv.yml` を安全な場所に保管し、セキュアなチャネルで共有してください
- パブリックリポジトリへの誤コミットを防ぐため、`.gitignore` に `.syncenv.yml` を追加することを検討してください

## クラウドプロバイダーの設定

お好みのクラウドプロバイダーを選択し、以下の設定手順に従ってください。

---

### オプション1: AWS S3

**おすすめ:** AWSインフラを既に使用しているチーム

**設定方法:**

環境変数でAWS認証情報を設定:

```bash
export AWS_ACCESS_KEY_ID=your_access_key
export AWS_SECRET_ACCESS_KEY=your_secret_key
```

または、AWS CLIで設定:

```bash
aws configure
```

**設定ファイルの例:**
```yaml
storage:
  type: s3
  bucket: my-syncenv-bucket
  region: us-west-2
  prefix: envs/  # オプション
```

---

### オプション2: Azure Blob Storage

**おすすめ:** Microsoft Azureエコシステムを使用しているチーム

**設定方法:**

Azure接続文字列を設定:

```bash
export AZURE_STORAGE_CONNECTION_STRING="DefaultEndpointsProtocol=https;AccountName=...;AccountKey=...;EndpointSuffix=core.windows.net"
```

**設定ファイルの例:**
```yaml
storage:
  type: azure
  container_name: my-syncenv-container
```

---

### オプション3: Google Cloud

**おすすめ:** Google Cloud Platformを使用しているチーム

**設定方法:**

Google Cloud認証情報を設定:

```bash
export GOOGLE_APPLICATION_CREDENTIALS=/path/to/credentials.json
```

または、gcloud CLIで認証:

```bash
gcloud auth application-default login
```

**設定ファイルの例:**
```yaml
storage:
  type: gcs
  project_id: my-project
  bucket_name: my-syncenv-bucket
```

## コマンド

| コマンド | 説明 |
|---------|------|
| `syncenv init` | 設定ファイルを作成 |
| `syncenv push [--tag TAG]` | 環境設定ファイルをアップロード |
| `syncenv pull [--tag TAG] [-f]` | 環境設定ファイルをダウンロード |
| `syncenv list` | 保存されているバージョン一覧 |
| `syncenv diff TAG1 TAG2` | 2つのバージョン間の差分表示 |

## ユースケース

- **バージョン別の環境**: v1.5、v1.6などの異なる設定を維持
- **チーム協働**: 新しいメンバーが素早く正しい環境をセットアップ
- **複数環境**: 開発、ステージング、本番環境の異なる設定
- **デバッグ**: 古いバージョンに簡単に切り替えて正しい環境設定ファイルで実行

## セキュリティ上の考慮事項

- 機密データを保存する際は常に暗号化を使用してください
- 暗号化キーは `.syncenv.yml` に保存されます - このファイルをチームと安全に共有してください
- `.gitignore` を使用して `.syncenv.yml` と `.env` ファイルのパブリックリポジトリへのコミットを防いでください
- クラウドプロバイダーのIAMロールと権限を適切に使用してください
- 最大限のセキュリティを確保するには、`.syncenv.yml` をパスワードマネージャーやシークレットボルトに保存してください

## ビルド方法

### オプション1: ビルドスクリプト

```bash
./build.sh
```

### オプション2: 直接ビルド

```bash
unset GOROOT && unset GOPATH && go build -o syncenv ./cmd/syncenv
```

### オプション3: Makefile

```bash
make build
```

## テスト

```bash
# テストを実行
make test

# または直接goコマンドで
go test ./...

# カバレッジ付きでテスト
make test-coverage
```

## プロジェクト構造

```
syncenv/
├── cmd/syncenv/          # メインエントリーポイント
├── internal/
│   ├── archive/         # 複数ファイル用のtar.gz処理
│   ├── config/          # 設定管理
│   ├── git/             # Git連携
│   ├── crypto/          # AES-256-GCM暗号化
│   ├── storage/         # クラウドストレージ実装
│   │   ├── s3.go       # AWS S3
│   │   ├── azure.go    # Azure Blob
│   │   ├── gcs.go      # Google Cloud
│   │   └── mock.go     # テスト用モックストレージ
│   └── cli/            # CLIコマンド
├── README.md           # 英語ドキュメント
├── README.ja.md        # このファイル（日本語）
├── LICENSE             # MITライセンス
├── Makefile           # ビルドタスク
└── build.sh           # ビルドスクリプト
```

## コントリビューション

プルリクエストを歓迎します！お気軽にプルリクエストを提出してください。

## ライセンス

MITライセンス - 詳細は[LICENSE](./LICENSE)ファイルを参照してください

## サポート

問題、質問、機能リクエストがある場合は、GitHubでissueを開いてください。

---

## 詳細な背景（1分版）

### 問題
プロダクトがv1.5からv1.6にアップデートされると、環境設定ファイルのスキーマも変わります。本番で問題が起きてv1.5で調査したいとき、新しいメンバーは「v1.5の環境設定ファイルがわからない」「動かせない」という状況に陥ります。

### 解決策
**syncenv**は、環境設定ファイルをGitのバージョン（タグ/ブランチ）と紐付けてクラウドストレージで管理するOSSツールです。

```bash
git checkout v1.5
syncenv pull  # v1.5の環境設定ファイルを自動取得
```

### ターゲット
バージョン管理されたプロダクトを開発するチーム、特にマイクロサービスやAPI開発者

**OSSとして公開。あなたのチームの環境設定ファイル管理を変えませんか？**

---

**プロジェクトステータス**: プロダクション準備完了 ✅

テスト済み:
- ✅ Google Cloud
- ✅ Azure Blob Storage
- ⏳ AWS S3（実装完了、実環境テスト待ち）
