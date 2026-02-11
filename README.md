# Transjakarta Fleet Backend

Backend service untuk technical test manajemen armada:
- Menerima data lokasi kendaraan dari MQTT.
- Menyimpan data lokasi ke PostgreSQL.
- Menyediakan API lokasi terakhir dan riwayat lokasi.
- Mengirim event geofence ke RabbitMQ.
- Menjalankan service via Docker Compose.

## Stack
- Go 1.25
- Fiber
- PostgreSQL 16
- Eclipse Mosquitto (MQTT)
- RabbitMQ 3 (management)
- Docker + Docker Compose

## Arsitektur Singkat
1. `scripts/publisher/main.go` publish data lokasi ke topic `/fleet/vehicle/{vehicle_id}/location` tiap 2 detik.
2. `cmd/mqtt-subscriber/main.go` subscribe topic, parse JSON, simpan ke tabel `vehicle_locations`.
3. Subscriber hitung jarak ke titik geofence (Haversine). Jika dalam radius, publish event ke exchange `fleet.events`.
4. Queue `geofence_alerts` menerima event, lalu `cmd/worker/main.go` consume dan log payload event.
5. `cmd/api/main.go` expose endpoint untuk ambil lokasi terakhir dan riwayat.

## Struktur Proyek
```text
transjakarta-fleet-backend/
|-- cmd/
|   |-- api/main.go
|   |-- mqtt-subscriber/main.go
|   `-- worker/main.go
|-- internal/
|   |-- db/db.go
|   |-- geofence/geofence.go
|   |-- handlers/location_handler.go
|   |-- models/vehicle_location.go
|   `-- repository/location_repository.go
|-- scripts/
|   |-- publisher/main.go
|   `-- sql/init.sql
|-- docker/mosquitto/mosquitto.conf
|-- docker-compose.yml
|-- Dockerfile
`-- .env.example
```

## Environment Variables
Buat file `.env` dari `.env.example`:

```bash
# PowerShell
Copy-Item .env.example .env
```

Isi `.env` untuk mode Docker Compose:

```env
DB_HOST=postgres
DB_PORT=5432
DB_USER=your_db_user
DB_PASSWORD=your_db_password
DB_NAME=your_db_name

API_PORT=3000

MQTT_BROKER_URL=tcp://mosquitto:1883
RABBITMQ_URL=amqp://guest:guest@rabbitmq:5672/

GEOFENCE_LAT=-6.2088
GEOFENCE_LNG=106.8456
GEOFENCE_RADIUS=50
```

## Menjalankan Semua Service dengan Docker
```bash
docker compose up -d --build
docker compose ps
```

Service yang dijalankan:
- `api`
- `mqtt-subscriber`
- `worker`
- `postgres`
- `rabbitmq`
- `mosquitto`

RabbitMQ management UI:
- URL: `http://localhost:15672`
- Username: `guest`
- Password: `guest`

## Inisialisasi Database
Schema belum auto-migrate saat startup, jadi jalankan manual:

```bash
docker exec -i tj_postgres psql -U tj_user -d tj_fleet < scripts/sql/init.sql
```

Jika nilai DB berbeda, sesuaikan `-U` dan `-d` dengan `.env`.

## Menjalankan Mock Publisher
Publisher dijalankan dari host (bukan dari container runtime image).

### Opsi 1: Set env hanya untuk command (PowerShell)
```powershell
$env:MQTT_BROKER_URL="tcp://localhost:1883"; go run ./scripts/publisher
```

### Opsi 2: Pakai env aktif terminal
Pastikan `MQTT_BROKER_URL=tcp://localhost:1883`, lalu:
```bash
go run ./scripts/publisher
```

## API Endpoints
Base URL: `http://localhost:3000`

### 1) Get latest location
`GET /vehicles/{vehicle_id}/location`

Contoh:
```bash
curl http://localhost:3000/vehicles/B1234XYZ/location
```

Contoh response sukses:
```json
{
  "vehicle_id": "B1234XYZ",
  "latitude": -6.2088,
  "longitude": 106.8456,
  "timestamp": 1715003456
}
```

Contoh response saat data belum ada:
```json
{
  "error": "Location not found"
}
```

### 2) Get location history
`GET /vehicles/{vehicle_id}/history?start=<unix>&end=<unix>`

Contoh:
```bash
curl "http://localhost:3000/vehicles/B1234XYZ/history?start=1715000000&end=1715009999"
```

Contoh error query parameter:
```json
{
  "error": "Invalid start timestamp"
}
```

## Verifikasi End-to-End
1. Jalankan stack: `docker compose up -d --build`
2. Inisialisasi DB: `scripts/sql/init.sql`
3. Jalankan publisher di host.
4. Cek data masuk DB:
```bash
docker exec -it tj_postgres psql -U tj_user -d tj_fleet -c "SELECT * FROM vehicle_locations ORDER BY timestamp DESC LIMIT 5;"
```
5. Cek API latest/history via browser atau Postman.
6. Cek log geofence:
```bash
docker compose logs -f mqtt-subscriber
docker compose logs -f worker
```

## Catatan Implementasi Saat Ini
- Worker saat ini hanya log event geofence yang diterima.
- Tidak ada batch insert, transaction orchestration, atau benchmark performa formal.
- Geofence event yang dipublish mengikuti format:
```json
{
  "vehicle_id": "B1234XYZ",
  "event": "geofence_entry",
  "location": {
    "latitude": -6.2088,
    "longitude": 106.8456
  },
  "timestamp": 1715003456
}
```

## Postman
Buat collection minimal berisi:
1. `GET /vehicles/:vehicle_id/location`
2. `GET /vehicles/:vehicle_id/history?start=&end=`

Environment variable yang disarankan:
- `base_url` = `http://localhost:3000`
- `vehicle_id` = `B1234XYZ`
- `start`
- `end`
