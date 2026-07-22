# Global Resumable Upload Guide (Tus + RustFS)

This document provides a technical guide for developers on how to use and extend the **Global Resumable Upload Service** implemented using the **Tus Protocol** and **RustFS** (S3-compatible).

## 1. Overview

Traditional multipart/form-data uploads are often unreliable for large files or unstable networks. This project implements the [Tus Protocol](https://tus.io/), an open standard for resumable file uploads.

### Why Use Resumable Uploads?

- **Reliability**: If an upload is interrupted, it can resume exactly where it left off.
- **Efficiency**: Files are streamed directly to S3-compatible storage (RustFS), minimizing memory usage on the API server.
- **Centralized Logic**: A single endpoint handles all uploads, delegating business logic to specific modules via hooks.

---

## 2. Architecture: The Registry Pattern

The upload service is designed to be **feature-agnostic**. It doesn't know about "Avatars" or "Documents". Instead, it uses a **Registry Pattern**:

1.  **Tus Handler**: Manages the protocol, chunks, and storage.
2.  **Event Registry**: Maps a `type` (sent in metadata) to a specific **Upload Hook**.
3.  **Upload Hook**: A piece of business logic that runs after a file is successfully uploaded (e.g., updating a user's avatar URL in the database).

---

## 3. Backend Implementation

### Step 1: Create an Upload Hook

To handle a new type of upload, you must implement the `tus.UploadHook` interface in your module's `usecase` or `delivery` layer.

```go
// Example: internal/modules/user/usecase/avatar_hook.go
package usecase

import (
    "context"
    "github.com/Roisfaozi/queue-base/pkg/tus"
)

type AvatarHook struct {
    UseCase UserUseCase
}

func (h *AvatarHook) HandleUpload(ctx context.Context, event tus.UploadEvent) error {
    // 1. Extract metadata sent by the frontend
    userID := event.Metadata["user_id"]

    // 2. Perform business logic (e.g., save URL to DB)
    return h.UseCase.SaveAvatarURL(ctx, userID, event.FileURL)
}
```

### Step 2: Register the Hook

Register your hook in `internal/config/app.go` during application initialization.

```go
// internal/config/app.go

// 1. Initialize your hook
avatarHook := &userUseCase.AvatarHook{UseCase: userModule.UserUseCase}

// 2. Register it with a unique key (e.g., "project_doc")
sseManager := sse.NewManager() // Existing
tusRegistry := tus.NewRegistry()
tusRegistry.Register("avatar", avatarHook)
```

---

## 4. Frontend Integration

Clients should use standard Tus libraries like `tus-js-client` or `uppy`. The key requirement is providing the `type` in the `Upload-Metadata` header.

### Example using `tus-js-client`:

```javascript
import * as tus from "tus-js-client";

const file = document.getElementById("file-input").files[0];

const upload = new tus.Upload(file, {
  endpoint: "http://localhost:8080/api/v1/upload/files/",
  retryDelays: [0, 1000, 3000, 5000],
  headers: {
    Authorization: `Bearer ${jwtToken}`,
  },
  metadata: {
    filename: file.name,
    filetype: file.type,
    type: "avatar", // MANDATORY: Must match backend registration
    user_id: "12345", // Custom metadata for your Hook
  },
  onError: (error) => console.log("Failed:", error),
  onProgress: (bytesUploaded, bytesTotal) => {
    const percentage = ((bytesUploaded / bytesTotal) * 100).toFixed(2);
    console.log(percentage + "%");
  },
  onSuccess: () => {
    console.log("Download %s from %s", upload.file.name, upload.url);
  },
});

upload.start();
```

---

## 5. Configuration & Setup

### Environment Variables (.env)

```ini
# S3/RustFS Configuration
STORAGE_DRIVER=s3
STORAGE_S3_ENDPOINT=http://localhost:9000
STORAGE_S3_BUCKET=uploads
STORAGE_S3_ACCESS_KEY=rustfsadmin
STORAGE_S3_SECRET_KEY=rustfsadmin
STORAGE_S3_FORCE_PATH_STYLE=true

# TUS Configuration
TUS_BASE_PATH=/api/v1/upload/files/
```

### Infrastructure (Docker)

Ensure **RustFS** is running as part of your `docker-compose.dev.yml`. You can access the console at `http://localhost:9001` to verify file persistence.

---

## 6. Endpoints Reference

All TUS interactions happen under the `/api/v1/upload/files/` prefix.

| Method     | Purpose                                                 |
| :--------- | :------------------------------------------------------ |
| **POST**   | Create a new upload session. Returns `Location` header. |
| **HEAD**   | Get the current offset (where to resume).               |
| **PATCH**  | Send a chunk of binary data.                            |
| **DELETE** | Terminate and remove a partial upload.                  |

---

## 7. Troubleshooting & CORS

Tus requires several specific headers to be exposed to the browser. These are automatically handled in `internal/middleware/cors_middleware.go`:

- **Allowed Headers**: `Tus-Resumable`, `Upload-Length`, `Upload-Metadata`, `Upload-Offset`, `Upload-Protocol`.
- **Exposed Headers**: `Location`, `Tus-Version`, `Tus-Resumable`, `Upload-Offset`, `Upload-Length`.

**Common Issue**: If you get a `404 Not Found` on a PATCH request, ensure your `TUS_BASE_PATH` env variable exactly matches the route registered in `router.go` and that `http.StripPrefix` is used correctly.
