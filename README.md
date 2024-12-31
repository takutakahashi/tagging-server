# Tagging Server with Likes

## Overview
This server provides functionality for managing tags associated with targets, along with the ability to like targets. Each user is identified via an Authorization key, which is hashed using SHA-256 and used to segregate data. The server is implemented in Go and uses SQLite as the database backend.

---

## Features
- Add tags to specific targets.
- Retrieve tags associated with a specific target.
- Retrieve targets associated with a specific tag.
- Like a specific target.
- Retrieve the like count for a specific target.
- Data is isolated per user based on their Authorization key.

---

## Setup

### Prerequisites
- Go 1.18+
- SQLite

### Installation
1. Clone the repository:
   ```bash
   git clone <repository-url>
   cd <repository-directory>
   ```
2. Build the server:
   ```bash
   go build -o tagging-server
   ```
3. Run the server:
   ```bash
   ./tagging-server
   ```

By default, the server will start on `localhost:8080`.

---

## API Documentation

### Authorization
Each API request requires an `Authorization` header. This header should contain a user-specific key. The server uses the SHA-256 hash of this key to manage data isolation.

---

### Endpoints

#### 1. Add a Tag to a Target
**URL:** `/add-tag`

**Method:** `POST`

**Request Headers:**
- `Authorization: <your-auth-key>`

**Request Body:**
```json
{
  "target": "<target-name>",
  "tag": "<tag-name>"
}
```

**Response:**
- `200 OK` on success.

---

#### 2. Retrieve Tags for a Target
**URL:** `/get-tags?target=<target-name>`

**Method:** `GET`

**Request Headers:**
- `Authorization: <your-auth-key>`

**Response Body:**
```json
{
  "tags": ["<tag1>", "<tag2>", ...]
}
```

---

#### 3. Retrieve Targets for a Tag
**URL:** `/get-targets?tag=<tag-name>`

**Method:** `GET`

**Request Headers:**
- `Authorization: <your-auth-key>`

**Response Body:**
```json
{
  "targets": ["<target1>", "<target2>", ...]
}
```

---

#### 4. Like a Target
**URL:** `/like-target`

**Method:** `POST`

**Request Headers:**
- `Authorization: <your-auth-key>`

**Request Body:**
```json
{
  "target": "<target-name>"
}
```

**Response:**
- `200 OK` on success.

---

#### 5. Retrieve Like Count for a Target
**URL:** `/get-likes?target=<target-name>`

**Method:** `GET`

**Request Headers:**
- `Authorization: <your-auth-key>`

**Response Body:**
```json
{
  "like_count": <number-of-likes>
}
```

---

## Example Usage

### Add a Tag to a Target
```bash
curl -X POST "http://localhost:8080/add-tag" \
     -H "Authorization: my-secret-key" \
     -H "Content-Type: application/json" \
     -d '{"target": "myTarget", "tag": "myTag"}'
```

### Retrieve Tags for a Target
```bash
curl -X GET "http://localhost:8080/get-tags?target=myTarget" \
     -H "Authorization: my-secret-key"
```

### Retrieve Targets for a Tag
```bash
curl -X GET "http://localhost:8080/get-targets?tag=myTag" \
     -H "Authorization: my-secret-key"
```

### Like a Target
```bash
curl -X POST "http://localhost:8080/like-target" \
     -H "Authorization: my-secret-key" \
     -H "Content-Type: application/json" \
     -d '{"target": "myTarget"}'
```

### Retrieve Like Count for a Target
```bash
curl -X GET "http://localhost:8080/get-likes?target=myTarget" \
     -H "Authorization: my-secret-key"
```

**Response Example:**
```json
{
  "like_count": 5
}
```

---

## Database Schema

### `targets_tags` Table
| Column              | Type    | Description                                     |
|---------------------|---------|-------------------------------------------------|
| `encryption_key_hash` | TEXT    | SHA-256 hash of the Authorization key           |
| `target`            | TEXT    | Name of the target                              |
| `tag`               | TEXT    | Name of the tag                                 |

### `likes` Table
| Column              | Type    | Description                                     |
|---------------------|---------|-------------------------------------------------|
| `encryption_key_hash` | TEXT    | SHA-256 hash of the Authorization key           |
| `target`            | TEXT    | Name of the target                              |
| `like_count`        | INTEGER | Number of likes for the target                 |

---

## Notes
- Ensure the `Authorization` key is kept secure to prevent unauthorized access to your data.
- The server is designed to be lightweight and simple, suitable for single-user or small-scale use cases.


