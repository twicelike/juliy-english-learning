# Coding Style & Conventions

Dự án áp dụng chặt chẽ các nguyên tắc của **Clean Architecture** và các quy ước viết code chuẩn của ngôn ngữ Go.

## 1. Nguyên tắc cốt lõi (Core Principles)
- **Không dùng Magic Number/String:** Mọi giá trị hằng số (constants) phải được định nghĩa rõ ràng ở tầng `domain` hoặc trong package cấu hình.
- **Không dùng Global Variable:** Tránh trạng thái toàn cục (global state) làm ảnh hưởng đến unit test và tính đồng thời (concurrency) của Go. Sử dụng Dependency Injection (truyền qua constructor/struct) thay thế.
- **Dependency Rule:** Mã nguồn ở tầng bên ngoài (như Handler) có thể phụ thuộc vào tầng bên trong (Service, Domain), nhưng tầng bên trong TUYỆT ĐỐI KHÔNG phụ thuộc vào tầng bên ngoài.

## 2. Trách nhiệm các tầng (Layer Boundaries)
- **Handler Layer (`internal/handler`):**
  - Chỉ làm nhiệm vụ parse request (JSON, query, param) từ HTTP.
  - Tuyệt đối không gọi trực tiếp Database hay xử lý logic nghiệp vụ.
  - Phải gọi sang Service layer.
- **Service Layer (`internal/service`):**
  - Chứa logic nghiệp vụ.
  - Tuyệt đối không biết gì về HTTP (không chứa đối tượng `gin.Context` hay `http.Request`).
  - Tuyệt đối không gọi trực tiếp các lệnh của MongoDB driver (không import `go.mongodb.org/mongo-driver` ở tầng này).
- **Domain Layer (`internal/domain`):**
  - Định nghĩa Interface (ví dụ: `UserRepository`, `UserService`).
  - Chứa các Struct model thuần tuý mô tả thực thể.
- **Repository Layer (`internal/repository`):**
  - Là tầng duy nhất chịu trách nhiệm giao tiếp với Database.
  - Chứa các logic truy vấn (Find, Insert, Update, Delete).

## 3. Quy ước Đặt tên (Naming Conventions)
- **Biến/Hàm (Variables & Functions):** Sử dụng `camelCase` (ví dụ: `getUserByID`).
- **Structs/Interfaces:** Sử dụng `PascalCase` và có ý nghĩa rõ ràng (ví dụ: `UserRepository`, `FlashcardService`).
- **File:** Tên file nên viết thường, dùng dấu gạch dưới (ví dụ: `user_handler.go`, `auth_middleware.go`).
- **Package:** Tên package phải ngắn gọn, viết thường và phản ánh đúng nội dung thư mục (ví dụ: `package handler`).

## 4. Error Handling
- Bắt lỗi sớm và trả về (Return early).
- Sử dụng Custom Error để định nghĩa rõ HTTP Status code (tham khảo `docs/api.md`).
- Luôn log lỗi chi tiết ở backend, nhưng trả về Frontend message an toàn cho người dùng.