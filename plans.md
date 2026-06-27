# connectionpool-sandbox 構築計画

## 目的
MySQL コネクションプールの有無が性能に与える影響を、k6 負荷試験 + Grafana で可視化する

## スタック（全 OSS・ローカル完結）
| コンポーネント | 役割 |
|---|---|
| MySQL 8.0 | 検証対象 DB |
| Go (net/http + database/sql) | with-pool / without-pool の2エンドポイント |
| k6 | 負荷試験（20 VU × 1min × 2シナリオ） |
| InfluxDB 1.8 | k6 メトリクス格納 |
| Prometheus + mysqld-exporter | MySQL メトリクス収集 |
| Grafana | 統合ダッシュボード |

## ファイル構成
```
connectionpool-sandbox/
├── docker-compose.yml
├── Makefile
├── plans.md
├── app/
│   ├── Dockerfile
│   ├── go.mod
│   └── main.go          ← with-pool handler 実装済み / without-pool は TODO(human)
├── k6/
│   └── scenario.js      ← 2シナリオ（with_pool → without_pool）
├── mysql/
│   └── init.sql
├── prometheus/
│   └── prometheus.yml
└── grafana/
    ├── provisioning/
    │   ├── datasources/datasources.yml
    │   └── dashboards/provider.yml
    └── dashboards/
        └── connectionpool.json
```

## タスク
- [x] ディレクトリ作成
- [x] plans.md 作成
- [ ] docker-compose.yml
- [ ] app/main.go (with-pool 実装 + TODO(human))
- [ ] app/go.mod + Dockerfile
- [ ] k6/scenario.js
- [ ] mysql/init.sql
- [ ] prometheus/prometheus.yml
- [ ] grafana provisioning (datasources + dashboards)
- [ ] grafana/dashboards/connectionpool.json
- [ ] Makefile + .gitignore
- [ ] git init
- [ ] Learn by Doing: without-pool handler 実装依頼

## 比較観点
- レスポンスタイム p50 / p95（Grafana で並べて比較）
- MySQL Active Connections（プールなしは急増する）
- スループット（req/s）

---

## 追加調査: Grafana vs Datadog の機能比較

### APM（分散トレーシング）

| 機能 | Datadog APM | Grafana 相当 |
|---|---|---|
| 分散トレース | 組み込み | **Grafana Tempo**（別コンポーネント） |
| サービスマップ | 自動生成 | Tempo + Grafana で手動構築 |
| トレース→ログ相関 | 自動（trace_id） | Loki + Tempo で設定が必要 |
| Go SDK | `dd-trace-go` 1行差し替え | `go.opentelemetry.io/otel` + exporter 設定 |
| セットアップ難易度 | 低（Agent + ライブラリ差し替え） | 高（Tempo / OTel Collector / Loki の個別構築） |

Datadog APM は Agent が自動でトレースを収集・相関させる。
Grafana スタックでは OpenTelemetry → Tempo → Grafana の pipeline を自前で組む必要がある。

### Profiler（継続プロファイリング）

| 機能 | Datadog Profiler | Grafana 相当 |
|---|---|---|
| CPU / メモリ / goroutine プロファイル | 組み込み | **Grafana Pyroscope**（旧 Phlare） |
| Go SDK | `dd-trace-go` に同梱 | `github.com/grafana/pyroscope-go` |
| flame graph UI | Datadog UI | Grafana の Flame Graph パネル |

```go
// Grafana Pyroscope を使う場合
import "github.com/grafana/pyroscope-go"

pyroscope.Start(pyroscope.Config{
    ApplicationName: "connectionpool-sandbox",
    ServerAddress:   "http://pyroscope:4040",
    ProfileTypes: []pyroscope.ProfileType{
        pyroscope.ProfileCPU,
        pyroscope.ProfileAllocObjects,
        pyroscope.ProfileInuseObjects,
    },
})
```

### まとめ

- Datadog: APM / Profiler / メトリクス / ログが**1つの Agent + SDK**に統合
- Grafana: 各機能が独立した OSS（Prometheus / Tempo / Loki / Pyroscope）を組み合わせる必要がある。自由度は高いがセットアップコストが高い

---

## 追加実装: レコード数の可視化

### アプローチ比較

| 方法 | 精度 | 実装コスト |
|---|---|---|
| `mysql_info_schema_table_rows` (mysqld-exporter) | 低（InnoDB の推定値） | ゼロ（すでに取得中） |
| Go アプリから `SELECT COUNT(*) FROM requests` を定期実行 | 正確 | 低 |
| Prometheus Recording Rule | 中 | 低 |

### 実装方針: Go アプリに custom gauge を追加

`main.go` に以下を追加し、5秒ごとに `requests` テーブルの行数を gauge として expose する。

```go
var tableRowCount = promauto.NewGauge(prometheus.GaugeOpts{
    Name: "db_table_row_count",
    Help: "Exact row count of sandbox.requests table",
})

// main() 内
go func() {
    for range time.Tick(5 * time.Second) {
        var count float64
        poolDB.QueryRow("SELECT COUNT(*) FROM requests").Scan(&count)
        tableRowCount.Set(count)
    }
}()
```

Grafana クエリ: `db_table_row_count{job="app"}`

### タスク
- [ ] `main.go` に `db_table_row_count` gauge 追加
- [ ] Grafana ダッシュボードに "requests テーブル行数" パネル追加
- [ ] `make app` でリビルド・デプロイ
