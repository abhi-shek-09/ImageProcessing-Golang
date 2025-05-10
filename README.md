
# 🖼️ Go Image Transformation Service

A simple Go-based microservice to upload, transform, and serve images with resizing, cropping, rotation, format conversion, watermarking, and filters.


## ✨ Features

- Upload and store images
- Resize, crop, rotate, flip/mirror
- Convert image formats (jpeg, png)
- Apply filters (grayscale, sepia)
- Watermarking (basic block)
- JPEG compression control
- Track transformations in DB (optional)


## 📦 Environment Variables

Create a `.env` file in the root:

```env
PORT=8080
UPLOAD_DIR=./uploads
DATABASE_URL=your_postgres_or_sqlite_connection
```

Make sure `UPLOAD_DIR` exists or will be created by your app using `os.MkdirAll`.

---
## 📸 API Endpoints

### 🔐 Image Upload

| Method | Endpoint      | Description                      |
|--------|---------------|----------------------------------|
| POST   | api/upload    | Upload an image (multipart form) |
 
### 🖼️ Image Fetch

| Method | Endpoint           | Description                       |
|--------|--------------------|-----------------------------------|
| GET    | api/image/{id}     | Get original and transformed URLs |
 
### 🖼️ Images Fetch

| Method | Endpoint       | Description                                     |
|--------|----------------|-------------------------------------------------|
| GET    | api/images     | Get original and transformed URLs for all images|
 
### 🛠️ Image Transformation

| Method | Endpoint                   | Description                          |
|--------|----------------------------|--------------------------------------|
| POST   | api/images/{id}/transform  | Apply transformations to image by ID |

**Request Body (JSON)**:
```json
{
  "resize": { "width": 200, "height": 200 },
  "crop": { "x": 10, "y": 10, "width": 100, "height": 100 },
  "rotate": 90,
  "flip": true,
  "mirror": true,
  "compress": 80,
  "format": "jpeg",
  "filter": "grayscale",
  "watermark": { "x": 10, "y": 20 }
}
```
---

## 🧪 Test Flow

1. **Upload an image**  
   - Use Postman `POST api/upload` with `form-data` (key = `image`, type = `File`)  
2. **Copy the `id` from the response**
3. **Call `POST api/images/{id}/transform`**  
   - Use raw JSON body with transformation config
4. **Call `GET api/image/{id}`**  
   - See both original and transformed URLs
5. **Call `GET api/images`**  
   - See both original and transformed URLs for all images

---

## ⚠️ Notes

- Transformed images are saved as:  
  `/uploads/<originalname>_transformed<format>_<timestamp>.<ext>`
- All paths are relative to `UPLOAD_DIR`
- On Windows, use `filepath.ToSlash()` when saving image paths
- If path errors like `Failed to open original image` occur:
  - Ensure `img.OriginalURL` is a relative path like `/uploads/...`
  - Ensure `UPLOAD_DIR` exists and is correctly configured
- Uses `github.com/disintegration/imaging` for transformations

---

## 🔧 Tech Stack

- Golang
- net/http standard library
- imaging for image processing
- Postgres