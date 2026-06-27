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
