# Demo Script

## Foundation Demo

1. Start MySQL with `docker compose up -d mysql`.
2. Start the API:

```bash
cd apps/api
go run ./cmd/server
```

3. Confirm the API responds:

```bash
curl http://localhost:8080/api/health
curl http://localhost:8080/api/version
```

4. Start the frontend:

```bash
cd apps/web
npm install
npm run dev
```

5. Open `http://localhost:5173`.
6. Resize the browser to verify the desktop sidebar and mobile drawer.
7. Navigate through each placeholder page.
8. Confirm the topbar, user menu placeholder, and alert badge placeholder are visible.
