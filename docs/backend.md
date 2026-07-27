# Backend Technical Guidelines

## Ngôn ngữ & Framework
- **Ngôn ngữ:** Go (Golang) 1.21+
- **Web Framework:** Gin (https://github.com/gin-gonic/gin) dùng để điều hướng API và middleware.
- **Database Driver:** `go.mongodb.org/mongo-driver` (Sử dụng official driver của MongoDB).

## Dependency Injection (DI)
- Sử dụng DI thủ công (Constructor injection) để kết nối các tầng.
- Mẫu khởi tạo tiêu chuẩn:
  ```go
  // Ở file main.go hoặc server.go
  userRepo := repository.NewUserRepository(mongoClient)
  userService := service.NewUserService(userRepo)
  userHandler := handler.NewUserHandler(userService)
  
  // Đăng ký routes
  api := router.Group("/api/v1")
  api.GET("/users/:id", userHandler.GetUser)
  ```

## Context Lifecycle
- Truyền `context.Context` xuyên suốt từ Handler xuống Service và Repository.
- Việc truyền context đảm bảo có thể huỷ (cancel) các truy vấn DB hoặc API ngoại vi (OpenAI) khi Client ngắt kết nối đột ngột.

## Validation
- Sử dụng thư viện Validator (tích hợp sẵn trong Gin) thông qua struct tags `binding:"required,email"` v.v.
- Xử lý lỗi validate tại tầng Handler trước khi truyền Data Transfer Object (DTO) xuống Service.