# お知らせくん (Oshirasekun)

<p align="center">
  <img src="fig/app_icon.png" alt="お知らせくんアイコン" width="120" />
</p>

<p align="center">
  日用品の買い忘れをなくす、スマートなリマインダー・在庫管理アプリ
</p>

---

## 概要

**お知らせくん**は、シャンプーや歯磨き粉などの日用消耗品が「なくなりそうなタイミング」を自動で計算し、プッシュ通知でお知らせするアプリです。

商品の内容量・ジャンル・使用人数を登録するだけで、消費ペースから残り日数を算出。カラーコードで緊急度を視覚化し、買い忘れをゼロにします。

---

## スクリーンショット

<p align="center">
  <img src="fig/sign.png" alt="サインイン" width="200" />
  <img src="fig/home.png" alt="ホーム" width="200" />
  <img src="fig/search.png" alt="商品検索" width="200" />
  <img src="fig/scan.png" alt="バーコードスキャン" width="200" />
  <img src="fig/list.png" alt="登録一覧" width="200" />
</p>

---

## 主な機能

| 機能 | 説明 |
|------|------|
| **消費日数の自動計算** | ジャンル・内容量・使用人数から残り日数を算出 (`days = ceil(volume / (people × daily_usage))`) |
| **カレンダー表示** | 月別ビューで期限日をカラーコード表示（緑・黄・赤・期限切れ） |
| **バーコードスキャン** | JAN コードで商品を素早く検索・登録 |
| **楽天 API 連動** | 商品名・画像・容量を自動取得してラクラク登録 |
| **プッシュ通知** | 毎朝 08:00 JST に期限 7 日前・3 日前・当日・期限切れを通知 |
| **通知日数カスタマイズ** | 通知するタイミングを 1〜14 日前で変更可能 |
| **期限切れ商品の一括削除** | まとめてリセット |

---

## 技術スタック

### Frontend — Flutter

| カテゴリ | 技術 |
|----------|------|
| Framework | Flutter 3.3.0+ / Dart 3.3.0+ |
| 状態管理 | flutter_riverpod 2.5.1 |
| ルーティング | go_router 14.2.7 |
| HTTP クライアント | dio 5.4.3 |
| 認証 | Firebase Auth (Email / Google / Apple) |
| プッシュ通知 | Firebase Cloud Messaging |
| カレンダー | table_calendar 3.1.2 |
| バーコードスキャン | mobile_scanner 5.2.3 |
| 画像キャッシュ | cached_network_image 3.3.1 |

### Backend — Go

| カテゴリ | 技術 |
|----------|------|
| Language | Go 1.26.2 |
| Web Framework | Gin 1.12.0 |
| Database | PostgreSQL 16 |
| 認証 | Firebase Admin SDK (ID トークン検証) |
| プッシュ配信 | Firebase Cloud Messaging (FCM) |
| スケジューラ | robfig/cron v3（毎朝 08:00 JST） |
| 外部 API | 楽天市場 API |
| デプロイ | Railway (Nixpacks) |

---

## アーキテクチャ

```
┌─────────────────────────────────────────────────────┐
│                   Flutter App                        │
│  (iOS / Android / Web / macOS / Windows / Linux)    │
│                                                      │
│  Riverpod ─ GoRouter ─ Dio (Firebase Token)         │
└───────────────────┬──────────────────────────────────┘
                    │ HTTPS
┌───────────────────▼──────────────────────────────────┐
│                 Go REST API (Gin)                     │
│  /api/v1                                             │
│  ├── Firebase Middleware (Auth)                      │
│  ├── Products CRUD                                   │
│  ├── Calendar                                        │
│  ├── Rakuten API Proxy                               │
│  ├── FCM Token Management                            │
│  └── Cron Notification Service (08:00 JST)          │
└──────────┬──────────────────────────────┬────────────┘
           │                              │
    ┌──────▼──────┐              ┌────────▼────────┐
    │ PostgreSQL  │              │    Firebase      │
    │     16      │              │ Auth + FCM + SDK │
    └─────────────┘              └─────────────────┘
```

---

## データベーススキーマ

```sql
users
  id UUID PK
  firebase_uid TEXT UNIQUE
  email TEXT
  fcm_token TEXT
  notification_days_before INT DEFAULT 7
  created_at, updated_at TIMESTAMPTZ

products
  id UUID PK
  user_id UUID FK(users)
  item_code TEXT
  name TEXT
  genre TEXT
  image_url TEXT
  content_volume FLOAT
  content_unit TEXT
  days_to_consume INT
  registered_date DATE
  next_due_date DATE
  is_deleted BOOL DEFAULT false

purchase_logs
  id UUID PK
  product_id UUID FK(products)
  purchased_date DATE

genre_daily_usage  -- マスタデータ
  id INT PK
  genre TEXT
  daily_usage_per_person FLOAT
  unit TEXT
```

