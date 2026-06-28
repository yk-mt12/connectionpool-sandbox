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
         │
         ▼
       Tempo (OpenTelemetry トレース)
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
| Tempo | 分散トレース（OpenTelemetry） | 3100 / 4318 |
| Pyroscope | 継続的プロファイリング | 4040 |

## デプロイ方式

- **Docker Compose**: ローカル開発・動作確認用
- **Kubernetes (minikube)**: k8s 構成の検証用
- **Argo CD (GitOps)**: k8s 環境の GitOps 運用（App of Apps パターン）

## 前提条件

### Docker Compose

- Docker Desktop
- Go 1.21+（ローカル開発時のみ）
- k6（ローカル実行する場合）

### Kubernetes / Argo CD

- minikube
- kubectl
- helm

---

## Docker Compose

### セットアップ

```bash
# go.sum を生成
make setup

# 全サービスを起動
make up
```

### 起動確認

```bash
curl http://localhost:8080/health
curl http://localhost:8080/with-pool
```

### 負荷試験

```bash
make k6-run
```

シナリオ内容（`k6/scenario.js`）:

| シナリオ | 開始 | VU 数 | 時間 | エンドポイント |
|---|---|---|---|---|
| with_pool | 0s | 20 | 60s | `/with-pool` |
| without_pool | 70s | 20 | 60s | `/without-pool` |

### Grafana でメトリクスを確認

1. http://localhost:3000 を開く（admin / admin）
2. Dashboards → **Connection Pool Sandbox** を選択

確認できるパネル:

- Active Virtual Users（k6 VU 推移）
- Request Rate（with-pool vs without-pool）
- Response Time: with-pool vs without-pool（p50 / p95 比較）
- MySQL Active Connections
- MySQL Queries/sec
- DB Connection Pool: Open / In-Use / Idle
- requests テーブル行数
- CPU / Memory / Goroutines Flame Graph（Pyroscope）

### トレースを確認する（Tempo）

負荷試験中に `/with-pool` または `/without-pool` を叩くと、リクエストごとにトレースが Tempo に送信される。

**Grafana Explore でトレースを検索する手順:**

1. http://localhost:3000 を開く
2. 左メニュー → **Explore**
3. データソースを **Tempo** に切り替える
4. **Search** タブを選択し、以下を設定して **Run query**:
   - Service Name: `connectionpool-sandbox`
5. トレース一覧が表示される。クリックすると Span の詳細（階層・所要時間）が展開される

**スパン構成（1リクエストあたり）:**

```
HTTP /with-pool  (otelhttp middleware)
└── RecordUsecase.Execute
    └── pool.Insert          ← SQL実行のみ（接続はプールから取得）

HTTP /without-pool
└── RecordUsecase.Execute
    └── nopool.Insert        ← TCP接続 + 認証 + SQL実行
```

`nopool.Insert` のスパンが `pool.Insert` より大幅に長い場合、その差がコネクション確立コスト。

### プロファイルを確認する（Pyroscope）

アプリ起動後、Pyroscope が継続的に CPU・メモリ・ゴルーチンのプロファイルを収集する。

**Grafana でフレームグラフを見る手順:**

1. http://localhost:3000 を開く
2. Dashboards → **Connection Pool Sandbox** → 下部の **Continuous Profiling (Pyroscope)** セクション
3. または左メニュー → **Explore** → データソースを **Pyroscope** に切り替えて直接クエリ

**Explore でのクエリ例:**

| 目的 | Profile Type | Label Selector |
|---|---|---|
| CPU ホットパス | `process_cpu:cpu:nanoseconds:cpu:nanoseconds` | `{service_name="connectionpool-sandbox"}` |
| メモリ使用量 | `memory:inuse_space:bytes:space:bytes` | `{service_name="connectionpool-sandbox"}` |
| ゴルーチン数 | `goroutine:goroutines:count:goroutine:count` | `{service_name="connectionpool-sandbox"}` |

**フレームグラフの読み方:**

```
横幅 = CPU時間の割合（幅が広いほどホット）
縦  = コールスタック（上が呼び出し元、下が葉）
```

