# connectionpool-sandbox

MySQL コネクションプールの有無が性能に与える影響を、k6 負荷試験 + Grafana で可視化する検証環境。

## 構成

```
k6 ──────────────────► InfluxDB ──────► Grafana
         │                                  ▲
         ▼                                  │
       Go App (with-pool / without-pool)    │
         │                                  │
         ▼                                  │
       MySQL ◄── mysqld-exporter ◄── Prometheus
```

| サービス | 役割 | ポート |
|---|---|---|
| MySQL 8.0 | 検証対象 DB | 3306 |
| Go App | with-pool / without-pool エンドポイント | 8080 |
| k6 | 負荷試験 | - |
| InfluxDB 1.8 | k6 メトリクス格納 | 8086 |
| Prometheus | MySQL メトリクス収集 | 9090 |
| mysqld-exporter | MySQL → Prometheus | 9104 |
| Grafana | 統合ダッシュボード | 3000 |

## 前提条件

- Docker Desktop
- Go 1.21+（ローカル開発時のみ）
- k6（ローカル実行する場合）

## セットアップ

### 1. go.sum を生成

```bash
make setup
```

### 2. 全サービスを起動

```bash
make up
```

初回は Docker イメージのビルドがあるため 2〜3 分かかります。

### 3. 起動確認

```bash
# App ヘルスチェック
curl http://localhost:8080/health

# MySQL 疎通
curl http://localhost:8080/with-pool
```

## 負荷試験の実行

```bash
make k6-run
```

シナリオ内容（`k6/scenario.js`）:

| シナリオ | 開始 | VU 数 | 時間 | エンドポイント |
|---|---|---|---|---|
| with_pool | 0s | 20 | 60s | `/with-pool` |
| without_pool | 70s | 20 | 60s | `/without-pool` |

## Grafana でメトリクスを確認

1. http://localhost:3000 を開く（admin / admin）
2. Dashboards → **Connection Pool Sandbox** を選択

確認できるパネル:

- Active Virtual Users（k6 VU 推移）
- Request Rate（with-pool vs without-pool）
- Response Time: with-pool vs without-pool（p50 / p95 比較）
- MySQL Active Connections
- MySQL Queries/sec

## 主要コマンド

```bash
# 起動
make up

# 停止（ボリューム保持）
make down

# App だけ再起動（コード変更後）
make restart

# App ログを tail
make logs

# 負荷試験
make k6-run

# go.sum 生成（初回 or go.mod 変更後）
make setup
```

## エンドポイント

| パス | 説明 |
|---|---|
| `GET /with-pool` | package-level の `sql.DB`（プール）を使用。`SetMaxOpenConns(10)` 設定済み |
| `GET /without-pool` | リクエストごとに新規接続を生成して閉じる |
| `GET /health` | ヘルスチェック |

## ディレクトリ構成

```
connectionpool-sandbox/
├── docker-compose.yml
├── Makefile
├── app/
│   ├── Dockerfile
│   ├── go.mod
│   └── main.go
├── k6/
│   └── scenario.js
├── mysql/
│   └── init.sql          # exporter ユーザー作成
├── prometheus/
│   └── prometheus.yml
└── grafana/
    ├── provisioning/
    │   ├── datasources/  # InfluxDB + Prometheus 自動設定
    │   └── dashboards/
    └── dashboards/
        └── connectionpool.json
```
