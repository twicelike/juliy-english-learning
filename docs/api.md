# API Convention & Error Handling

Tất cả các API của dự án đều phải tuân thủ chuẩn RESTful và sử dụng định dạng dữ liệu trả về duy nhất (Standard Wrapper) để Frontend dễ dàng xử lý.

## 1. RESTful Endpoints
- Sử dụng danh từ số nhiều cho Resource (ví dụ: `/api/v1/users`, `/api/v1/flashcards`).
- **GET:** Lấy dữ liệu.
- **POST:** Tạo dữ liệu mới.
- **PUT / PATCH:** Cập nhật dữ liệu.
- **DELETE:** Xóa dữ liệu.

## 2. Response Convention (Dữ liệu thành công)
Mọi response thành công (HTTP 200, 201) đều bọc trong một đối tượng chung:
```json
{
  "success": true,
  "data": {
    "id": "60d5ecb54b1234567890",
    "word": "ubiquitous",
    "meaning": "có mặt ở khắp nơi"
  },
  "message": "Từ vựng đã được lưu thành công",
  "error": null
}
```
*Note: `message` là tuỳ chọn, thường dùng cho thông báo hiển thị Toast trên UI.*

## 3. Error Convention (Dữ liệu lỗi)
Khi có lỗi (HTTP 4xx, 5xx), định dạng trả về bắt buộc phải tuân theo:
```json
{
  "success": false,
  "data": null,
  "error": {
    "code": "VALIDATION_ERROR",
    "message": "Email không hợp lệ.",
    "details": [
      {
        "field": "email",
        "issue": "must be a valid email address"
      }
    ]
  }
}
```
- `code`: Chuỗi string viết HOA (VD: `UNAUTHORIZED`, `NOT_FOUND`, `INTERNAL_SERVER_ERROR`).
- `message`: Thông báo lỗi thân thiện dành cho người dùng (có thể đem hiển thị UI).
- `details`: (Tuỳ chọn) Chi tiết lỗi dành cho debug hoặc báo lỗi form cụ thể.

## 4. Phân trang (Pagination)
Sử dụng offset/limit hoặc cursor-based. Chuẩn response trả về cho List:
```json
{
  "success": true,
  "data": {
    "items": [...],
    "meta": {
      "total": 150,
      "page": 1,
      "limit": 20,
      "has_next": true
    }
  }
}
```