---

## API エンドポイント

| Method | Path | 説明 |
|--------|------|------|
| `POST` | `/auth/register` | ユーザー登録 |
| `GET` | `/api/v1/products` | 商品一覧 (ステータス付き) |
| `POST` | `/api/v1/products` | 商品登録 |
| `POST` | `/api/v1/products/calculate-days` | 消費日数計算 |
| `DELETE` | `/api/v1/products/:id` | 商品削除（ソフト） |
| `PATCH` | `/api/v1/products/:id/days` | 消費日数更新 |
| `POST` | `/api/v1/products/:id/purchase` | 購入済みマーク |
| `GET` | `/api/v1/calendar?year=&month=` | 月別カレンダーデータ |
| `GET` | `/api/v1/items/search?keyword=&page=&hits=` | 楽天商品検索 |
| `GET` | `/api/v1/items/barcode?jan_code=` | JAN コード検索 |
| `PUT` | `/api/v1/settings` | 通知設定更新 |
| `PUT` | `/api/v1/fcm-token` | FCM トークン更新 |
| `DELETE` | `/api/v1/settings/products/expired` | 期限切れ商品一括削除 |
| `GET` | `/health` | ヘルスチェック |

すべての `/api/v1/*` エンドポイントは Firebase ID トークンによる認証が必要です。

---

## ステータスカラー

| カラー | 条件 |
|--------|------|
| 緑 | 残り日数 > 通知日数設定値 |
| 黄 | 残り日数 ≤ 通知日数設定値 かつ > 3 |
| 赤 | 残り日数 0〜3 日 |
| グレー | 残り日数 < 0（期限切れ） |

---

## セットアップ

### 必要要件

- Flutter 3.3.0+
- Go 1.22+
- Docker & Docker Compose
- Firebase プロジェクト（Auth + FCM 有効化）
- 楽天 API アプリケーション ID

### Backend 起動

```bash
# PostgreSQL をDockerで起動
docker compose up -d

# 環境変数を設定
cd backend
cp .env.example .env
# .env を編集して各種キー・パスを設定

# サーバー起動（マイグレーション自動実行）
go run ./cmd/server
```

### 環境変数（backend/.env）

```env
DB_HOST=localhost
DB_PORT=5432
DB_USER=inform_user
DB_PASSWORD=inform_pass
DB_NAME=inform_db

RAKUTEN_APPLICATION_ID=<楽天APIアプリID>

FIREBASE_PROJECT_ID=<Firebaseプロジェクト名>
FIREBASE_SERVICE_ACCOUNT_JSON=<サービスアカウントJSONのパス>

PORT=8080
```

### Frontend 起動

```bash
cd frontend
flutter pub get

# FlutterFireCLI でFirebase設定を生成済みであること
flutter run
```

---

## デプロイ

### Backend (Railway)

`railway.toml` が設定済みです。Railway にリポジトリを接続し、環境変数を設定するだけで自動デプロイされます。

```toml
[build]
builder = "nixpacks"
buildCommand = "go build -o server ./cmd/server"

[deploy]
startCommand = "./server"
healthcheckPath = "/health"
```

### Frontend (Firebase Hosting / App Store)

```bash
# Web ビルド
flutter build web
firebase deploy

# iOS / Android は通常のストアビルドフローに従ってください
```

---

## プロジェクト構成

```
inform/
├── frontend/               # Flutter アプリ
│   └── lib/
│       ├── core/           # router, theme, constants
│       ├── screens/        # 6 画面
│       ├── services/       # API / Auth / FCM
│       ├── providers/      # Riverpod プロバイダ
│       └── models/         # データモデル
├── backend/                # Go REST API
│   ├── cmd/server/         # エントリポイント
│   └── internal/
│       ├── config/         # 環境変数
│       ├── db/migrations/  # SQL マイグレーション
│       ├── handlers/       # HTTPハンドラ
│       ├── middleware/      # Firebase 認証
│       ├── repository/     # DB アクセス層
│       ├── services/       # ビジネスロジック・通知
│       └── router/         # ルーティング定義
├── fig/                    # アイコン・スクリーンショット
├── docker-compose.yml      # ローカル PostgreSQL
└── SPEC.md                 # 詳細仕様書
```

---

## ライセンス

MIT License
