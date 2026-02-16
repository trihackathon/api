# PWA統合ガイド

このAPIはPWA（Progressive Web App）での利用を想定した設計になっています。

## 実装済みのPWA対応機能

### 1. CORS設定
PWAから異なるオリジンでAPIを呼び出せるよう、CORSが設定されています。

**本番環境での設定**
```go
// main.go
e.Use(middleware.CORSWithConfig(middleware.CORSConfig{
    AllowOrigins: []string{"https://your-pwa-domain.com"}, // 本番では具体的なドメインを指定
    AllowMethods: []string{http.MethodGet, http.MethodPost, http.MethodPut, http.MethodDelete},
    AllowHeaders: []string{echo.HeaderOrigin, echo.HeaderContentType, echo.HeaderAccept, echo.HeaderAuthorization},
    ExposeHeaders: []string{echo.HeaderContentLength},
    AllowCredentials: true,
}))
```

### 2. GPSポイント重複防止
PWAのオフライン同期時に同じGPSポイントが複数回送信される可能性に対応しています。

**クライアント側でのUUID生成**
```javascript
import { v4 as uuidv4 } from 'uuid';

const gpsPoint = {
  client_id: uuidv4(), // クライアント側でユニークIDを生成
  latitude: position.coords.latitude,
  longitude: position.coords.longitude,
  accuracy: position.coords.accuracy,
  timestamp: new Date(position.timestamp).toISOString()
};
```

**サーバー側での重複チェック**
- `client_id`が指定されている場合、既に保存済みのポイントは自動的にスキップされます
- データベースレベルでもユニーク制約が設定されています

### 3. バッチ送信対応
オフライン時に蓄積したGPSポイントを一括送信できます。

**リクエスト例**
```json
POST /api/activities/running/{activityId}/gps
{
  "points": [
    {
      "client_id": "550e8400-e29b-41d4-a716-446655440001",
      "latitude": 35.6815,
      "longitude": 139.7672,
      "accuracy": 10.5,
      "timestamp": "2026-02-16T10:00:10Z"
    },
    {
      "client_id": "550e8400-e29b-41d4-a716-446655440002",
      "latitude": 35.6817,
      "longitude": 139.7673,
      "accuracy": 8.2,
      "timestamp": "2026-02-16T10:00:20Z"
    }
  ]
}
```

**レスポンス**
```json
{
  "saved_count": 2,
  "current_distance_km": 0.145
}
```

## PWA実装例

### Service Workerでのオフライン対応

```javascript
// sw.js
self.addEventListener('sync', (event) => {
  if (event.tag === 'gps-sync') {
    event.waitUntil(syncGPSData());
  }
});

async function syncGPSData() {
  const db = await openIndexedDB();
  const pendingPoints = await db.getAll('pending_gps_points');
  
  if (pendingPoints.length === 0) return;
  
  const response = await fetch(`${API_BASE}/api/activities/running/${activityId}/gps`, {
    method: 'POST',
    headers: {
      'Authorization': `Bearer ${idToken}`,
      'Content-Type': 'application/json'
    },
    body: JSON.stringify({ points: pendingPoints })
  });
  
  if (response.ok) {
    await db.clear('pending_gps_points');
  }
}
```

### GPS追跡の実装例

```javascript
// gps-tracker.js
class GPSTracker {
  constructor(activityId, idToken) {
    this.activityId = activityId;
    this.idToken = idToken;
    this.gpsBuffer = [];
    this.watchId = null;
  }

  start() {
    this.watchId = navigator.geolocation.watchPosition(
      (position) => this.handlePosition(position),
      (error) => this.handleError(error),
      {
        enableHighAccuracy: true,
        maximumAge: 0,
        timeout: 5000
      }
    );

    // 30秒ごとに自動送信
    this.syncInterval = setInterval(() => this.syncGPS(), 30000);
  }

  async handlePosition(position) {
    const point = {
      client_id: this.generateClientId(),
      latitude: position.coords.latitude,
      longitude: position.coords.longitude,
      accuracy: position.coords.accuracy,
      timestamp: new Date(position.timestamp).toISOString()
    };

    this.gpsBuffer.push(point);

    // オフライン対応: IndexedDBに保存
    await this.saveToLocalDB(point);
  }

  async syncGPS() {
    if (this.gpsBuffer.length === 0) return;

    try {
      const response = await fetch(
        `${API_BASE}/api/activities/running/${this.activityId}/gps`,
        {
          method: 'POST',
          headers: {
            'Authorization': `Bearer ${this.idToken}`,
            'Content-Type': 'application/json'
          },
          body: JSON.stringify({ points: this.gpsBuffer })
        }
      );

      if (response.ok) {
        const data = await response.json();
        console.log(`Distance: ${data.current_distance_km}km`);
        
        // 送信成功後はバッファをクリア
        this.gpsBuffer = [];
        await this.clearLocalDB();
      }
    } catch (error) {
      console.error('GPS sync failed, will retry later:', error);
      // オフライン時はService Workerに任せる
    }
  }

  generateClientId() {
    return `${Date.now()}-${Math.random().toString(36).substr(2, 9)}`;
  }

  stop() {
    if (this.watchId) {
      navigator.geolocation.clearWatch(this.watchId);
    }
    if (this.syncInterval) {
      clearInterval(this.syncInterval);
    }
  }
}
```

### 使用例

```javascript
// ランニング開始
const response = await fetch(`${API_BASE}/api/activities/running/start`, {
  method: 'POST',
  headers: {
    'Authorization': `Bearer ${idToken}`,
    'Content-Type': 'application/json'
  },
  body: JSON.stringify({
    latitude: currentPosition.coords.latitude,
    longitude: currentPosition.coords.longitude
  })
});

const activity = await response.json();
const tracker = new GPSTracker(activity.id, idToken);
tracker.start();

// ランニング終了時
tracker.stop();
await fetch(`${API_BASE}/api/activities/running/${activity.id}/finish`, {
  method: 'POST',
  headers: {
    'Authorization': `Bearer ${idToken}`,
    'Content-Type': 'application/json'
  },
  body: JSON.stringify({
    latitude: currentPosition.coords.latitude,
    longitude: currentPosition.coords.longitude
  })
});
```

## 注意事項

### 1. client_idの扱い
- `client_id`は**オプション**です。指定しない場合も動作します
- 指定する場合は、クライアント側で**必ずユニーク**なIDを生成してください
- UUID v4の使用を推奨します

### 2. バッチサイズ
- 1回のリクエストで送信するGPSポイントは**10〜50点**程度を推奨
- あまり大きすぎるとタイムアウトのリスクがあります

### 3. 精度フィルタ
- `accuracy > 50m`のポイントは距離計算から除外されます（保存はされます）
- 高精度モードでGPSを取得することを推奨します

### 4. HTTPS必須
- PWAの多くの機能（Service Worker、Geolocation APIなど）は**HTTPS環境**でのみ動作します
- 本番環境では必ずHTTPSを使用してください

## トラブルシューティング

### GPSポイントが保存されない
1. `client_id`が重複していないか確認
2. タイムスタンプがRFC3339形式か確認
3. 認証トークンが有効か確認

### 距離が計算されない
1. `accuracy`が50m以下のポイントがあるか確認
2. 連続する2点間が1km以内か確認
3. 最低2点以上のGPSポイントが必要です