`without-pool` 負荷をかけている時間帯に絞ると、`net.(*Dialer).DialContext` が幅広く表示される。これがコネクション確立コストの正体。

**時間帯を絞る方法:**

Grafana のダッシュボード右上の時間範囲セレクタで `without_pool` シナリオの 70s〜130s の区間を選択すると、その期間のプロファイルのみ表示される。

### 主要コマンド

```bash
make up          # 起動
make down        # 停止（ボリューム保持）
make restart     # App だけ再起動（コード変更後）
make logs        # App ログを tail
make k6-run      # 負荷試験
make setup       # go.sum 生成（初回 or go.mod 変更後）
```

---

## Kubernetes (minikube)

### セットアップ

```bash
cd k8s
make setup
```

### 負荷試験

```bash
make k6-run
```

### ポートフォワード

```bash
make port-forward
# Grafana:    http://localhost:3000
# Prometheus: http://localhost:9090
```

### クリーンアップ

```bash
make clean
```

---

## Argo CD (GitOps)

App of Apps パターンで全コンポーネントを GitOps 管理する。  
git push するだけで Argo CD がクラスタを自動で同期する。

```
Git (k8s/argocd/applications/)
        ↓ watch
connectionpool-root (Argo CD Application)
        ↓ manages
connectionpool-app / mysql / influxdb / tempo / ...
        ↓ manages
実際の k8s リソース（Deployment / StatefulSet / ...）
```

### セットアップ

```bash
cd k8s

# Argo CD インストール + ルート Application 適用
make argocd-bootstrap

# UI を開く（https://localhost:8081）
make argocd-ui

# 初期パスワード確認
make argocd-password
```

> ブラウザで TLS 警告が出る場合はそのまま続行（自己署名証明書）。

### Application 一覧

| Application | 管理対象 |
|---|---|
| connectionpool-root | k8s/argocd/applications/ 配下の全 Application |
| connectionpool-app | k8s/app/ |
| connectionpool-mysql | k8s/mysql/ |
| connectionpool-influxdb | k8s/influxdb/ |
| connectionpool-tempo | k8s/tempo/ |
| connectionpool-mysql-exporter | k8s/mysql-exporter/ |
| connectionpool-grafana-dashboard | k8s/grafana-dashboard/ |
| connectionpool-pyroscope | k8s/pyroscope/ |
| connectionpool-prometheus | Helm: prometheus-community/prometheus |
| connectionpool-grafana | Helm: grafana/grafana |

> k6 Job は Git 管理外（`make k6-run` で手動実行）。

### 手動同期（即時反映したい場合）

```bash
kubectl annotate application connectionpool-app \
  -n argocd argocd.argoproj.io/refresh=hard --overwrite
```

### クリーンアップ

```bash
make argocd-clean
```

---

## エンドポイント

| パス | 説明 |
|---|---|
| `GET /with-pool` | package-level の `sql.DB`（プール）を使用。`SetMaxOpenConns(25)` 設定済み |
| `GET /without-pool` | リクエストごとに新規接続を生成して閉じる |
| `GET /health` | ヘルスチェック |
| `GET /metrics` | Prometheus メトリクス |

## ディレクトリ構成

```
connectionpool-sandbox/
├── docker-compose.yml
├── Makefile
├── app/
│   ├── Dockerfile
│   ├── go.mod
│   ├── main.go
│   ├── adapter/
│   │   ├── handler/        # HTTP ハンドラ
│   │   └── repository/     # pool.go / nopool.go
│   ├── port/               # インターフェース定義
│   ├── usecase/            # ビジネスロジック
│   └── infrastructure/     # OpenTelemetry 設定
├── k6/
│   └── scenario.js
├── mysql/
│   └── init.sql
├── prometheus/
│   └── prometheus.yml
├── grafana/
│   ├── provisioning/
│   └── dashboards/
└── k8s/
    ├── Makefile
    ├── namespace.yaml
    ├── app/
    ├── mysql/
    ├── influxdb/
    ├── tempo/
    ├── mysql-exporter/
    ├── grafana-dashboard/
    ├── helm-values/
    ├── k6/
    └── argocd/
        ├── namespace.yaml
        ├── root-application.yaml
        └── applications/   # App of Apps の子 Application 群
```
